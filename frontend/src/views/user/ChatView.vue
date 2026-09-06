<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { useChatStore } from '../../stores/chat'
import { useAuthStore } from '../../stores/auth'
import { consumeFreshChatFlag, hasDraftSessionFlag, shouldStartFreshOnChatEntry } from '../../utils/chat-entry'
import { renderMarkdownToSafeHtml } from '../../utils/markdown'
import { copyText } from '../../utils/clipboard'
import { updatePreferences, getCurrentAnnouncement, ackAnnouncement } from '../../api/user-extras'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import Popover from '@/components/ui/popover/Popover.vue'
import PopoverTrigger from '@/components/ui/popover/PopoverTrigger.vue'
import PopoverContent from '@/components/ui/popover/PopoverContent.vue'
import Sheet from '@/components/ui/sheet/Sheet.vue'
import SheetContent from '@/components/ui/sheet/SheetContent.vue'
import SheetHeader from '@/components/ui/sheet/SheetHeader.vue'
import SheetTitle from '@/components/ui/sheet/SheetTitle.vue'
import AlertDialog from '@/components/ui/alert-dialog/AlertDialog.vue'
import AlertDialogContent from '@/components/ui/alert-dialog/AlertDialogContent.vue'
import AlertDialogHeader from '@/components/ui/alert-dialog/AlertDialogHeader.vue'
import AlertDialogTitle from '@/components/ui/alert-dialog/AlertDialogTitle.vue'
import AlertDialogDescription from '@/components/ui/alert-dialog/AlertDialogDescription.vue'
import AlertDialogFooter from '@/components/ui/alert-dialog/AlertDialogFooter.vue'
import AlertDialogAction from '@/components/ui/alert-dialog/AlertDialogAction.vue'
import AlertDialogCancel from '@/components/ui/alert-dialog/AlertDialogCancel.vue'
import Dialog from '@/components/ui/dialog/Dialog.vue'
import DialogContent from '@/components/ui/dialog/DialogContent.vue'
import DialogHeader from '@/components/ui/dialog/DialogHeader.vue'
import DialogTitle from '@/components/ui/dialog/DialogTitle.vue'
import DialogFooter from '@/components/ui/dialog/DialogFooter.vue'
import { Sparkles, Copy, Check, SlidersHorizontal, History, Square, Send, Plus, Trash2 } from 'lucide-vue-next'

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

// 回答模式（长度 × 风格 × 格式）
const showAnswerPrefs = ref(false)
const answerLength = ref(authStore.profile?.answerLength || 'standard')
const answerStyle = ref(authStore.profile?.answerStyle || 'standard')
const answerFormat = ref(authStore.profile?.answerFormat || 'standard')
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
const answerFormatOptions = [
  { value: 'standard', label: '标准' },
  { value: 'plain', label: '纯文字' }
]

// 一次性公告
const announcement = ref(null)
const announcementTitle = ref('')

// 删除对话的确认弹层（Plan §3.3：删除 AlertDialog）
const pendingDeleteId = ref(null)

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

