package com.unixcs.aichat.ui.session

/**
 * What the root composable needs to decide which surface to show. There is no
 * explicit navigation between login and chat: the token flow drives this.
 */
sealed interface SessionState {
    data object Loading : SessionState

    data class LoggedOut(val message: String? = null) : SessionState

    data object LoggedIn : SessionState
}
