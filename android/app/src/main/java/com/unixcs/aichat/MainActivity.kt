package com.unixcs.aichat

import android.graphics.Color
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.unixcs.aichat.ui.navigation.AppRoot
import com.unixcs.aichat.ui.settings.SettingsViewModel
import com.unixcs.aichat.ui.theme.AIChatTheme
import com.unixcs.aichat.ui.theme.ThemeMode
import org.koin.androidx.compose.koinViewModel

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            val settingsViewModel: SettingsViewModel = koinViewModel()
            val appearance by settingsViewModel.appearance.collectAsStateWithLifecycle()
            val mode = ThemeMode.fromKey(appearance.themeMode)
            val dark = when (mode) {
                ThemeMode.System -> isSystemInDarkTheme()
                ThemeMode.Light -> false
                ThemeMode.Dark -> true
            }
            // Edge-to-edge derives status bar icon contrast from the system
            // theme; re-assert on every mode switch so icons stay readable.
            LaunchedEffect(dark) {
                enableEdgeToEdge(
                    statusBarStyle = if (dark) {
                        SystemBarStyle.dark(Color.TRANSPARENT)
                    } else {
                        SystemBarStyle.light(Color.TRANSPARENT, Color.TRANSPARENT)
                    },
                )
            }
            AIChatTheme(themeMode = mode, accentKey = appearance.themeColor) {
                AppRoot()
            }
        }
    }
}
