package com.unixcs.aichat.ui.announcement

import com.unixcs.aichat.data.api.ApiClient
import com.unixcs.aichat.data.api.AnnouncementDto
import com.unixcs.aichat.data.api.ApiResult

/**
 * Announcement fetch/ack. Token comes from the caller (ViewModel reads
 * SessionStore); failures are deliberately silent — an announcement is a
 * courtesy popup, never a reason to disturb the session, and an unacked one
 * simply reappears on the next launch (server-side read state is the source
 * of truth). 401 is NOT fed back into SessionStore.expire: session-check and
 * profile calls already own that path and racing them adds nothing.
 */
class AnnouncementRepository(private val api: ApiClient) {

    suspend fun current(token: String): AnnouncementDto? =
        when (val result = api.currentAnnouncement(token)) {
            is ApiResult.Success -> result.data
            is ApiResult.Failure -> null
        }

    suspend fun ack(token: String, announcement: AnnouncementDto): Boolean =
        when (val result = api.ackAnnouncement(token, announcement.id, announcement.updatedAt)) {
            is ApiResult.Success -> true
            is ApiResult.Failure -> false
        }
}
