package com.unixcs.aichat.data.chat

import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.api.map
import com.unixcs.aichat.data.session.SessionStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flow

/**
 * Chat-side API facade. Every call resolves the token at call time so a kicked
 * session is noticed by the next request, not only at startup.
 */
class ChatRepository(
    private val api: ApiClient,
    private val streamClient: ChatStreamClient,
    private val store: SessionStore,
) {

    suspend fun conversations(): ApiResult<List<ChatConversation>> =
        withToken { api.listConversations(it) }.map { list ->
            list.map { dto ->
                ChatConversation(
                    id = dto.id,
                    title = dto.title,
                    createdAt = dto.createdAt,
                    updatedAt = dto.updatedAt,
                )
            }
        }

    suspend fun createConversation(): ApiResult<ChatConversation> =
        withToken { api.createConversation(it) }.map { dto ->
            ChatConversation(
                id = dto.id,
                title = dto.title,
                createdAt = dto.createdAt,
                updatedAt = dto.updatedAt,
            )
        }

    suspend fun messages(conversationId: String): ApiResult<List<ChatMessage>> =
        withToken { api.fetchMessages(it, conversationId) }.map { list ->
            list.map { dto ->
                ChatMessage(
                    id = dto.id,
                    conversationId = dto.conversationId,
                    role = dto.role,
                    content = dto.content,
                    createdAt = dto.createdAt,
                )
            }
        }

    suspend fun deleteConversation(conversationId: String): ApiResult<Boolean> =
        withToken { api.deleteConversation(it, conversationId) }

    fun stream(conversationId: String, content: String): Flow<StreamEvent> = flow {
        val token = store.token.first()
        if (token.isNullOrBlank()) {
            emit(StreamEvent.Failed("登录已失效，请重新登录", "SESSION_KICKED", 401))
            return@flow
        }
        streamClient.stream(token, conversationId, content).collect { event ->
            // SSE errors arrive on an HTTP 200 response, so the kicked code is
            // the only signal that the session is gone.
            if (event is StreamEvent.Failed &&
                (event.status == 401 || event.code == "SESSION_KICKED")
            ) {
                store.expire(event.message)
            }
            emit(event)
        }
    }

    private suspend fun <T> withToken(block: suspend (String) -> ApiResult<T>): ApiResult<T> {
        val token = store.token.first()
        if (token.isNullOrBlank()) {
            return ApiResult.Failure(401, "登录已失效，请重新登录", "SESSION_KICKED")
        }
        val result = block(token)
        if (result is ApiResult.Failure && result.unauthorized) {
            store.expire(result.message)
        }
        return result
    }
}
