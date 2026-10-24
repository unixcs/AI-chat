package com.unixcs.aichat.ui.theme

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ThemeOptionsTest {

    @Test
    fun themeModeDecodesKnownKeysAndFallsBackToSystem() {
        assertEquals(ThemeMode.System, ThemeMode.fromKey("system"))
        assertEquals(ThemeMode.Light, ThemeMode.fromKey("light"))
        assertEquals(ThemeMode.Dark, ThemeMode.fromKey("dark"))
        assertEquals(ThemeMode.System, ThemeMode.fromKey("unknown"))
        assertEquals(ThemeMode.System, ThemeMode.fromKey(null))
    }

    @Test
    fun accentPalettesHaveUniqueKeysAndResolveWithFallback() {
        val keys = AccentPalettes.map { it.key }
        assertEquals(keys.size, keys.distinct().size)
        assertTrue(AccentPalettes.size >= 5)

        val first = AccentPalettes.first()
        assertEquals(first, accentPalette(first.key))
        // Unknown keys fall back to the first palette instead of crashing.
        assertEquals(first, accentPalette("nonexistent"))
    }
}
