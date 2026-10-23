package com.unixcs.aichat.data.session

import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.api.LoginData
import com.unixcs.aichat.data.api.UserProfile
import kotlinx.coroutines.flow.first

/**
 * Auth flows shared by the login screen and the session gate. Saving the token
 * here is what flips the app from Login to Chat — no navigation callbacks.
 */
class AuthRepository(
    private val api: ApiClient,
    private val store: SessionStore,
) {
    suspend fun login(phone: String, password: String): ApiResult<LoginData> {
        val result = api.login(phone.trim(), password)
        if (result is ApiResult.Success) {
            store.saveSession(result.data.token, phone.trim())
        }
        return result
    }

    suspend fun fetchProfile(token: String): ApiResult<UserProfile> = api.fetchProfile(token)

    suspend fun logout() {
        val token = store.token.first()
        // Clear locally first: logout must feel instant even with no network.
        store.clear()
        if (!token.isNullOrBlank()) {
            api.logout(token) // best effort, result intentionally ignored
        }
    }
}
