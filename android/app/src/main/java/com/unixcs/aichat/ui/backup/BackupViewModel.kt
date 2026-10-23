package com.unixcs.aichat.ui.backup

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.unixcs.aichat.data.backup.BackupResult
import com.unixcs.aichat.data.backup.ChatBackupRepository
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class BackupUiState(
    val busy: Boolean = false,
    val message: String? = null,
)

class BackupViewModel(
    private val repo: ChatBackupRepository,
) : ViewModel() {

    private val _ui = MutableStateFlow(BackupUiState())
    val ui: StateFlow<BackupUiState> = _ui.asStateFlow()

    fun export(uri: Uri, password: CharArray) = execute(password) { repo.export(uri, password) }

    fun import(uri: Uri, password: CharArray) = execute(password) { repo.import(uri, password) }

    fun consumeMessage(shown: String) {
        _ui.update { if (it.message == shown) it.copy(message = null) else it }
    }

    private fun execute(password: CharArray, action: suspend () -> BackupResult) {
        if (_ui.value.busy) {
            password.fill('\u0000')
            return
        }
        _ui.value = BackupUiState(busy = true)
        viewModelScope.launch {
            val result = try {
                action()
            } catch (e: CancellationException) {
                throw e
            } catch (_: Exception) {
                BackupResult.Failure("备份操作失败，请重试")
            } finally {
                password.fill('\u0000')
            }
            val message = when (result) {
                is BackupResult.Exported -> "已导出 ${result.conversations} 个会话、${result.messages} 条消息"
                is BackupResult.Imported -> "已导入 ${result.conversations} 个会话、${result.messages} 条消息"
                is BackupResult.Failure -> result.message
            }
            _ui.value = BackupUiState(busy = false, message = message)
        }
    }
}
