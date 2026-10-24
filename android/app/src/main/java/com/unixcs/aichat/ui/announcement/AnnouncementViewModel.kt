package com.unixcs.aichat.ui.announcement

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.unixcs.aichat.data.api.AnnouncementDto
import com.unixcs.aichat.data.session.SessionStore
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Serves the unread-announcement popup on the chat screen. Fetched once per
 * ChatScreen entry (the VM is scoped to the CHAT back stack entry, so it dies
 * with logout and survives the round-trip to Settings — no repeat popups, no
 * cross-account leaks). The popup waits ~600ms after entry so the chat list
 * paints first; acknowledging hides the card, and a failed ack self-heals by
 * popping again next launch.
 */
class AnnouncementViewModel(
    private val repo: AnnouncementRepository,
    private val store: SessionStore,
) : ViewModel() {

    private val _pending = MutableStateFlow<AnnouncementDto?>(null)
    val pending: StateFlow<AnnouncementDto?> = _pending.asStateFlow()

    private var acking = false

    init {
        viewModelScope.launch {
            delay(600)
            val token = store.token.first()
            if (!token.isNullOrBlank()) {
                _pending.value = repo.current(token)
            }
        }
    }

    /** 「我知道了」: hide now, ack in the fire-and-forget background. */
    fun acknowledge() {
        val announcement = _pending.value ?: return
        _pending.value = null
        if (acking) return
        acking = true
        viewModelScope.launch {
            val token = store.token.first()
            if (!token.isNullOrBlank()) {
                repo.ack(token, announcement)
            }
            acking = false
        }
    }

    /** Scrim/back dismiss without ack: the announcement stays pending server-side. */
    fun dismiss() {
        _pending.value = null
    }
}
