package com.unixcs.aichat.data.backup

import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertThrows
import org.junit.Test

class BackupCryptoTest {

    private val password = "correct horse battery".toCharArray()
    private val plaintext = """{"hello":"世界","n":42}""".toByteArray()

    @Test
    fun roundTripReturnsOriginalBytes() {
        val encrypted = BackupCrypto.encrypt(plaintext, password)

        assertFalse(encrypted.contentEquals(plaintext))
        assertArrayEquals(plaintext, BackupCrypto.decrypt(encrypted, password))
    }

    @Test
    fun wrongPasswordIsReportedAsPasswordFailure() {
        val encrypted = BackupCrypto.encrypt(plaintext, password)

        assertThrows(BackupPasswordException::class.java) {
            BackupCrypto.decrypt(encrypted, "wrong password".toCharArray())
        }
    }

    @Test
    fun tamperedCiphertextIsRejected() {
        val encrypted = BackupCrypto.encrypt(plaintext, password)
        val tampered = encrypted.copyOf()
        tampered[tampered.lastIndex] = (tampered.last().toInt() xor 0x01).toByte()

        assertThrows(BackupPasswordException::class.java) {
            BackupCrypto.decrypt(tampered, password)
        }
    }

    @Test
    fun badMagicAndTruncatedFilesAreFormatErrors() {
        val encrypted = BackupCrypto.encrypt(plaintext, password)

        val badMagic = encrypted.copyOf().also { it[0] = 'X'.code.toByte() }
        assertThrows(BackupFormatException::class.java) {
            BackupCrypto.decrypt(badMagic, password)
        }

        assertThrows(BackupFormatException::class.java) {
            BackupCrypto.decrypt(encrypted.copyOf(20), password)
        }
    }

    @Test
    fun everyEncryptionUsesFreshSaltAndIv() {
        val first = BackupCrypto.encrypt(plaintext, password)
        val second = BackupCrypto.encrypt(plaintext, password)

        // Header = magic(7) + version(1) + salt(16) + iv(12).
        val firstHeader = first.copyOfRange(0, 36)
        val secondHeader = second.copyOfRange(0, 36)
        assertFalse(firstHeader.contentEquals(secondHeader))
        assertFalse(first.contentEquals(second))
    }
}
