package com.unixcs.aichat.ui.screens

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.FilledTonalIconButton
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.unixcs.aichat.R
import com.unixcs.aichat.data.chat.ChatConversation
import com.unixcs.aichat.ui.util.groupConversations
import com.unixcs.aichat.ui.util.maskPhone

private const val MAX_TITLE_CODE_POINTS = 50

/**
 * Chat-history drawer body, styled after DeepSeek: borderless rows in soft
 * full-round pills, a filled round search field, date buckets, a bottom user
 * strip, and a selection mode with bulk delete. Rows carry no timestamps —
 * DeepSeek keeps them to one quiet line. Both the active row's ⋯ button and a
 * long-press anywhere on a row open the same anchored popup menu
 * (rename/delete); multi-select is entered from the first bucket header.
 */
@OptIn(ExperimentalFoundationApi::class)
@Composable
fun HistoryDrawerContent(
    conversations: List<ChatConversation>,
    currentId: String?,
    userPhone: String?,
    bulkProgress: Pair<Int, Int>?,
    onNewChat: () -> Unit,
    onSelect: (String) -> Unit,
    onRename: (String, String) -> Unit,
    onDelete: (String) -> Unit,
    onDeleteMany: (List<String>) -> Unit,
    onOpenSettings: () -> Unit,
) {
    var query by rememberSaveable { mutableStateOf("") }
    var renameTarget by remember { mutableStateOf<ChatConversation?>(null) }
    var deleteTarget by remember { mutableStateOf<ChatConversation?>(null) }
    // Plain remember: an open popup is not worth restoring after process
    // death — a restored target id could match a since-deleted conversation
    // and leave a phantom menuExpanded match.
    var menuTargetId by remember { mutableStateOf<String?>(null) }
    var selectMode by rememberSaveable { mutableStateOf(false) }
    var selected by rememberSaveable(stateSaver = StringSetSaver) { mutableStateOf(setOf<String>()) }
    var confirmBulkDelete by remember { mutableStateOf(false) }
    val sections = groupConversations(conversations, query)

    Column(modifier = Modifier.fillMaxSize()) {
        if (selectMode) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(start = 20.dp, end = 8.dp, top = 14.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "选择对话",
                    style = MaterialTheme.typography.titleLarge,
                    modifier = Modifier.weight(1f),
                )
                IconButton(onClick = {
                    selectMode = false
                    selected = emptySet()
                }) {
                    Icon(Icons.Default.Close, contentDescription = "退出多选")
                }
            }
        } else {
            Text(
                text = "聊天历史",
                style = MaterialTheme.typography.titleLarge,
                modifier = Modifier.padding(start = 20.dp, end = 8.dp, top = 20.dp),
            )
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TextField(
                value = query,
                onValueChange = { query = it },
                modifier = Modifier
                    .weight(1f)
                    .height(52.dp),
                placeholder = {
                    Text("搜索对话标题", style = MaterialTheme.typography.bodyMedium)
                },
                leadingIcon = {
                    Icon(
                        Icons.Default.Search,
                        contentDescription = null,
                        modifier = Modifier.size(20.dp),
                    )
                },
                singleLine = true,
                shape = RoundedCornerShape(26.dp),
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                    unfocusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                    disabledContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent,
                    disabledIndicatorColor = Color.Transparent,
                ),
            )
            if (selectMode) {
                Spacer(Modifier.width(8.dp))
                FilledTonalIconButton(
                    onClick = {
                        selected = if (selected.size == conversations.size) {
                            emptySet()
                        } else {
                            conversations.map { it.id }.toSet()
                        }
                    },
                ) {
                    Icon(
                        painter = painterResource(R.drawable.ic_select_all),
                        contentDescription = "全选/全不选",
                        modifier = Modifier.size(20.dp),
                    )
                }
            }
        }

        if (sections.isEmpty()) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = if (query.isBlank()) "还没有对话记录" else "没有匹配的对话",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
                contentPadding = PaddingValues(start = 10.dp, end = 10.dp, bottom = 12.dp),
            ) {
                sections.forEachIndexed { sectionIndex, section ->
                    item(key = "header-${section.label}") {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(start = 10.dp, end = 2.dp, top = 12.dp, bottom = 4.dp),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text(
                                text = section.label,
                                style = MaterialTheme.typography.labelMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.weight(1f),
                            )
                            // DeepSeek keeps its batch-entry icon on the first
                            // bucket's header row; hidden while a bulk delete
                            // runs so a second run cannot silently drop a new
                            // selection (the ViewModel early-returns on a
                            // concurrent run).
                            if (sectionIndex == 0 && !selectMode && bulkProgress == null) {
                                IconButton(
                                    onClick = {
                                        menuTargetId = null
                                        selectMode = true
                                        selected = emptySet()
                                    },
                                    modifier = Modifier.size(28.dp),
                                ) {
                                    Icon(
                                        painter = painterResource(R.drawable.ic_select),
                                        contentDescription = "批量管理",
                                        modifier = Modifier.size(18.dp),
                                    )
                                }
                            }
                        }
                    }
                    items(section.items, key = { it.id }) { conversation ->
                        ConversationRow(
                            conversation = conversation,
                            active = conversation.id == currentId,
                            selectMode = selectMode,
                            checked = conversation.id in selected,
                            menuExpanded = menuTargetId == conversation.id,
                            onClick = {
                                if (selectMode) {
                                    selected = if (conversation.id in selected) {
                                        selected - conversation.id
                                    } else {
                                        selected + conversation.id
                                    }
                                } else {
                                    onSelect(conversation.id)
                                }
                            },
                            onLongPress = {
                                // Long-press = the same popup as the ⋯ button.
                                if (!selectMode) menuTargetId = conversation.id
                            },
                            onMenuDismiss = { menuTargetId = null },
                            onRenameFromMenu = {
                                menuTargetId = null
                                renameTarget = conversation
                            },
                            onDeleteFromMenu = {
                                menuTargetId = null
                                deleteTarget = conversation
                            },
                        )
                    }
                }
            }
        }

        // Bottom user strip, DeepSeek-style: avatar dot + masked phone.
        HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clickable(onClick = onOpenSettings)
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Box(
                modifier = Modifier
                    .size(34.dp)
                    .background(MaterialTheme.colorScheme.surfaceVariant, CircleShape),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = userPhone?.takeIf { it.isNotBlank() }?.let(::maskPhone)?.take(1) ?: "客",
                    style = MaterialTheme.typography.titleSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            Spacer(Modifier.width(10.dp))
            Text(
                text = userPhone?.takeIf { it.isNotBlank() }?.let(::maskPhone) ?: "未登录",
                style = MaterialTheme.typography.bodyLarge,
                modifier = Modifier.weight(1f),
            )
        }

        // Stays up after selection mode exits so a running bulk delete keeps
        // its progress feedback until it finishes.
        if (selectMode || bulkProgress != null) {
            BulkDeleteBar(
                selectedCount = selected.size,
                progress = bulkProgress,
                onDelete = { confirmBulkDelete = true },
            )
        }
    }

    renameTarget?.let { target ->
        RenameDialog(
            conversation = target,
            onDismiss = { renameTarget = null },
            onConfirm = { title ->
                renameTarget = null
                onRename(target.id, title)
            },
        )
    }

    deleteTarget?.let { target ->
        AlertDialog(
            onDismissRequest = { deleteTarget = null },
            title = { Text("删除对话") },
            text = { Text("确定删除「${target.title.ifBlank { "未命名对话" }}」吗？删除后无法恢复。") },
            confirmButton = {
                TextButton(
                    onClick = {
                        deleteTarget = null
                        onDelete(target.id)
                    },
                ) {
                    Text("删除", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { deleteTarget = null }) { Text("取消") }
            },
        )
    }

    if (confirmBulkDelete) {
        AlertDialog(
            onDismissRequest = { confirmBulkDelete = false },
            title = { Text("删除对话") },
            text = { Text("确定删除选中的 ${selected.size} 个对话吗？删除后无法恢复。") },
            confirmButton = {
                TextButton(
                    onClick = {
                        confirmBulkDelete = false
                        val ids = selected.toList()
                        selectMode = false
                        selected = emptySet()
                        onDeleteMany(ids)
                    },
                ) {
                    Text("删除", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { confirmBulkDelete = false }) { Text("取消") }
            },
        )
    }
}

/** 「已选 N 个」→「正在删除 i/N」+ disabled button: feedback instead of a freeze. */
@Composable
private fun BulkDeleteBar(
    selectedCount: Int,
    progress: Pair<Int, Int>?,
    onDelete: () -> Unit,
) {
    Column(Modifier.fillMaxWidth()) {
        HorizontalDivider()
        if (progress != null) {
            LinearProgressIndicator(
                progress = { progress.first.toFloat() / progress.second.coerceAtLeast(1) },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
            )
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = if (progress != null) {
                    "正在删除 ${progress.first}/${progress.second}"
                } else {
                    "已选 $selectedCount 个对话"
                },
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier
                    .weight(1f)
                    .padding(start = 20.dp),
            )
            TextButton(
                onClick = onDelete,
                enabled = progress == null && selectedCount > 0,
            ) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = null,
                    modifier = Modifier.size(18.dp),
                )
                Spacer(Modifier.width(4.dp))
                Text("删除")
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun ConversationRow(
    conversation: ChatConversation,
    active: Boolean,
    selectMode: Boolean,
    checked: Boolean,
    menuExpanded: Boolean,
    onClick: () -> Unit,
    onLongPress: () -> Unit,
    onMenuDismiss: () -> Unit,
    onRenameFromMenu: () -> Unit,
    onDeleteFromMenu: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 2.dp)
            .background(
                color = when {
                    selectMode && checked -> MaterialTheme.colorScheme.secondaryContainer
                    active -> MaterialTheme.colorScheme.secondaryContainer
                    else -> Color.Transparent
                },
                shape = RoundedCornerShape(24.dp),
            )
            .combinedClickable(onClick = onClick, onLongClick = onLongPress)
            .padding(start = 12.dp, end = 4.dp)
            .heightIn(min = 44.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (selectMode) {
            Box(
                modifier = Modifier
                    .padding(end = 10.dp)
                    .size(20.dp)
                    .background(
                        if (checked) {
                            MaterialTheme.colorScheme.primary
                        } else {
                            Color.Transparent
                        },
                        CircleShape,
                    )
                    .border(
                        width = 1.5.dp,
                        color = if (checked) {
                            MaterialTheme.colorScheme.primary
                        } else {
                            MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
                        },
                        shape = CircleShape,
                    ),
                contentAlignment = Alignment.Center,
            ) {
                if (checked) {
                    Icon(
                        painter = painterResource(R.drawable.ic_check),
                        contentDescription = null,
                        modifier = Modifier.size(12.dp),
                        tint = MaterialTheme.colorScheme.onPrimary,
                    )
                }
            }
        }
        Text(
            text = conversation.title.ifBlank { "未命名对话" },
            style = MaterialTheme.typography.bodyLarge,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            color = if (active && !selectMode) {
                MaterialTheme.colorScheme.primary
            } else {
                MaterialTheme.colorScheme.onSurface
            },
            modifier = Modifier.weight(1f),
        )
        if (!selectMode) {
            // The ⋯ button stays visible only on the active row (DeepSeek
            // pattern); the menu itself anchors to this trailing slot on every
            // row so long-press pops it right where the finger is.
            Box(contentAlignment = Alignment.Center) {
                if (active) {
                    IconButton(onClick = onLongPress, modifier = Modifier.size(32.dp)) {
                        Icon(
                            Icons.Default.MoreVert,
                            contentDescription = "更多操作",
                            modifier = Modifier.size(18.dp),
                            tint = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                } else {
                    Spacer(Modifier.size(32.dp))
                }
                DropdownMenu(expanded = menuExpanded, onDismissRequest = onMenuDismiss) {
                    DropdownMenuItem(
                        text = { Text("重命名") },
                        onClick = onRenameFromMenu,
                    )
                    DropdownMenuItem(
                        text = { Text("删除", color = MaterialTheme.colorScheme.error) },
                        onClick = onDeleteFromMenu,
                    )
                }
            }
        }
    }
}

private val StringSetSaver = androidx.compose.runtime.saveable.listSaver<Set<String>, String>(
    save = { it.toList() },
    restore = { it.toSet() },
)

@Composable
private fun RenameDialog(
    conversation: ChatConversation,
    onDismiss: () -> Unit,
    onConfirm: (String) -> Unit,
) {
    var text by rememberSaveable(conversation.id) { mutableStateOf(conversation.title) }
    val trimmed = text.trim()
    // The backend limits titles by Unicode code points (Go runes); using
    // String.length here would count an emoji twice and reject valid input.
    val used = text.codePointCount(0, text.length)
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("重命名对话") },
        text = {
            OutlinedTextField(
                value = text,
                onValueChange = {
                    if (it.codePointCount(0, it.length) <= MAX_TITLE_CODE_POINTS) text = it
                },
                singleLine = true,
                supportingText = { Text("$used/$MAX_TITLE_CODE_POINTS") },
            )
        },
        confirmButton = {
            TextButton(
                onClick = { onConfirm(trimmed) },
                enabled = trimmed.isNotEmpty() && trimmed != conversation.title,
            ) {
                Text("保存")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("取消") }
        },
    )
}
