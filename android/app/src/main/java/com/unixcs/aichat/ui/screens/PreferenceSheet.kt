package com.unixcs.aichat.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.unixcs.aichat.data.prefs.ChatPreferences
import com.unixcs.aichat.data.prefs.PreferenceOptions

/**
 * Three-dimension answer preference editor. Tapping a chip saves immediately;
 * there is no explicit "apply" step, matching the web settings sheet. Uses
 * material3's ModalBottomSheet directly, as in v1.0.0.
 */
@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun PreferenceSheet(
    preferences: ChatPreferences,
    onLengthChange: (String) -> Unit,
    onStyleChange: (String) -> Unit,
    onFormatChange: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    ModalBottomSheet(onDismissRequest = onDismiss, sheetState = sheetState) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .navigationBarsPadding()
                .padding(horizontal = 20.dp)
                .padding(bottom = 24.dp),
        ) {
            Text("回答偏好", style = MaterialTheme.typography.titleLarge)
            Spacer(Modifier.height(4.dp))
            Text(
                text = "修改会立即保存，并用于之后的新回复；历史消息保持不变。",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )

            PreferenceGroup(
                title = "回答长度",
                options = PreferenceOptions.lengths,
                selected = preferences.answerLength,
                onSelect = onLengthChange,
            )
            PreferenceGroup(
                title = "回答风格",
                options = PreferenceOptions.styles,
                selected = preferences.answerStyle,
                onSelect = onStyleChange,
            )
            PreferenceGroup(
                title = "回答格式",
                options = PreferenceOptions.formats,
                selected = preferences.answerFormat,
                onSelect = onFormatChange,
            )
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun PreferenceGroup(
    title: String,
    options: List<PreferenceOptions.Option>,
    selected: String,
    onSelect: (String) -> Unit,
) {
    Text(
        text = title,
        style = MaterialTheme.typography.titleSmall,
        modifier = Modifier.padding(top = 20.dp, bottom = 8.dp),
    )
    FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        options.forEach { option ->
            FilterChip(
                selected = option.value == selected,
                onClick = { onSelect(option.value) },
                label = { Text(option.label) },
            )
        }
    }
}
