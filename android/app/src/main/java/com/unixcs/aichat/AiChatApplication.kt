package com.unixcs.aichat

import android.app.Application
import com.unixcs.aichat.di.appModule
import org.koin.android.ext.koin.androidContext
import org.koin.core.context.startKoin

class AiChatApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        startKoin {
            androidContext(this@AiChatApplication)
            modules(appModule)
        }
    }
}
