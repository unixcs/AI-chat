package com.unixcs.aichat.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilledIconButton
import androidx.compose.material3.FilledTonalIconButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.unixcs.aichat.data.chat.ChatMessage
import com.unixcs.aichat.ui.chat.ChatViewModel

private const val DRAFT_ID = "__draft__"

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatScreen(
    viewModel: ChatViewModel,
    onOpenHistory: () -> Unit,
    onOpenSettings: () -> Unit,
) {
    val ui by viewModel.ui.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    val display = remember(ui.messages, ui.streaming, ui.streamingText) {
        buildList {
            addAll(ui.messages)
            if (ui.streaming) {
                add(
                    ChatMessage(
                        id = DRAFT_ID,
                        role = "assistant",
                        content = ui.streamingText,
                        createdAt = "",
                    ),
                )
            }
        }
    }
    val lastLength = display.lastOrNull()?.content?.length ?: 0
    val nearBottom by remember {
        derivedStateOf {
            val info = listState.layoutInfo
            val lastVisible = info.visibleItemsInfo.lastOrNull()?.index ?: -1
            lastVisible >= info.totalItemsCount - 2
        }
    }
    LaunchedEffect(display.size, lastLength, nearBottom) {
        if (display.isNotEmpty() && nearBottom) {
            listState.scrollToItem(display.lastIndex)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        text = ui.title.ifBlank { "AI-chat" },
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                },
                navigationIcon = {
                    IconButton(onClick = onOpenHistory) {
                        Icon(Icons.Default.Menu, contentDescription = "打开历史")
                    }
                },
                actions = {
                    IconButton(onClick = viewModel::startNewChat) {
                        Icon(Icons.Default.Add, contentDescription = "新建对话")
                    }
                    IconButton(onClick = onOpenSettings) {
                        Icon(Icons.Default.Settings, contentDescription = "设置")
                    }
                },
            )
        },
        bottomBar = {
            Surface(tonalElevation = 2.dp) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .navigationBarsPadding()
                        .imePadding(),
                ) {
                    ui.error?.let { message ->
                        ErrorBanner(message = message, onDismiss = viewModel::dismissError)
                    }
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 12.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.Bottom,
                    ) {
                        OutlinedTextField(
                            value = ui.input,
                            onValueChange = viewModel::onInputChange,
                            modifier = Modifier.weight(1f),
                            placeholder = { Text("输入消息…") },
                            maxLines = 4,
                            shape = MaterialTheme.shapes.extraLarge,
                            enabled = !ui.loading,
                        )
                        Spacer(Modifier.size(8.dp))
                        if (ui.streaming) {
                            FilledTonalIconButton(onClick = viewModel::stopStreaming) {
                                Icon(Icons.Default.Close, contentDescription = "停止生成")
                            }
                        } else {
                            FilledIconButton(
                                onClick = viewModel::send,
                                enabled = ui.canSend,
                            ) {
                                Icon(
                                    Icons.AutoMirrored.Filled.Send,
                                    contentDescription = "发送",
                                )
                            }
                        }
                    }
                }
            }
        },
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding),
        ) {
            if (ui.loading) {
                LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
            }
            if (display.isEmpty() && !ui.loading) {
                EmptyChatHint(modifier = Modifier.align(Alignment.Center))
            } else {
                LazyColumn(
                    state = listState,
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    items(display, key = { it.id }) { message ->
                        MessageBubble(
                            message = message,
                            thinking = message.id == DRAFT_ID && message.content.isEmpty(),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun MessageBubble(message: ChatMessage, thinking: Boolean) {
    val isUser = message.role == "user"
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start,
    ) {
        Surface(
            color = if (isUser) {
                MaterialTheme.colorScheme.primaryContainer
            } else {
                MaterialTheme.colorScheme.surfaceVariant
            },
            shape = RoundedCornerShape(
                topStart = 16.dp,
                topEnd = 16.dp,
                bottomStart = if (isUser) 16.dp else 4.dp,
                bottomEnd = if (isUser) 4.dp else 16.dp,
            ),
            modifier = Modifier.widthIn(max = 340.dp),
        ) {
            if (thinking) {
                Row(
                    modifier = Modifier.padding(horizontal = 14.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    CircularProgressIndicator(modifier = Modifier.size(14.dp), strokeWidth = 2.dp)
                    Spacer(Modifier.size(8.dp))
                    Text(
                        text = "正在思考…",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            } else {
                SelectionContainer {
                    Text(
                        text = message.content,
                        style = MaterialTheme.typography.bodyLarge,
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
                    )
                }
            }
        }
    }
}

@Composable
private fun EmptyChatHint(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(horizontal = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = "开始新的对话",
            style = MaterialTheme.typography.titleMedium,
        )
        Spacer(Modifier.size(8.dp))
        Text(
            text = "输入内容后发送，回复会实时显示在这里。",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}

@Composable
private fun ErrorBanner(message: String, onDismiss: () -> Unit) {
    Surface(color = MaterialTheme.colorScheme.errorContainer) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 16.dp, end = 4.dp, top = 4.dp, bottom = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = message,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onErrorContainer,
                modifier = Modifier.weight(1f),
            )
            TextButton(onClick = onDismiss) {
                Text("关闭", color = MaterialTheme.colorScheme.onErrorContainer)
            }
        }
    }
}
