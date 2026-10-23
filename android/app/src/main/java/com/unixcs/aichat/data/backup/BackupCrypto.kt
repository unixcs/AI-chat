package com.unixcs.aichat.data.backup

import java.nio.charset.StandardCharsets
import java.security.GeneralSecurityException
import java.security.SecureRandom
import javax.crypto.AEADBadTagException
import javax.crypto.Cipher
import javax.crypto.SecretKeyFactory
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.PBEKeySpec

class BackupFormatException(message: String) : Exception(message)

/** Wrong password and a corrupted ciphertext are indistinguishable by design. */
class BackupPasswordException : Exception("密码错误或文件已损坏")

/**
 * Encrypted file layout:
 *
 * ```
 * "AICHAT1" (7) | version (1) | salt (16) | iv (12) | AES-256-GCM ciphertext
 * ```
 *
 * The key is derived from the user passphrase with PBKDF2-HMAC-SHA256. The
 * passphrase itself is never stored, logged or sent anywhere; callers remain
 * responsible for wiping their char arrays.
 */
object BackupCrypto {
    private val MAGIC = "AICHAT1".toByteArray(StandardCharsets.US_ASCII)
    private const val VERSION_OFFSET = 7
    private const val SALT_OFFSET = 8
    private const val SALT_LENGTH = 16
    private const val IV_OFFSET = SALT_OFFSET + SALT_LENGTH
    private const val IV_LENGTH = 12
    private const val HEADER_LENGTH = IV_OFFSET + IV_LENGTH
    private const val TAG_BITS = 128
    private const val ITERATIONS = 120_000
    private const val KEY_BITS = 256
    private const val MIN_PAYLOAD_LENGTH = 16 // GCM tag
    private val random = SecureRandom()

    fun encrypt(plaintext: ByteArray, password: CharArray): ByteArray {
        val salt = ByteArray(SALT_LENGTH).also(random::nextBytes)
        val iv = ByteArray(IV_LENGTH).also(random::nextBytes)
        val key = deriveKey(password, salt)
        try {
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.ENCRYPT_MODE, key, GCMParameterSpec(TAG_BITS, iv))
            return header(salt, iv) + cipher.doFinal(plaintext)
        } finally {
            key.fill(0)
        }
    }

    fun decrypt(payload: ByteArray, password: CharArray): ByteArray {
        if (payload.size < HEADER_LENGTH + MIN_PAYLOAD_LENGTH) {
            throw BackupFormatException("备份文件不完整")
        }
        if (!payload.copyOfRange(0, MAGIC.size).contentEquals(MAGIC)) {
            throw BackupFormatException("不是 AI-chat 备份文件")
        }
        val version = payload[VERSION_OFFSET].toInt() and 0xFF
        if (version != BACKUP_VERSION) {
            throw BackupFormatException("不支持的备份版本：$version")
        }
        val salt = payload.copyOfRange(SALT_OFFSET, SALT_OFFSET + SALT_LENGTH)
        val iv = payload.copyOfRange(IV_OFFSET, HEADER_LENGTH)
        val ciphertext = payload.copyOfRange(HEADER_LENGTH, payload.size)
        val key = deriveKey(password, salt)
        try {
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.DECRYPT_MODE, key, GCMParameterSpec(TAG_BITS, iv))
            return cipher.doFinal(ciphertext)
        } catch (_: AEADBadTagException) {
            throw BackupPasswordException()
        } catch (_: GeneralSecurityException) {
            throw BackupPasswordException()
        } finally {
            key.fill(0)
        }
    }

    private fun header(salt: ByteArray, iv: ByteArray): ByteArray {
        val header = ByteArray(HEADER_LENGTH)
        MAGIC.copyInto(header, 0)
        header[VERSION_OFFSET] = BACKUP_VERSION.toByte()
        salt.copyInto(header, SALT_OFFSET)
        iv.copyInto(header, IV_OFFSET)
        return header
    }

    private fun deriveKey(password: CharArray, salt: ByteArray): ByteArray {
        val spec = PBEKeySpec(password, salt, ITERATIONS, KEY_BITS)
        return try {
            SecretKeyFactory.getInstance("PBKDF2WithHmacSHA256")
                .generateSecret(spec)
                .encoded
        } catch (e: GeneralSecurityException) {
            throw IllegalStateException("PBKDF2-HMAC-SHA256 unavailable", e)
        } finally {
            spec.clearPassword()
        }
    }
}
