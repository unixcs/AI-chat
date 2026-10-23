package com.unixcs.aichat.data.chat

data class ChatConversation(
    val id: String,
    val title: String,
    val createdAt: String,
    val updatedAt: String,
)

data class ChatMessage(
    val id: String,
    val conversationId: String = "",
    val role: String,
    val content: String,
    val createdAt: String,
)

/**
 * One terminal signal per stream: either [Completed] or [Failed]. Deltas may
 * arrive before a failure — the backend persists partial text on abort, so the
 * ViewModel always refreshes messages afterwards.
 */
sealed interface StreamEvent {
    data class Delta(val text: String) : StreamEvent

    data object Completed : StreamEvent

    data class Failed(
        val message: String,
        val code: String? = null,
        val status: Int = 0,
    ) : StreamEvent
}
