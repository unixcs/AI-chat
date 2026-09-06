<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { useChatStore } from '../../stores/chat'
import { useAuthStore } from '../../stores/auth'
import { consumeFreshChatFlag, hasDraftSessionFlag, shouldStartFreshOnChatEntry } from '../../utils/chat-entry'
import { renderMarkdownToSafeHtml } from '../../utils/markdown'
import { copyText } from '../../utils/clipboard'
import { updatePreferences, getCurrentAnnouncement, ackAnnouncement } from '../../api/user-extras'

const chatStore = useChatStore()
const authStore = useAuthStore()
const inputValue = ref('')
const errorText = ref('')
const showHistoryDrawer = ref(false)
const messageListRef = ref(null)
const composerTextareaRef = ref(null)

// 桌面端侧栏 vs 移动端抽屉
const isDesktop = ref(window.innerWidth > 960)

// 复制反馈：记录刚复制成功的消息 id，1.5s 后还原
const copiedMessageId = ref(null)
let copiedTimer = null

// 智能滚动：只有用户本来就在底部时才跟随输出
const isNearBottom = ref(true)

// 回答模式（长度 × 风格）
const showAnswerPrefs = ref(false)
const answerLength = ref(authStore.profile?.answerLength || 'standard')
const answerStyle = ref(authStore.profile?.answerStyle || 'standard')
const answerLengthOptions = [
  { value: 'concise', label: '精简' },
  { value: 'standard', label: '适中' },
  { value: 'detailed', label: '详细' }
]
const answerStyleOptions = [
  { value: 'plain', label: '大白话' },
  { value: 'standard', label: '标准' },
  { value: 'professional', label: '专业' },
  { value: 'rigorous', label: '严谨' },
  { value: 'encouraging', label: '鼓励' }
]

// 一次性公告
const announcement = ref(null)
const announcementTitle = ref('')

const activeMessages = computed(() => {
  const id = chatStore.activeConversationId
  if (!id) {
    return []
  }
  return chatStore.messagesMap[id] || []
})

// 流式过程中最后一个消息的内容长度，作为跟随滚动的信号
const lastMessageLength = computed(() => {
  const last = activeMessages.value[activeMessages.value.length - 1]
  return last ? last.content.length : 0
})

const canChat = computed(() => {
  const expireAt = authStore.profile?.memberExpireAt
  if (!expireAt) {
    return false
  }
  return dayjs(expireAt).isAfter(dayjs())
})

const formatTime = (time) => {
  return dayjs(time).format('MM-DD HH:mm')
}

const renderAssistantContent = (content) => {
  return renderMarkdownToSafeHtml(content)
}

const selectConversation = async (conversationId) => {
  await chatStore.fetchMessages(conversationId)
  isNearBottom.value = true
  await scrollToBottom(true)
}

const addConversation = async () => {
  await chatStore.addConversation()
  inputValue.value = ''
  focusComposer()
}

const removeConversation = async (conversationId) => {
  try {
    await chatStore.deleteConversation(conversationId)
  } catch (error) {
    errorText.value = error.response?.data?.message || '删除失败'
  }
}

const openHistoryDrawer = () => {
  showHistoryDrawer.value = true
}

const closeHistoryDrawer = () => {
  showHistoryDrawer.value = false
}

const stopGenerating = () => {
  chatStore.stopStreaming()
}

const focusComposer = async () => {
  await nextTick()
  composerTextareaRef.value?.focus()
}

const messageIsNearBottom = (el) => {
  return el.scrollHeight - el.scrollTop - el.clientHeight < 80
}

const scrollToBottom = async (force = false) => {
  await nextTick()
  const el = messageListRef.value
  if (!el) {
    return
  }
  if (!force && !isNearBottom.value) {
    return
  }
  el.scrollTop = el.scrollHeight
}

const onMessageScroll = () => {
  const el = messageListRef.value
  if (!el) {
    return
  }
  isNearBottom.value = messageIsNearBottom(el)
}

