package com.unixcs.aichat.ui.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.api.UserProfile
import com.unixcs.aichat.data.session.SessionStore
import com.unixcs.aichat.data.settings.AppSettingsStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class AppearanceState(
    val themeMode: String = "system",
    val themeColor: String = "violet",
)

data class AccountUiState(
    val busy: Boolean = false,
    val message: String? = null,
)

/**
 * Owns the device-local appearance settings and the account maintenance calls
 * (redeem code, nickname update). Profile changes flow back through the session
 * gate so chat and settings stay consistent.
 */
class SettingsViewModel(
    private val settings: AppSettingsStore,
    private val store: SessionStore,
    private val accountRepo: AccountRepository,
) : ViewModel() {

    val appearance: StateFlow<AppearanceState> = combine(
        settings.themeMode,
        settings.themeColor,
    ) { mode, color -> AppearanceState(mode, color) }
        .stateIn(viewModelScope, SharingStarted.Eagerly, AppearanceState())

    private val _account = MutableStateFlow(AccountUiState())
    val account: StateFlow<AccountUiState> = _account.asStateFlow()

    fun setThemeMode(mode: String) {
        viewModelScope.launch { settings.setThemeMode(mode) }
    }

    fun setThemeColor(color: String) {
        viewModelScope.launch { settings.setThemeColor(color) }
    }

    fun consumeMessage(shown: String) {
        _account.update { if (it.message == shown) it.copy(message = null) else it }
    }

    fun redeem(code: String, onProfileUpdated: (UserProfile) -> Unit) {
        execute(
            call = { withTokenAndReport { token -> accountRepo.redeem(token, code) } },
            onSuccess = onProfileUpdated,
        )
    }

    fun updateNickname(
        nickname: String,
        currentProfile: UserProfile?,
        onProfileUpdated: (UserProfile) -> Unit,
    ) {
        execute(
            call = {
                withTokenAndReport { token ->
                    // avatarUrl is a full overwrite on the backend: echo the
                    // current value so a nickname-only edit never wipes it.
                    accountRepo.updateProfile(token, nickname, currentProfile?.avatarUrl ?: "")
                }
            },
            onSuccess = onProfileUpdated,
        )
    }

    private fun execute(
        call: suspend () -> ApiResult<UserProfile>,
        onSuccess: (UserProfile) -> Unit,
    ) {
        if (_account.value.busy) return
        _account.value = AccountUiState(busy = true)
        viewModelScope.launch {
            when (val result = call()) {
                is ApiResult.Success -> {
                    _account.value = AccountUiState(busy = false)
                    onSuccess(result.data)
                }

                is ApiResult.Failure -> {
                    _account.value = AccountUiState(busy = false, message = result.message)
                }
            }
        }
    }

    private suspend fun withTokenAndReport(
        block: suspend (String) -> ApiResult<UserProfile>,
    ): ApiResult<UserProfile> {
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
