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
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import Textarea from '@/components/ui/textarea/Textarea.vue'
import AlertDialog from '@/components/ui/alert-dialog/AlertDialog.vue'
import AlertDialogContent from '@/components/ui/alert-dialog/AlertDialogContent.vue'
import AlertDialogHeader from '@/components/ui/alert-dialog/AlertDialogHeader.vue'
import AlertDialogTitle from '@/components/ui/alert-dialog/AlertDialogTitle.vue'
import AlertDialogDescription from '@/components/ui/alert-dialog/AlertDialogDescription.vue'
import AlertDialogFooter from '@/components/ui/alert-dialog/AlertDialogFooter.vue'
import AlertDialogAction from '@/components/ui/alert-dialog/AlertDialogAction.vue'
import AlertDialogCancel from '@/components/ui/alert-dialog/AlertDialogCancel.vue'
import { toast } from 'vue-sonner'
import { History, RotateCcw, Save, SlidersHorizontal } from 'lucide-vue-next'

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
    toast.success(`「${card.label}」卡片已保存`)
  } catch (error) {
    prefError.value = error.response?.data?.message || '保存失败'
    toast.error(prefError.value)
  } finally {
    prefSavingKey.value = ''
  }
}

const resetPrefCard = async (card) => {
  prefSavingKey.value = prefKey(card)
  prefError.value = ''
  try {
    await resetAdminPrefPrompt(card.dimension, card.value)
    const rev = await getAdminPrefPrompts()
    prefCards.value = (rev.data.data || []).map((c) => ({ ...c, draft: c.content }))
    prefDirty.value = {}
    prefNotice.value = `「${card.label}」已恢复预置文案`
    toast.success(`「${card.label}」已恢复预置文案`)
  } catch (error) {
    prefError.value = error.response?.data?.message || '恢复失败'
    toast.error(prefError.value)
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
    toast.success(noticeText.value)
    const rev = await getAdminPromptRevisions()
    revisions.value = rev.data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '保存失败'
    toast.error(errorText.value)
  } finally {
    saving.value = false
  }
}