const syncComposerHeight = async () => {
  await nextTick()
  const el = composerTextareaRef.value
  if (!el) {
    return
  }

  const isMobileViewport = window.matchMedia('(max-width: 960px)').matches
  const mobileMaxHeight = 96

  el.style.height = 'auto'
  const nextHeight = isMobileViewport ? Math.min(el.scrollHeight, mobileMaxHeight) : el.scrollHeight
  el.style.height = `${nextHeight}px`
}

const submitMessage = async () => {
  errorText.value = ''
  const content = inputValue.value.trim()
  if (!content) {
    return
  }

  if (content.length > 1000) {
    errorText.value = '最多只能占卜 1,000 字以内哦~'
    return
  }

  if (!canChat.value) {
    errorText.value = '会员过期，请续费后使用。'
    return
  }

  inputValue.value = ''
  isNearBottom.value = true
  await syncComposerHeight()
  try {
    const streamPromise = chatStore.postStreamMessage(content)
    await scrollToBottom(true)
    await streamPromise
  } catch (error) {
    errorText.value = error.message || '请联系管理员 ⚠️E0'
  }
}

const onComposerKeydown = (event) => {
  if (event.key !== 'Enter') {
    return
  }
  if (event.shiftKey) {
    return
  }
  event.preventDefault()
  if (chatStore.streaming) {
    return
  }
  submitMessage()
}

const copyMessage = async (msg) => {
  const ok = await copyText(msg.content)
  if (!ok) {
    errorText.value = '复制失败，请长按文本手动复制'
    return
  }
  copiedMessageId.value = msg.id
  clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => {
    copiedMessageId.value = null
  }, 1500)
}

const setAnswerLength = async (value) => {
  answerLength.value = value
  await persistPreferences()
}

const setAnswerStyle = async (value) => {
  answerStyle.value = value
  await persistPreferences()
}

const persistPreferences = async () => {
  try {
    await updatePreferences({ answerLength: answerLength.value, answerStyle: answerStyle.value })
    if (authStore.profile) {
      authStore.setProfile({
        ...authStore.profile,
        answerLength: answerLength.value,
        answerStyle: answerStyle.value
      })
    }
  } catch (error) {
    errorText.value = error.response?.data?.message || '回答模式保存失败'
  }
}

const acknowledgeAnnouncement = async () => {
  const current = announcement.value
  announcement.value = null
  if (!current) {
    return
  }
  try {
    await ackAnnouncement(current.id, current.updatedAt)
  } catch (error) {
    // 确认失败不打扰用户：下次进入再提示
  }
}

const handleResize = () => {
  isDesktop.value = window.innerWidth > 960
  if (isDesktop.value) {
    showHistoryDrawer.value = false
  }
}

onMounted(async () => {
  window.addEventListener('resize', handleResize)

  // 未读公告（一次性展示）
  try {
    const { data } = await getCurrentAnnouncement()
    if (data?.data) {
      announcement.value = data.data
    }
  } catch (error) {
    // 公告拉取失败不影响聊天
  }

  const startFresh = shouldStartFreshOnChatEntry({
    hasFreshChatFlag: consumeFreshChatFlag(),
    hasDraftSession: hasDraftSessionFlag()
  })
  await chatStore.fetchConversations({ startFresh })
  await syncComposerHeight()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  clearTimeout(copiedTimer)
})

watch(
  () => chatStore.uiResetKey,
  () => {
    inputValue.value = ''
    errorText.value = ''
    closeHistoryDrawer()
    syncComposerHeight()
  }
)

watch(inputValue, () => {
  syncComposerHeight()
})

// profile 加载晚于本组件挂载时，回填服务器端已存的回答模式（UI 不说谎）
watch(() => authStore.profile, (p) => {
  if (!p) {
    return
  }
  if (p.answerLength) {
    answerLength.value = p.answerLength
  }
  if (p.answerStyle) {
    answerStyle.value = p.answerStyle
  }
})

// 输出跟随时持续滚到底部；用户上滑后不打扰
watch(lastMessageLength, async () => {
  if (chatStore.streaming && isNearBottom.value) {
    await scrollToBottom(true)
  }
})
</script>

