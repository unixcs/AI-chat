package com.unixcs.aichat.data.backup

import java.io.ByteArrayOutputStream
import java.util.zip.ZipEntry
import java.util.zip.ZipOutputStream
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class BackupCodecTest {

    private val json = Json {
        ignoreUnknownKeys = true
        explicitNulls = false
    }

    private val data = BackupData(
        conversations = listOf(
            BackupConversation(
                id = "c1",
                title = "项目周报",
                createdAt = "2026-10-01T08:00:00Z",
                updatedAt = "2026-10-02T09:30:00Z",
            ),
        ),
        messages = listOf(
            BackupMessage(
                id = "m1",
                conversationId = "c1",
                role = "user",
                content = "你好",
                createdAt = "2026-10-01T08:00:00Z",
            ),
            BackupMessage(
                id = "m2",
                conversationId = "c1",
                role = "assistant",
                content = "你好，有什么可以帮你？",
                createdAt = "2026-10-01T08:00:05Z",
            ),
        ),
    )

    @Test
    fun zipAndUnzipRoundTrip() {
        val zip = BackupCodec.zip(data, "2026-10-24T12:00:00Z", "0.6.0", json)

        assertEquals(data, BackupCodec.unzip(zip, json))
    }

    @Test
    fun emptyBackupRoundTrips() {
        val empty = BackupData()
        val zip = BackupCodec.zip(empty, "2026-10-24T12:00:00Z", "0.6.0", json)

        assertEquals(empty, BackupCodec.unzip(zip, json))
    }

    @Test
    fun manifestRecordsCountsAndAppVersion() {
        val zip = BackupCodec.zip(data, "2026-10-24T12:00:00Z", "0.6.0", json)
        val manifest = decodeManifest(zip)

        assertEquals("aichat-backup", manifest.format)
        assertEquals(1, manifest.version)
        assertEquals("2026-10-24T12:00:00Z", manifest.exportedAt)
        assertEquals("0.6.0", manifest.appVersion)
        assertEquals(1, manifest.conversationCount)
        assertEquals(2, manifest.messageCount)
    }

    @Test
    fun unknownEntriesAreRejected() {
        val zip = zipOf(
            "manifest.json" to validManifest(),
            "data.json" to json.encodeToString(BackupData()),
            "payload.bin" to "x",
        )

        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip(zip, json)
        }
    }

    @Test
    fun missingEntriesAreRejected() {
        val withoutManifest = zipOf("data.json" to json.encodeToString(BackupData()))
        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip(withoutManifest, json)
        }

        val withoutData = zipOf("manifest.json" to validManifest())
        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip(withoutData, json)
        }
    }

    @Test
    fun unsupportedFormatOrVersionIsRejected() {
        val wrongFormat = zipOf(
            "manifest.json" to """{"format":"other","version":1}""",
            "data.json" to json.encodeToString(BackupData()),
        )
        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip(wrongFormat, json)
        }

        val wrongVersion = zipOf(
            "manifest.json" to """{"format":"aichat-backup","version":99}""",
            "data.json" to json.encodeToString(BackupData()),
        )
        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip(wrongVersion, json)
        }
    }

    @Test
    fun malformedJsonIsReportedAsFormatError() {
        val broken = zipOf(
            "manifest.json" to validManifest(),
            "data.json" to "{not json",
        )

        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip(broken, json)
        }
    }

    @Test
    fun nonZipPayloadIsReportedAsFormatError() {
        assertThrows(BackupFormatException::class.java) {
            BackupCodec.unzip("definitely not a zip".toByteArray(), json)
        }
    }

    @Test
    fun unknownJsonFieldsAreIgnoredForForwardCompatibility() {
        val zip = zipOf(
            "manifest.json" to validManifest(),
            "data.json" to
                """{"conversations":[{"id":"c1","title":"t","createdAt":"a","updatedAt":"b","futureField":1}],"messages":[],"extra":true}""",
        )

        val decoded = BackupCodec.unzip(zip, json)
        assertEquals(1, decoded.conversations.size)
        assertEquals("t", decoded.conversations.single().title)
    }

    private fun validManifest(): String =
        """{"format":"aichat-backup","version":1,"exportedAt":"2026-10-24T12:00:00Z","appVersion":"0.6.0","conversationCount":0,"messageCount":0}"""

    private fun zipOf(vararg entries: Pair<String, String>): ByteArray {
        val buffer = ByteArrayOutputStream()
        ZipOutputStream(buffer).use { zip ->
            for ((name, content) in entries) {
                zip.putNextEntry(ZipEntry(name))
                zip.write(content.toByteArray(Charsets.UTF_8))
                zip.closeEntry()
            }
        }
        return buffer.toByteArray()
    }

    private fun decodeManifest(zip: ByteArray): BackupManifest {
        var bytes: ByteArray? = null
        java.util.zip.ZipInputStream(zip.inputStream()).use { input ->
            while (true) {
                val entry = input.nextEntry ?: break
                if (entry.name == "manifest.json") {
                    bytes = input.readBytes()
                    break
                }
                input.closeEntry()
            }
        }
        return json.decodeFromString(String(bytes ?: error("missing manifest"), Charsets.UTF_8))
    }
}
