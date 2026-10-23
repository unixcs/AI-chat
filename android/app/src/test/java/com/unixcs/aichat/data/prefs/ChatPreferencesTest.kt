package com.unixcs.aichat.data.prefs

import com.unixcs.aichat.data.api.UserProfile
import org.junit.Assert.assertEquals
import org.junit.Test

class ChatPreferencesTest {

    @Test
    fun unknownOrLegacyValuesFallBackToStandard() {
        val profile = UserProfile(
            answerLength = "",
            answerStyle = "legacy-value",
            answerFormat = "not-a-format",
        )

        assertEquals(
            ChatPreferences(
                answerLength = "standard",
                answerStyle = "standard",
                answerFormat = "standard",
            ),
            ChatPreferences.fromProfile(profile),
        )
    }

    @Test
    fun validServerValuesArePreserved() {
        val profile = UserProfile(
            answerLength = "detailed",
            answerStyle = "encouraging",
            answerFormat = "plain",
        )

        assertEquals(
            ChatPreferences(
                answerLength = "detailed",
                answerStyle = "encouraging",
                answerFormat = "plain",
            ),
            ChatPreferences.fromProfile(profile),
        )
    }

    @Test
    fun everyOptionHasALabelAndNormalizesToItself() {
        PreferenceOptions.lengths.forEach {
            assertEquals(it.value, PreferenceOptions.normalize(it.value, PreferenceOptions.lengths))
            assertEquals(it.label, PreferenceOptions.lengthLabel(it.value))
        }
        PreferenceOptions.styles.forEach {
            assertEquals(it.value, PreferenceOptions.normalize(it.value, PreferenceOptions.styles))
            assertEquals(it.label, PreferenceOptions.styleLabel(it.value))
        }
        PreferenceOptions.formats.forEach {
            assertEquals(it.value, PreferenceOptions.normalize(it.value, PreferenceOptions.formats))
            assertEquals(it.label, PreferenceOptions.formatLabel(it.value))
        }
    }
}
