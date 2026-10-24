package com.unixcs.aichat.ui.screens

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.unixcs.aichat.BuildConfig
import com.unixcs.aichat.R
import com.unixcs.aichat.data.api.UserProfile
import com.unixcs.aichat.ui.backup.BackupViewModel
import com.unixcs.aichat.ui.settings.SettingsViewModel
import com.unixcs.aichat.ui.theme.AccentPalettes
import com.unixcs.aichat.ui.theme.ThemeMode
import com.unixcs.aichat.ui.util.maskPhone
import java.time.LocalDate
import org.koin.androidx.compose.koinViewModel

private enum class BackupAction { Export, Import }

// Phase 1: account + logout. Phase 4 adds the preference-visibility switch;
// Phase 5 adds encrypted export/import; Phase 6 shows the production endpoint.
// R15 adds redeem-code top-up, nickname editing, and appearance settings.
@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun SettingsScreen(
    profile: UserProfile?,
    showPreferenceEntry: Boolean,
    onShowPreferenceEntryChange: (Boolean) -> Unit,
    onProfileUpdated: (UserProfile) -> Unit,
    onLogout: () -> Unit,
    onBack: () -> Unit,
) {
    var confirmLogout by remember { mutableStateOf(false) }
    var dialogAction by remember { mutableStateOf<BackupAction?>(null) }
    var pendingPassword by remember { mutableStateOf<CharArray?>(null) }
    var showRedeemDialog by remember { mutableStateOf(false) }
    var showNicknameDialog by remember { mutableStateOf(false) }
    val snackbarHostState = remember { SnackbarHostState() }
    val backupViewModel: BackupViewModel = koinViewModel()
    val backupState by backupViewModel.ui.collectAsStateWithLifecycle()
    val settingsViewModel: SettingsViewModel = koinViewModel()
    val accountState by settingsViewModel.account.collectAsStateWithLifecycle()
    val appearance by settingsViewModel.appearance.collectAsStateWithLifecycle()

    val exportLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.CreateDocument("application/octet-stream"),
    ) { uri: Uri? ->
        val password = pendingPassword
        pendingPassword = null
        if (uri != null && password != null) {
            backupViewModel.export(uri, password.copyOf())
        }
        password?.fill('\u0000')
    }

    val importLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.OpenDocument(),
    ) { uri: Uri? ->
        val password = pendingPassword
        pendingPassword = null
        if (uri != null && password != null) {
            backupViewModel.import(uri, password.copyOf())
        }
        password?.fill('\u0000')
    }

    LaunchedEffect(backupState.message) {
        val message = backupState.message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        backupViewModel.consumeMessage(message)
    }

    LaunchedEffect(accountState.message) {
        val message = accountState.message ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        settingsViewModel.consumeMessage(message)
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
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
                headlineContent = { Text("昵称") },
                supportingContent = {
                    Text(profile?.nickname?.takeIf { it.isNotBlank() } ?: "未设置昵称")
                },
                trailingContent = {
                    TextButton(onClick = { showNicknameDialog = true }, enabled = !accountState.busy) {
                        Text("修改")
                    }
                },
            )
            HorizontalDivider()
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
            HorizontalDivider()
            ListItem(
                headlineContent = { Text("兑换码充值") },
                supportingContent = { Text("输入卡密为会员续期，兑换后立即生效。") },
                trailingContent = {
                    TextButton(onClick = { showRedeemDialog = true }, enabled = !accountState.busy) {
                        Text("去兑换")
                    }
                },
            )
            OutlinedButton(
                onClick = { confirmLogout = true },
                modifier = Modifier.padding(vertical = 16.dp),
            ) {
                Text("退出登录", color = MaterialTheme.colorScheme.error)
            }

            SectionTitle("外观")
            ListItem(
                headlineContent = { Text("深色模式") },
                supportingContent = { Text("聊天页顶部可一键切换深浅色。") },
            )
            FlowRow(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 4.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                ThemeMode.entries.forEach { mode ->
                    FilterChip(
                        selected = appearance.themeMode == mode.key,
                        onClick = { settingsViewModel.setThemeMode(mode.key) },
                        label = { Text(mode.label) },
                    )
                }
            }
            HorizontalDivider()
            Text(
                text = "主题色",
                style = MaterialTheme.typography.bodyLarge,
                modifier = Modifier.padding(top = 12.dp, bottom = 8.dp),
            )
            FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                AccentPalettes.forEach { palette ->
                    FilterChip(
                        selected = appearance.themeColor == palette.key,
                        onClick = { settingsViewModel.setThemeColor(palette.key) },
                        label = { Text(palette.label) },
                        leadingIcon = if (appearance.themeColor == palette.key) {
                            {
                                Icon(
                                    painter = painterResource(R.drawable.ic_check),
                                    contentDescription = null,
                                    modifier = Modifier.size(16.dp),
                                )
                            }
                        } else {
                            null
                        },
                    )
                }
            }
            Spacer(Modifier.height(8.dp))

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

            SectionTitle("数据")
            Text(
                text = "备份包含本机缓存的全部会话和消息，使用 AES-256-GCM 加密后保存到你选择的位置。",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(bottom = 8.dp),
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                OutlinedButton(
                    onClick = { dialogAction = BackupAction.Export },
                    enabled = !backupState.busy,
                    modifier = Modifier.weight(1f),
                ) {
                    Text("导出备份")
                }
                OutlinedButton(
                    onClick = { dialogAction = BackupAction.Import },
                    enabled = !backupState.busy,
                    modifier = Modifier.weight(1f),
                ) {
                    Text("导入备份")
                }
            }
            if (backupState.busy) {
                LinearProgressIndicator(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                )
            }
            Text(
                text = "导入只会合并，不会删除本机已有会话。备份密码不会上传或保存，遗忘后无法恢复。",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp, bottom = 8.dp),
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

    dialogAction?.let { action ->
        BackupPasswordDialog(
            exporting = action == BackupAction.Export,
            onDismiss = { dialogAction = null },
            onConfirm = { password ->
                dialogAction = null
                pendingPassword = password.toCharArray()
                when (action) {
                    BackupAction.Export -> exportLauncher.launch(
                        "aichat-backup-${LocalDate.now()}.aichat",
                    )

                    BackupAction.Import -> importLauncher.launch(
                        arrayOf(
                            "application/octet-stream",
                            "application/zip",
                            "application/x-zip-compressed",
                        ),
                    )
                }
            },
        )
    }

    if (showRedeemDialog) {
        RedeemDialog(
            busy = accountState.busy,
            onDismiss = { showRedeemDialog = false },
            onConfirm = { code ->
                settingsViewModel.redeem(code, onProfileUpdated)
                showRedeemDialog = false
            },
        )
    }

    if (showNicknameDialog) {
        NicknameDialog(
            current = profile?.nickname.orEmpty(),
            busy = accountState.busy,
            onDismiss = { showNicknameDialog = false },
            onConfirm = { nickname ->
                settingsViewModel.updateNickname(nickname, profile, onProfileUpdated)
                showNicknameDialog = false
            },
        )
    }
}

