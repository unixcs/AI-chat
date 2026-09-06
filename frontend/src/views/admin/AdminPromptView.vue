<script setup>
import { computed, onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminPrompt,
  getAdminPromptRevisions,
  restoreAdminPrompt,
  updateAdminPrompt,
  getAdminPrefPrompts,
  updateAdminPrefPrompt,
  resetAdminPrefPrompt
} from '../../api/admin'

const info = ref({ content: '', version: null, updatedAt: null, operator: null, source: 'env' })
const revisions = ref([])
const draft = ref('')
const noticeText = ref('')
const errorText = ref('')
const saving = ref(false)
const loading = ref(false)

const dirty = computed(() => draft.value !== info.value.content)

// 后端上限按字节（20000），中文每字 3 字节，字符数会低估占用
const draftBytes = computed(() => new TextEncoder().encode(draft.value).length)

// ---------- 偏好提示词卡片（长度/风格/格式 × 各选项） ----------
const prefCards = ref([])
const prefLoading = ref(false)
const prefSavingKey = ref('')
const prefNotice = ref('')
const prefError = ref('')
const prefDirty = ref({})

const dimensionTitles = { answerLength: '回答长度', answerStyle: '回答风格', answerFormat: '输出格式' }
const prefGroups = computed(() => {
  const groups = []
  for (const [dimension, title] of Object.entries(dimensionTitles)) {
    groups.push({ dimension, title, cards: prefCards.value.filter((c) => c.dimension === dimension) })
  }
  return groups
})

const prefKey = (card) => `${card.dimension}.${card.value}`

const loadPrefCards = async () => {
  prefLoading.value = true
  try {
    const { data } = await getAdminPrefPrompts()
    prefCards.value = (data.data || []).map((c) => ({ ...c, draft: c.content }))
    prefDirty.value = {}
  } catch (error) {
    prefError.value = error.response?.data?.message || '偏好卡片加载失败'
  } finally {
    prefLoading.value = false
  }
}

const markPrefDirty = (card) => {
  prefDirty.value[prefKey(card)] = card.draft !== card.content
}

const savePrefCard = async (card) => {
  prefSavingKey.value = prefKey(card)
  prefNotice.value = ''
  prefError.value = ''
  try {
    await updateAdminPrefPrompt(card.dimension, card.value, card.draft)
    const rev = await getAdminPrefPrompts()
    prefCards.value = (rev.data.data || []).map((c) => ({ ...c, draft: c.content }))
    prefDirty.value = {}
    prefNotice.value = `「${card.label}」卡片已保存，下一个新聊天立即生效`
  } catch (error) {
    prefError.value = error.response?.data?.message || '保存失败'
  } finally {
    prefSavingKey.value = ''
  }
}

const resetPrefCard = async (card) => {
  if (!window.confirm(`把「${card.label}」恢复为系统预置文案？`)) {
    return
  }
  prefSavingKey.value = prefKey(card)
  prefError.value = ''
  try {
    await resetAdminPrefPrompt(card.dimension, card.value)
    const rev = await getAdminPrefPrompts()
    prefCards.value = (rev.data.data || []).map((c) => ({ ...c, draft: c.content }))
    prefDirty.value = {}
    prefNotice.value = `「${card.label}」已恢复预置文案`
  } catch (error) {
    prefError.value = error.response?.data?.message || '恢复失败'
  } finally {
    prefSavingKey.value = ''
  }
}

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

onMounted(() => {
  load()
  loadPrefCards()
})

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
  // 并发保存下"下一个版本号"由服务端决定，确认文案不预推算
  if (!window.confirm(`恢复到 v${item.version}？将以其内容创建一个新版本（历史保留，可随时再恢复）。`)) {
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
        <span class="mutedText">{{ draftBytes }} / 20000 字节</span>
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

    <h3 class="historyTitle">偏好提示词卡片</h3>
    <p class="mutedText promptHint">
      用户在聊天里选择「回答长度 / 回答风格 / 输出格式」后，实际发给 AI 的指令 =
      上方基础内置 Prompt + 对应的三张卡片文案（按长度 → 风格 → 格式顺序拼接）。
      卡片已预置默认文案，修改保存后立即生效，无需重启；进行中的请求保持开始时的版本。
    </p>
    <div v-if="prefLoading" class="mutedText">加载中...</div>
    <div v-for="group in prefGroups" :key="group.dimension" class="prefCardGroup">
      <h4 class="prefGroupTitle">{{ group.title }}</h4>
      <div v-for="card in group.cards" :key="prefKey(card)" class="prefCard">
        <div class="prefCardHead">
          <span class="prefCardLabel">{{ card.label }}</span>
          <span v-if="card.customized" class="tag tagActive">已自定义</span>
          <span v-else class="tag tagOff">预置</span>
          <span class="prefSpacer"></span>
          <button
            v-if="card.customized"
            class="ghostBtn"
            :disabled="prefSavingKey === prefKey(card)"
            @click="resetPrefCard(card)"
          >恢复预置</button>
          <button
            class="primaryBtn"
            :disabled="!prefDirty[prefKey(card)] || prefSavingKey === prefKey(card)"
            @click="savePrefCard(card)"
          >{{ prefSavingKey === prefKey(card) ? '保存中...' : '保存' }}</button>
        </div>
        <textarea v-model="card.draft" rows="3" spellcheck="false" @input="markPrefDirty(card)"></textarea>
        <small class="mutedText prefPresetLine">预置：{{ card.preset }}</small>
      </div>
    </div>
    <p v-if="prefNotice" class="okText">{{ prefNotice }}</p>
    <p v-if="prefError" class="dangerText">{{ prefError }}</p>

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

.prefCardGroup {
  margin-top: 14px;
}

.prefGroupTitle {
  margin: 0 0 8px;
  font-size: 14px;
}

.prefCard {
  display: grid;
  gap: 8px;
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  padding: 12px 14px;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.5);
}

.prefCard textarea {
  width: 100%;
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.65;
  font-family: Consolas, 'Courier New', monospace;
  background: rgba(255, 255, 255, 0.6);
  color: var(--text-main);
  resize: vertical;
}

.prefCardHead {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.prefCardLabel {
  font-weight: 600;
  font-size: 13px;
}

.prefSpacer {
  flex: 1;
}

.prefPresetLine {
  line-height: 1.5;
}

[data-theme='dark'] .prefCard,
[data-theme='dark'] .prefCard textarea {
  background: rgba(255, 255, 255, 0.04);
}

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
