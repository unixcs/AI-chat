package com.unixcs.aichat.di

import com.unixcs.aichat.BuildConfig
import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.session.AuthRepository
import com.unixcs.aichat.data.session.SessionStore
import com.unixcs.aichat.ui.session.SessionViewModel
import com.unixcs.aichat.ui.login.LoginViewModel
import java.util.concurrent.TimeUnit
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import org.koin.android.ext.koin.androidContext
import org.koin.androidx.viewmodel.dsl.viewModel
import org.koin.dsl.module

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

    single { ApiClient(BuildConfig.API_BASE_URL, get(), get()) }
    single { SessionStore(androidContext()) }
    single { AuthRepository(get(), get()) }

    viewModel { LoginViewModel(get()) }
    viewModel { SessionViewModel(get(), get()) }
}
