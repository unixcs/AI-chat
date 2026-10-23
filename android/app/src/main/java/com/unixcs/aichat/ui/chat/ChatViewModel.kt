package com.unixcs.aichat.ui.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.unixcs.aichat.data.api.ApiResult
import com.unixcs.aichat.data.chat.ChatConversation
import com.unixcs.aichat.data.chat.ChatMessage
import com.unixcs.aichat.data.chat.ChatRepository
import com.unixcs.aichat.data.chat.StreamEvent
import com.unixcs.aichat.data.local.ChatDao
import com.unixcs.aichat.data.local.ConversationEntity
import com.unixcs.aichat.data.local.MessageEntity
import com.unixcs.aichat.data.session.SessionStore
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.drop
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ChatUiState(
    val conversationId: String? = null,
    val title: String = "",
    val input: String = "",
    val messages: List<ChatMessage> = emptyList(),
    val loading: Boolean = false,
    val streaming: Boolean = false,
    val streamingText: String = "",
    val error: String? = null,
) {
    val canSend: Boolean get() = input.isNotBlank() && !streaming && !loading
}

class ChatViewModel(
    private val repo: ChatRepository,
    private val dao: ChatDao,
    private val store: SessionStore,
) : ViewModel() {

    private val _ui = MutableStateFlow(ChatUiState())
    val ui: StateFlow<ChatUiState> = _ui.asStateFlow()

    val conversations: StateFlow<List<ChatConversation>> = dao.observeConversations()
        .map { list ->
            list.map { ChatConversation(it.id, it.title, it.createdAt, it.updatedAt) }
        }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), emptyList())

    private var streamJob: Job? = null

    init {
        viewModelScope.launch { refreshConversations() }
        viewModelScope.launch {
            conversations.collect { list ->
                val id = _ui.value.conversationId ?: return@collect
                val title = list.firstOrNull { it.id == id }?.title
                if (!title.isNullOrBlank() && title != _ui.value.title) {
                    _ui.update { if (it.conversationId == id) it.copy(title = title) else it }
                }
            }
        }
        // Logout or a kicked session must not leak the previous account's
        // local cache or draft into the next login.
        viewModelScope.launch {
            store.token.drop(1).collect { token ->
                if (token.isNullOrBlank()) resetForLogout()
            }
        }
    }

    fun onInputChange(value: String) {
        _ui.update { it.copy(input = value, error = null) }
    }

    fun dismissError() {
        _ui.update { it.copy(error = null) }
    }

    fun startNewChat() {
        streamJob?.cancel()
        streamJob = null
        _ui.value = ChatUiState()
    }

    fun openConversation(conversationId: String) {
        if (_ui.value.conversationId == conversationId && !_ui.value.loading) return
        streamJob?.cancel()
        streamJob = null
        _ui.value = ChatUiState(
            conversationId = conversationId,
            title = conversations.value.firstOrNull { it.id == conversationId }?.title.orEmpty(),
            loading = true,
        )
        viewModelScope.launch {
            val cached = dao.getMessages(conversationId)
            if (_ui.value.conversationId != conversationId) return@launch
            if (cached.isNotEmpty()) {
                _ui.update { it.copy(messages = cached.map { it.toChatMessage() }, loading = false) }
            }
            when (val result = repo.messages(conversationId)) {
                is ApiResult.Success -> {
                    dao.replaceMessages(conversationId, result.data.map { it.toEntity() })
                    if (_ui.value.conversationId == conversationId) {
                        _ui.update { it.copy(messages = result.data, loading = false, error = null) }
                    }
                }

                is ApiResult.Failure -> if (_ui.value.conversationId == conversationId) {
                    _ui.update { it.copy(loading = false, error = result.message) }
                }
            }
        }
    }

    fun send() {
        val state = _ui.value
        val text = state.input.trim()
        if (text.isEmpty() || state.streaming || state.loading) return
        _ui.update { it.copy(input = "", error = null) }

        viewModelScope.launch {
            val conversationId = ensureConversation(text) ?: return@launch
            _ui.update {
                if (it.conversationId != conversationId) return@update it
                it.copy(
                    messages = it.messages + ChatMessage(
                        id = "local-${System.nanoTime()}",
                        conversationId = conversationId,
                        role = "user",
                        content = text,
                        createdAt = "",
                    ),
                    streaming = true,
                    streamingText = "",
                    error = null,
                )
            }
            startStream(conversationId, text)
        }
    }

    fun stopStreaming() {
        val state = _ui.value
        val conversationId = state.conversationId ?: return
        if (!state.streaming) return
        streamJob?.cancel()
        streamJob = null
        val partial = state.streamingText
        _ui.update {
            it.copy(
                streaming = false,
                streamingText = "",
                messages = if (partial.isNotBlank()) {
                    it.messages + ChatMessage(
                        id = "local-assistant-${System.nanoTime()}",
                        conversationId = conversationId,
                        role = "assistant",
                        content = partial,
                        createdAt = "",
                    )
                } else {
                    it.messages
                },
            )
        }
        viewModelScope.launch {
            // The server persists the aborted partial reply just after cancel;
            // give it a moment before replacing the optimistic bubble.
            delay(500)
            if (_ui.value.conversationId == conversationId) {
                refreshMessages(conversationId)
            }
            refreshConversations()
        }
    }

    fun deleteConversation(conversationId: String) {
        viewModelScope.launch {
            when (val result = repo.deleteConversation(conversationId)) {
                is ApiResult.Success -> {
                    dao.deleteConversationCascade(conversationId)
                    if (_ui.value.conversationId == conversationId) {
                        startNewChat()
                    }
                    refreshConversations()
                }

                is ApiResult.Failure -> _ui.update { it.copy(error = result.message) }
            }
        }
    }

    private suspend fun ensureConversation(text: String): String? {
        _ui.value.conversationId?.let { return it }
        return when (val result = repo.createConversation()) {
            is ApiResult.Success -> {
                dao.upsertConversations(listOf(result.data.toEntity()))
                _ui.update { it.copy(conversationId = result.data.id, title = result.data.title) }
                result.data.id
            }

            is ApiResult.Failure -> {
                _ui.update { it.copy(input = text, error = result.message) }
                null
            }
        }
    }

    private fun startStream(conversationId: String, text: String) {
        streamJob?.cancel()
        streamJob = viewModelScope.launch {
            repo.stream(conversationId, text).collect { event ->
                when (event) {
                    is StreamEvent.Delta -> _ui.update {
                        if (it.conversationId == conversationId) {
                            it.copy(streamingText = it.streamingText + event.text)
                        } else {
                            it
                        }
                    }

                    StreamEvent.Completed -> Unit

                    is StreamEvent.Failed -> _ui.update {
                        if (it.conversationId != conversationId) return@update it
                        it.copy(
                            error = event.message,
                            // Membership is checked before the server inserts
                            // the user message, so it is safe to restore input.
                            input = if (event.status == 403 && it.input.isBlank()) text else it.input,
                        )
                    }
                }
            }
            finishStream(conversationId)
        }
    }

    private suspend fun finishStream(conversationId: String) {
        if (_ui.value.conversationId != conversationId) return
        val partial = _ui.value.streamingText
        when (val result = repo.messages(conversationId)) {
            is ApiResult.Success -> {
                dao.replaceMessages(conversationId, result.data.map { it.toEntity() })
                _ui.update {
                    if (it.conversationId != conversationId) {
                        it
                    } else {
                        it.copy(messages = result.data, streaming = false, streamingText = "")
                    }
                }
            }

            is ApiResult.Failure -> _ui.update {
                if (it.conversationId != conversationId) return@update it
                val draft = if (partial.isNotBlank()) {
                    listOf(
                        ChatMessage(
                            id = "local-assistant-${System.nanoTime()}",
                            conversationId = conversationId,
                            role = "assistant",
                            content = partial,
                            createdAt = "",
                        ),
                    )
                } else {
                    emptyList()
                }
                it.copy(messages = it.messages + draft, streaming = false, streamingText = "")
            }
        }
        refreshConversations()
    }

    private suspend fun refreshMessages(conversationId: String) {
        when (val result = repo.messages(conversationId)) {
            is ApiResult.Success -> {
                dao.replaceMessages(conversationId, result.data.map { it.toEntity() })
                _ui.update {
                    if (it.conversationId == conversationId) it.copy(messages = result.data) else it
                }
            }

            is ApiResult.Failure -> Unit
        }
    }

    private suspend fun refreshConversations() {
        when (val result = repo.conversations()) {
            is ApiResult.Success -> {
                val items = result.data.map { it.toEntity() }
                if (items.isEmpty()) {
                    dao.clearAll()
                } else {
                    dao.upsertConversations(items)
                    dao.deleteConversationsNotIn(items.map { it.id })
                    dao.pruneOrphanMessages()
                }
            }

            is ApiResult.Failure -> Unit
        }
    }

    private fun resetForLogout() {
        streamJob?.cancel()
        streamJob = null
        _ui.value = ChatUiState()
        viewModelScope.launch { dao.clearAll() }
    }

    private fun MessageEntity.toChatMessage() = ChatMessage(id, conversationId, role, content, createdAt)

    private fun ChatConversation.toEntity() = ConversationEntity(id, title, createdAt, updatedAt)

    private fun ChatMessage.toEntity() = MessageEntity(id, conversationId, role, content, createdAt)
}
