package com.unixcs.aichat.data.backup

import android.content.ContentResolver
import android.net.Uri
import com.unixcs.aichat.data.local.ChatDao
import com.unixcs.aichat.data.local.ConversationEntity
import com.unixcs.aichat.data.local.MessageEntity
import java.io.ByteArrayOutputStream
import java.io.IOException
import java.io.InputStream
import java.time.Instant
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json

sealed interface BackupResult {
    data class Exported(val conversations: Int, val messages: Int) : BackupResult

    data class Imported(val conversations: Int, val messages: Int) : BackupResult

    data class Failure(val message: String) : BackupResult
}

/**
 * Local-only backup. The account password is never involved: the user chooses a
 * separate backup passphrase, which encrypts the archive before it leaves the
 * app through the Storage Access Framework.
 *
 * Import merges instead of replacing, so restoring an old backup cannot silently
 * delete newer local conversations. The server stays authoritative: on the next
 * refresh, conversations the account no longer owns are pruned.
 */
class ChatBackupRepository(
    private val dao: ChatDao,
    private val contentResolver: ContentResolver,
    private val json: Json,
    private val appVersion: String,
) {

    suspend fun export(uri: Uri, password: CharArray): BackupResult = withContext(Dispatchers.IO) {
        try {
            val snapshot = dao.snapshot()
            val data = BackupData(
                conversations = snapshot.conversations.map {
                    BackupConversation(it.id, it.title, it.createdAt, it.updatedAt)
                },
                messages = snapshot.messages.map {
                    BackupMessage(it.id, it.conversationId, it.role, it.content, it.createdAt)
                },
            )
            val zip = BackupCodec.zip(
                data = data,
                exportedAt = Instant.now().toString(),
                appVersion = appVersion,
                json = json,
            )
            val encrypted = BackupCrypto.encrypt(zip, password)
            val output = contentResolver.openOutputStream(uri, "wt")
                ?: return@withContext BackupResult.Failure("无法写入所选位置")
            output.use { it.write(encrypted) }
            BackupResult.Exported(data.conversations.size, data.messages.size)
        } catch (_: IOException) {
            BackupResult.Failure("导出失败，请检查存储空间或更换保存位置")
        } catch (_: SecurityException) {
            BackupResult.Failure("没有写入该位置的权限")
        } finally {
            password.fill('\u0000')
        }
    }

    suspend fun import(uri: Uri, password: CharArray): BackupResult = withContext(Dispatchers.IO) {
        try {
            val encrypted = contentResolver.openInputStream(uri)?.use {
                readBounded(it, MAX_ENCRYPTED_BYTES)
            } ?: return@withContext BackupResult.Failure("无法读取所选文件")
            val zip = BackupCrypto.decrypt(encrypted, password)
            val data = BackupCodec.unzip(zip, json)

            val conversations = data.conversations
                .filter { it.id.isNotBlank() }
                .distinctBy { it.id }
                .map { ConversationEntity(it.id, it.title, it.createdAt, it.updatedAt) }
            val knownIds = conversations.mapTo(HashSet()) { it.id }
            val messages = data.messages
                .filter {
                    it.id.isNotBlank() &&
                        it.conversationId in knownIds &&
                        (it.role == "user" || it.role == "assistant")
                }
                .distinctBy { it.id }
                .map { MessageEntity(it.id, it.conversationId, it.role, it.content, it.createdAt) }

            dao.mergeBackup(conversations, messages)
            BackupResult.Imported(conversations.size, messages.size)
        } catch (_: BackupPasswordException) {
            BackupResult.Failure("密码错误或文件已损坏")
        } catch (e: BackupFormatException) {
            BackupResult.Failure(e.message ?: "备份文件格式不正确")
        } catch (_: IOException) {
            BackupResult.Failure("读取备份失败")
        } catch (_: SecurityException) {
            BackupResult.Failure("没有读取该文件的权限")
        } finally {
            password.fill('\u0000')
        }
    }

    private fun readBounded(input: InputStream, limit: Int): ByteArray {
        val output = ByteArrayOutputStream()
        val buffer = ByteArray(8 * 1024)
        while (true) {
            val read = input.read(buffer)
            if (read < 0) break
            output.write(buffer, 0, read)
            if (output.size() > limit) {
                throw BackupFormatException("备份文件过大")
            }
        }
        return output.toByteArray()
    }

    private companion object {
        const val MAX_ENCRYPTED_BYTES = 64 * 1024 * 1024
    }
}
