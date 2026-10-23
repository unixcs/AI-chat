package com.unixcs.aichat.data.chat

import kotlinx.serialization.json.Json
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

internal sealed interface SseEvent {
    data class Delta(val text: String) : SseEvent

    data class Error(val code: String?, val message: String) : SseEvent

    data object Done : SseEvent
}

/**
 * Parses one SSE line. The backend only emits `data:` frames; comments, event
 * names and blank lines are ignored. Unknown or truncated JSON is dropped
 * rather than failing the whole reply.
 */
internal class SseParser(private val json: Json) {

    fun parse(line: String): SseEvent? {
        if (!line.startsWith("data:")) return null
        val payload = line.removePrefix("data:").trim()
        if (payload.isEmpty()) return null
        if (payload == "[DONE]") return SseEvent.Done
        return try {
            val obj = json.parseToJsonElement(payload).jsonObject
            val delta = obj["delta"]?.jsonPrimitive?.contentOrNull
            if (delta != null) {
                SseEvent.Delta(delta)
            } else {
                val error = obj["error"]?.jsonPrimitive?.contentOrNull
                if (error != null) {
                    SseEvent.Error(
                        code = obj["code"]?.jsonPrimitive?.contentOrNull,
                        message = error,
                    )
                } else {
                    null
                }
            }
        } catch (_: Exception) {
            null
        }
    }
}
