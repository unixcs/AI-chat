package com.unixcs.aichat.ui.components

import org.junit.Assert.assertEquals
import org.junit.Test

/**
 * Regression tests for the swipe-to-close bounce: a light quick left-swipe
 * from the fully-open drawer used to settle back OPEN because the settle rule
 * looked only at position (progress > 0.5) and ignored release velocity.
 */
class ShellSettleTest {

    private val flickPx = 150f

    @Test
    fun lightQuickLeftSwipeFromFullyOpenCloses() {
        // The reported bug: swipe closed a little, fast — must not bounce back.
        assertEquals(0f, settleTarget(value = 0.9f, velocity = -flickPx - 1f, flickPx = flickPx))
    }

    @Test
    fun lightQuickRightSwipeFromClosedOpens() {
        assertEquals(1f, settleTarget(value = 0.1f, velocity = flickPx + 1f, flickPx = flickPx))
    }

    @Test
    fun slowReleasePastHalfStaysOpen() {
        assertEquals(1f, settleTarget(value = 0.6f, velocity = 0f, flickPx = flickPx))
    }

    @Test
    fun slowReleaseUnderHalfCloses() {
        assertEquals(0f, settleTarget(value = 0.4f, velocity = 0f, flickPx = flickPx))
    }

    @Test
    fun boundaryHalfWithNoVelocityCloses() {
        assertEquals(0f, settleTarget(value = SETTLE_THRESHOLD, velocity = 0f, flickPx = flickPx))
    }
}
