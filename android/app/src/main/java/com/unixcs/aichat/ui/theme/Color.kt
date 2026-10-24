package com.unixcs.aichat.ui.theme

import androidx.compose.ui.graphics.Color

// ─── Soft accent palettes for chat ──────────────────────────────────────────
// Each accent ships a light primary/onPrimary/primaryContainer set and a dark
// one. Saturation stays low and contrast stays comfortable so long reading
// sessions don't fatigue — the palettes were picked against Material 3 tonal
// targets (primary 40 light / 80 dark).

data class AccentPalette(
    val key: String,
    val label: String,
    val lightPrimary: Color,
    val lightOnPrimary: Color,
    val lightPrimaryContainer: Color,
    val lightOnPrimaryContainer: Color,
    val darkPrimary: Color,
    val darkOnPrimary: Color,
    val darkPrimaryContainer: Color,
    val darkOnPrimaryContainer: Color,
)

val AccentPalettes: List<AccentPalette> = listOf(
    AccentPalette(
        key = "violet",
        label = "雾紫",
        lightPrimary = Color(0xFF6750A4),
        lightOnPrimary = Color(0xFFFFFFFF),
        lightPrimaryContainer = Color(0xFFEADDFF),
        lightOnPrimaryContainer = Color(0xFF21005D),
        darkPrimary = Color(0xFFD0BCFF),
        darkOnPrimary = Color(0xFF381E72),
        darkPrimaryContainer = Color(0xFF4F378B),
        darkOnPrimaryContainer = Color(0xFFEADDFF),
    ),
    AccentPalette(
        key = "starlight",
        label = "星夜蓝",
        lightPrimary = Color(0xFF4A6572),
        lightOnPrimary = Color(0xFFFFFFFF),
        lightPrimaryContainer = Color(0xFFD7E3EA),
        lightOnPrimaryContainer = Color(0xFF0F1D26),
        darkPrimary = Color(0xFFAEC7D2),
        darkOnPrimary = Color(0xFF1A333D),
        darkPrimaryContainer = Color(0xFF334B57),
        darkOnPrimaryContainer = Color(0xFFD7E3EA),
    ),
    AccentPalette(
        key = "moonlight",
        label = "月白青",
        lightPrimary = Color(0xFF4A6269),
        lightOnPrimary = Color(0xFFFFFFFF),
        lightPrimaryContainer = Color(0xFFCCE8E8),
        lightOnPrimaryContainer = Color(0xFF051F1F),
        darkPrimary = Color(0xFFB0CCCB),
        darkOnPrimary = Color(0xFF1B3534),
        darkPrimaryContainer = Color(0xFF324B4A),
        darkOnPrimaryContainer = Color(0xFFCCE8E8),
    ),
    AccentPalette(
        key = "tarot",
        label = "秘仪金",
        lightPrimary = Color(0xFF7A5900),
        lightOnPrimary = Color(0xFFFFFFFF),
        lightPrimaryContainer = Color(0xFFFFE08D),
        lightOnPrimaryContainer = Color(0xFF251A00),
        darkPrimary = Color(0xFFECC248),
        darkOnPrimary = Color(0xFF3F2E00),
        darkPrimaryContainer = Color(0xFF5B4300),
        darkOnPrimaryContainer = Color(0xFFFFE08D),
    ),
    AccentPalette(
        key = "rosewood",
        label = "暮玫",
        lightPrimary = Color(0xFF8A4A5E),
        lightOnPrimary = Color(0xFFFFFFFF),
        lightPrimaryContainer = Color(0xFFFFD9E1),
        lightOnPrimaryContainer = Color(0xFF37091C),
        darkPrimary = Color(0xFFFFB1C0),
        darkOnPrimary = Color(0xFF51112B),
        darkPrimaryContainer = Color(0xFF6C2B44),
        darkOnPrimaryContainer = Color(0xFFFFD9E1),
    ),
    AccentPalette(
        key = "pine",
        label = "松间绿",
        lightPrimary = Color(0xFF4C6355),
        lightOnPrimary = Color(0xFFFFFFFF),
        lightPrimaryContainer = Color(0xFFCFE9D6),
        lightOnPrimaryContainer = Color(0xFF0A1F13),
        darkPrimary = Color(0xFFB1CCB9),
        darkOnPrimary = Color(0xFF1D3526),
        darkPrimaryContainer = Color(0xFF334B3B),
        darkOnPrimaryContainer = Color(0xFFCFE9D6),
    ),
)

fun accentPalette(key: String): AccentPalette =
    AccentPalettes.firstOrNull { it.key == key } ?: AccentPalettes.first()
