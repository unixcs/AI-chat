package com.unixcs.aichat.data.session

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.sessionDataStore by preferencesDataStore(name = "session")

/**
 * Persists the JWT and the phone number shown in the drawer/settings.
 * DataStore keeps the session across process death; clearing the token is the
 * single source of truth for "logged out".
 */
class SessionStore(context: Context) {
    private val appContext = context.applicationContext

    val token: Flow<String?> = appContext.sessionDataStore.data.map { it[KEY_TOKEN] }
    val phone: Flow<String?> = appContext.sessionDataStore.data.map { it[KEY_PHONE] }

    suspend fun saveSession(token: String, phone: String) {
        appContext.sessionDataStore.edit { prefs ->
            prefs[KEY_TOKEN] = token
            prefs[KEY_PHONE] = phone
        }
    }

    suspend fun savePhone(phone: String) {
        appContext.sessionDataStore.edit { prefs ->
            if (phone.isNotBlank()) prefs[KEY_PHONE] = phone
        }
    }

    suspend fun clear() {
        appContext.sessionDataStore.edit { prefs ->
            prefs.remove(KEY_TOKEN)
            prefs.remove(KEY_PHONE)
        }
    }

    private companion object {
        val KEY_TOKEN = stringPreferencesKey("token")
        val KEY_PHONE = stringPreferencesKey("phone")
    }
}