<template>
  <section class="chatStage">
    <section class="chatPanel card panelShell">
      <div ref="messageListRef" class="messageViewport" @scroll="onMessageScroll">
        <div v-if="activeMessages.length === 0" class="emptyState">
          <h2>告诉我你有什么想法</h2>
          <p class="emptyHint">开始新的对话吧</p>
        </div>

        <article
          v-for="msg in activeMessages"
          :key="msg.id"
          class="messageRow"
          :class="msg.role === 'user' ? 'isUser' : 'isBot'"
        >
          <img
            class="messageAvatar"
            :src="msg.role === 'user' ? '/assets/user-avatar.png' : '/assets/ai-avatar.png'"
            :alt="msg.role === 'user' ? '用户头像' : 'AI头像'"
          />

          <div class="messageMeta">
            <span class="messageRole">{{ msg.role === 'user' ? '你' : 'Thallo' }}</span>
            <div class="messageBubble">
              <div v-if="msg.role === 'assistant'">
                <span
                  v-if="chatStore.streaming && msg.id === chatStore.streamingAssistantId"
                  class="streamingPlainText"
                >{{ msg.content }}</span>
                <div
                  v-else
                  class="markdownBody"
                  v-html="renderAssistantContent(msg.content)"
                />
              </div>
              <p v-else>{{ msg.content }}</p>
            </div>
            <div class="messageFooter">
              <time>{{ formatTime(msg.createdAt) }}</time>
              <button
                class="copyBtn"
                :class="{ copied: copiedMessageId === msg.id }"
                :title="copiedMessageId === msg.id ? '已复制' : '复制'"
                :aria-label="copiedMessageId === msg.id ? '已复制' : '复制消息'"
                @click="copyMessage(msg)"
              >
                <svg v-if="copiedMessageId !== msg.id" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M8 3h11a1 1 0 0 1 1 1v12h-2V5H8V3zM5 7h11a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V8a1 1 0 0 1 1-1zm1 2v10h9V9H6z" />
                </svg>
                <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M9 16.2 4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2z" />
                </svg>
                <span class="copyLabel">{{ copiedMessageId === msg.id ? '已复制' : '复制' }}</span>
              </button>
            </div>
          </div>
        </article>

        <div v-if="chatStore.streaming" class="typingTips mutedText">
          灵感开启中，请稍等片刻...
        </div>
      </div>

      <div class="composerShell">
        <div v-if="showAnswerPrefs" class="answerPrefsPanel card">
          <div class="prefGroup">
            <span class="prefLabel">回答长度</span>
            <div class="prefChips">
              <button
                v-for="opt in answerLengthOptions"
                :key="opt.value"
                class="prefChip"
                :class="{ active: answerLength === opt.value }"
                @click="setAnswerLength(opt.value)"
              >{{ opt.label }}</button>
            </div>
          </div>
          <div class="prefGroup">
            <span class="prefLabel">回答风格</span>
            <div class="prefChips">
              <button
                v-for="opt in answerStyleOptions"
                :key="opt.value"
                class="prefChip"
                :class="{ active: answerStyle === opt.value }"
                @click="setAnswerStyle(opt.value)"
              >{{ opt.label }}</button>
            </div>
          </div>
        </div>

        <div class="composerSurface">
          <textarea
            ref="composerTextareaRef"
            v-model="inputValue"
            rows="1"
            placeholder="把你此刻最想问的内容写下来..."
            @keydown="onComposerKeydown"
          />

          <div class="composerActions">
            <button
              class="circleIconBtn historyIconBtn"
              title="历史对话"
              aria-label="历史对话"
              @click="openHistoryDrawer"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M5 5.75A1.75 1.75 0 0 1 6.75 4h10.5A1.75 1.75 0 0 1 19 5.75v12.5A1.75 1.75 0 0 1 17.25 20H6.75A1.75 1.75 0 0 1 5 18.25V5.75zm2.5.75v2h9v-2h-9zm0 4v2h9v-2h-9zm0 4v2h6v-2h-6z"
                />
              </svg>
              <span class="historyBtnLabel">历史</span>
            </button>
            <button
              class="circleIconBtn prefIconBtn"
              :class="{ active: showAnswerPrefs }"
              title="回答模式"
              aria-label="回答模式"
              @click="showAnswerPrefs = !showAnswerPrefs"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M4 5h16v2H4V5zm0 6h10v2H4v-2zm0 6h7v2H4v-2zm13-.2 2.1-2.1 1.4 1.4L18.4 18l2.1 2.1-1.4 1.4L17 19.4l-2.1 2.1-1.4-1.4 2.1-2.1-2.1-2.1 1.4-1.4 2.1 2.1z" />
              </svg>
            </button>
            <button
              v-if="chatStore.streaming"
              class="circleIconBtn stopIconBtn"
              title="停止生成"
              aria-label="停止生成"
              @click="stopGenerating"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M7 7h10v10H7z" />
              </svg>
            </button>
            <button
              v-else
              class="circleIconBtn sendIconBtn"
              title="发送消息"
              aria-label="发送消息"
              @click="submitMessage"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M3 20l18-8L3 4v6l12 2-12 2v6z" />
              </svg>
            </button>
          </div>
        </div>

        <p v-if="errorText" class="dangerText composerError">{{ errorText }}</p>
      </div>
    </section>

    <div v-if="showHistoryDrawer" class="historyDrawerMask" @click.self="closeHistoryDrawer">
      <aside class="historyDrawerPanel card" :class="'panelShell'">
        <div class="historyDrawerHead">
          <div>
            <span class="sectionLabel">History</span>
            <h3>历史对话</h3>
          </div>
          <div class="historyActions">
            <button class="ghostBtn" @click="closeHistoryDrawer">关闭</button>
            <button class="primaryBtn" @click="addConversation">新建</button>
          </div>
        </div>

        <div
          v-for="item in chatStore.list"
          :key="item.id"
          class="historyItem"
          :class="{ active: chatStore.activeConversationId === item.id }"
          role="button"
          tabindex="0"
          @click="selectConversation(item.id); closeHistoryDrawer()"
          @keydown.enter="selectConversation(item.id); closeHistoryDrawer()"
        >
          <span class="historyTitle">{{ item.title }}</span>
          <small>{{ formatTime(item.updatedAt) }}</small>
          <button
            class="historyDeleteBtn"
            title="删除对话"
            aria-label="删除对话"
            @click.stop="removeConversation(item.id)"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M9 3h6l1 2h4v2H4V5h4l1-2zm-3 6h12l-1 12a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2L6 9zm4 3v7h2v-7h-2zm4 0v7h2v-7h-2z" />
            </svg>
          </button>
        </div>
        <p v-if="chatStore.list.length === 0" class="mutedText historyEmpty">还没有对话记录</p>
      </aside>
    </div>

    <div v-if="announcement" class="announcementMask" @click.self="acknowledgeAnnouncement">
      <div class="announcementPanel card panelShell" role="dialog" aria-modal="true">
        <span class="sectionLabel">Announcement</span>
        <h3>{{ announcement.title || '系统公告' }}</h3>
        <p class="announcementContent">{{ announcement.content }}</p>
        <button class="primaryBtn announcementAck" @click="acknowledgeAnnouncement">我知道了</button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.chatStage {
  display: grid;
  height: calc(100dvh - 44px);
  min-height: 0;
  width: 100%;
}

