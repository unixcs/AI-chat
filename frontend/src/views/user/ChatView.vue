<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { useChatStore } from '../../stores/chat'
import { useAuthStore } from '../../stores/auth'
import { consumeFreshChatFlag, hasDraftSessionFlag, shouldStartFreshOnChatEntry } from '../../utils/chat-entry'
import { renderMarkdownToSafeHtml } from '../../utils/markdown'

const chatStore = useChatStore()
const authStore = useAuthStore()
const inputValue = ref('')
const errorText = ref('')
const showHistoryDrawer = ref(false)
const messageListRef = ref(null)
const composerTextareaRef = ref(null)

const activeMessages = computed(() => {
  const id = chatStore.activeConversationId
  if (!id) {
    return []
  }
  return chatStore.messagesMap[id] || []
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
}

const addConversation = async () => {
  await chatStore.addConversation()
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

const scrollToBottom = async () => {
  await nextTick()
  const el = messageListRef.value
  if (!el) {
    return
  }
  el.scrollTop = el.scrollHeight
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
  await syncComposerHeight()
  try {
    const streamPromise = chatStore.postStreamMessage(content)
    await scrollToBottom()
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

onMounted(async () => {
  const startFresh = shouldStartFreshOnChatEntry({
    hasFreshChatFlag: consumeFreshChatFlag(),
    hasDraftSession: hasDraftSessionFlag()
  })
  await chatStore.fetchConversations({ startFresh })
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

onMounted(async () => {
  await syncComposerHeight()
})
</script>

<template>
  <section class="chatStage">
    <section class="chatPanel card panelShell">
      <div ref="messageListRef" class="messageViewport">
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
              <div
                v-if="msg.role === 'assistant'"
                class="markdownBody"
                v-html="renderAssistantContent(msg.content)"
              />
              <p v-else>{{ msg.content }}</p>
            </div>
            <time>{{ formatTime(msg.createdAt) }}</time>
          </div>
        </article>

        <div v-if="chatStore.streaming" class="typingTips mutedText">
          灵感开启中，请稍等片刻...
        </div>
      </div>

      <div class="composerShell">
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
              class="circleIconBtn sendIconBtn"
              :disabled="chatStore.streaming"
              :title="chatStore.streaming ? '生成中' : '发送消息'"
              :aria-label="chatStore.streaming ? '生成中' : '发送消息'"
              @click="submitMessage"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M3 20l18-8L3 4v6l12 2-12 2v6z" />
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

        <button
          v-for="item in chatStore.list"
          :key="item.id"
          class="historyItem"
          :class="{ active: chatStore.activeConversationId === item.id }"
          @click="selectConversation(item.id); closeHistoryDrawer()"
        >
          <span>{{ item.title }}</span>
          <small>{{ formatTime(item.updatedAt) }}</small>
        </button>
      </aside>
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

.messageRow time {
  display: block;
  font-size: 12px;
  color: var(--text-soft);
}

.markdownBody {
  font-size: 15px;
  line-height: 1.72;
  color: var(--text-main);
  word-break: break-word;
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

.sendIconBtn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.historyIconBtn {
  color: var(--text-soft);
}

.historyBtnLabel {
  display: none;
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
  width: 100%;
  border: 1px solid var(--line-soft);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.42);
  margin-bottom: 10px;
  padding: 15px 16px;
  text-align: left;
  display: flex;
  justify-content: space-between;
  gap: 10px;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
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
