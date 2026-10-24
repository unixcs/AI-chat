package com.unixcs.aichat.data.local

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import kotlinx.coroutines.flow.Flow

data class BackupSnapshot(
    val conversations: List<ConversationEntity>,
    val messages: List<MessageEntity>,
)

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

    @Query("SELECT * FROM conversations")
    abstract suspend fun getAllConversations(): List<ConversationEntity>

    @Query("SELECT * FROM messages")
    abstract suspend fun getAllMessages(): List<MessageEntity>

    /** Consistent snapshot for a backup export (both queries in one transaction). */
    @Transaction
    open suspend fun snapshot(): BackupSnapshot =
        BackupSnapshot(getAllConversations(), getAllMessages())

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    abstract suspend fun upsertConversations(items: List<ConversationEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    abstract suspend fun insertMessages(items: List<MessageEntity>)

    /** Merge a backup into the cache; existing rows with the same ids win. */
    @Transaction
    open suspend fun mergeBackup(
        conversations: List<ConversationEntity>,
        messages: List<MessageEntity>,
    ) {
        upsertConversations(conversations)
        insertMessages(messages)
    }

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

    @Query("DELETE FROM messages WHERE conversationId IN (:ids)")
    abstract suspend fun deleteMessagesIn(ids: List<String>)

    @Query("DELETE FROM conversations WHERE id IN (:ids)")
    abstract suspend fun deleteConversationsIn(ids: List<String>)

    /** One transaction for a whole bulk delete so the cache never half-clears. */
    @Transaction
    open suspend fun deleteConversationsCascade(ids: List<String>) {
        if (ids.isEmpty()) return
        deleteMessagesIn(ids)
        deleteConversationsIn(ids)
    }

    @Query("SELECT * FROM messages WHERE conversationId = :conversationId ORDER BY createdAt, id")
    abstract suspend fun getMessages(conversationId: String): List<MessageEntity>

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
