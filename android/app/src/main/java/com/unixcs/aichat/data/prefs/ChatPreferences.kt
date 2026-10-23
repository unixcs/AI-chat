package com.unixcs.aichat.data.prefs

import com.unixcs.aichat.data.api.UserProfile

/**
 * Values accepted by PUT /api/user/preferences. The backend rejects unknown
 * strings, so every UI label maps to one of these exact codes.
 */
object PreferenceOptions {
    data class Option(val value: String, val label: String)

    val lengths = listOf(
        Option("concise", "精简"),
        Option("standard", "适中"),
        Option("detailed", "详细"),
    )
    val styles = listOf(
        Option("plain", "大白话"),
        Option("standard", "标准"),
        Option("professional", "专业"),
        Option("rigorous", "严谨"),
        Option("encouraging", "鼓励"),
    )
    val formats = listOf(
        Option("standard", "标准"),
        Option("plain", "纯文字"),
    )

    fun normalize(raw: String?, options: List<Option>): String {
        val standard = options.first { it.value == "standard" }.value
        return options.firstOrNull { it.value == raw }?.value ?: standard
    }

    fun lengthLabel(value: String): String =
        lengths.firstOrNull { it.value == value }?.label ?: lengths[1].label

    fun styleLabel(value: String): String =
        styles.firstOrNull { it.value == value }?.label ?: styles[1].label

    fun formatLabel(value: String): String =
        formats.firstOrNull { it.value == value }?.label ?: formats[0].label
}

/**
 * The three answer-shaping dimensions. Legacy accounts may carry empty/null
 * strings; those normalize to "standard" exactly like the web UI.
 */
data class ChatPreferences(
    val answerLength: String = "standard",
    val answerStyle: String = "standard",
    val answerFormat: String = "standard",
) {
    companion object {
        fun fromProfile(profile: UserProfile): ChatPreferences = ChatPreferences(
            answerLength = PreferenceOptions.normalize(profile.answerLength, PreferenceOptions.lengths),
            answerStyle = PreferenceOptions.normalize(profile.answerStyle, PreferenceOptions.styles),
            answerFormat = PreferenceOptions.normalize(profile.answerFormat, PreferenceOptions.formats),
        )
    }
}
