package com.unixcs.aichat.data.backup

import kotlinx.serialization.Serializable

const val BACKUP_FORMAT = "aichat-backup"
const val BACKUP_VERSION = 1

/**
 * Stable on-disk shape of a backup. It is intentionally separate from the Room
 * entities so a local schema change can never silently alter the file format.
 */
@Serializable
data class BackupConversation(
    val id: String,
    val title: String,
    val createdAt: String,
    val updatedAt: String,
)

@Serializable
data class BackupMessage(
    val id: String,
    val conversationId: String,
    val role: String,
    val content: String,
    val createdAt: String,
)

@Serializable
data class BackupData(
    val conversations: List<BackupConversation> = emptyList(),
    val messages: List<BackupMessage> = emptyList(),
)

@Serializable
data class BackupManifest(
    val format: String = BACKUP_FORMAT,
    val version: Int = BACKUP_VERSION,
    val exportedAt: String = "",
    val appVersion: String = "",
    val conversationCount: Int = 0,
    val messageCount: Int = 0,
) {
    companion object {
        const val FORMAT = BACKUP_FORMAT
        const val VERSION = BACKUP_VERSION
    }
}
