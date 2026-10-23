package com.unixcs.aichat.ui.util

import java.time.LocalDate
import java.time.OffsetDateTime
import java.time.ZoneId
import java.time.format.DateTimeFormatter

private val timeOnly = DateTimeFormatter.ofPattern("HH:mm")
private val monthDay = DateTimeFormatter.ofPattern("MM-dd HH:mm")

/**
 * Formats a backend ISO timestamp for the history list: time for today,
 * "昨天 HH:mm" for yesterday, "MM-dd HH:mm" otherwise.
 */
fun formatTimestamp(raw: String): String {
    if (raw.isBlank()) return ""
    return try {
        val zone = ZoneId.systemDefault()
        val local = OffsetDateTime.parse(raw).atZoneSameInstant(zone)
        val today = LocalDate.now(zone)
        when (local.toLocalDate()) {
            today -> local.format(timeOnly)
            today.minusDays(1) -> "昨天 " + local.format(timeOnly)
            else -> local.format(monthDay)
        }
    } catch (_: Exception) {
        raw
    }
}
