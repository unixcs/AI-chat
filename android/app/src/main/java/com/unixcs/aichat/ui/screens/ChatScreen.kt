package com.unixcs.aichat.ui.screens

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.animation.core.Animatable
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
import androidx.compose.material3.IconButtonDefaults
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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.unixcs.aichat.R
import com.unixcs.aichat.data.chat.ChatMessage
import com.unixcs.aichat.ui.announcement.AnnouncementViewModel
import com.unixcs.aichat.ui.chat.ChatViewModel
import com.unixcs.aichat.ui.components.DrawerShell
import com.unixcs.aichat.ui.components.MarkdownText
import com.unixcs.aichat.ui.components.ShellSpring
import kotlinx.coroutines.launch
import org.koin.androidx.compose.koinViewModel

private const val DRAFT_ID = "__draft__"

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatScreen(
    viewModel: ChatViewModel,
    onOpenSettings: () -> Unit,
) {
    val ui by viewModel.ui.collectAsStateWithLifecycle()
    val conversations by viewModel.conversations.collectAsStateWithLifecycle()
    val preferences by viewModel.preferences.collectAsStateWithLifecycle()
    val showPreferenceEntry by viewModel.showPreferenceEntry.collectAsStateWithLifecycle()
    val appearance by viewModel.appearance.collectAsStateWithLifecycle()
    val userPhone by viewModel.userPhone.collectAsStateWithLifecycle()
    val bulkDelete by viewModel.bulkDelete.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    val drawerProgress = remember { Animatable(0f) }
    val scope = rememberCoroutineScope()
    var showPreferences by rememberSaveable { mutableStateOf(false) }
    // Chat-screen-scoped: survives the Settings round-trip (no repeat popup)
    // and dies with logout (no cross-account leak).
    val announcementViewModel: AnnouncementViewModel = koinViewModel()
    val pendingAnnouncement by announcementViewModel.pending.collectAsStateWithLifecycle()

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

    // Wraps the whole Scaffold (bottom bar included) so the scrim also covers
    // the input row. Caller-owned Animatable progress: no AnchoredDraggable,
    // no NaN-anchor state machine — the class of wedge that froze the drawer.
    DrawerShell(
        progress = drawerProgress,
        onDismiss = { scope.launch { drawerProgress.animateTo(0f, ShellSpring) } },
        sheet = {
            HistoryDrawerContent(
                conversations = conversations,
                currentId = ui.conversationId,
                userPhone = userPhone,
                bulkProgress = bulkDelete,
                onNewChat = {
                    viewModel.startNewChat()
                    scope.launch { drawerProgress.animateTo(0f, ShellSpring) }
                },
                onSelect = { conversationId ->
                    viewModel.openConversation(conversationId)
                    scope.launch { drawerProgress.animateTo(0f, ShellSpring) }
                },
                onRename = viewModel::renameConversation,
                onDelete = viewModel::deleteConversation,
                onDeleteMany = viewModel::deleteConversations,
                onOpenSettings = {
                    scope.launch { drawerProgress.animateTo(0f, ShellSpring) }
                    onOpenSettings()
                },
            )
        },
    ) {
        Scaffold(
            topBar = {
                TopAppBar(
                    title = {
                        Text(
                            text = ui.title.ifBlank { "塔罗" },
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                        )
                    },
                    navigationIcon = {
                        IconButton(onClick = { scope.launch { drawerProgress.animateTo(1f, ShellSpring) } }) {
                            Icon(Icons.Default.Menu, contentDescription = "打开历史")
                        }
                    },
                    actions = {
                        // Quick light/dark toggle, DeepSeek-style one-tap switch.
                        IconButton(onClick = viewModel::toggleDarkMode) {
                            Icon(
                                painter = painterResource(
                                    if (appearance.themeMode == "dark") R.drawable.ic_sun else R.drawable.ic_moon,
                                ),
                                contentDescription = if (appearance.themeMode == "dark") "切换浅色" else "切换深色",
                            )
                        }
                        if (showPreferenceEntry) {
                            TextButton(onClick = { showPreferences = true }) {
                                Text("偏好")
                            }
                        }
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
                        contentPadding = PaddingValues(horizontal = 14.dp, vertical = 16.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                    ) {
                        items(display, key = { it.id }) { message ->
                            MessageCard(
                                message = message,
                                thinking = message.id == DRAFT_ID && message.content.isEmpty(),
                                onCopied = viewModel::notifyCopied,
                            )
                        }
                    }
                }
            }
        }
    }

    if (showPreferences) {
        PreferenceSheet(
            preferences = preferences,
            onLengthChange = viewModel::setAnswerLength,
            onStyleChange = viewModel::setAnswerStyle,
            onFormatChange = viewModel::setAnswerFormat,
            onDismiss = { showPreferences = false },
        )
    }

    pendingAnnouncement?.let { announcement ->
        AnnouncementCard(
            announcement = announcement,
            onAcknowledge = announcementViewModel::acknowledge,
            onDismiss = announcementViewModel::dismiss,
        )
    }
}

@Composable
private fun MessageCard(message: ChatMessage, thinking: Boolean, onCopied: () -> Unit) {
    val isUser = message.role == "user"
    Column(modifier = Modifier.fillMaxWidth()) {
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
                    topStart = 20.dp,
                    topEnd = 20.dp,
                    bottomStart = if (isUser) 20.dp else 6.dp,
                    bottomEnd = if (isUser) 6.dp else 20.dp,
                ),
                modifier = Modifier.widthIn(max = 340.dp),
            ) {
                if (thinking) {
                    Row(
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 13.dp),
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
                } else if (isUser || message.id == DRAFT_ID) {
                    // While streaming, render plain text: re-parsing Markdown on
                    // every delta is what made the web client stutter. The final
                    // persisted message is rendered as Markdown.
                    SelectionContainer {
                        Text(
                            text = message.content,
                            style = MaterialTheme.typography.bodyLarge,
                            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                        )
                    }
                } else {
                    MarkdownText(
                        markdown = message.content,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    )
                }
            }
        }
        // Action bar sits below the bubble (DeepSeek/ChatGPT pattern): copy is
        // one tap away without stealing long-press selection inside the text.
        if (!thinking && message.content.isNotEmpty()) {
            CopyButton(
                text = message.content,
                alignEnd = isUser,
                onCopied = onCopied,
            )
        }
    }
}

@Composable
private fun CopyButton(text: String, alignEnd: Boolean, onCopied: () -> Unit) {
    val context = LocalContext.current
    var copied by remember { mutableStateOf(false) }
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (alignEnd) Arrangement.End else Arrangement.Start,
    ) {
        IconButton(
            onClick = {
                val clipboard =
                    context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                clipboard.setPrimaryClip(ClipData.newPlainText("message", text))
                copied = true
                onCopied()
            },
            modifier = Modifier.size(32.dp),
            colors = IconButtonDefaults.iconButtonColors(
                contentColor = MaterialTheme.colorScheme.onSurfaceVariant,
            ),
        ) {
            Icon(
                painter = painterResource(if (copied) R.drawable.ic_check else R.drawable.ic_copy),
                contentDescription = if (copied) "已复制" else "复制",
                modifier = Modifier.size(16.dp),
            )
        }
    }
    if (copied) {
        LaunchedEffect(Unit) {
            kotlinx.coroutines.delay(1500)
            copied = false
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
