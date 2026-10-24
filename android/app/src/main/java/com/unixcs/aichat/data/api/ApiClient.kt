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

    suspend fun listConversations(token: String): ApiResult<List<ConversationDto>> =
        get(path = "/api/chat/conversations?page=1&pageSize=500", token = token)

    suspend fun createConversation(token: String): ApiResult<ConversationDto> =
        post(path = "/api/chat/conversations", body = "{}", token = token)

    suspend fun fetchMessages(token: String, conversationId: String): ApiResult<List<MessageDto>> =
        get(path = "/api/chat/conversations/$conversationId/messages", token = token)

    suspend fun deleteConversation(token: String, conversationId: String): ApiResult<Boolean> =
        delete(path = "/api/chat/conversations/$conversationId", token = token)

    suspend fun renameConversation(
        token: String,
        conversationId: String,
        title: String,
    ): ApiResult<ConversationDto> =
        put(
            path = "/api/chat/conversations/$conversationId/title",
            body = json.encodeToString(RenameConversationRequest(title = title)),
            token = token,
        )

    /**
     * Partial update: omitted fields are left untouched by the backend, so a
     * single chip tap never clobbers the other two dimensions.
     */
    suspend fun updatePreferences(
        token: String,
        answerLength: String? = null,
        answerStyle: String? = null,
        answerFormat: String? = null,
    ): ApiResult<Boolean> =
        put(
            path = "/api/user/preferences",
            body = json.encodeToString(
                PreferencesUpdate(
                    answerLength = answerLength,
                    answerStyle = answerStyle,
                    answerFormat = answerFormat,
                ),
            ),
            token = token,
        )

    /** Redeem a card code; the backend returns the refreshed profile. */
    suspend fun redeem(token: String, code: String): ApiResult<RedeemResponse> =
        post(
            path = "/api/user/redeem",
            body = json.encodeToString(RedeemRequest(code = code.trim())),
            token = token,
        )

    /**
     * Full-overwrite profile update: the backend keeps the old nickname only
     * when the new one is blank, and always stores the given avatarUrl.
     */
    suspend fun updateProfile(
        token: String,
        nickname: String,
        avatarUrl: String,
    ): ApiResult<UserProfile> =
        put(
            path = "/api/user/profile",
            body = json.encodeToString(
                UpdateProfileRequest(nickname = nickname, avatarUrl = avatarUrl),
            ),
            token = token,
        )

    /**
     * Newest unread announcement, or null when there is none — the backend
     * answers `data: null` in that case, so this path must accept null data.
     */
    suspend fun currentAnnouncement(token: String): ApiResult<AnnouncementDto?> =
        getNullable(path = "/api/announcements/current", token = token)

    /**
     * asOf echoes the announcement's updatedAt verbatim: the backend drops a
     * stale ack when the admin replaced the content in between.
     */
    suspend fun ackAnnouncement(token: String, announcementId: String, asOf: String): ApiResult<Boolean> =
        post(
            path = "/api/announcements/$announcementId/ack",
            body = json.encodeToString(AckAnnouncementRequest(asOf = asOf)),
            token = token,
        )

    private suspend inline fun <reified T> get(path: String, token: String): ApiResult<T> {
        val request = Request.Builder()
            .url(baseUrl + path)
            .header("Authorization", "Bearer $token")
            .get()
            .build()
        return execute(request)
    }

    /** Same as [get] but accepts a `data: null` envelope as a successful null. */
    private suspend inline fun <reified T> getNullable(path: String, token: String): ApiResult<T?> {
        val request = Request.Builder()
            .url(baseUrl + path)
            .header("Authorization", "Bearer $token")
            .get()
            .build()
        return executeNullable(request)
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

    private suspend inline fun <reified T> delete(path: String, token: String): ApiResult<T> {
        val request = Request.Builder()
            .url(baseUrl + path)
            .header("Authorization", "Bearer $token")
            .delete()
            .build()
        return execute(request)
    }

    private suspend inline fun <reified T> put(path: String, body: String, token: String): ApiResult<T> {
        val request = Request.Builder()
            .url(baseUrl + path)
            .header("Authorization", "Bearer $token")
            .put(body.toRequestBody(jsonMediaType))
            .build()
        return execute(request)
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
                    val envelope = decodeEnvelope<T>(text)
                        ?: return@withContext ApiResult.Failure(response.code, "服务器响应异常")
                    if (envelope.code != 0) {
                        return@withContext ApiResult.Failure(
                            status = response.code,
                            message = envelope.message ?: "请求失败",
                        )
                    }
                    ApiResult.Success(
                        envelope.data
                            ?: return@withContext ApiResult.Failure(response.code, "服务器响应异常"),
                    )
                }
            } catch (_: IOException) {
                ApiResult.Failure(status = -1, message = "网络连接失败，请检查网络")
            }
        }

    /**
     * Nullable twin of [execute]: here `data: null` inside a code-0 envelope
     * is a legitimate answer (the announcement endpoint signals "no unread
     * announcement" exactly that way), not a malformed response.
     */
    private suspend inline fun <reified T> executeNullable(request: Request): ApiResult<T?> =
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
                    val envelope = decodeEnvelope<T>(text)
                        ?: return@withContext ApiResult.Failure(response.code, "服务器响应异常")
                    if (envelope.code != 0) {
                        return@withContext ApiResult.Failure(
                            status = response.code,
                            message = envelope.message ?: "请求失败",
                        )
                    }
                    ApiResult.Success(envelope.data)
                }
            } catch (_: IOException) {
                ApiResult.Failure(status = -1, message = "网络连接失败，请检查网络")
            }
        }

    private inline fun <reified T> decodeEnvelope(text: String): Envelope<T>? = try {
        json.decodeFromString<Envelope<T>>(text)
    } catch (_: Exception) {
        null
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
