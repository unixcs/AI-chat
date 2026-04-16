<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { watch } from 'vue'
import dayjs from 'dayjs'
import { useChatStore } from '../../stores/chat'
import { useAuthStore } from '../../stores/auth'
import { consumeFreshChatFlag, shouldCreateFreshConversationOnEntry } from '../../utils/chat-entry'
import { renderMarkdownToSafeHtml } from '../../utils/markdown'

const chatStore = useChatStore()
const authStore = useAuthStore()
const inputValue = ref('')
const errorText = ref('')
const showHistoryPanel = ref(false)
const showHistoryDrawer = ref(false)
const messageListRef = ref(null)

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
  const startFresh = shouldCreateFreshConversationOnEntry(consumeFreshChatFlag())
  await chatStore.fetchConversations({ startFresh })
})

watch(
  () => chatStore.uiResetKey,
  () => {
    inputValue.value = ''
    errorText.value = ''
    closeHistoryDrawer()
  }
)
</script>

<template>
  <section class="chatPage">
    <section class="card chatPanel">
      <div ref="messageListRef" class="messageList">
        <div v-if="activeMessages.length === 0" class="emptyTips mutedText">
          开始新的对话吧
        </div>

        <article
          v-for="msg in activeMessages"
          :key="msg.id"
          class="msgRow"
          :class="msg.role === 'user' ? 'isUser' : 'isBot'"
        >
          <img
            class="msgAvatar"
            :src="msg.role === 'user' ? '/assets/user-avatar.png' : '/assets/ai-avatar.png'"
            :alt="msg.role === 'user' ? '用户头像' : 'AI头像'"
          />
          <div class="msgBody">
            <div
              v-if="msg.role === 'assistant'"
              class="markdownBody"
              v-html="renderAssistantContent(msg.content)"
            />
            <p v-else>{{ msg.content }}</p>
            <time>{{ formatTime(msg.createdAt) }}</time>
          </div>
        </article>

        <div v-if="chatStore.streaming" class="typingTips mutedText">
          灵感开启中，请稍等片刻...
        </div>
      </div>

      <div class="composer">
        <div class="composerRow">
          <textarea
            v-model="inputValue"
            rows="2"
            placeholder="请输入内容..."
            @keydown="onComposerKeydown"
          />
          <div class="composerButtons">
            <button
              class="circleIconBtn historyIconBtn"
              title="历史对话"
              aria-label="历史对话"
              @click="openHistoryDrawer"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M4 6a2 2 0 0 1 2-2h3v2H6v12h12v-3h2v3a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6zm8-2h8v8h-2V7.41l-5.3 5.3-1.4-1.42L16.58 6H12V4z"
                />
              </svg>
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
      <aside class="historyDrawer card">
        <div class="historyHead">
          <h3>历史对话</h3>
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
.chatPage {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  align-items: stretch;
  min-height: 0;
  width: 100%;
}

.chatPanel {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 28px);
  min-height: calc(100dvh - 28px);
  width: 100%;
  overflow: hidden;
  position: relative;
}

.messageList {
  flex: 1;
  overflow: auto;
  padding: 14px;
  background: var(--chat-surface);
}

.emptyTips {
  text-align: center;
  margin-top: 60px;
}

.typingTips {
  text-align: center;
  padding: 6px;
  font-size: 12px;
}

.msgRow {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 9px;
  width: 100%;
}

.msgAvatar {
  width: 26px;
  height: 26px;
  border-radius: 999px;
  object-fit: cover;
  flex: 0 0 auto;
  border: 1px solid rgba(148, 163, 184, 0.34);
}

.msgBody {
  min-width: 0;
  width: fit-content;
  max-width: min(80%, 680px);
  border-radius: 12px;
  padding: 9px 11px;
  border: 1px solid transparent;
  box-shadow: var(--shadow-soft);
}

.msgRow p {
  margin: 0;
  white-space: pre-wrap;
  font-size: 14px;
  line-height: 1.45;
}

.markdownBody {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-main);
  word-break: break-word;
}

