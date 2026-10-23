package com.unixcs.aichat.data.chat

import java.io.IOException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request

/**
 * Reads the chat SSE stream with a plain OkHttp call. EventSource is not used:
 * it cannot send an Authorization header and Android has no built-in client.
 *
 * Deltas are coalesced into ~48 ms batches to keep recomposition cheap. The
 * flow is cold; cancelling collection cancels the HTTP call, which is also how
 * the server learns about a user-initiated stop.
 */
class ChatStreamClient(
    baseUrl: String,
    private val client: OkHttpClient,
    private val json: Json,
) {
    private val baseUrl = baseUrl.trimEnd('/')
    private val parser = SseParser(json)

    fun stream(
        token: String,
        conversationId: String,
        content: String,
    ): Flow<StreamEvent> = callbackFlow {
        val url = "$baseUrl/api/chat/conversations/$conversationId/stream"
            .toHttpUrl()
            .newBuilder()
            .addQueryParameter("content", content)
            .build()
        val request = Request.Builder()
            .url(url)
            .header("Authorization", "Bearer $token")
            .get()
            .build()
        val call = client.newCall(request)

        val job = launch(Dispatchers.IO) {
            var pending = StringBuilder()
            var lastEmit = System.nanoTime()

            fun flush() {
                if (pending.isNotEmpty()) {
                    trySend(StreamEvent.Delta(pending.toString()))
                    pending = StringBuilder()
                }
                lastEmit = System.nanoTime()
            }

            try {
                call.execute().use { response ->
                    if (!response.isSuccessful) {
                        val body = response.body?.string().orEmpty()
                        val error = parseHttpError(body)
                        trySend(
                            StreamEvent.Failed(
                                message = error?.first ?: defaultMessage(response.code),
                                code = error?.second,
                                status = response.code,
                            ),
                        )
                        return@use
                    }
                    val source = response.body?.source() ?: return@use
                    while (isActive) {
                        val line = source.readUtf8Line() ?: break
                        when (val event = parser.parse(line)) {
                            is SseEvent.Delta -> {
                                pending.append(event.text)
                                if (System.nanoTime() - lastEmit >= BATCH_NANOS) flush()
                            }

                            is SseEvent.Error -> {
                                flush()
                                trySend(StreamEvent.Failed(event.message, event.code))
                                close()
                                return@use
                            }

                            SseEvent.Done -> {
                                flush()
                                trySend(StreamEvent.Completed)
                                close()
                                return@use
                            }

                            null -> Unit
                        }
                    }
                    // A proxy may close a finished stream without [DONE];
                    // treat EOF as completion rather than a network error.
                    flush()
                    trySend(StreamEvent.Completed)
                }
            } catch (_: IOException) {
                if (isActive) {
                    trySend(StreamEvent.Failed("网络连接失败，请检查网络", status = -1))
                }
            } finally {
                close()
            }
        }

        awaitClose {
            call.cancel()
            job.cancel()
        }
    }

    private fun parseHttpError(body: String): Pair<String, String?>? = try {
        val obj = json.parseToJsonElement(body).jsonObject
        val message = obj["message"]?.jsonPrimitive?.contentOrNull
        if (message == null) null else message to obj["code"]?.jsonPrimitive?.contentOrNull
    } catch (_: Exception) {
        null
    }

    private fun defaultMessage(status: Int): String = when {
        status == 401 -> "登录已失效，请重新登录"
        status == 403 -> "没有权限"
        status >= 500 -> "服务器繁忙"
        else -> "请求失败"
    }

    private companion object {
        const val BATCH_NANOS = 48_000_000L
    }
}
