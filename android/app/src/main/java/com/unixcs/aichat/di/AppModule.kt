package com.unixcs.aichat.di

import com.unixcs.aichat.BuildConfig
import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.backup.ChatBackupRepository
import com.unixcs.aichat.data.chat.ChatRepository
import com.unixcs.aichat.data.chat.ChatStreamClient
import com.unixcs.aichat.data.local.AppDatabase
import com.unixcs.aichat.data.prefs.PreferencesRepository
import com.unixcs.aichat.data.session.AuthRepository
import com.unixcs.aichat.data.session.SessionStore
import com.unixcs.aichat.data.settings.AppSettingsStore
import com.unixcs.aichat.ui.chat.ChatViewModel
import com.unixcs.aichat.ui.backup.BackupViewModel
import com.unixcs.aichat.ui.session.SessionViewModel
import com.unixcs.aichat.ui.login.LoginViewModel
import com.unixcs.aichat.ui.settings.AccountRepository
import com.unixcs.aichat.ui.settings.SettingsViewModel
import java.util.concurrent.TimeUnit
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import org.koin.android.ext.koin.androidContext
import org.koin.androidx.viewmodel.dsl.viewModel
import org.koin.core.qualifier.named
import org.koin.dsl.module

private const val STREAM_CLIENT = "stream-client"

val appModule = module {
    single {
        Json {
            ignoreUnknownKeys = true
            explicitNulls = false
        }
    }

    single {
        OkHttpClient.Builder()
            .connectTimeout(10, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()
    }

    // SSE replies can stay silent while the model thinks, so the streaming
    // client must not inherit the 30 s JSON read timeout.
    single(named(STREAM_CLIENT)) {
        get<OkHttpClient>().newBuilder()
            .readTimeout(0, TimeUnit.MILLISECONDS)
            .build()
    }

    single { ApiClient(BuildConfig.API_BASE_URL, get(), get()) }
    single { SessionStore(androidContext()) }
    single { AppSettingsStore(androidContext()) }
    single { AuthRepository(get(), get()) }
    single { PreferencesRepository(get(), get()) }
    single {
        ChatBackupRepository(
            dao = get(),
            contentResolver = androidContext().contentResolver,
            json = get(),
            appVersion = BuildConfig.VERSION_NAME,
        )
    }

    single { AppDatabase.build(androidContext()) }
    single { get<AppDatabase>().chatDao() }
    single { ChatStreamClient(BuildConfig.API_BASE_URL, get(named(STREAM_CLIENT)), get()) }
    single { ChatRepository(get(), get(), get()) }
    single { AccountRepository(get()) }

    viewModel { LoginViewModel(get()) }
    viewModel { SessionViewModel(get(), get()) }
    viewModel { ChatViewModel(get(), get(), get(), get(), get()) }
    viewModel { BackupViewModel(get()) }
    viewModel { SettingsViewModel(get(), get()) }
}
