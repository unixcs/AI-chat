package com.unixcs.aichat.ui.screens

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.unixcs.aichat.BuildConfig
import com.unixcs.aichat.data.api.UserProfile
import com.unixcs.aichat.ui.util.maskPhone

// Phase 1: account + logout. Phase 4 adds the preference-visibility switch;
// Phase 5 adds encrypted export/import; Phase 6 shows the production endpoint.
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    profile: UserProfile?,
    showPreferenceEntry: Boolean,
    onShowPreferenceEntryChange: (Boolean) -> Unit,
    onLogout: () -> Unit,
    onBack: () -> Unit,
) {
    var confirmLogout by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("设置") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回")
                    }
                },
            )
        },
    ) { innerPadding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp),
        ) {
            SectionTitle("账户")
            ListItem(
                headlineContent = { Text("当前账号") },
                supportingContent = {
                    Text(
                        profile?.phone?.takeIf { it.isNotBlank() }?.let(::maskPhone)
                            ?: "未获取到账号信息",
                    )
                },
            )
            HorizontalDivider()
            ListItem(
                headlineContent = { Text("会员到期") },
                supportingContent = {
                    Text(profile?.memberExpireAt?.take(10)?.takeIf { it.isNotBlank() } ?: "未开通")
                },
            )
            OutlinedButton(
                onClick = { confirmLogout = true },
                modifier = Modifier.padding(vertical = 16.dp),
            ) {
                Text("退出登录", color = MaterialTheme.colorScheme.error)
            }

            SectionTitle("对话")
            ListItem(
                headlineContent = { Text("显示回答偏好入口") },
                supportingContent = {
                    Text("在聊天页顶部显示「偏好」按钮，可随时切换回答长度、风格和格式。本开关只影响本机。")
                },
                trailingContent = {
                    Switch(
                        checked = showPreferenceEntry,
                        onCheckedChange = onShowPreferenceEntryChange,
                    )
                },
            )
            HorizontalDivider()

            SectionTitle("关于")
            ListItem(
                headlineContent = { Text("服务器地址") },
                supportingContent = { Text(BuildConfig.API_BASE_URL) },
            )
            HorizontalDivider()
            ListItem(
                headlineContent = { Text("版本") },
                supportingContent = { Text("${BuildConfig.VERSION_NAME} (${BuildConfig.VERSION_CODE})") },
            )
        }
    }

    if (confirmLogout) {
        AlertDialog(
            onDismissRequest = { confirmLogout = false },
            title = { Text("退出登录") },
            text = { Text("退出后需要重新输入密码。本机缓存的聊天记录会一并清除，服务器上的记录保留。") },
            confirmButton = {
                TextButton(
                    onClick = {
                        confirmLogout = false
                        onLogout()
                    },
                ) {
                    Text("退出", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { confirmLogout = false }) {
                    Text("取消")
                }
            },
        )
    }
}

@Composable
private fun SectionTitle(text: String) {
    Text(
        text = text,
        style = MaterialTheme.typography.titleSmall,
        color = MaterialTheme.colorScheme.primary,
        modifier = Modifier.padding(top = 24.dp, bottom = 4.dp),
    )
}
