package com.unixcs.aichat.data.api

import java.io.IOException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody

/**
 * Thin OkHttp client for the JSON endpoints. SSE gets its own reader in
 * Phase 3 — this class deliberately knows nothing about streaming.
 */
class ApiClient(
    baseUrl: String,
    private val client: OkHttpClient,
    private val json: Json,
) {
    private val baseUrl = baseUrl.trimEnd('/')
    private val jsonMediaType = "application/json; charset=utf-8".toMediaType()

    suspend fun login(phone: String, password: String): ApiResult<LoginData> =
        post(
            path = "/api/auth/login",
            body = json.encodeToString(LoginRequest(phone = phone, password = password)),
            token = null,
        )

    suspend fun sessionCheck(token: String): ApiResult<SessionCheck> =
        get(path = "/api/user/session-check", token = token)

    suspend fun fetchProfile(token: String): ApiResult<UserProfile> =
        get(path = "/api/user/profile", token = token)

    suspend fun logout(token: String): ApiResult<Boolean> =
        post(path = "/api/auth/logout", body = "{}", token = token)

    private suspend inline fun <reified T> get(path: String, token: String): ApiResult<T> {
        val request = Request.Builder()
            .url(baseUrl + path)
            .header("Authorization", "Bearer $token")
            .get()
            .build()
        return execute(request)
    }

    private suspend inline fun <reified T> post(path: String, body: String, token: String?): ApiResult<T> {
        val builder = Request.Builder()
            .url(baseUrl + path)
            .post(body.toRequestBody(jsonMediaType))
        if (token != null) {
            builder.header("Authorization", "Bearer $token")
        }
        return execute(builder.build())
    }

    private suspend inline fun <reified T> execute(request: Request): ApiResult<T> =
        withContext(Dispatchers.IO) {
            try {
                client.newCall(request).execute().use { response ->
                    val text = response.body?.string().orEmpty()
                    if (!response.isSuccessful) {
                        val error = decodeError(text)
                        return@withContext ApiResult.Failure(
                            status = response.code,
                            message = error?.message ?: defaultMessage(response.code),
                            code = error?.code,
                        )
                    }
                    val envelope = try {
                        json.decodeFromString<Envelope<T>>(text)
                    } catch (_: Exception) {
                        null
                    } ?: return@withContext ApiResult.Failure(response.code, "服务器响应异常")

                    if (envelope.code != 0) {
                        return@withContext ApiResult.Failure(
                            status = response.code,
                            message = envelope.message ?: "请求失败",
                        )
                    }
                    val data = envelope.data
                        ?: return@withContext ApiResult.Failure(response.code, "服务器响应异常")
                    ApiResult.Success(data)
                }
            } catch (_: IOException) {
                ApiResult.Failure(status = -1, message = "网络连接失败，请检查网络")
            }
        }

    private fun decodeError(text: String): ErrorBody? = try {
        json.decodeFromString<ErrorBody>(text)
    } catch (_: Exception) {
        null
    }

    private fun defaultMessage(status: Int): String = when {
        status == 401 -> "登录已失效，请重新登录"
        status == 403 -> "没有权限"
        status >= 500 -> "服务器繁忙"
        else -> "请求失败"
    }
}