.chatPanel {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  position: relative;
}

.messageViewport {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 26px;
  background: var(--chat-surface);
}

.emptyState {
  min-height: 100%;
  display: grid;
  align-content: center;
  justify-items: center;
  text-align: center;
  gap: 8px;
  padding: 50px 24px;
}

.emptyState h2 {
  margin: 0;
  color: var(--text-title);
  font-size: clamp(24px, 4vw, 36px);
}

.typingTips {
  text-align: center;
  padding: 14px;
  font-size: 12px;
}

.messageRow {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 18px;
  width: 100%;
}

.messageAvatar {
  width: 38px;
  height: 38px;
  border-radius: 16px;
  object-fit: cover;
  flex: 0 0 auto;
  border: 1px solid rgba(148, 163, 184, 0.24);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.08);
}

.messageMeta {
  min-width: 0;
  width: min(100%, 860px);
  display: grid;
  gap: 8px;
}

.messageRole {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-soft);
  text-transform: uppercase;
}

.messageBubble {
  min-width: 0;
  width: fit-content;
  max-width: min(82%, 760px);
  border-radius: 24px;
  padding: 16px 18px;
  border: 1px solid transparent;
  box-shadow: var(--shadow-soft);
}

.messageRow p {
  margin: 0;
  white-space: pre-wrap;
  font-size: 15px;
  line-height: 1.75;
}

