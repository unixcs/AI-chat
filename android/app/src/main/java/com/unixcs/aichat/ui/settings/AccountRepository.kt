package com.unixcs.aichat.ui.settings

import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.api.UserProfile
import com.unixcs.aichat.data.api.map

/**
 * Account maintenance calls that are not part of the streaming chat path:
 * redeem-code top-up and profile (nickname) updates.
 */
class AccountRepository(private val api: ApiClient) {

    suspend fun redeem(token: String, code: String): ApiResult<UserProfile> =
        api.redeem(token, code).map { it.profile }

    suspend fun updateProfile(token: String, nickname: String, avatarUrl: String): ApiResult<UserProfile> =
        api.updateProfile(token, nickname, avatarUrl)
}
