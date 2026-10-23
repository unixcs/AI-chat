package com.unixcs.aichat.data.local

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase

/**
 * Local cache of server-owned conversations. The server is authoritative; if
 * the schema ever changes the cache is dropped rather than migrated.
 */
@Database(
    entities = [ConversationEntity::class, MessageEntity::class],
    version = 1,
    exportSchema = false,
)
abstract class AppDatabase : RoomDatabase() {
    abstract fun chatDao(): ChatDao

    companion object {
        fun build(context: Context): AppDatabase =
            Room.databaseBuilder(context, AppDatabase::class.java, "aichat.db")
                .fallbackToDestructiveMigration()
                .build()
    }
}
