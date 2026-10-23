package com.unixcs.aichat.ui.util

import com.unixcs.aichat.data.chat.ChatConversation
import java.time.LocalDate
import java.time.OffsetDateTime
import java.time.ZoneId

data class HistorySection(val label: String, val items: List<ChatConversation>)

/**
 * Filters by title and buckets conversations the way chat apps usually do.
 * Input is already newest-first, and each bucket preserves that order.
 */
fun groupConversations(
    conversations: List<ChatConversation>,
    query: String,
    today: LocalDate = LocalDate.now(),
    zone: ZoneId = ZoneId.systemDefault(),
): List<HistorySection> {
    val needle = query.trim()
    val filtered = if (needle.isEmpty()) {
        conversations
    } else {
        conversations.filter { it.title.contains(needle, ignoreCase = true) }
    }

    val buckets = linkedMapOf<String, MutableList<ChatConversation>>(
        "今天" to mutableListOf(),
        "昨天" to mutableListOf(),
        "近 7 天" to mutableListOf(),
        "更早" to mutableListOf(),
    )
    filtered.forEach { conversation ->
        val date = localDate(conversation.updatedAt, zone)
        val label = when {
            date == null -> "更早"
            date == today -> "今天"
            date == today.minusDays(1) -> "昨天"
            date.isAfter(today.minusDays(7)) -> "近 7 天"
            else -> "更早"
        }
        buckets.getValue(label).add(conversation)
    }
    return buckets
        .filterValues { it.isNotEmpty() }
        .map { (label, items) -> HistorySection(label, items) }
}

private fun localDate(raw: String, zone: ZoneId): LocalDate? = try {
    OffsetDateTime.parse(raw).atZoneSameInstant(zone).toLocalDate()
} catch (_: Exception) {
    null
}
