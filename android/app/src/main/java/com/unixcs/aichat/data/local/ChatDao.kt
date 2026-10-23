package com.unixcs.aichat.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import kotlinx.coroutines.flow.Flow

/**
 * Abstract class instead of an interface: Room's KSP processor handles
 * concrete @Transaction bodies reliably in classes across Kotlin JVM-default
 * modes, while still letting the generated implementation provide the
 * abstract queries.
 */
@Dao
abstract class ChatDao {

    @Query("SELECT * FROM conversations ORDER BY updatedAt DESC")
    abstract fun observeConversations(): Flow<List<ConversationEntity>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    abstract suspend fun upsertConversations(items: List<ConversationEntity>)

    @Query("DELETE FROM conversations WHERE id NOT IN (:ids)")
    abstract suspend fun deleteConversationsNotIn(ids: List<String>)

    @Query("DELETE FROM conversations")
    abstract suspend fun clearConversations()

    @Query("SELECT * FROM conversations WHERE id = :conversationId")
    abstract suspend fun getConversation(conversationId: String): ConversationEntity?

    @Query("DELETE FROM messages WHERE conversationId = :conversationId")
    abstract suspend fun deleteMessages(conversationId: String)

    @Query("DELETE FROM conversations WHERE id = :conversationId")
    abstract suspend fun deleteConversation(conversationId: String)

    @Transaction
    open suspend fun deleteConversationCascade(conversationId: String) {
        deleteMessages(conversationId)
        deleteConversation(conversationId)
    }

    @Query("SELECT * FROM messages WHERE conversationId = :conversationId ORDER BY createdAt, id")
    abstract suspend fun getMessages(conversationId: String): List<MessageEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    abstract suspend fun insertMessages(items: List<MessageEntity>)

    @Transaction
    open suspend fun replaceMessages(conversationId: String, items: List<MessageEntity>) {
        deleteMessages(conversationId)
        insertMessages(items)
    }

    @Query("DELETE FROM messages WHERE conversationId NOT IN (SELECT id FROM conversations)")
    abstract suspend fun pruneOrphanMessages()

    @Transaction
    open suspend fun clearAll() {
        deleteMessagesAll()
        clearConversations()
    }

    @Query("DELETE FROM messages")
    abstract suspend fun deleteMessagesAll()
}
