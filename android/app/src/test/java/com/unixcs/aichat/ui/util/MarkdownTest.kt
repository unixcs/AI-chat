package com.unixcs.aichat.ui.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class MarkdownTest {

    @Test
    fun parsesHeadingsParagraphsAndLists() {
        val blocks = parseMarkdown(
            """
            # 标题

            第一行
            第二行

            - 项目一
              - 子项目
            1. 第一步
            2. 第二步

            ---
            """.trimIndent(),
        )

        assertEquals(7, blocks.size)
        assertEquals(MarkdownBlock.Heading(1, listOf(InlineSpan("标题"))), blocks[0])

        val paragraph = blocks[1] as MarkdownBlock.Paragraph
        assertEquals(listOf(InlineSpan("第一行\n第二行")), paragraph.spans)

        assertEquals(MarkdownBlock.Bullet(0, listOf(InlineSpan("项目一"))), blocks[2])
        assertEquals(MarkdownBlock.Bullet(1, listOf(InlineSpan("子项目"))), blocks[3])
        assertEquals(MarkdownBlock.Ordered(1, 0, listOf(InlineSpan("第一步"))), blocks[4])
        assertEquals(MarkdownBlock.Ordered(2, 0, listOf(InlineSpan("第二步"))), blocks[5])
        assertEquals(MarkdownBlock.Divider, blocks[6])
    }

    @Test
    fun fencedCodeKeepsMarkersAndLanguage() {
        val blocks = parseMarkdown(
            """
            说明

            ```kotlin
            val a = "**不是粗体**"
            // - 不是列表
            ```

            结尾
            """.trimIndent(),
        )

        assertEquals(3, blocks.size)
        assertEquals(MarkdownBlock.Paragraph(listOf(InlineSpan("说明"))), blocks[0])
        val code = blocks[1] as MarkdownBlock.Code
        assertEquals("kotlin", code.language)
        assertEquals("val a = \"**不是粗体**\"\n// - 不是列表", code.code)
        assertEquals(MarkdownBlock.Paragraph(listOf(InlineSpan("结尾"))), blocks[2])
    }

    @Test
    fun unclosedFenceConsumesTheRemainder() {
        val blocks = parseMarkdown("```\nline 1\nline 2")

        assertEquals(1, blocks.size)
        assertEquals(MarkdownBlock.Code("", "line 1\nline 2"), blocks[0])
    }

    @Test
    fun parsesInlineEmphasisCodeAndStrike() {
        val spans = parseInline("**粗体** 和 *斜体* 以及 ~~删除~~ 和 `code`")

        assertEquals(
            listOf(
                InlineSpan("粗体", bold = true),
                InlineSpan(" 和 "),
                InlineSpan("斜体", italic = true),
                InlineSpan(" 以及 "),
                InlineSpan("删除", strike = true),
                InlineSpan(" 和 "),
                InlineSpan("code", code = true),
            ),
            spans,
        )
    }

    @Test
    fun codeSpanIsNotReinterpretedAsEmphasis() {
        val spans = parseInline("**a `b * c` d**")

        assertEquals(
            listOf(
                InlineSpan("a ", bold = true),
                InlineSpan("b * c", bold = true, code = true),
                InlineSpan(" d", bold = true),
            ),
            spans,
        )
    }

    @Test
    fun parsesLinksAndKeepsEscapesLiteral() {
        val link = parseInline("[DeepSeek](https://chat.deepseek.com)")
        assertEquals(
            listOf(InlineSpan("DeepSeek", linkUrl = "https://chat.deepseek.com")),
            link,
        )

        val escaped = parseInline("\\*不是斜体\\*")
        assertEquals(listOf(InlineSpan("*不是斜体*")), escaped)
    }

    @Test
    fun underscoresInsideIdentifiersStayLiteral() {
        assertEquals(
            listOf(InlineSpan("snake_case_name")),
            parseInline("snake_case_name"),
        )
    }

    @Test
    fun groupsConsecutiveQuoteLines() {
        val blocks = parseMarkdown("> 第一行\n> 第二行")

        assertEquals(1, blocks.size)
        assertEquals(
            MarkdownBlock.Quote(listOf(InlineSpan("第一行\n第二行"))),
            blocks[0],
        )
    }

    @Test
    fun blankInputProducesNoBlocks() {
        assertTrue(parseMarkdown("").isEmpty())
        assertTrue(parseMarkdown("\n\n").isEmpty())
    }
}