const confirmRemoveConversation = async () => {
  const id = pendingDeleteId.value
  pendingDeleteId.value = null
  if (id) {
    await removeConversation(id)
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

const setAnswerFormat = async (value) => {
  answerFormat.value = value
  await persistPreferences()
}

const persistPreferences = async () => {
  try {
    await updatePreferences({ answerLength: answerLength.value, answerStyle: answerStyle.value, answerFormat: answerFormat.value })
    if (authStore.profile) {
      authStore.setProfile({
        ...authStore.profile,
        answerLength: answerLength.value,
        answerStyle: answerStyle.value,
        answerFormat: answerFormat.value
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
  // 旧账号 answerFormat 为 null → 回落 standard
  answerFormat.value = p.answerFormat || 'standard'
})

// 输出跟随时持续滚到底部；用户上滑后不打扰
watch(lastMessageLength, async () => {
  if (chatStore.streaming && isNearBottom.value) {
    await scrollToBottom(true)
  }
})
</script>

<template>
  <section class="flex h-full min-h-0 flex-col md:px-5 md:py-4">
    <div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-none border-border bg-card/45 backdrop-blur-xl md:rounded-2xl md:border md:shadow-[var(--shadow-elevated)]">
      <div ref="messageListRef" class="min-h-0 flex-1 overflow-y-auto px-3.5 py-5 sm:px-6" @scroll="onMessageScroll">
        <div v-if="activeMessages.length === 0" class="grid min-h-full place-content-center justify-items-center gap-2 px-6 py-12 text-center">
          <span class="mb-2 inline-flex size-14 items-center justify-center rounded-2xl bg-accent text-accent-foreground">
            <Sparkles class="size-7" />
          </span>
          <h2 class="m-0 text-2xl font-bold text-foreground sm:text-3xl">告诉我你有什么想法</h2>
          <p class="text-sm text-muted-foreground">开始新的对话吧</p>
        </div>

        <article
          v-for="msg in activeMessages"
          :key="msg.id"
          class="mb-4.5 flex w-full items-start gap-2.5 sm:gap-3.5"
          :class="msg.role === 'user' ? 'flex-row-reverse justify-start' : 'flex-row'"
        >
          <img
            class="size-9 shrink-0 rounded-xl object-cover shadow-sm sm:size-10"
            :src="msg.role === 'user' ? '/assets/user-avatar.png' : '/assets/ai-avatar.png'"
            :alt="msg.role === 'user' ? '用户头像' : 'AI头像'"
          />

          <div class="grid w-full max-w-full min-w-0 gap-1.5 sm:max-w-[860px]" :class="msg.role === 'user' ? 'justify-items-end' : 'justify-items-start'">
            <span class="text-xs font-bold tracking-wide text-muted-foreground">{{ msg.role === 'user' ? '你' : 'Thallo' }}</span>
            <div
              class="min-w-0 w-fit max-w-full rounded-2xl px-4 py-3 text-[15px] leading-[1.75] shadow-sm sm:max-w-[min(82%,760px)]"
              :class="msg.role === 'user' ? 'chatUserBubble' : 'chatBotBubble'"
            >
              <div v-if="msg.role === 'assistant'">
                <span
                  v-if="chatStore.streaming && msg.id === chatStore.streamingAssistantId"
                  class="block break-words whitespace-pre-wrap"
                >{{ msg.content }}</span>
                <div
                  v-else
                  class="markdownBody"
                  v-html="renderAssistantContent(msg.content)"
                />
              </div>
              <p v-else class="m-0 break-words whitespace-pre-wrap">{{ msg.content }}</p>
            </div>
            <div class="flex items-center gap-2.5">
              <time class="text-xs text-faint">{{ formatTime(msg.createdAt) }}</time>
              <button
                class="inline-flex min-h-11 cursor-pointer items-center gap-1.5 rounded-lg border px-3 text-xs font-semibold transition-colors duration-150"
                :class="copiedMessageId === msg.id
                  ? 'border-primary/35 bg-accent text-primary'
                  : 'border-border bg-card/70 text-muted-foreground hover:text-foreground'"
                :aria-label="copiedMessageId === msg.id ? '已复制' : '复制消息'"
              >
                <Check v-if="copiedMessageId === msg.id" class="size-3.5" />
                <Copy v-else class="size-3.5" />
                <span>{{ copiedMessageId === msg.id ? '已复制' : '复制' }}</span>
              </button>
            </div>
          </div>
        </article>

        <div v-if="chatStore.streaming" class="py-3.5 text-center text-xs text-muted-foreground">
          灵感开启中，请稍等片刻...
        </div>
      </div>

      <div class="shrink-0 border-t border-border bg-card/70 px-3 pt-2.5 pb-[max(12px,env(safe-area-inset-bottom))] sm:px-5">
        <Popover :open="showAnswerPrefs" @update:open="showAnswerPrefs = $event">
          <div class="mx-auto max-w-3xl">
            <div class="flex items-end gap-2.5 rounded-2xl border border-border bg-card px-3 py-2.5 shadow-sm transition-colors focus-within:border-primary/50 sm:gap-3.5 sm:px-4">
              <textarea
                ref="composerTextareaRef"
                v-model="inputValue"
                rows="1"
                placeholder="把你此刻最想问的内容写下来..."
                class="composerTextarea"
                @keydown="onComposerKeydown"
              />

              <div class="flex shrink-0 items-center gap-1.5 sm:gap-2">
                <PopoverTrigger as-child>
                  <Button
                    variant="ghost"
                    size="icon"
                    title="回答模式"
                    aria-label="回答模式"
                    :class="showAnswerPrefs ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
                  >
                    <SlidersHorizontal class="size-[18px]" />
                  </Button>
                </PopoverTrigger>
                <Button
                  variant="ghost"
                  size="icon"
                  title="历史对话"
                  aria-label="历史对话"
                  class="hidden text-muted-foreground sm:inline-flex"
                  @click="openHistoryDrawer"
                >
                  <History class="size-[18px]" />
                </Button>
                <Button
                  v-if="chatStore.streaming"
                  variant="outline"
                  size="icon"
                  title="停止生成"
                  aria-label="停止生成"
                  class="border-destructive/30 bg-destructive/10 text-destructive hover:bg-destructive/20"
                  @click="stopGenerating"
                >
                  <Square class="size-4" />
                </Button>
                <Button
                  v-else
                  size="icon"
                  title="发送消息"
                  aria-label="发送消息"
                  class="shadow-md"
                  @click="submitMessage"
                >
                  <Send class="size-4" />
                </Button>
              </div>
            </div>

            <!-- 移动端：历史入口并入操作行下方 -->
            <button
              class="mx-auto mt-2 flex min-h-9 cursor-pointer items-center gap-1.5 rounded-lg px-3 text-xs font-semibold text-muted-foreground transition-colors hover:text-foreground sm:hidden"
              @click="openHistoryDrawer"
            >
              <History class="size-4" />
              历史对话
            </button>
          </div>

          <!-- 偏好弹层（三维度：长度 × 风格 × 格式） -->
          <PopoverContent side="top" align="end" :side-offset="10" class="w-[min(92vw,380px)] rounded-xl">
            <div class="grid gap-3.5">
              <div>
                <p class="mb-2 text-xs font-bold text-muted-foreground">回答长度</p>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="opt in answerLengthOptions"
                    :key="opt.value"
                    class="min-h-9 cursor-pointer rounded-lg border px-3 text-[13px] font-medium transition-colors duration-150"
                    :class="answerLength === opt.value
                      ? 'border-primary/40 bg-accent font-bold text-accent-foreground'
                      : 'border-border bg-card/60 text-muted-foreground hover:bg-muted'"
                    @click="setAnswerLength(opt.value)"
                  >{{ opt.label }}</button>
                </div>
              </div>
              <div>
                <p class="mb-2 text-xs font-bold text-muted-foreground">回答风格</p>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="opt in answerStyleOptions"
                    :key="opt.value"
                    class="min-h-9 cursor-pointer rounded-lg border px-3 text-[13px] font-medium transition-colors duration-150"
                    :class="answerStyle === opt.value
                      ? 'border-primary/40 bg-accent font-bold text-accent-foreground'
                      : 'border-border bg-card/60 text-muted-foreground hover:bg-muted'"
                    @click="setAnswerStyle(opt.value)"
                  >{{ opt.label }}</button>
                </div>
              </div>
              <div>
                <p class="mb-2 text-xs font-bold text-muted-foreground">输出格式</p>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="opt in answerFormatOptions"
                    :key="opt.value"
                    class="min-h-9 cursor-pointer rounded-lg border px-3 text-[13px] font-medium transition-colors duration-150"
                    :class="answerFormat === opt.value
                      ? 'border-primary/40 bg-accent font-bold text-accent-foreground'
                      : 'border-border bg-card/60 text-muted-foreground hover:bg-muted'"
                    @click="setAnswerFormat(opt.value)"
                  >{{ opt.label }}</button>
                </div>
              </div>
              <p class="text-xs text-faint">三个维度独立生效，自由组合，对所有新对话生效。</p>
            </div>
          </PopoverContent>
        </Popover>

        <Alert v-if="errorText" variant="destructive" class="mx-auto mt-2.5 max-w-3xl">{{ errorText }}</Alert>
      </div>
    </div>

    <!-- 历史对话 Sheet（桌面右侧抽屉 / 移动端全屏） -->
    <Sheet :open="showHistoryDrawer" @update:open="showHistoryDrawer = $event">
      <SheetContent side="right" class="w-full gap-0 p-0 sm:max-w-[400px]">
        <SheetHeader class="flex-row items-center justify-between border-b border-border p-5">
          <SheetTitle class="flex items-center gap-2 text-lg">
            <History class="size-5 text-primary" />
            历史对话
          </SheetTitle>
          <Button size="sm" class="min-h-9" @click="addConversation">
            <Plus class="size-4" />
            新建
          </Button>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto p-4">
          <button
            v-for="item in chatStore.list"
            :key="item.id"
            class="group relative mb-2.5 flex min-h-14 w-full cursor-pointer items-center justify-between gap-2.5 rounded-xl border px-4 py-3 text-left transition-colors duration-150"
            :class="chatStore.activeConversationId === item.id
              ? 'border-primary/30 bg-accent'
              : 'border-border bg-card/60 hover:bg-muted'"
            @click="selectConversation(item.id); closeHistoryDrawer()"
          >
            <span class="min-w-0 flex-1 truncate text-sm font-medium text-foreground">{{ item.title }}</span>
            <small class="shrink-0 text-xs text-faint">{{ formatTime(item.updatedAt) }}</small>
            <span
              class="inline-flex size-11 shrink-0 cursor-pointer items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
              title="删除对话"
              aria-label="删除对话"
              role="button"
              tabindex="0"
              @click.stop="pendingDeleteId = item.id"
              @keydown.enter.stop="pendingDeleteId = item.id"
            >
              <Trash2 class="size-4" />
            </span>
          </button>
          <p v-if="chatStore.list.length === 0" class="py-8 text-center text-[13px] text-muted-foreground">还没有对话记录</p>
        </div>
      </SheetContent>
    </Sheet>

    <!-- 删除对话确认 -->
    <AlertDialog :open="pendingDeleteId !== null" @update:open="(v) => { if (!v) pendingDeleteId = null }">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>删除这段对话？</AlertDialogTitle>
          <AlertDialogDescription>删除后对话与其中全部消息不可恢复。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction class="bg-destructive text-destructive-foreground hover:opacity-90" @click="confirmRemoveConversation">
            删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 一次性公告 -->
    <Dialog :open="!!announcement" @update:open="(v) => { if (!v && announcement) acknowledgeAnnouncement() }">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <Badge class="w-fit">Announcement</Badge>
          <DialogTitle>{{ announcement?.title || '系统公告' }}</DialogTitle>
        </DialogHeader>
        <p class="m-0 max-h-[50vh] overflow-y-auto text-sm leading-relaxed break-words whitespace-pre-wrap text-foreground">
          {{ announcement?.content }}
        </p>
        <DialogFooter>
          <Button @click="acknowledgeAnnouncement">我知道了</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
</template>

<style scoped>
.composerTextarea {
  resize: none;
  min-height: 24px;
  width: 100%;
  flex: 1 1 auto;
  background: transparent;
  color: var(--foreground);
  border: none;
  padding: 0;
  font-size: 15px;
  line-height: 1.7;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

@media (min-width: 961px) {
  .composerTextarea {
    min-height: 44px;
    padding-top: 10px;
  }
}

.composerTextarea::selection {
  background: rgba(63, 125, 78, 0.18);
}

.composerTextarea:focus {
  outline: none;
}

.chatUserBubble {
  background: var(--chat-user-bg);
  border: 1px solid var(--chat-user-border);
  color: var(--chat-user-text);
}

.chatUserBubble :deep(*) {
  color: inherit;
}

.chatBotBubble {
  background: var(--chat-bot-bg);
  border: 1px solid var(--chat-bot-border);
  color: var(--foreground);
}

.markdownBody {
  font-size: 15px;
  line-height: 1.72;
  color: inherit;
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
  padding: 2px 6px;
  border-radius: 6px;
}

.markdownBody :deep(pre) {
  overflow-x: auto;
  padding: 14px;
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
