package com.unixcs.aichat.ui.session

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.api.UserProfile
import com.unixcs.aichat.data.session.AuthRepository
import com.unixcs.aichat.data.session.SessionStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Session gate. On startup it validates the cached token against the backend
 * (which also reveals a session kicked by another device), then follows the
 * token flow so login/logout switch surfaces without navigation calls.
 *
 * A network outage is not a reason to log the user out: the cached token is
 * trusted and the next authenticated request reports the real status.
 */
class SessionViewModel(
    private val repo: AuthRepository,
    private val store: SessionStore,
) : ViewModel() {

    private val _state = MutableStateFlow<SessionState>(SessionState.Loading)
    val state: StateFlow<SessionState> = _state.asStateFlow()

    private val _profile = MutableStateFlow<UserProfile?>(null)
    val profile: StateFlow<UserProfile?> = _profile.asStateFlow()

    init {
        // A 401 or kick discovered by any later request must land on the login
        // screen with the server's reason instead of a silent bounce.
        viewModelScope.launch {
            store.expirations.collect { message ->
                if (_state.value is SessionState.LoggedOut) return@collect
                _profile.value = null
                _state.value = SessionState.LoggedOut(message)
                store.clear()
            }
        }

        viewModelScope.launch {
            val token = store.token.first()
            if (token.isNullOrBlank()) {
                _state.value = SessionState.LoggedOut()
            } else {
                resolveCachedSession(token)
            }

            store.token.distinctUntilChanged().collect { current ->
                when {
                    current.isNullOrBlank() && _state.value !is SessionState.LoggedOut -> {
                        _profile.value = null
                        _state.value = SessionState.LoggedOut()
                    }

                    !current.isNullOrBlank() && _state.value !is SessionState.LoggedIn -> {
                        _state.value = SessionState.LoggedIn
                        refreshProfile(current)
                    }
                }
            }
        }
    }

    fun logout() {
        viewModelScope.launch { repo.logout() }
    }

    private suspend fun resolveCachedSession(token: String) {
        when (val result = repo.fetchProfile(token)) {
            is ApiResult.Success -> {
                _profile.value = result.data
                store.savePhone(result.data.phone)
                _state.value = SessionState.LoggedIn
            }

            is ApiResult.Failure -> {
                // 401: token expired / kicked. 403 here can only mean the
                // account was disabled (membership is enforced on chat).
                if (result.unauthorized || result.status == 403) {
                    _state.value = SessionState.LoggedOut(result.message)
                    store.clear()
                } else {
                    _state.value = SessionState.LoggedIn
                }
            }
        }
    }

    private suspend fun refreshProfile(token: String) {
        when (val result = repo.fetchProfile(token)) {
            is ApiResult.Success -> {
                _profile.value = result.data
                store.savePhone(result.data.phone)
            }

            is ApiResult.Failure -> {
                if (result.unauthorized) {
                    _state.value = SessionState.LoggedOut(result.message)
                    store.clear()
                }
            }
        }
    }
}
