package com.unixcs.aichat.ui.util

import org.junit.Assert.assertEquals
import org.junit.Test

class MaskTest {

    @Test
    fun masksTheMiddleFourDigits() {
        assertEquals("138****8000", maskPhone("13800138000"))
    }

    @Test
    fun leavesNonStandardLengthUntouched() {
        assertEquals("12345", maskPhone("12345"))
    }
}