.messageFooter {
  display: flex;
  align-items: center;
  gap: 10px;
}

.messageFooter time {
  display: block;
  font-size: 12px;
  color: var(--text-soft);
}

.copyBtn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.5);
  color: var(--text-soft);
  cursor: pointer;
  padding: 5px 10px;
  font-size: 12px;
  line-height: 1;
  transition: transform 0.15s ease, color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
}

.copyBtn svg {
  width: 14px;
  height: 14px;
  fill: currentColor;
}

.copyBtn:hover {
  transform: translateY(-1px);
  color: var(--text-main);
  border-color: var(--line-strong);
}

.copyBtn.copied {
  color: #3f7d4e;
  border-color: rgba(63, 125, 78, 0.35);
  background: rgba(63, 125, 78, 0.08);
}

.copyLabel {
  font-weight: 600;
}

.markdownBody {
  font-size: 15px;
  line-height: 1.72;
  color: var(--text-main);
  word-break: break-word;
}

.streamingPlainText {
  display: block;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 15px;
  line-height: 1.72;
  color: var(--text-main);
}

.markdownBody :deep(p),
.markdownBody :deep(ul),
.markdownBody :deep(ol),
.markdownBody :deep(blockquote),
.markdownBody :deep(pre) {
  margin: 0 0 12px;
}

.markdownBody :deep(*:last-child) {
  margin-bottom: 0;
}

.markdownBody :deep(ul),
.markdownBody :deep(ol) {
  padding-left: 20px;
}

.markdownBody :deep(table) {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  margin-bottom: 10px;
  display: block;
  overflow-x: auto;
}

