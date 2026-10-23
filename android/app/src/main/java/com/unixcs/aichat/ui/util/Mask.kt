package com.unixcs.aichat.ui.util

/** 138****1234 — the only phone form shown anywhere in the UI or logs. */
fun maskPhone(phone: String): String {
    if (phone.length != 11) return phone
    return phone.take(3) + "****" + phone.takeLast(4)
}
