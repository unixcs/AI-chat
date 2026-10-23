package com.unixcs.aichat.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.ExperimentalTextApi
import androidx.compose.ui.text.LinkAnnotation
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextLinkStyles
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.withLink
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import com.unixcs.aichat.ui.util.InlineSpan
import com.unixcs.aichat.ui.util.MarkdownBlock
import com.unixcs.aichat.ui.util.parseMarkdown

/**
 * Read-only Markdown renderer used for assistant replies. User messages stay
 * plain text: they are sent verbatim, so styling them would misrepresent them.
 */
@OptIn(ExperimentalTextApi::class)
@Composable
fun MarkdownText(
    markdown: String,
    modifier: Modifier = Modifier,
) {
    val blocks = remember(markdown) { parseMarkdown(markdown) }
    val codeBackground = MaterialTheme.colorScheme.surface.copy(alpha = 0.7f)

    SelectionContainer {
        Column(
            modifier = modifier,
            verticalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            blocks.forEach { block ->
                when (block) {
                    is MarkdownBlock.Paragraph -> RichText(
                        spans = block.spans,
                        style = MaterialTheme.typography.bodyLarge,
                        codeBackground = codeBackground,
                    )

                    is MarkdownBlock.Heading -> RichText(
                        spans = block.spans,
                        style = when (block.level) {
                            1 -> MaterialTheme.typography.titleLarge
                            2 -> MaterialTheme.typography.titleMedium
                            else -> MaterialTheme.typography.titleSmall
                        }.copy(fontWeight = FontWeight.Bold),
                        codeBackground = codeBackground,
                    )

                    is MarkdownBlock.Bullet -> ListRow(
                        marker = "•",
                        depth = block.depth,
                        spans = block.spans,
                        codeBackground = codeBackground,
                    )

                    is MarkdownBlock.Ordered -> ListRow(
                        marker = "${block.number}.",
                        depth = block.depth,
                        spans = block.spans,
                        codeBackground = codeBackground,
                    )

                    is MarkdownBlock.Quote -> QuoteBlock(
                        spans = block.spans,
                        codeBackground = codeBackground,
                    )

                    is MarkdownBlock.Code -> CodeBlock(
                        language = block.language,
                        code = block.code,
                        background = codeBackground,
                    )

                    MarkdownBlock.Divider -> HorizontalDivider(
                        color = MaterialTheme.colorScheme.outlineVariant,
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalTextApi::class)
@Composable
private fun RichText(
    spans: List<InlineSpan>,
    style: TextStyle,
    codeBackground: Color,
    modifier: Modifier = Modifier,
    color: Color = Color.Unspecified,
) {
    val linkColor = MaterialTheme.colorScheme.primary
    val annotated = remember(spans, codeBackground, linkColor) {
        buildAnnotatedString {
            spans.forEach { span ->
                val spanStyle = span.toSpanStyle(codeBackground, linkColor)
                withStyle(spanStyle) {
                    val url = span.linkUrl
                    if (url != null) {
                        withLink(
                            LinkAnnotation.Url(
                                url = url,
                                styles = TextLinkStyles(style = spanStyle),
                            ),
                        ) {
                            append(span.text)
                        }
                    } else {
                        append(span.text)
                    }
                }
            }
        }
    }
    Text(text = annotated, style = style, modifier = modifier, color = color)
}

private fun InlineSpan.toSpanStyle(codeBackground: Color, linkColor: Color): SpanStyle = SpanStyle(
    fontWeight = if (bold) FontWeight.SemiBold else null,
    fontStyle = if (italic) FontStyle.Italic else null,
    textDecoration = when {
        strike -> TextDecoration.LineThrough
        linkUrl != null -> TextDecoration.Underline
        else -> null
    },
    fontFamily = if (code) FontFamily.Monospace else null,
    background = if (code) codeBackground else Color.Unspecified,
    color = if (linkUrl != null) linkColor else Color.Unspecified,
)

@OptIn(ExperimentalTextApi::class)
@Composable
private fun ListRow(
    marker: String,
    depth: Int,
    spans: List<InlineSpan>,
    codeBackground: Color,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = (depth * 16).dp),
    ) {
        Text(
            text = marker,
            style = MaterialTheme.typography.bodyLarge,
            modifier = Modifier.widthIn(min = 22.dp),
        )
        Spacer(Modifier.width(4.dp))
        RichText(
            spans = spans,
            style = MaterialTheme.typography.bodyLarge,
            codeBackground = codeBackground,
            modifier = Modifier.weight(1f),
        )
    }
}

@OptIn(ExperimentalTextApi::class)
@Composable
private fun QuoteBlock(
    spans: List<InlineSpan>,
    codeBackground: Color,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(IntrinsicSize.Min)
            .padding(vertical = 2.dp),
    ) {
        Box(
            modifier = Modifier
                .width(3.dp)
                .fillMaxHeight()
                .clip(RoundedCornerShape(2.dp))
                .background(MaterialTheme.colorScheme.outline),
        )
        Spacer(Modifier.width(10.dp))
        RichText(
            spans = spans,
            style = MaterialTheme.typography.bodyMedium,
            codeBackground = codeBackground,
            modifier = Modifier.weight(1f),
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}

@Composable
private fun CodeBlock(language: String, code: String, background: Color) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .background(background)
            .padding(horizontal = 12.dp, vertical = 10.dp),
    ) {
        if (language.isNotBlank()) {
            Text(
                text = language,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(bottom = 4.dp),
            )
        }
        Text(
            text = code,
            style = MaterialTheme.typography.bodySmall.copy(fontFamily = FontFamily.Monospace),
            modifier = Modifier
                .horizontalScroll(rememberScrollState())
                .heightIn(min = 16.dp),
        )
    }
}