.markdownBody :deep(th),
.markdownBody :deep(td) {
  padding: 10px 12px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  text-align: left;
  vertical-align: top;
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.markdownBody :deep(thead th) {
  background: rgba(148, 163, 184, 0.08);
}

.markdownBody :deep(li + li) {
  margin-top: 4px;
}

.markdownBody :deep(a) {
  color: var(--accent-strong);
  text-decoration: underline;
}

.markdownBody :deep(code) {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 13px;
  background: rgba(17, 24, 39, 0.08);
  padding: 2px 6px;
  border-radius: 8px;
}

.markdownBody :deep(pre) {
  overflow-x: auto;
  padding: 14px;
  border-radius: 18px;
  background: rgba(17, 24, 39, 0.08);
}

.markdownBody :deep(pre code) {
  background: transparent;
  padding: 0;
  border-radius: 0;
}

.markdownBody :deep(blockquote) {
  padding-left: 12px;
  border-left: 3px solid rgba(95, 111, 133, 0.35);
  color: var(--text-soft);
}

.isUser {
  justify-content: flex-start;
  flex-direction: row-reverse;
}

.isUser .messageMeta {
  justify-items: end;
}

.isUser .messageBubble {
  background: var(--chat-user-bg);
  border-color: var(--chat-user-border);
  color: var(--chat-user-text);
}

.isUser .messageBubble p,
.isUser .messageBubble :deep(*) {
  color: inherit;
}

.isUser .messageBubble::selection,
.isUser .messageBubble p::selection,
.isUser .messageBubble span::selection,
.isUser .messageBubble strong::selection,
.isUser .messageBubble em::selection,
.isUser .messageBubble b::selection,
.isUser .messageBubble i::selection,
.isUser .messageBubble code::selection,
.isUser .messageBubble a::selection,
.isUser .messageBubble :deep(*)::selection {
  background: var(--chat-user-selection-bg);
  color: var(--chat-user-selection-text);
}

.isBot .messageBubble {
  background: var(--chat-bot-bg);
  border-color: var(--chat-bot-border);
}

.isBot .messageBubble::selection,
.isBot .messageBubble p::selection,
.isBot .messageBubble span::selection,
.isBot .messageBubble strong::selection,
.isBot .messageBubble em::selection,
.isBot .messageBubble b::selection,
.isBot .messageBubble i::selection,
.isBot .messageBubble code::selection,
.isBot .messageBubble a::selection,
.isBot .messageBubble li::selection,
.isBot .messageBubble blockquote::selection,
.isBot .messageBubble th::selection,
.isBot .messageBubble td::selection,
.isBot .messageBubble :deep(*)::selection {
  background: var(--chat-bot-selection-bg);
  color: var(--chat-bot-selection-text);
}

.composerShell {
  flex: 0 0 auto;
  border-top: 1px solid var(--line-soft);
  padding: 18px 20px 20px;
  background: var(--surface-base);
}

.composerSurface {
  display: flex;
  align-items: flex-end;
  gap: 14px;
  border: 1px solid var(--line-soft);
  border-radius: 26px;
  padding: 14px 14px 14px 18px;
  background: rgba(255, 255, 255, 0.58);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

[data-theme='dark'] .composerSurface {
  background: rgba(255, 255, 255, 0.04);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.composerSurface textarea {
  resize: none;
  min-height: 68px;
  width: 100%;
  flex: 1 1 auto;
  background: transparent;
  color: var(--text-main);
  border: none;
  padding: 0;
  font-size: 15px;
  line-height: 1.7;
  overflow-y: auto;
}

.composerSurface textarea::selection {
  background: var(--chat-input-selection-bg);
  color: var(--chat-input-selection-text);
}

.composerSurface textarea:focus {
  outline: none;
}

.composerActions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.circleIconBtn {
  width: 46px;
  height: 46px;
  border-radius: 18px;
  border: 1px solid var(--line-strong);
  background: rgba(255, 255, 255, 0.6);
  color: var(--text-main);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

[data-theme='dark'] .circleIconBtn {
  background: rgba(255, 255, 255, 0.04);
}

.circleIconBtn svg {
  width: 18px;
  height: 18px;
  fill: currentColor;
}

.circleIconBtn:hover {
  transform: translateY(-1px);
  border-color: rgba(100, 111, 125, 0.3);
  background: rgba(255, 255, 255, 0.82);
}

[data-theme='dark'] .circleIconBtn:hover {
  background: rgba(255, 255, 255, 0.08);
}

.sendIconBtn {
  border-color: transparent;
  background: linear-gradient(135deg, #5e6d81 0%, #4b596d 100%);
  color: #f7f6f2;
  box-shadow: 0 16px 26px rgba(86, 98, 116, 0.24);
}

.sendIconBtn:hover {
  background: linear-gradient(135deg, #4f5d71 0%, #424f61 100%);
}

.prefIconBtn.active {
  border-color: rgba(95, 111, 133, 0.45);
  background: rgba(95, 111, 133, 0.14);
}

.stopIconBtn {
  color: var(--danger);
  background: rgba(198, 93, 75, 0.08);
  border-color: rgba(198, 93, 75, 0.16);
}

.composerError {
  margin: 10px 2px 0;
  font-size: 12px;
}

.answerPrefsPanel {
  display: grid;
  gap: 12px;
  padding: 14px 16px;
  margin-bottom: 10px;
  border-radius: 18px;
}

.prefGroup {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.prefLabel {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-soft);
  min-width: 60px;
}

.prefChips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.prefChip {
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.5);
  color: var(--text-soft);
  font-size: 12px;
  padding: 6px 12px;
  cursor: pointer;
  transition: border-color 0.15s ease, background 0.15s ease, color 0.15s ease;
}

.prefChip.active {
  border-color: rgba(95, 111, 133, 0.45);
  background: rgba(95, 111, 133, 0.14);
  color: var(--text-main);
  font-weight: 700;
}

.historyDrawerMask {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
  display: flex;
  justify-content: flex-end;
  z-index: 90;
}

.historyDrawerPanel {
  width: min(360px, 94vw);
  height: 100dvh;
  padding: 24px 20px;
  border-radius: 0;
  overflow: auto;
}

.historyDrawerHead {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 18px;
}

.historyDrawerHead h3 {
  margin: 12px 0 0;
  font-size: 24px;
  color: var(--text-title);
}

.historyActions {
  display: flex;
  gap: 8px;
}

.historyItem {
  position: relative;
  width: 100%;
  border: 1px solid var(--line-soft);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.42);
  margin-bottom: 10px;
  padding: 15px 44px 15px 16px;
  text-align: left;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

.historyTitle {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.historyDeleteBtn {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 30px;
  height: 30px;
  border-radius: 10px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-soft);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s ease, background 0.15s ease, border-color 0.15s ease;
}

.historyDeleteBtn svg {
  width: 15px;
  height: 15px;
  fill: currentColor;
}

.historyDeleteBtn:hover {
  color: var(--danger, #c65d4b);
  background: rgba(198, 93, 75, 0.08);
  border-color: rgba(198, 93, 75, 0.2);
}

.historyEmpty {
  text-align: center;
  padding: 18px 0;
  font-size: 13px;
}

.announcementMask {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
  display: grid;
  place-items: center;
  z-index: 120;
  padding: 20px;
}

.announcementPanel {
  width: min(460px, 94vw);
  padding: 26px 24px;
  display: grid;
  gap: 12px;
  justify-items: start;
}

.announcementPanel h3 {
  margin: 0;
  font-size: 22px;
  color: var(--text-title);
}

.announcementContent {
  margin: 0;
  white-space: pre-wrap;
  color: var(--text-main);
  line-height: 1.7;
  font-size: 14px;
  max-height: 50vh;
  overflow: auto;
}

.announcementAck {
  justify-self: end;
}

[data-theme='dark'] .historyItem {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
}

.historyItem:hover {
  transform: translateY(-1px);
  border-color: var(--line-strong);
  background: rgba(255, 255, 255, 0.72);
}

[data-theme='dark'] .historyItem:hover {
  background: rgba(255, 255, 255, 0.07);
}

.historyItem small {
  color: var(--text-soft);
  flex: 0 0 auto;
}

.historyItem.active {
  border-color: rgba(95, 111, 133, 0.3);
  background: rgba(95, 111, 133, 0.12);
  box-shadow: var(--shadow-soft);
}

@media (max-width: 960px) {
  .chatStage {
    height: calc(100dvh - 84px);
  }

  .chatPanel {
    height: 100%;
  }

  .messageViewport {
    padding: 18px 14px;
  }

  .messageRow {
    gap: 10px;
  }

  .messageAvatar {
    width: 34px;
    height: 34px;
    border-radius: 14px;
  }

  .messageBubble {
    max-width: 100%;
    border-radius: 20px;
    padding: 14px 15px;
  }

  .copyBtn {
    padding: 7px 12px;
    font-size: 13px;
    border-radius: 14px;
  }

  .copyBtn svg {
    width: 15px;
    height: 15px;
  }

  .composerSurface {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    border-radius: 22px;
    padding: 10px 10px 10px 12px;
  }

  .composerShell {
    padding: 10px 12px 12px;
  }

  .composerSurface textarea {
    min-height: 22px;
    max-height: 96px;
    font-size: 14px;
    line-height: 1.4;
    -webkit-overflow-scrolling: touch;
  }

  .composerActions {
    justify-content: flex-end;
    align-items: center;
    gap: 6px;
  }

  .circleIconBtn {
    width: 34px;
    height: 34px;
    border-radius: 12px;
  }

  .circleIconBtn svg {
    width: 15px;
    height: 15px;
  }

  .historyIconBtn {
    width: auto;
    min-width: 34px;
    padding: 0 10px;
    gap: 4px;
  }

  .historyBtnLabel {
    display: inline;
    font-size: 12px;
    font-weight: 600;
    line-height: 1;
  }

  .historyDrawerPanel {
    width: 100%;
    padding: 18px 14px;
  }

  .historyDrawerHead,
  .historyActions {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
