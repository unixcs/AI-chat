package com.unixcs.aichat.ui.util

/**
 * Minimal Markdown subset for chat replies. It deliberately avoids a heavy
 * CommonMark dependency: the backend steers models towards headings, lists,
 * emphasis and fenced code, which is all this renders.
 *
 * Parsing is pure so it can be unit-tested without Compose.
 */
sealed interface MarkdownBlock {
    data class Paragraph(val spans: List<InlineSpan>) : MarkdownBlock

    data class Heading(val level: Int, val spans: List<InlineSpan>) : MarkdownBlock

    data class Bullet(val depth: Int, val spans: List<InlineSpan>) : MarkdownBlock

    data class Ordered(val number: Int, val depth: Int, val spans: List<InlineSpan>) : MarkdownBlock

    data class Quote(val spans: List<InlineSpan>) : MarkdownBlock

    data class Code(val language: String, val code: String) : MarkdownBlock

    data object Divider : MarkdownBlock
}

data class InlineSpan(
    val text: String,
    val bold: Boolean = false,
    val italic: Boolean = false,
    val code: Boolean = false,
    val strike: Boolean = false,
    val linkUrl: String? = null,
)

private val headingPattern = Regex("^(#{1,6})\\s+(.*)$")
private val bulletPattern = Regex("^(\\s*)[-*+]\\s+(.*)$")
private val orderedPattern = Regex("^(\\s*)(\\d{1,9})[.)]\\s+(.*)$")

fun parseMarkdown(source: String): List<MarkdownBlock> {
    val lines = source.replace("\r\n", "\n").split('\n')
    val blocks = mutableListOf<MarkdownBlock>()
    val paragraph = mutableListOf<String>()

    fun flushParagraph() {
        if (paragraph.isEmpty()) return
        blocks += MarkdownBlock.Paragraph(parseInline(paragraph.joinToString("\n")))
        paragraph.clear()
    }

    var i = 0
    while (i < lines.size) {
        val line = lines[i]
        val trimmedStart = line.trimStart()

        val fence = fenceOpening(trimmedStart)
        if (fence != null) {
            flushParagraph()
            val code = mutableListOf<String>()
            i++
            while (i < lines.size && !isFenceClosing(lines[i].trimStart(), fence)) {
                code += lines[i]
                i++
            }
            if (i < lines.size) i++ // consume the closing fence
            blocks += MarkdownBlock.Code(fence.language, code.joinToString("\n"))
            continue
        }

        if (line.isBlank()) {
            flushParagraph()
            i++
            continue
        }

        val heading = headingPattern.matchEntire(line.trimEnd())
        if (heading != null) {
            flushParagraph()
            val text = heading.groupValues[2].trimEnd().removeSuffix("#").trimEnd()
            blocks += MarkdownBlock.Heading(
                level = heading.groupValues[1].length,
                spans = parseInline(text),
            )
            i++
            continue
        }

        val ordered = orderedPattern.matchEntire(line)
        if (ordered != null) {
            flushParagraph()
            blocks += MarkdownBlock.Ordered(
                number = ordered.groupValues[2].toInt(),
                depth = indentDepth(ordered.groupValues[1]),
                spans = parseInline(ordered.groupValues[3]),
            )
            i++
            continue
        }

        val bullet = bulletPattern.matchEntire(line)
        if (bullet != null) {
            flushParagraph()
            blocks += MarkdownBlock.Bullet(
                depth = indentDepth(bullet.groupValues[1]),
                spans = parseInline(bullet.groupValues[2]),
            )
            i++
            continue
        }

        if (trimmedStart.startsWith(">")) {
            flushParagraph()
            val quote = mutableListOf<String>()
            while (i < lines.size && lines[i].trimStart().startsWith(">")) {
                quote += lines[i].trimStart().removePrefix(">").removePrefix(" ")
                i++
            }
            blocks += MarkdownBlock.Quote(parseInline(quote.joinToString("\n")))
            continue
        }

        if (isDivider(trimmedStart)) {
            flushParagraph()
            blocks += MarkdownBlock.Divider
            i++
            continue
        }

        paragraph += trimmedStart.trimEnd()
        i++
    }
    flushParagraph()
    return blocks
}

private data class Fence(val marker: Char, val count: Int, val language: String)

private fun fenceOpening(line: String): Fence? {
    val marker = line.firstOrNull() ?: return null
    if (marker != '`' && marker != '~') return null
    val count = line.takeWhile { it == marker }.length
    if (count < 3) return null
    val info = line.drop(count).trim()
    // A backtick fence may not carry backticks in its info string.
    if (marker == '`' && info.contains('`')) return null
    return Fence(marker, count, info.substringBefore(' ').trim())
}

private fun isFenceClosing(line: String, fence: Fence): Boolean {
    val count = line.takeWhile { it == fence.marker }.length
    return count >= fence.count && line.drop(count).isBlank()
}

private fun isDivider(line: String): Boolean {
    val trimmed = line.trim()
    if (trimmed.length < 3) return false
    return trimmed.all { it == '-' } || trimmed.all { it == '*' } || trimmed.all { it == '_' }
}

