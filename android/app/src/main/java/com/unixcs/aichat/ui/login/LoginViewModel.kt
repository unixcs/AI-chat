package com.unixcs.aichat.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.session.AuthRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class LoginUiState(
    val phone: String = "",
    val password: String = "",
    val loading: Boolean = false,
    val error: String? = null,
) {
    val canSubmit: Boolean get() = phone.length == 11 && password.isNotBlank()
}

class LoginViewModel(private val repo: AuthRepository) : ViewModel() {

    private val _ui = MutableStateFlow(LoginUiState())
    val ui: StateFlow<LoginUiState> = _ui.asStateFlow()

    fun onPhoneChange(value: String) {
        val digits = value.filter { it.isDigit() }.take(11)
        _ui.update { it.copy(phone = digits, error = null) }
    }

    fun onPasswordChange(value: String) {
        _ui.update { it.copy(password = value, error = null) }
    }

    fun submit() {
        val current = _ui.value
        if (current.loading || !current.canSubmit) return

        _ui.update { it.copy(loading = true, error = null) }
        viewModelScope.launch {
            when (val result = repo.login(current.phone, current.password)) {
                // Saving the token flips the session gate; drop the password
                // from memory immediately.
                is ApiResult.Success -> _ui.update { it.copy(loading = false, password = "") }
                is ApiResult.Failure -> _ui.update { it.copy(loading = false, error = result.message) }
            }
        }
    }
}
