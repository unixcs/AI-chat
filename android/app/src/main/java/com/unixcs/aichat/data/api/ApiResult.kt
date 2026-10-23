package com.unixcs.aichat.data.api

/**
 * Result of a single API call. HTTP failures and transport failures share one
 * type so ViewModels never need to know about OkHttp or the envelope format.
 */
sealed interface ApiResult<out T> {
    data class Success<T>(val data: T) : ApiResult<T>

    data class Failure(
        val status: Int,
        val message: String,
        val code: String? = null,
    ) : ApiResult<Nothing> {
        /** 401 means the session is gone (expired or kicked by another device). */
        val unauthorized: Boolean get() = status == 401
    }
}
