package com.unixcs.aichat.data.settings

import android.content.Context
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.appSettingsDataStore by preferencesDataStore(name = "app_settings")

/**
 * Device-local UI preference (not synced to the account): the visibility of
 * the answer-preference entry. Theming is static (v1.0.0 baseline) — the
 * theme-mode/accent machinery was the recurring freeze trigger and is gone.
 */
class AppSettingsStore(context: Context) {
    private val appContext = context.applicationContext

    val showPreferenceEntry: Flow<Boolean> =
        appContext.appSettingsDataStore.data.map { it[KEY_SHOW_PREFERENCE_ENTRY] ?: true }

    suspend fun setShowPreferenceEntry(visible: Boolean) {
        appContext.appSettingsDataStore.edit { it[KEY_SHOW_PREFERENCE_ENTRY] = visible }
    }

    private companion object {
        val KEY_SHOW_PREFERENCE_ENTRY = booleanPreferencesKey("show_preference_entry")
    }
}
