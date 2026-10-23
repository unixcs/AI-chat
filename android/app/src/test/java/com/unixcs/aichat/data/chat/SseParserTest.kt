package com.unixcs.aichat.data.chat

import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class SseParserTest {

    private val parser = SseParser(
        Json {
            ignoreUnknownKeys = true
            explicitNulls = false
        },
    )

    @Test
    fun deltaFrameIsParsed() {
        val event = parser.parse("""data: {"delta":"你好"}""")

        assertEquals(SseEvent.Delta("你好"), event)
    }

    @Test
    fun doneMarkerIsRecognized() {
        assertEquals(SseEvent.Done, parser.parse("data: [DONE]"))
    }

    @Test
    fun errorFrameKeepsCodeAndMessage() {
        val event = parser.parse(
            """data: {"code":"SESSION_KICKED","error":"账号已在其他设备登录"}""",
        )

        assertEquals(SseEvent.Error("SESSION_KICKED", "账号已在其他设备登录"), event)
    }

    @Test
    fun blankLinesCommentsAndEventNamesAreIgnored() {
        assertNull(parser.parse(""))
        assertNull(parser.parse(": keep-alive"))
        assertNull(parser.parse("event: message"))
        assertNull(parser.parse("data:"))
    }

    @Test
    fun malformedPayloadIsDroppedInsteadOfFailingTheStream() {
        assertNull(parser.parse("""data: {"delta":"""))
        assertNull(parser.parse("data: <html>gateway</html>"))
        assertNull(parser.parse("""data: {"unknown":1}"""))
    }
}