@Composable
private fun RedeemDialog(
    busy: Boolean,
    onDismiss: () -> Unit,
    onConfirm: (String) -> Unit,
) {
    var code by remember { mutableStateOf("") }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("兑换码充值") },
        text = {
            Column {
                Text(
                    text = "输入获得的卡密。兑换成功后会员时长立即累加，兑换码只能使用一次。",
                    style = MaterialTheme.typography.bodyMedium,
                )
                Spacer(Modifier.height(12.dp))
                OutlinedTextField(
                    value = code,
                    onValueChange = { code = it },
                    label = { Text("兑换码") },
                    singleLine = true,
                    enabled = !busy,
                )
            }
        },
        confirmButton = {
            TextButton(
                onClick = { onConfirm(code.trim()) },
                enabled = !busy && code.isNotBlank(),
            ) {
                Text("兑换")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss, enabled = !busy) {
                Text("取消")
            }
        },
    )
}

@Composable
private fun NicknameDialog(
    current: String,
    busy: Boolean,
    onDismiss: () -> Unit,
    onConfirm: (String) -> Unit,
) {
    var nickname by remember { mutableStateOf(current) }
    val changed = nickname.trim() != current && nickname.isNotBlank()

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("修改昵称") },
        text = {
            Column {
                Text(
                    text = "昵称用于展示，最多 50 个字符；留空则保持原昵称不变。",
                    style = MaterialTheme.typography.bodyMedium,
                )
                Spacer(Modifier.height(12.dp))
                OutlinedTextField(
                    value = nickname,
                    onValueChange = { nickname = it },
                    label = { Text("昵称") },
                    singleLine = true,
                    enabled = !busy,
                )
            }
        },
        confirmButton = {
            TextButton(
                onClick = { onConfirm(nickname.trim()) },
                enabled = !busy && changed,
            ) {
                Text("保存")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss, enabled = !busy) {
                Text("取消")
            }
        },
    )
}

@Composable
private fun BackupPasswordDialog(
    exporting: Boolean,
    onDismiss: () -> Unit,
    onConfirm: (String) -> Unit,
) {
    var password by remember { mutableStateOf("") }
    var confirmation by remember { mutableStateOf("") }
    val minimumLength = 8
    val valid = password.length >= minimumLength && (!exporting || password == confirmation)

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(if (exporting) "导出加密备份" else "导入加密备份") },
        text = {
            Column {
                Text(
                    text = if (exporting) {
                        "设置至少 8 位密码。密码不会上传或保存，遗忘后备份无法恢复。"
                    } else {
                        "输入导出备份时设置的密码。连续输错不会锁定，但无法跳过校验。"
                    },
                    style = MaterialTheme.typography.bodyMedium,
                )
                Spacer(Modifier.height(12.dp))
                OutlinedTextField(
                    value = password,
                    onValueChange = { password = it },
                    label = { Text("备份密码") },
                    singleLine = true,
                    visualTransformation = PasswordVisualTransformation(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
                )
                if (exporting) {
                    Spacer(Modifier.height(8.dp))
                    OutlinedTextField(
                        value = confirmation,
                        onValueChange = { confirmation = it },
                        label = { Text("确认密码") },
                        singleLine = true,
                        isError = confirmation.isNotEmpty() && confirmation != password,
                        supportingText = {
                            if (confirmation.isNotEmpty() && confirmation != password) {
                                Text("两次输入的密码不一致")
                            }
                        },
                        visualTransformation = PasswordVisualTransformation(),
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
                    )
                }
                if (password.isNotEmpty() && password.length < minimumLength) {
                    Text(
                        text = "密码至少 8 位",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.error,
                        modifier = Modifier.padding(top = 4.dp),
                    )
                }
            }
        },
        confirmButton = {
            TextButton(onClick = { onConfirm(password) }, enabled = valid) {
                Text(if (exporting) "选择保存位置" else "选择备份文件")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        },
    )
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
