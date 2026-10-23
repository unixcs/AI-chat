package com.unixcs.aichat.data.chat

import java.util.concurrent.TimeUnit
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class ChatStreamClientTest {

    private lateinit var server: MockWebServer
    private lateinit var client: ChatStreamClient

    @Before
    fun setUp() {
        server = MockWebServer().apply { start() }
        client = ChatStreamClient(
            baseUrl = server.url("/").toString(),
            client = OkHttpClient.Builder()
                .readTimeout(5, TimeUnit.SECONDS)
                .build(),
            json = Json {
                ignoreUnknownKeys = true
                explicitNulls = false
            },
        )
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    @Test
    fun deltasAreCoalescedAndStreamCompletesOnDone() = runBlocking {
        server.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "text/event-stream")
                .setBody(
                    buildString {
                        append("data: {\"delta\":\"你\"}\n\n")
                        append("data: {\"delta\":\"好\"}\n\n")
                        append("data: [DONE]\n\n")
                    },
                ),
        )

        val events = withTimeout(10_000) {
            client.stream("jwt-abc", "c1", "打个招呼").toList()
        }

        val text = events.filterIsInstance<StreamEvent.Delta>().joinToString("") { it.text }
        assertEquals("你好", text)
        assertTrue(events.last() is StreamEvent.Completed)

        val recorded = server.takeRequest()
        assertEquals("GET", recorded.method)
        assertEquals(
            "/api/chat/conversations/c1/stream?content=%E6%89%93%E4%B8%AA%E6%8B%9B%E5%91%BC",
            recorded.path,
        )
        assertEquals("Bearer jwt-abc", recorded.getHeader("Authorization"))
    }

    @Test
    fun partnerErrorFrameBecomesFailureWithCode() = runBlocking {
        server.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "text/event-stream")
                .setBody("data: {\"code\":\"SESSION_KICKED\",\"error\":\"账号已在其他设备登录\"}\n\n"),
        )

        val events = withTimeout(10_000) {
            client.stream("jwt-abc", "c1", "hi").toList()
        }

        val failure = events.last() as StreamEvent.Failed
        assertEquals("SESSION_KICKED", failure.code)
        assertEquals("账号已在其他设备登录", failure.message)
        assertEquals(0, failure.status)
    }

    @Test
    fun httpErrorCarriesStatusAndJsonMessage() = runBlocking {
        server.enqueue(
            MockResponse()
                .setResponseCode(403)
                .setBody("""{"message":"会员过期，请续费后使用"}"""),
        )

        val events = withTimeout(10_000) {
            client.stream("jwt-abc", "c1", "hi").toList()
        }

        val failure = events.single() as StreamEvent.Failed
        assertEquals(403, failure.status)
        assertEquals("会员过期，请续费后使用", failure.message)
    }
}
