package com.unixcs.aichat.data.api

import kotlinx.serialization.Serializable

@Serializable
internal data class LoginRequest(
    val phone: String,
    val password: String,
)

@Serializable
data class LoginData(val token: String = "")

/** GET /api/user/session-check wraps a boolean in an object. */
@Serializable
data class SessionCheck(val ok: Boolean = false)

/**
 * Subset of GET /api/user/profile that the App needs. Every field has a
 * default so additive backend changes never break deserialization.
 */
@Serializable
data class UserProfile(
    val id: String = "",
    val phone: String = "",
    val nickname: String = "",
    val avatarUrl: String = "",
    val status: String = "",
    val role: String = "",
    val memberExpireAt: String? = null,
    val createdAt: String = "",
    val answerLength: String = "",
    val answerStyle: String = "",
    val answerFormat: String = "",
)

@Serializable
internal data class Envelope<T>(
    val code: Int = -1,
    val message: String? = null,
    val data: T? = null,
)

@Serializable
internal data class ErrorBody(
    val message: String? = null,
    val code: String? = null,
)
