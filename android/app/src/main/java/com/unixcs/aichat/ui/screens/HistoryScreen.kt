package com.unixcs.aichat.ui.screens

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.unixcs.aichat.data.chat.ChatConversation
import com.unixcs.aichat.ui.chat.ChatViewModel
import com.unixcs.aichat.ui.util.formatTimestamp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HistoryScreen(
    viewModel: ChatViewModel,
    onBack: () -> Unit,
    onSelect: (String) -> Unit,
) {
    val conversations by viewModel.conversations.collectAsStateWithLifecycle()
    var pendingDelete by remember { mutableStateOf<ChatConversation?>(null) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("历史对话") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回")
                    }
                },
            )
        },
    ) { innerPadding ->
        if (conversations.isEmpty()) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = "还没有对话记录",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding),
            ) {
                items(conversations, key = { it.id }) { conversation ->
                    ListItem(
                        headlineContent = {
                            Text(
                                text = conversation.title.ifBlank { "未命名对话" },
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        },
                        supportingContent = { Text(formatTimestamp(conversation.updatedAt)) },
                        trailingContent = {
                            IconButton(onClick = { pendingDelete = conversation }) {
                                Icon(Icons.Default.Delete, contentDescription = "删除对话")
                            }
                        },
                        modifier = Modifier.clickable { onSelect(conversation.id) },
                    )
                    HorizontalDivider()
                }
            }
        }
    }

    pendingDelete?.let { target ->
        AlertDialog(
            onDismissRequest = { pendingDelete = null },
            title = { Text("删除对话") },
            text = { Text("将删除「${target.title.ifBlank { "未命名对话" }}」及其全部消息，且无法恢复。") },
            confirmButton = {
                TextButton(
                    onClick = {
                        pendingDelete = null
                        viewModel.deleteConversation(target.id)
                    },
                ) {
                    Text("删除", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { pendingDelete = null }) {
                    Text("取消")
                }
            },
        )
    }
}
