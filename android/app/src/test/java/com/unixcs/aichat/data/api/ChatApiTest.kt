package com.unixcs.aichat.data.api

import java.util.concurrent.TimeUnit
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class ChatApiTest {

    private lateinit var server: MockWebServer
    private lateinit var api: ApiClient

    @Before
    fun setUp() {
        server = MockWebServer().apply { start() }
        api = ApiClient(
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
    fun listConversationsRequestsFirstPageAndParsesArray() = runTest {
        server.enqueue(
            MockResponse().setResponseCode(200).setBody(
                """
                {"code":0,"data":[
                  {"id":"c1","userId":"u1","title":"新对话 10-24 12:00",
                   "createdAt":"2026-10-24T04:00:00.000Z","updatedAt":"2026-10-24T04:10:00.000Z",
                   "userPhone":"13900000001"}
                ]}
                """.trimIndent(),
            ),
        )

        val result = api.listConversations("jwt-abc")

        assertTrue(result is ApiResult.Success)
        val list = (result as ApiResult.Success<List<ConversationDto>>).data
        assertEquals(1, list.size)
        assertEquals("c1", list[0].id)
        assertEquals("新对话 10-24 12:00", list[0].title)

        val recorded = server.takeRequest()
        assertEquals("GET", recorded.method)
        assertEquals("/api/chat/conversations?page=1&pageSize=500", recorded.path)
        assertEquals("Bearer jwt-abc", recorded.getHeader("Authorization"))
    }

    @Test
    fun createConversationPostsEmptyObject() = runTest {
        server.enqueue(
            MockResponse().setResponseCode(200).setBody(
                """{"code":0,"data":{"id":"c2","userId":"u1","title":"新对话 10-24 12:00"}}""",
            ),
        )

        val result = api.createConversation("jwt-abc")

        assertTrue(result is ApiResult.Success)
        assertEquals("c2", (result as ApiResult.Success<ConversationDto>).data.id)

        val recorded = server.takeRequest()
        assertEquals("POST", recorded.method)
        assertEquals("/api/chat/conversations", recorded.path)
        assertEquals("{}", recorded.body.readUtf8())
    }

    @Test
    fun fetchMessagesParsesUserAndAssistantRoles() = runTest {
        server.enqueue(
            MockResponse().setResponseCode(200).setBody(
                """
                {"code":0,"data":[
                  {"id":"m1","conversationId":"c1","role":"user","content":"你好","createdAt":"2026-10-24T04:00:01.000Z"},
                  {"id":"m2","conversationId":"c1","role":"assistant","content":"你好！","createdAt":"2026-10-24T04:00:02.000Z"}
                ]}
                """.trimIndent(),
            ),
        )

        val result = api.fetchMessages("jwt-abc", "c1")

        assertTrue(result is ApiResult.Success)
        val messages = (result as ApiResult.Success<List<MessageDto>>).data
        assertEquals(listOf("user", "assistant"), messages.map { it.role })
        assertEquals("你好！", messages[1].content)
        assertEquals("/api/chat/conversations/c1/messages", server.takeRequest().path)
    }

    @Test
    fun deleteConversationUsesDeleteAndParsesBoolean() = runTest {
        server.enqueue(MockResponse().setResponseCode(200).setBody("""{"code":0,"data":true}"""))

        val result = api.deleteConversation("jwt-abc", "c1")

        assertTrue(result is ApiResult.Success)
        assertEquals(true, (result as ApiResult.Success<Boolean>).data)

        val recorded = server.takeRequest()
        assertEquals("DELETE", recorded.method)
        assertEquals("/api/chat/conversations/c1", recorded.path)
    }

    @Test
    fun renameConversationPutsTitleAndParsesUpdatedConversation() = runTest {
        server.enqueue(
            MockResponse().setResponseCode(200).setBody(
                """
                {"code":0,"data":{"id":"c1","userId":"u1","title":"项目周报",
                 "createdAt":"2026-10-24T04:00:00.000Z","updatedAt":"2026-10-24T05:00:00.000Z"}}
                """.trimIndent(),
            ),
        )

        val result = api.renameConversation("jwt-abc", "c1", "项目周报")

        assertTrue(result is ApiResult.Success)
        val conversation = (result as ApiResult.Success<ConversationDto>).data
        assertEquals("项目周报", conversation.title)
        assertEquals("2026-10-24T05:00:00.000Z", conversation.updatedAt)

        val recorded = server.takeRequest()
        assertEquals("PUT", recorded.method)
        assertEquals("/api/chat/conversations/c1/title", recorded.path)
        assertEquals("""{"title":"项目周报"}""", recorded.body.readUtf8())
        assertEquals("Bearer jwt-abc", recorded.getHeader("Authorization"))
    }

    @Test
    fun preferenceUpdateSendsOnlyTheTouchedDimension() = runTest {
        server.enqueue(MockResponse().setResponseCode(200).setBody("""{"code":0,"data":true}"""))

        val result = api.updatePreferences("jwt-abc", answerFormat = "plain")

        assertTrue(result is ApiResult.Success)
        assertEquals(true, (result as ApiResult.Success<Boolean>).data)

        val recorded = server.takeRequest()
        assertEquals("PUT", recorded.method)
        assertEquals("/api/user/preferences", recorded.path)
        // Null dimensions must be omitted, never sent as null: the backend
        // treats an explicit null as a validation error.
        assertEquals("""{"answerFormat":"plain"}""", recorded.body.readUtf8())
    }

    @Test
    fun expiredMembershipSurfacesServerMessage() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(403)
                .setBody("""{"message":"会员过期，请续费后使用"}"""),
        )

        val result = api.listConversations("jwt-abc")

        assertTrue(result is ApiResult.Failure)
        result as ApiResult.Failure
        assertEquals(403, result.status)
        assertEquals("会员过期，请续费后使用", result.message)
    }
}
