<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { getAdminConversationMessages, getAdminConversationsByQuery } from '../../api/admin'

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
    <article class="card leftPanel">
      <h2 class="sectionTitle">会话列表</h2>
      <div class="toolbarRow">
        <input v-model="queryState.phone" placeholder="按手机号筛选" />
        <input v-model="queryState.keyword" placeholder="按标题筛选" />
        <input v-model="queryState.search" placeholder="搜索手机号/标题/消息内容" />
        <button class="ghostBtn" @click="queryState.page = 1; loadConversations()">查询</button>
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

    <article class="card rightPanel">
      <h2 class="sectionTitle">对话详情（完整内容）</h2>
      <div class="msgList">
        <article v-for="msg in detailMessages" :key="msg.id" class="msgRow" :class="msg.role">
          <header>
            <strong>{{ msg.role }}</strong>
            <span class="mutedText">{{ formatTime(msg.createdAt) }}</span>
          </header>
          <p>{{ msg.content }}</p>
        </article>
      </div>
    </article>
  </section>
</template>

<style scoped>
.conversationPanel {
  display: grid;
  grid-template-columns: 340px 1fr;
  gap: 14px;
}

.leftPanel,
.rightPanel {
  padding: 16px;
}

.historyItem {
  width: 100%;
  border: 1px solid var(--border-main);
  border-radius: 10px;
  background: var(--bg-panel);
  margin-bottom: 8px;
  padding: 10px;
  display: flex;
  justify-content: space-between;
  text-align: left;
  cursor: pointer;
  transition: all 0.18s ease;
}

.historyItem:hover {
  border-color: #c4d2eb;
  background: var(--bg-soft);
}

.historyItem.active {
  border-color: var(--bg-accent);
  background: rgba(47, 124, 246, 0.08);
}

.msgList {
  max-height: calc(100vh - 170px);
  overflow: auto;
}

.msgRow {
  border: 1px solid var(--border-main);
  border-radius: 10px;
  padding: 10px;
  margin-bottom: 10px;
  box-shadow: var(--shadow-soft);
}

.msgRow header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.msgRow.user {
  background: var(--chat-user-bg);
}

.msgRow.assistant {
  background: var(--chat-bot-bg);
}

.msgRow.system {
  background: var(--bg-soft);
}

.msgRow p {
  margin: 0;
  white-space: pre-wrap;
}

@media (max-width: 980px) {
  .conversationPanel {
    grid-template-columns: 1fr;
  }
}
</style>
