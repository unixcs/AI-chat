package com.unixcs.aichat.data.settings

import android.content.Context
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.appSettingsDataStore by preferencesDataStore(name = "app_settings")

/**
 * Device-local UI preferences (not synced to the account): the visibility of
 * the answer-preference entry plus the appearance (theme mode and accent).
 */
class AppSettingsStore(context: Context) {
    private val appContext = context.applicationContext

    val showPreferenceEntry: Flow<Boolean> =
        appContext.appSettingsDataStore.data.map { it[KEY_SHOW_PREFERENCE_ENTRY] ?: true }

    suspend fun setShowPreferenceEntry(visible: Boolean) {
        appContext.appSettingsDataStore.edit { it[KEY_SHOW_PREFERENCE_ENTRY] = visible }
    }

    /** "system" | "light" | "dark" — see ThemeMode.fromKey for decoding. */
    val themeMode: Flow<String> =
        appContext.appSettingsDataStore.data.map { it[KEY_THEME_MODE] ?: "system" }

    /** Accent palette key, one of ui.theme.AccentPalettes. */
    val themeColor: Flow<String> =
        appContext.appSettingsDataStore.data.map { it[KEY_THEME_COLOR] ?: "violet" }

    suspend fun setThemeMode(mode: String) {
        appContext.appSettingsDataStore.edit { it[KEY_THEME_MODE] = mode }
    }

    suspend fun setThemeColor(color: String) {
        appContext.appSettingsDataStore.edit { it[KEY_THEME_COLOR] = color }
    }

    private companion object {
        val KEY_SHOW_PREFERENCE_ENTRY = booleanPreferencesKey("show_preference_entry")
        val KEY_THEME_MODE = stringPreferencesKey("theme_mode")
        val KEY_THEME_COLOR = stringPreferencesKey("theme_color")
    }
}
