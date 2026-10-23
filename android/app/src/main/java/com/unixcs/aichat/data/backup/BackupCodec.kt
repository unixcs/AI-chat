package com.unixcs.aichat.data.backup

import java.io.ByteArrayOutputStream
import java.io.InputStream
import java.util.zip.ZipEntry
import java.util.zip.ZipInputStream
import java.util.zip.ZipOutputStream
import kotlinx.serialization.SerializationException
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

/**
 * Plaintext container inside the encrypted envelope: a two-entry ZIP holding
 * `manifest.json` and `data.json`. Unknown entries are rejected rather than
 * skipped so a manipulated archive cannot smuggle unbounded data through.
 */
object BackupCodec {
    private const val MANIFEST_ENTRY = "manifest.json"
    private const val DATA_ENTRY = "data.json"
    private const val MAX_ENTRY_BYTES = 32 * 1024 * 1024
    private const val MAX_TOTAL_BYTES = 64 * 1024 * 1024
    private const val MAX_CONVERSATIONS = 20_000
    private const val MAX_MESSAGES = 500_000

    fun zip(
        data: BackupData,
        exportedAt: String,
        appVersion: String,
        json: Json,
    ): ByteArray {
        val manifest = BackupManifest(
            exportedAt = exportedAt,
            appVersion = appVersion,
            conversationCount = data.conversations.size,
            messageCount = data.messages.size,
        )
        val buffer = ByteArrayOutputStream()
        ZipOutputStream(buffer).use { zip ->
            zip.putNextEntry(ZipEntry(MANIFEST_ENTRY))
            zip.write(json.encodeToString(manifest).toByteArray(Charsets.UTF_8))
            zip.closeEntry()

            zip.putNextEntry(ZipEntry(DATA_ENTRY))
            zip.write(json.encodeToString(data).toByteArray(Charsets.UTF_8))
            zip.closeEntry()
        }
        return buffer.toByteArray()
    }

    fun unzip(bytes: ByteArray, json: Json): BackupData {
        var manifestBytes: ByteArray? = null
        var dataBytes: ByteArray? = null
        var total = 0

        ZipInputStream(bytes.inputStream()).use { zip ->
            while (true) {
                val entry = zip.nextEntry ?: break
                if (entry.isDirectory) {
                    zip.closeEntry()
                    continue
                }
                if (entry.name != MANIFEST_ENTRY && entry.name != DATA_ENTRY) {
                    throw BackupFormatException("备份文件包含未知内容")
                }
                val entryBytes = readBounded(zip, MAX_ENTRY_BYTES)
                total += entryBytes.size
                if (total > MAX_TOTAL_BYTES) {
                    throw BackupFormatException("备份内容过大")
                }
                if (entry.name == MANIFEST_ENTRY) {
                    manifestBytes = entryBytes
                } else {
                    dataBytes = entryBytes
                }
                zip.closeEntry()
            }
        }

        val manifest = decode<BackupManifest>(
            json,
            manifestBytes ?: throw BackupFormatException("备份文件缺少清单"),
        )
        if (manifest.format != BACKUP_FORMAT || manifest.version != BACKUP_VERSION) {
            throw BackupFormatException("不支持的备份格式")
        }

        val data = decode<BackupData>(
            json,
            dataBytes ?: throw BackupFormatException("备份文件缺少数据"),
        )
        if (data.conversations.size > MAX_CONVERSATIONS || data.messages.size > MAX_MESSAGES) {
            throw BackupFormatException("备份数据量超出限制")
        }
        return data
    }

    private inline fun <reified T> decode(json: Json, bytes: ByteArray): T = try {
        json.decodeFromString<T>(String(bytes, Charsets.UTF_8))
    } catch (_: SerializationException) {
        throw BackupFormatException("备份数据无法解析")
    } catch (_: IllegalArgumentException) {
        throw BackupFormatException("备份数据无法解析")
    }

    private fun readBounded(input: InputStream, limit: Int): ByteArray {
        val output = ByteArrayOutputStream()
        val buffer = ByteArray(8 * 1024)
        while (true) {
            val read = input.read(buffer)
            if (read < 0) break
            output.write(buffer, 0, read)
            if (output.size() > limit) {
                throw BackupFormatException("备份条目过大")
            }
        }
        return output.toByteArray()
    }
}