const restore = async (item) => {
  errorText.value = ''
  noticeText.value = ''
  try {
    const { data } = await restoreAdminPrompt(item.version)
    info.value = data.data
    draft.value = data.data.content
    noticeText.value = `已恢复 v${item.version} 的内容为新版本 v${data.data.version}`
    toast.success(noticeText.value)
    const rev = await getAdminPromptRevisions()
    revisions.value = rev.data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '恢复失败'
    toast.error(errorText.value)
  }
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const preview = (content) => {
  const oneLine = content.replace(/\s+/g, ' ').trim()
  return oneLine.length > 80 ? oneLine.slice(0, 80) + '…' : oneLine
}

// ---------- 确认弹层（替代 window.confirm） ----------
const pendingRestoreRevision = ref(null)
const pendingResetCard = ref(null)

const confirmRestoreRevision = () => {
  const item = pendingRestoreRevision.value
  pendingRestoreRevision.value = null
  if (item) {
    restore(item)
  }
}

const confirmResetCard = () => {
  const card = pendingResetCard.value
  pendingResetCard.value = null
  if (card) {
    resetPrefCard(card)
  }
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-5">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <div class="flex flex-wrap items-center gap-2.5">
          <Badge>内置提示词</Badge>
          <Badge :variant="info.source === 'db' ? 'default' : 'outline'">
            {{ info.source === 'db' ? `SQLite · v${info.version}` : '环境/出厂配置' }}
          </Badge>
          <small class="text-xs text-muted-foreground">
            最后修改：{{ formatTime(info.updatedAt) }}
            <template v-if="info.operator"> · 操作人：{{ info.operator }}</template>
          </small>
        </div>
        <p class="mt-3 text-[13px] leading-relaxed text-muted-foreground">
          修改保存后无需重启，下一个新聊天立即使用最新提示词；生成进行中的请求沿用开始时的版本。
          用户自己的「回答长度 / 回答风格 / 输出格式」会叠加在这段提示词之后，互不冲突。
        </p>

        <div class="mt-4 grid gap-3">
          <Textarea v-model="draft" rows="12" spellcheck="false" placeholder="输入系统内置提示词…" class="min-h-64 font-mono text-sm" />
          <div class="flex flex-wrap items-center gap-2.5">
            <span class="text-xs text-muted-foreground">{{ draftBytes }} / 20000 字节</span>
            <span class="flex-1"></span>
            <Button v-if="dirty" variant="ghost" class="min-h-9" @click="draft = info.content">放弃修改</Button>
            <Button class="min-h-9" :disabled="!dirty || saving" @click="save">
              <Save class="size-4" />
              {{ saving ? '保存中...' : '保存并生效' }}
            </Button>
          </div>
          <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>
          <p v-if="noticeText" class="m-0 text-[13px] text-primary">{{ noticeText }}</p>
          <p v-if="loading" class="m-0 text-[13px] text-muted-foreground">加载中...</p>
        </div>
      </CardContent>
    </Card>

    <!-- 偏好提示词卡片 -->
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h3 class="flex items-center gap-2 text-[15px] font-semibold text-card-foreground">
          <SlidersHorizontal class="size-4 text-primary" />
          偏好提示词卡片
        </h3>
        <p class="mt-2 text-[13px] leading-relaxed text-muted-foreground">
          用户在聊天里选择「回答长度 / 回答风格 / 输出格式」后，实际发给 AI 的指令 =
          上方基础内置 Prompt + 对应的三张卡片文案（按长度 → 风格 → 格式顺序拼接）。
          卡片已预置默认文案，修改保存后立即生效，无需重启；进行中的请求保持开始时的版本。
        </p>

        <p v-if="prefLoading" class="mt-3 text-[13px] text-muted-foreground">加载中...</p>

        <div v-for="group in prefGroups" :key="group.dimension" class="mt-5">
          <h4 class="mb-2.5 text-sm font-semibold text-muted-foreground">{{ group.title }}</h4>
          <div class="grid gap-3">
            <div v-for="card in group.cards" :key="prefKey(card)" class="grid gap-2 rounded-xl border border-border bg-muted/40 p-3.5">
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-[13px] font-semibold text-foreground">{{ card.label }}</span>
                <Badge v-if="card.customized">已自定义</Badge>
                <Badge v-else variant="outline">预置</Badge>
                <span class="flex-1"></span>
                <Button
                  v-if="card.customized"
                  variant="ghost"
                  size="sm"
                  class="text-muted-foreground"
                  :disabled="prefSavingKey === prefKey(card)"
                  @click="pendingResetCard = card"
                >
                  <RotateCcw class="size-3.5" />
                  恢复预置
                </Button>
                <Button
                  size="sm"
                  :disabled="!prefDirty[prefKey(card)] || prefSavingKey === prefKey(card)"
                  @click="savePrefCard(card)"
                >
                  {{ prefSavingKey === prefKey(card) ? '保存中...' : '保存' }}
                </Button>
              </div>
              <Textarea v-model="card.draft" rows="3" spellcheck="false" class="min-h-20 bg-card/70 font-mono text-[13px]" @input="markPrefDirty(card)" />
              <small class="text-xs leading-relaxed text-faint">预置：{{ card.preset }}</small>
            </div>
          </div>
        </div>

        <Alert v-if="prefError" variant="destructive" class="mt-3">{{ prefError }}</Alert>
        <p v-if="prefNotice" class="mt-3 mb-0 text-[13px] text-primary">{{ prefNotice }}</p>
      </CardContent>
    </Card>

    <!-- 版本历史 -->
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h3 class="flex items-center gap-2 text-[15px] font-semibold text-card-foreground">
          <History class="size-4 text-primary" />
          版本历史（最近 50 条）
        </h3>
        <p v-if="revisions.length === 0" class="mt-3 text-[13px] text-muted-foreground">
          还没有保存过版本——当前展示的是环境/出厂配置的内容。
        </p>
        <div class="mt-4 grid gap-2.5">
          <div v-for="item in revisions" :key="item.version" class="rounded-xl border border-border bg-muted/40 p-3.5">
            <div class="flex flex-wrap items-center gap-2.5">
              <span class="text-sm font-bold text-foreground">v{{ item.version }}</span>
              <Badge v-if="item.version === info.version">当前生效</Badge>
              <small class="text-xs text-muted-foreground">{{ formatTime(item.createdAt) }} · {{ item.operator }}</small>
              <span class="flex-1"></span>
              <Button variant="outline" size="sm" class="min-h-9" @click="pendingRestoreRevision = item">恢复此版本</Button>
            </div>
            <p class="mb-0 mt-2 text-[13px] break-words text-muted-foreground" style="overflow-wrap:anywhere">{{ preview(item.content) }}</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 恢复版本确认 -->
    <AlertDialog :open="pendingRestoreRevision !== null" @update:open="(v) => { if (!v) pendingRestoreRevision = null }">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>恢复到 v{{ pendingRestoreRevision?.version }}？</AlertDialogTitle>
          <AlertDialogDescription>将以其内容创建一个新版本（历史保留，可随时再恢复）。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="confirmRestoreRevision">恢复</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 恢复预置确认 -->
    <AlertDialog :open="pendingResetCard !== null" @update:open="(v) => { if (!v) pendingResetCard = null }">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>恢复预置文案？</AlertDialogTitle>
          <AlertDialogDescription>把「{{ pendingResetCard?.label }}」恢复为系统预置文案，当前自定义内容将被移除。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="confirmResetCard">恢复预置</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </section>
</template>