private fun indentDepth(indent: String): Int = (indent.count { it == ' ' } / 2).coerceIn(0, 3)

/**
 * Inline subset: `code`, **bold**, *italic*, ~~strike~~ and [label](url).
 * Unmatched markers stay literal, and the first backtick run wins so code can
 * never be re-interpreted as emphasis.
 */
fun parseInline(
    text: String,
    bold: Boolean = false,
    italic: Boolean = false,
    strike: Boolean = false,
    linkUrl: String? = null,
): List<InlineSpan> {
    val spans = mutableListOf<InlineSpan>()
    val literal = StringBuilder()

    fun flushLiteral() {
        if (literal.isEmpty()) return
        spans += InlineSpan(
            text = literal.toString(),
            bold = bold,
            italic = italic,
            strike = strike,
            linkUrl = linkUrl,
        )
        literal.clear()
    }

    var i = 0
    while (i < text.length) {
        val c = text[i]

        if (c == '\\' && i + 1 < text.length) {
            literal.append(text[i + 1])
            i += 2
            continue
        }

        if (c == '`') {
            val run = runLength(text, i, '`')
            val close = findRun(text, i + run, '`', run)
            if (close >= 0) {
                flushLiteral()
                spans += InlineSpan(
                    text = text.substring(i + run, close),
                    bold = bold,
                    italic = italic,
                    code = true,
                    strike = strike,
                    linkUrl = linkUrl,
                )
                i = close + run
                continue
            }
            literal.append(text, i, i + run)
            i += run
            continue
        }

        if (!strike && text.startsWith("~~", i)) {
            val close = text.indexOf("~~", i + 2)
            if (close > i + 2) {
                flushLiteral()
                spans += parseInline(
                    text = text.substring(i + 2, close),
                    bold = bold,
                    italic = italic,
                    strike = true,
                    linkUrl = linkUrl,
                )
                i = close + 2
                continue
            }
        }

        if (text.startsWith("**", i) || text.startsWith("__", i)) {
            val marker = text.substring(i, i + 2)
            val close = text.indexOf(marker, i + 2)
            if (close > i + 2) {
                flushLiteral()
                spans += parseInline(
                    text = text.substring(i + 2, close),
                    bold = true,
                    italic = italic,
                    strike = strike,
                    linkUrl = linkUrl,
                )
                i = close + 2
                continue
            }
        }

        if (c == '*' || c == '_') {
            val insideWord = c == '_' && i > 0 && (text[i - 1].isLetterOrDigit() || text[i - 1] == '_')
            if (!insideWord) {
                val close = findSingle(text, i + 1, c)
                if (close > i + 1) {
                    flushLiteral()
                    spans += parseInline(
                        text = text.substring(i + 1, close),
                        bold = bold,
                        italic = true,
                        strike = strike,
                        linkUrl = linkUrl,
                    )
                    i = close + 1
                    continue
                }
            }
            literal.append(c)
            i++
            continue
        }

        if (c == '[') {
            val link = parseLink(text, i)
            if (link != null) {
                flushLiteral()
                spans += parseInline(
                    text = link.label,
                    bold = bold,
                    italic = italic,
                    strike = strike,
                    linkUrl = link.url,
                )
                i = link.end
                continue
            }
        }

        literal.append(c)
        i++
    }
    flushLiteral()
    return spans
}

private fun runLength(text: String, start: Int, marker: Char): Int {
    var end = start
    while (end < text.length && text[end] == marker) end++
    return end - start
}

private fun findRun(text: String, start: Int, marker: Char, length: Int): Int {
    var i = start
    while (i < text.length) {
        if (text[i] == marker) {
            val run = runLength(text, i, marker)
            if (run == length) return i
            i += run
        } else {
            i++
        }
    }
    return -1
}

private fun findSingle(text: String, start: Int, marker: Char): Int {
    var i = start
    while (i < text.length) {
        if (text[i] == '\\') {
            i += 2
            continue
        }
        if (text[i] == marker) {
            if (marker == '_' && i + 1 < text.length && text[i + 1].isLetterOrDigit()) {
                i++
                continue
            }
            return i
        }
        i++
    }
    return -1
}

private data class ParsedLink(val label: String, val url: String, val end: Int)

private fun parseLink(text: String, start: Int): ParsedLink? {
    var depth = 1
    var i = start + 1
    var close = -1
    while (i < text.length) {
        when (text[i]) {
            '\\' -> i++
            '[' -> depth++
            ']' -> {
                depth--
                if (depth == 0) {
                    close = i
                    break
                }
            }
        }
        i++
    }
    if (close < 0 || close + 1 >= text.length || text[close + 1] != '(') return null
    val end = text.indexOf(')', close + 2)
    if (end < 0) return null
    val label = text.substring(start + 1, close)
    val url = text.substring(close + 2, end).trim()
    if (label.isEmpty() || url.isEmpty()) return null
    return ParsedLink(label, url, end + 1)
}
