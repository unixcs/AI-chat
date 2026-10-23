package com.unixcs.aichat.data.api

import java.util.concurrent.TimeUnit
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class ApiClientTest {

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
    fun loginSuccessReturnsTokenAndPostsJson() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody("""{"code":0,"data":{"token":"jwt-abc"}}"""),
        )

        val result = api.login("13800138000", "secret123")

        assertTrue(result is ApiResult.Success)
        assertEquals("jwt-abc", (result as ApiResult.Success<LoginData>).data.token)

        val recorded = server.takeRequest()
        assertEquals("POST", recorded.method)
        assertEquals("/api/auth/login", recorded.path)
        assertTrue(recorded.getHeader("Content-Type").orEmpty().startsWith("application/json"))
        val body = recorded.body.readUtf8()
        assertTrue(body.contains("\"phone\":\"13800138000\""))
        assertTrue(body.contains("\"password\":\"secret123\""))
    }

    @Test
    fun loginFailureSurfacesServerMessage() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(400)
                .setBody("""{"message":"密码错误"}"""),
        )

        val result = api.login("13800138000", "wrong")

        assertTrue(result is ApiResult.Failure)
        result as ApiResult.Failure
        assertEquals(400, result.status)
        assertEquals("密码错误", result.message)
        assertNull(result.code)
    }

    @Test
    fun kickedSessionIsReportedAsUnauthorized() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(401)
                .setBody("""{"message":"账号已在其他设备登录","code":"SESSION_KICKED"}"""),
        )

        val result = api.fetchProfile("stale-token")

        assertTrue(result is ApiResult.Failure)
        result as ApiResult.Failure
        assertTrue(result.unauthorized)
        assertEquals("SESSION_KICKED", result.code)
        assertEquals("账号已在其他设备登录", result.message)
    }

    @Test
    fun profileParsesPreferencesAndIgnoresUnknownFields() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setBody(
                    """
                    {"code":0,"data":{
                      "id":"u1","phone":"13800138000","nickname":"tester",
                      "avatarUrl":"","status":"active","role":"user",
                      "memberExpireAt":"2027-01-01T00:00:00.000Z","createdAt":"2026-01-01T00:00:00.000Z",
                      "answerLength":"short","answerStyle":"concise","answerFormat":"markdown",
                      "unknownField":"ignored"
                    }}
                    """.trimIndent(),
                ),
        )

        val result = api.fetchProfile("jwt-abc")

        assertTrue(result is ApiResult.Success)
        val profile = (result as ApiResult.Success<UserProfile>).data
        assertEquals("13800138000", profile.phone)
        assertEquals("short", profile.answerLength)
        assertEquals("concise", profile.answerStyle)
        assertEquals("markdown", profile.answerFormat)
        assertEquals("2027-01-01T00:00:00.000Z", profile.memberExpireAt)
    }

    @Test
    fun authenticatedRequestCarriesBearerToken() = runTest {
        server.enqueue(MockResponse().setResponseCode(200).setBody("""{"code":0,"data":{"ok":true}}"""))

        val result = api.sessionCheck("jwt-abc")

        assertTrue(result is ApiResult.Success)
        assertEquals(true, (result as ApiResult.Success<SessionCheck>).data.ok)
        assertEquals("Bearer jwt-abc", server.takeRequest().getHeader("Authorization"))
    }

    @Test
    fun malformedEnvelopeFailsWithoutCrashing() = runTest {
        server.enqueue(MockResponse().setResponseCode(200).setBody("<html>gateway</html>"))

        val result = api.login("13800138000", "secret123")

        assertTrue(result is ApiResult.Failure)
        assertEquals("服务器响应异常", (result as ApiResult.Failure).message)
    }
}
