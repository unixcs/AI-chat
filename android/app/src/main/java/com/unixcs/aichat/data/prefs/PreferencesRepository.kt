package com.unixcs.aichat.data.prefs

import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.api.map
import com.unixcs.aichat.data.session.SessionStore
import kotlinx.coroutines.flow.first

/**
 * Reads and writes the answer preferences. Updates are partial: the UI sends
 * only the dimension the user touched, so a stale copy of the other two can
 * never overwrite a change made on another device.
 */
class PreferencesRepository(
    private val api: ApiClient,
    private val store: SessionStore,
) {
    suspend fun load(): ApiResult<ChatPreferences> = withToken { token ->
        api.fetchProfile(token).map(ChatPreferences::fromProfile)
    }

    suspend fun updateLength(value: String): ApiResult<Boolean> =
        withToken { token -> api.updatePreferences(token, answerLength = value) }

    suspend fun updateStyle(value: String): ApiResult<Boolean> =
        withToken { token -> api.updatePreferences(token, answerStyle = value) }

    suspend fun updateFormat(value: String): ApiResult<Boolean> =
        withToken { token -> api.updatePreferences(token, answerFormat = value) }

    private suspend fun <T> withToken(block: suspend (String) -> ApiResult<T>): ApiResult<T> {
        val token = store.token.first()
        if (token.isNullOrBlank()) {
            return ApiResult.Failure(401, "登录已失效，请重新登录", "SESSION_KICKED")
        }
        val result = block(token)
        if (result is ApiResult.Failure && result.unauthorized) {
            store.expire(result.message)
        }
        return result
    }
}
