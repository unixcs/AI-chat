<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { getAdminConversationMessages, getAdminConversationsByQuery } from '../../api/admin'
import { renderMarkdownToSafeHtml } from '../../utils/markdown'

const conversations = ref([])
const detailMessages = ref([])
const activeId = ref(null)
const total = ref(0)

const queryState = reactive({
  phone: '',
  keyword: '',
  search: '',
  page: 1,
  pageSize: 8
})

const maxPage = computed(() => {
  return Math.max(1, Math.ceil(total.value / queryState.pageSize))
})

const loadConversations = async () => {
  const { data } = await getAdminConversationsByQuery(queryState)
  conversations.value = data.data.items
  total.value = data.data.total
}

onMounted(async () => {
  await loadConversations()
  if (conversations.value.length > 0) {
    await loadDetail(conversations.value[0].id)
  }
})

const loadDetail = async (conversationId) => {
  activeId.value = conversationId
  const { data } = await getAdminConversationMessages(conversationId)
  detailMessages.value = data.data
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

const renderAssistantContent = (content) => {
  return renderMarkdownToSafeHtml(content)
}

const nextPage = async () => {
  if (queryState.page >= maxPage.value) {
    return
  }
  queryState.page += 1
  await loadConversations()
}

const prevPage = async () => {
  if (queryState.page <= 1) {
    return
  }
  queryState.page -= 1
  await loadConversations()
}
</script>

<template>
  <section class="conversationPanel">
    <article class="conversationHistoryCard" :class="['card', 'panelShell']">
      <span class="sectionLabel">Conversations</span>
      <h2 class="sectionTitle">会话列表</h2>

      <div class="queryCluster">
        <div class="toolbarRow">
          <input v-model="queryState.phone" placeholder="按手机号筛选" />
          <input v-model="queryState.keyword" placeholder="按标题筛选" />
          <input v-model="queryState.search" placeholder="搜索手机号/标题/消息内容" />
          <button class="ghostBtn" @click="queryState.page = 1; loadConversations()">查询</button>
        </div>
      </div>

      <button
        v-for="item in conversations"
        :key="item.id"
        class="historyItem"
        :class="{ active: activeId === item.id }"
        @click="loadDetail(item.id)"
      >
        <div>
          <strong>{{ item.title }}</strong>
          <p class="mutedText">{{ item.userPhone }}</p>
        </div>
        <small>{{ formatTime(item.updatedAt) }}</small>
      </button>

      <div class="pagerRow">
        <button class="ghostBtn" @click="prevPage">上一页</button>
        <span class="mutedText">第 {{ queryState.page }} 页 / 共 {{ maxPage }} 页</span>
        <button class="ghostBtn" @click="nextPage">下一页</button>
      </div>
    </article>

    <article class="transcriptPanel card panelShell">
      <span class="sectionLabel">Transcript</span>
      <h2 class="sectionTitle">对话详情</h2>

      <div class="msgList">
        <article v-for="msg in detailMessages" :key="msg.id" class="transcriptBubble" :class="msg.role">
          <header>
            <strong>{{ msg.role }}</strong>
            <span class="mutedText">{{ formatTime(msg.createdAt) }}</span>
          </header>
          <div v-if="msg.role === 'assistant'" class="markdownBody" v-html="renderAssistantContent(msg.content)" />
          <p v-else>{{ msg.content }}</p>
        </article>
      </div>
    </article>
  </section>
</template>

<style scoped>
.conversationPanel {
  display: grid;
  grid-template-columns: minmax(380px, 430px) minmax(0, 1fr);
  gap: 14px;
}

.conversationHistoryCard,
.transcriptPanel {
  padding: 20px;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.conversationHistoryCard .sectionTitle,
.transcriptPanel .sectionTitle {
  margin: 14px 0 18px;
}

.queryCluster {
  margin-bottom: 14px;
  padding: 14px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.46);
  border: 1px solid var(--line-soft);
}

[data-theme='dark'] .queryCluster {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
}

.historyItem {
  width: 100%;
  border: 1px solid var(--line-soft);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.4);
  margin-bottom: 10px;
  padding: 14px 15px;
  display: flex;
  justify-content: space-between;
  gap: 10px;
  text-align: left;
  cursor: pointer;
  transition: all 0.18s ease;
}

.historyItem strong,
.historyItem p,
.historyItem small {
  word-break: break-word;
  overflow-wrap: anywhere;
}

[data-theme='dark'] .historyItem {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
}

.historyItem:hover {
  border-color: var(--line-strong);
  background: rgba(255, 255, 255, 0.72);
}

[data-theme='dark'] .historyItem:hover {
  background: rgba(255, 255, 255, 0.07);
}

.historyItem.active {
  border-color: rgba(95, 111, 133, 0.26);
  background: rgba(95, 111, 133, 0.12);
}

[data-theme='dark'] .historyItem.active {
  border-color: rgba(163, 178, 198, 0.18);
  background: rgba(163, 178, 198, 0.12);
}

.msgList {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.transcriptBubble {
  border: 1px solid var(--line-soft);
  border-radius: 22px;
  padding: 14px 16px;
  margin-bottom: 12px;
  box-shadow: var(--shadow-soft);
}

.transcriptBubble header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.transcriptBubble.user {
  background: rgba(95, 111, 133, 0.12);
}

.transcriptBubble.assistant {
  background: rgba(255, 255, 255, 0.54);
}

[data-theme='dark'] .transcriptBubble.assistant {
  background: rgba(255, 255, 255, 0.04);
}

.transcriptBubble.system {
  background: rgba(243, 239, 233, 0.88);
}

[data-theme='dark'] .transcriptBubble.system {
  background: rgba(255, 255, 255, 0.025);
}

.transcriptBubble p {
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.7;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.markdownBody {
  line-height: 1.68;
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
  border: 1px solid rgba(148, 163, 184, 0.2);
  text-align: left;
  vertical-align: top;
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.markdownBody :deep(thead th) {
  background: rgba(148, 163, 184, 0.1);
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
  background: rgba(15, 23, 42, 0.08);
  padding: 1px 4px;
  border-radius: 6px;
}

[data-theme='dark'] .markdownBody :deep(code) {
  background: rgba(6, 10, 16, 0.52);
}

.markdownBody :deep(pre) {
  overflow-x: auto;
  padding: 12px;
  border-radius: 14px;
  background: rgba(15, 23, 42, 0.08);
}

[data-theme='dark'] .markdownBody :deep(pre) {
  background: rgba(6, 10, 16, 0.52);
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

@media (max-width: 980px) {
  .conversationPanel {
    grid-template-columns: 1fr;
  }

  .conversationHistoryCard,
  .transcriptPanel {
    padding: 16px;
  }

  .queryCluster {
    padding: 12px;
  }

  .historyItem {
    flex-direction: column;
  }

  .transcriptBubble header {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>
