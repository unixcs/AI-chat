<script setup>
import { computed, onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminPrompt,
  getAdminPromptRevisions,
  restoreAdminPrompt,
  updateAdminPrompt
} from '../../api/admin'

const info = ref({ content: '', version: null, updatedAt: null, operator: null, source: 'env' })
const revisions = ref([])
const draft = ref('')
const noticeText = ref('')
const errorText = ref('')
const saving = ref(false)
const loading = ref(false)

const dirty = computed(() => draft.value !== info.value.content)

const load = async () => {
  loading.value = true
  try {
    const { data } = await getAdminPrompt()
    info.value = data.data
    draft.value = data.data.content || ''
    const rev = await getAdminPromptRevisions()
    revisions.value = rev.data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '提示词加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)

const save = async () => {
  errorText.value = ''
  noticeText.value = ''
  if (!draft.value.trim()) {
    errorText.value = '提示词内容不能为空'
    return
  }
  if (!dirty.value) {
    return
  }
  saving.value = true
  try {
    const { data } = await updateAdminPrompt(draft.value)
    info.value = data.data
    draft.value = data.data.content
    noticeText.value = `已保存为新版本 v${data.data.version}，新聊天立即生效`
    const rev = await getAdminPromptRevisions()
    revisions.value = rev.data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '保存失败'
  } finally {
    saving.value = false
  }
}

const restore = async (item) => {
  if (!window.confirm(`恢复到 v${item.version}？将以其内容创建新版本 v${(info.value.version ?? 0) + 1}，历史保留。`)) {
    return
  }
  errorText.value = ''
  noticeText.value = ''
  try {
    const { data } = await restoreAdminPrompt(item.version)
    info.value = data.data
    draft.value = data.data.content
    noticeText.value = `已恢复 v${item.version} 的内容为新版本 v${data.data.version}`
    const rev = await getAdminPromptRevisions()
    revisions.value = rev.data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '恢复失败'
  }
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const preview = (content) => {
  const oneLine = content.replace(/\s+/g, ' ').trim()
  return oneLine.length > 80 ? oneLine.slice(0, 80) + '…' : oneLine
}
</script>

<template>
  <section class="card panelShell panel promptPanel">
    <span class="sectionLabel">Prompt</span>
    <h2 class="sectionTitle">内置提示词管理</h2>

    <div class="metaRow">
      <span class="tag" :class="info.source === 'db' ? 'tagActive' : 'tagOff'">
        {{ info.source === 'db' ? `SQLite · v${info.version}` : '环境/出厂配置' }}
      </span>
      <small class="mutedText">
        最后修改：{{ formatTime(info.updatedAt) }}
        <template v-if="info.operator"> · 操作人：{{ info.operator }}</template>
      </small>
    </div>
    <p class="mutedText promptHint">
      修改保存后无需重启，下一个新聊天立即使用最新提示词；生成进行中的请求沿用开始时的版本。
      用户自己的“回答长度 / 回答风格”会叠加在这段提示词之后，互不冲突。
    </p>

    <div class="editorBox">
      <textarea v-model="draft" rows="12" spellcheck="false" placeholder="输入系统内置提示词…"></textarea>
      <div class="editorActions">
        <span class="mutedText">{{ draft.length }} 字符</span>
        <span class="editorSpacer"></span>
        <button v-if="dirty" class="ghostBtn" @click="draft = info.content">放弃修改</button>
        <button class="primaryBtn" :disabled="!dirty || saving" @click="save">
          {{ saving ? '保存中...' : '保存并生效' }}
        </button>
      </div>
      <p v-if="noticeText" class="okText">{{ noticeText }}</p>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
      <p v-if="loading" class="mutedText">加载中...</p>
    </div>

    <h3 class="historyTitle">版本历史（最近 50 条）</h3>
    <div v-if="revisions.length === 0" class="mutedText emptyHistory">
      还没有保存过版本——当前展示的是环境/出厂配置的内容。
    </div>
    <div v-for="item in revisions" :key="item.version" class="revisionCard">
      <div class="revisionHead">
        <span class="revisionVersion">v{{ item.version }}</span>
        <span v-if="item.version === info.version" class="tag tagActive">当前生效</span>
        <small class="mutedText">{{ formatTime(item.createdAt) }} · {{ item.operator }}</small>
        <span class="revisionSpacer"></span>
        <button class="ghostBtn" @click="restore(item)">恢复此版本</button>
      </div>
      <p class="revisionPreview">{{ preview(item.content) }}</p>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 20px;
}

.promptPanel .sectionTitle {
  margin: 14px 0 14px;
}

.metaRow {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.promptHint {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.7;
}

.tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.tagActive {
  background: rgba(63, 125, 78, 0.12);
  color: #3f7d4e;
}

.tagOff {
  background: rgba(148, 163, 184, 0.16);
  color: var(--text-soft);
}

.editorBox {
  margin-top: 14px;
  display: grid;
  gap: 10px;
}

.editorBox textarea {
  width: 100%;
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  padding: 14px;
  font-size: 14px;
  line-height: 1.7;
  font-family: Consolas, 'Courier New', monospace;
  background: rgba(255, 255, 255, 0.5);
  color: var(--text-main);
  resize: vertical;
}

[data-theme='dark'] .editorBox textarea {
  background: rgba(255, 255, 255, 0.04);
}

.editorActions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.editorSpacer,
.revisionSpacer {
  flex: 1;
}

.okText {
  color: #3f7d4e;
  font-size: 13px;
  margin: 0;
}

.dangerText {
  color: var(--danger, #c65d4b);
  font-size: 13px;
  margin: 0;
}

.historyTitle {
  margin: 22px 0 12px;
  font-size: 16px;
  color: var(--text-title);
}

.emptyHistory {
  font-size: 13px;
}

.revisionCard {
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.4);
  padding: 12px 14px;
  margin-bottom: 10px;
}

[data-theme='dark'] .revisionCard {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
}

.revisionHead {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.revisionVersion {
  font-weight: 800;
  color: var(--text-title);
}

.revisionPreview {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--text-soft);
  word-break: break-word;
  overflow-wrap: anywhere;
}

@media (max-width: 640px) {
  .panel {
    padding: 12px;
  }

  .editorActions {
    flex-wrap: wrap;
  }

  .revisionHead {
    align-items: flex-start;
    flex-direction: column;
  }

  .revisionSpacer {
    display: none;
  }

  .revisionHead .ghostBtn {
    width: 100%;
  }
}
</style>