.markdownBody :deep(p),
.markdownBody :deep(ul),
.markdownBody :deep(ol),
.markdownBody :deep(blockquote),
.markdownBody :deep(pre) {
  margin: 0 0 10px;
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
  padding: 8px 10px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  text-align: left;
  vertical-align: top;
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.markdownBody :deep(thead th) {
  background: rgba(148, 163, 184, 0.12);
}

.markdownBody :deep(li + li) {
  margin-top: 4px;
}

.markdownBody :deep(a) {
  color: var(--bg-accent);
  text-decoration: underline;
}

.markdownBody :deep(code) {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 13px;
  background: rgba(15, 23, 42, 0.08);
  padding: 1px 4px;
  border-radius: 4px;
}

.markdownBody :deep(pre) {
  overflow-x: auto;
  padding: 10px;
  border-radius: 10px;
  background: rgba(15, 23, 42, 0.08);
}

.markdownBody :deep(pre code) {
  background: transparent;
  padding: 0;
  border-radius: 0;
}

.markdownBody :deep(blockquote) {
  padding-left: 12px;
  border-left: 3px solid rgba(47, 124, 246, 0.35);
  color: var(--text-soft);
}

.msgRow time {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-soft);
}

.isUser {
  justify-content: flex-start;
  flex-direction: row-reverse;
}

.isBot {
  justify-content: flex-start;
}

.isUser .msgBody {
  background: var(--chat-user-bg);
  border-color: var(--chat-user-border);
}

.isBot .msgBody {
  background: var(--chat-bot-bg);
  border-color: var(--chat-bot-border);
}

.composer {
  border-top: 1px solid var(--border-main);
  padding: 10px;
  background: var(--bg-panel);
}

.composerRow {
  display: flex;
  align-items: flex-end;
  gap: 10px;
}

.composer textarea {
  resize: none;
  min-height: 48px;
  width: 100%;
  flex: 1 1 auto;
  background: var(--bg-panel);
  color: var(--text-main);
  border: 1px solid var(--border-main);
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 14px;
  line-height: 1.4;
}

.composerButtons {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
  margin-left: auto;
}

.circleIconBtn {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  border: 1px solid var(--border-main);
  background: var(--bg-panel);
  color: var(--text-main);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.18s ease;
}

.circleIconBtn svg {
  width: 16px;
  height: 16px;
  fill: currentColor;
}

.circleIconBtn:hover {
  transform: translateY(-1px);
  border-color: #b8c6de;
  background: var(--bg-soft);
}

.sendIconBtn {
  border-color: var(--bg-accent);
  background: var(--bg-accent);
  color: var(--text-white);
  box-shadow: 0 6px 14px rgba(47, 124, 246, 0.28);
}

.sendIconBtn:hover {
  background: var(--bg-accent-dark);
  border-color: var(--bg-accent-dark);
}

.sendIconBtn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
}

.historyIconBtn {
  color: var(--text-soft);
}

.stopIconBtn {
  color: var(--danger);
  border-color: #fecaca;
  background: #fff5f5;
}

.composerError {
  margin: 6px 0 0;
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

.historyDrawer {
  width: min(320px, 92vw);
  height: 100dvh;
  padding: 16px;
  border-radius: 0;
  overflow: auto;
}

.historyHead {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.historyActions {
  display: flex;
  gap: 8px;
}

.historyHead h3 {
  margin: 0;
}

.historyItem {
  width: 100%;
  border: 1px solid var(--border-main);
  border-radius: 12px;
  background: var(--bg-panel);
  margin-bottom: 8px;
  padding: 10px;
  text-align: left;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  cursor: pointer;
  transition: all 0.18s ease;
}

.historyItem:hover {
  border-color: #c4d2eb;
  background: var(--bg-soft);
}

.historyItem small {
  color: var(--text-soft);
}

.historyItem.active {
  border-color: var(--bg-accent);
  background: rgba(47, 124, 246, 0.1);
}

@media (max-width: 960px) {
  .chatPage {
    grid-template-columns: 1fr;
    min-height: 0;
    width: 100%;
  }

  .chatPanel {
    height: calc(100dvh - 84px);
    min-height: calc(100dvh - 84px);
    width: 100%;
  }

  .composerRow {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .composerButtons {
    justify-content: flex-end;
    margin-left: 0;
    margin-top: 0;
    gap: 6px;
  }

  .circleIconBtn {
    width: 34px;
    height: 34px;
    border-radius: 10px;
  }

  .circleIconBtn svg {
    width: 14px;
    height: 14px;
  }

  .msgRow {
    gap: 6px;
  }

  .msgBody {
    max-width: 92%;
  }

  .messageList {
    padding: 10px;
  }

  .composer {
    padding: 8px;
  }
}
</style>
