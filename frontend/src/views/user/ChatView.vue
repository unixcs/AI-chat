<script setup>
import { computed, onMounted, ref } from 'vue'
import { watch } from 'vue'
import dayjs from 'dayjs'
import { useChatStore } from '../../stores/chat'
import { useAuthStore } from '../../stores/auth'

const chatStore = useChatStore()
const authStore = useAuthStore()
const inputValue = ref('')
const errorText = ref('')
const showHistoryPanel = ref(false)
const showHistoryDrawer = ref(false)

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

const submitMessage = async () => {
  errorText.value = ''
  const content = inputValue.value.trim()
  if (!content) {
    return
  }

  if (!canChat.value) {
    errorText.value = '会员过期，请续费后使用。'
    return
  }

  inputValue.value = ''
  try {
    await chatStore.postStreamMessage(content)
  } catch (error) {
    errorText.value = error.message || '请联系管理员 ⚠️E0'
  }
}

onMounted(async () => {
  await chatStore.fetchConversations()
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
      <div class="messageList">
        <div v-if="activeMessages.length === 0" class="emptyTips mutedText">
          开始新的对话吧
        </div>

        <article
          v-for="msg in activeMessages"
          :key="msg.id"
          class="msgRow"
          :class="msg.role === 'user' ? 'isUser' : 'isBot'"
        >
          <p>{{ msg.content }}</p>
          <time>{{ formatTime(msg.createdAt) }}</time>
        </article>

        <div v-if="chatStore.streaming" class="typingTips mutedText">AI 正在生成...</div>
      </div>

      <div class="composer">
        <div class="composerRow">
          <textarea v-model="inputValue" rows="2" placeholder="请输入内容..." />
          <div class="composerButtons">
            <button class="ghostBtn" @click="openHistoryDrawer">历史</button>
            <button class="primaryBtn" :disabled="chatStore.streaming" @click="submitMessage">
              {{ chatStore.streaming ? '生成中...' : '发送' }}
            </button>
            <button v-if="chatStore.streaming" class="ghostBtn" @click="stopGenerating">
              停止生成
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
}

.chatPanel {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 28px);
  overflow: hidden;
  position: relative;
}

.messageList {
  flex: 1;
  overflow: auto;
  padding: 18px;
  background: var(--chat-surface);
}

.emptyTips {
  text-align: center;
  margin-top: 60px;
}

.typingTips {
  text-align: center;
  padding: 8px;
  font-size: 13px;
}

.msgRow {
  max-width: 80%;
  border-radius: 12px;
  padding: 11px 13px;
  margin-bottom: 12px;
  border: 1px solid transparent;
  box-shadow: var(--shadow-soft);
}

.msgRow p {
  margin: 0;
  white-space: pre-wrap;
}

.msgRow time {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-soft);
}

.isUser {
  margin-left: auto;
  background: var(--chat-user-bg);
  border-color: var(--chat-user-border);
}

.isBot {
  margin-right: auto;
  background: var(--chat-bot-bg);
  border-color: var(--chat-bot-border);
}

.composer {
  border-top: 1px solid var(--border-main);
  padding: 14px;
  background: var(--bg-panel);
}

.composerRow {
  display: flex;
  align-items: flex-end;
  gap: 10px;
}

.composer textarea {
  resize: none;
  min-height: 56px;
  width: 100%;
  flex: 1 1 auto;
  background: var(--bg-panel);
  color: var(--text-main);
  border: 1px solid var(--border-main);
  border-radius: 10px;
  padding: 10px 12px;
}

.composerButtons {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
  margin-left: auto;
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
  height: 100vh;
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
  }

  .chatPanel {
    min-height: 62vh;
  }

  .composerRow {
    flex-direction: column;
    align-items: stretch;
  }

  .composerButtons {
    justify-content: flex-end;
    margin-left: 0;
    margin-top: 6px;
  }
}
</style>
