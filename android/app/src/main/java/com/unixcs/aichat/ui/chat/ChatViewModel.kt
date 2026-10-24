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
import com.unixcs.aichat.data.prefs.ChatPreferences
import com.unixcs.aichat.data.prefs.PreferencesRepository
import com.unixcs.aichat.data.session.SessionStore
import com.unixcs.aichat.data.settings.AppSettingsStore
import com.unixcs.aichat.ui.settings.AppearanceState
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
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
    private val prefsRepo: PreferencesRepository,
    private val settings: AppSettingsStore,
    private val dao: ChatDao,
    private val store: SessionStore,
) : ViewModel() {

    private val _ui = MutableStateFlow(ChatUiState())
    val ui: StateFlow<ChatUiState> = _ui.asStateFlow()

    private val _preferences = MutableStateFlow(ChatPreferences())
    val preferences: StateFlow<ChatPreferences> = _preferences.asStateFlow()

    /** Device-local toggle: hide the preference entry from the chat top bar. */
    val showPreferenceEntry: StateFlow<Boolean> = settings.showPreferenceEntry
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), true)

    /** Appearance (theme mode + accent), read straight from DataStore. */
    val appearance: StateFlow<AppearanceState> = combine(
        settings.themeMode,
        settings.themeColor,
    ) { mode, color -> AppearanceState(mode, color) }
        .stateIn(viewModelScope, SharingStarted.Eagerly, AppearanceState())

    /** One-tap light/dark switch on the chat top bar. */
    fun toggleDarkMode() {
        val next = if (appearance.value.themeMode == "dark") "light" else "dark"
        viewModelScope.launch { settings.setThemeMode(next) }
    }

    /** Clipboard feedback hook; reserved for future telemetry/haptics polish. */
    fun notifyCopied() = Unit

    val conversations: StateFlow<List<ChatConversation>> = dao.observeConversations()
        .map { list ->
            list.map { ChatConversation(it.id, it.title, it.createdAt, it.updatedAt) }
        }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), emptyList())

    private var streamJob: Job? = null

    init {
        viewModelScope.launch { refreshConversations() }
        viewModelScope.launch { refreshPreferences() }
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
                if (token.isNullOrBlank()) {
                    resetForLogout()
                } else {
                    refreshPreferences()
                }
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

    fun renameConversation(conversationId: String, newTitle: String) {
        val title = newTitle.trim()
        if (title.isEmpty()) return
        val known = conversations.value.firstOrNull { it.id == conversationId }?.title
        if (title == known) return
        viewModelScope.launch {
            when (val result = repo.renameConversation(conversationId, title)) {
                is ApiResult.Success -> {
                    dao.upsertConversations(listOf(result.data.toEntity()))
                    _ui.update {
                        if (it.conversationId == conversationId) it.copy(title = result.data.title) else it
                    }
                }

                is ApiResult.Failure -> {
                    _ui.update { it.copy(error = result.message) }
                    refreshConversations()
                }
            }
        }
    }

    fun setAnswerLength(value: String) = updatePreference(
        value = value,
        read = ChatPreferences::answerLength,
        apply = { prefs, v -> prefs.copy(answerLength = v) },
        request = { prefsRepo.updateLength(value) },
    )

    fun setAnswerStyle(value: String) = updatePreference(
        value = value,
        read = ChatPreferences::answerStyle,
        apply = { prefs, v -> prefs.copy(answerStyle = v) },
        request = { prefsRepo.updateStyle(value) },
    )

    fun setAnswerFormat(value: String) = updatePreference(
        value = value,
        read = ChatPreferences::answerFormat,
        apply = { prefs, v -> prefs.copy(answerFormat = v) },
        request = { prefsRepo.updateFormat(value) },
    )

    fun setPreferenceEntryVisible(visible: Boolean) {
        viewModelScope.launch { settings.setShowPreferenceEntry(visible) }
    }

    /**
     * Optimistic update: the chip flips immediately, and only this one
     * dimension is sent. A failure reverts that dimension (unless the user has
     * already picked something newer) and surfaces the server message.
     */
    private fun updatePreference(
        value: String,
        read: (ChatPreferences) -> String,
        apply: (ChatPreferences, String) -> ChatPreferences,
        request: suspend () -> ApiResult<Boolean>,
    ) {
        val previous = read(_preferences.value)
        if (value == previous) return
        _preferences.update { apply(it, value) }
        viewModelScope.launch {
            val result = request()
            if (result is ApiResult.Failure) {
                _preferences.update { if (read(it) == value) apply(it, previous) else it }
                _ui.update { it.copy(error = result.message) }
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

    private suspend fun refreshPreferences() {
        when (val result = prefsRepo.load()) {
            is ApiResult.Success -> _preferences.value = result.data
            // Defaults stay usable; a later successful load will replace them.
            is ApiResult.Failure -> Unit
        }
    }

    private fun resetForLogout() {
        streamJob?.cancel()
        streamJob = null
        _ui.value = ChatUiState()
        _preferences.value = ChatPreferences()
        viewModelScope.launch { dao.clearAll() }
    }

    private fun MessageEntity.toChatMessage() = ChatMessage(id, conversationId, role, content, createdAt)

    private fun ChatConversation.toEntity() = ConversationEntity(id, title, createdAt, updatedAt)

    private fun ChatMessage.toEntity() = MessageEntity(id, conversationId, role, content, createdAt)
}
