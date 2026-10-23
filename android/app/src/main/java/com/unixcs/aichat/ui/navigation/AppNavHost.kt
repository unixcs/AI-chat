package com.unixcs.aichat.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.unixcs.aichat.data.api.UserProfile
import com.unixcs.aichat.ui.login.LoginViewModel
import com.unixcs.aichat.ui.screens.ChatScreen
import com.unixcs.aichat.ui.screens.HistoryScreen
import com.unixcs.aichat.ui.screens.LoginScreen
import com.unixcs.aichat.ui.screens.SettingsScreen
import com.unixcs.aichat.ui.screens.SplashScreen
import com.unixcs.aichat.ui.session.SessionState
import com.unixcs.aichat.ui.session.SessionViewModel
import org.koin.androidx.compose.koinViewModel

object Routes {
    const val CHAT = "chat"
    const val HISTORY = "history"
    const val SETTINGS = "settings"
}

/**
 * Root gate. Login and Chat are mutually exclusive surfaces chosen by session
 * state, not routes — so token expiry can never leave a stale Chat on screen.
 */
@Composable
fun AppRoot(sessionViewModel: SessionViewModel = koinViewModel<SessionViewModel>()) {
    val state by sessionViewModel.state.collectAsStateWithLifecycle()
    val profile by sessionViewModel.profile.collectAsStateWithLifecycle()

    when (val current = state) {
        SessionState.Loading -> SplashScreen()

        is SessionState.LoggedOut -> LoginScreen(
            viewModel = koinViewModel<LoginViewModel>(),
            notice = current.message,
        )

        SessionState.LoggedIn -> AppNavHost(
            profile = profile,
            onLogout = sessionViewModel::logout,
        )
    }
}

@Composable
private fun AppNavHost(
    profile: UserProfile?,
    onLogout: () -> Unit,
) {
    val navController = rememberNavController()

    NavHost(
        navController = navController,
        startDestination = Routes.CHAT,
    ) {
        composable(Routes.CHAT) {
            ChatScreen(
                onOpenHistory = { navController.navigate(Routes.HISTORY) },
                onOpenSettings = { navController.navigate(Routes.SETTINGS) },
            )
        }

        composable(Routes.HISTORY) {
            HistoryScreen(onBack = { navController.popBackStack() })
        }

        composable(Routes.SETTINGS) {
            SettingsScreen(
                profile = profile,
                onLogout = onLogout,
                onBack = { navController.popBackStack() },
            )
        }
    }
}
