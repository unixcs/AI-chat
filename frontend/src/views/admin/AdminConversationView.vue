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

// ================= 视图层胶水（不改动上方业务逻辑） =================
import Button from '@/components/ui/button/Button.vue'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import { ChevronLeft, ChevronRight, FileText, MessagesSquare, Search } from 'lucide-vue-next'
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <div class="grid items-start gap-4 md:grid-cols-[minmax(320px,400px)_minmax(0,1fr)]">
      <!-- 会话列表 -->
      <Card>
        <CardContent class="flex flex-col gap-3.5 p-5">
          <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
            <MessagesSquare class="size-5 text-primary" />
            会话列表
          </h2>

          <form class="grid gap-2.5" @submit.prevent="queryState.page = 1; loadConversations()">
            <Input v-model="queryState.phone" placeholder="按手机号筛选" />
            <Input v-model="queryState.keyword" placeholder="按标题筛选" />
            <Input v-model="queryState.search" placeholder="搜索手机号/标题/消息内容" />
            <Button type="submit" variant="outline">
              <Search class="size-4" />
              查询
            </Button>
          </form>

          <div class="grid max-h-[60vh] content-start gap-2 overflow-y-auto md:max-h-[calc(100dvh-360px)] md:min-h-40">
            <button
              v-for="item in conversations"
              :key="item.id"
              type="button"
              class="w-full cursor-pointer rounded-xl border p-3.5 text-left transition-colors duration-150"
              :class="activeId === item.id
                ? 'border-primary/40 bg-accent'
                : 'border-border bg-card/60 hover:bg-accent/60'"
              @click="loadDetail(item.id)"
            >
              <span class="block break-words text-sm font-semibold text-foreground">{{ item.title }}</span>
              <span class="mt-1 flex items-center justify-between gap-2">
                <span class="text-[13px] break-all text-muted-foreground">{{ item.userPhone }}</span>
                <small class="shrink-0 text-xs text-faint">{{ formatTime(item.updatedAt) }}</small>
              </span>
            </button>
            <p v-if="conversations.length === 0" class="py-6 text-center text-[13px] text-muted-foreground">没有符合条件的会话</p>
          </div>

          <div class="flex flex-wrap items-center justify-end gap-2.5">
            <Button variant="outline" size="sm" :disabled="queryState.page <= 1" @click="prevPage">
              <ChevronLeft class="size-4" />
              上一页
            </Button>
            <span class="text-[13px] text-muted-foreground">第 {{ queryState.page }} 页 / 共 {{ maxPage }} 页</span>
            <Button variant="outline" size="sm" @click="nextPage">
              下一页
              <ChevronRight class="size-4" />
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- 对话详情 -->
      <Card>
        <CardContent class="flex flex-col gap-3.5 p-5">
          <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
            <FileText class="size-5 text-primary" />
            对话详情
          </h2>

          <div class="grid max-h-[70vh] content-start gap-3 overflow-y-auto pr-1 md:max-h-[calc(100dvh-220px)] md:min-h-40">
            <article
              v-for="msg in detailMessages"
              :key="msg.id"
              class="rounded-2xl border p-4"
              :class="msg.role === 'user'
                ? 'border-primary/20 bg-accent text-accent-foreground'
                : msg.role === 'system'
                  ? 'border-border bg-muted text-muted-foreground'
                  : 'border-border bg-card'"
            >
              <header class="mb-2 flex flex-wrap items-center justify-between gap-2">
                <strong class="text-[13px] font-semibold tracking-wide">{{ msg.role }}</strong>
                <span class="text-xs opacity-70">{{ formatTime(msg.createdAt) }}</span>
              </header>
              <div v-if="msg.role === 'assistant'" class="markdownBody" v-html="renderAssistantContent(msg.content)" />
              <p v-else class="m-0 break-words whitespace-pre-wrap">{{ msg.content }}</p>
            </article>
          </div>
        </CardContent>
      </Card>
    </div>
  </section>
</template>

<style scoped>
/* 助手消息为富文本 markdown，:deep() 样式无法用工具类表达（与 ChatView 同一套路） */
.markdownBody {
  font-size: 15px;
  line-height: 1.7;
  color: inherit;
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
  border: 1px solid var(--border);
  text-align: left;
  vertical-align: top;
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.markdownBody :deep(thead th) {
  background: rgba(127, 145, 130, 0.1);
}

.markdownBody :deep(li + li) {
  margin-top: 4px;
}

.markdownBody :deep(a) {
  color: var(--primary);
  text-decoration: underline;
}

.markdownBody :deep(code) {
  font-family: var(--font-mono);
  font-size: 13px;
  background: rgba(127, 145, 130, 0.12);
  padding: 1px 5px;
  border-radius: 6px;
}

.markdownBody :deep(pre) {
  overflow-x: auto;
  padding: 12px;
  border-radius: 12px;
  background: rgba(127, 145, 130, 0.12);
}

.markdownBody :deep(pre code) {
  background: transparent;
  padding: 0;
  border-radius: 0;
}

.markdownBody :deep(blockquote) {
  padding-left: 12px;
  border-left: 3px solid rgba(63, 125, 78, 0.4);
  color: var(--muted-foreground);
}
</style>
