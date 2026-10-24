package com.unixcs.aichat.ui.theme

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext

/**
 * Builds a full tonal scheme from one soft accent. Secondary and tertiary stay
 * neutral so the accent owns the identity — chat reading stays calm, which is
 * exactly what the DeepSeek-style look is after.
 */
private fun lightScheme(accent: AccentPalette) = lightColorScheme(
    primary = accent.lightPrimary,
    onPrimary = accent.lightOnPrimary,
    primaryContainer = accent.lightPrimaryContainer,
    onPrimaryContainer = accent.lightOnPrimaryContainer,
    secondary = accent.lightPrimary,
    onSecondary = accent.lightOnPrimary,
    secondaryContainer = accent.lightPrimaryContainer,
    onSecondaryContainer = accent.lightOnPrimaryContainer,
)

private fun darkScheme(accent: AccentPalette) = darkColorScheme(
    primary = accent.darkPrimary,
    onPrimary = accent.darkOnPrimary,
    primaryContainer = accent.darkPrimaryContainer,
    onPrimaryContainer = accent.darkOnPrimaryContainer,
    secondary = accent.darkPrimary,
    onSecondary = accent.darkOnPrimary,
    secondaryContainer = accent.darkPrimaryContainer,
    onSecondaryContainer = accent.darkOnPrimaryContainer,
)

/**
 * Dynamic (Material You) color is deliberately not used anymore: the user picks
 * an accent explicitly, and wallpaper-derived colors would override it.
 */
@Composable
fun AIChatTheme(
    themeMode: ThemeMode = ThemeMode.System,
    accentKey: String = AccentPalettes.first().key,
    content: @Composable () -> Unit,
) {
    val darkTheme = when (themeMode) {
        ThemeMode.System -> isSystemInDarkTheme()
        ThemeMode.Light -> false
        ThemeMode.Dark -> true
    }
    val colorScheme = if (darkTheme) {
        darkScheme(accentPalette(accentKey))
    } else {
        lightScheme(accentPalette(accentKey))
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = Typography,
        content = content,
    )
}
