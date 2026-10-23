package com.unixcs.aichat.ui.util

import com.unixcs.aichat.data.chat.ChatConversation
import java.time.LocalDate
import java.time.ZoneId
import org.junit.Assert.assertEquals
import org.junit.Test

class HistoryGroupingTest {

    private val zone = ZoneId.of("Asia/Shanghai")
    private val today = LocalDate.of(2026, 10, 24)

    private fun conversation(id: String, title: String, updatedAt: String) =
        ChatConversation(id = id, title = title, createdAt = updatedAt, updatedAt = updatedAt)

    @Test
    fun groupsByRecencyKeepingInputOrder() {
        val items = listOf(
            conversation("a", "今天的", "2026-10-24T09:00:00.000Z"),
            conversation("b", "昨天的", "2026-10-23T09:00:00.000Z"),
            conversation("c", "前天的", "2026-10-22T09:00:00.000Z"),
            conversation("d", "上个月的", "2026-09-20T09:00:00.000Z"),
        )

        val sections = groupConversations(items, query = "", today = today, zone = zone)

        assertEquals(listOf("今天", "昨天", "近 7 天", "更早"), sections.map { it.label })
        assertEquals(listOf("a"), sections[0].items.map { it.id })
        assertEquals(listOf("b"), sections[1].items.map { it.id })
        assertEquals(listOf("c"), sections[2].items.map { it.id })
        assertEquals(listOf("d"), sections[3].items.map { it.id })
    }

    @Test
    fun searchMatchesTitleCaseInsensitivelyAndDropsEmptySections() {
        val items = listOf(
            conversation("a", "Tarot Reading", "2026-10-24T09:00:00.000Z"),
            conversation("b", "塔罗运势", "2026-09-20T09:00:00.000Z"),
        )

        val hit = groupConversations(items, query = "tarot", today = today, zone = zone)
        assertEquals(listOf("今天"), hit.map { it.label })
        assertEquals(listOf("a"), hit[0].items.map { it.id })

        val chineseHit = groupConversations(items, query = "运势", today = today, zone = zone)
        assertEquals(listOf("更早"), chineseHit.map { it.label })
        assertEquals(listOf("b"), chineseHit[0].items.map { it.id })

        assertEquals(emptyList<HistorySection>(), groupConversations(items, query = "nope", today = today, zone = zone))
    }

    @Test
    fun malformedTimestampFallsIntoOlderBucket() {
        val items = listOf(conversation("x", "坏时间", "not-a-date"))

        val sections = groupConversations(items, query = "", today = today, zone = zone)

        assertEquals(listOf("更早"), sections.map { it.label })
        assertEquals(listOf("x"), sections[0].items.map { it.id })
    }

    @Test
    fun exactlySevenDaysAgoIsOlder() {
        val items = listOf(conversation("e", "八天前", "2026-10-16T09:00:00.000Z"))

        val sections = groupConversations(items, query = "", today = today, zone = zone)

        assertEquals(listOf("更早"), sections.map { it.label })
    }
}
