<script setup>
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminAnnouncements,
  createAdminAnnouncement,
  updateAdminAnnouncement,
  deleteAdminAnnouncement
} from '../../api/admin'

const announcements = ref([])
const noticeText = ref('')
const errorText = ref('')
const saving = ref(false)

const form = reactive({
  id: '',
  title: '',
  content: '',
  active: true
})

const editing = ref(false)

const load = async () => {
  try {
    const { data } = await getAdminAnnouncements()
    announcements.value = data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '公告列表加载失败'
  }
}

onMounted(load)

const resetForm = () => {
  form.id = ''
  form.title = ''
  form.content = ''
  form.active = true
  editing.value = false
}

const startEdit = (item) => {
  editing.value = true
  form.id = item.id
  form.title = item.title
  form.content = item.content
  form.active = Boolean(item.active)
}

const submit = async () => {
  errorText.value = ''
  noticeText.value = ''
  if (!form.title.trim() || !form.content.trim()) {
    errorText.value = '标题和内容不能为空'
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateAdminAnnouncement(form.id, {
        title: form.title,
        content: form.content,
        active: form.active
      })
      noticeText.value = '公告已更新'
    } else {
      await createAdminAnnouncement({
        title: form.title,
        content: form.content,
        active: form.active
      })
      noticeText.value = '公告已发布'
    }
    resetForm()
    await load()
  } catch (error) {
    errorText.value = error.response?.data?.message || '保存失败'
  } finally {
    saving.value = false
  }
}

const toggleActive = async (item) => {
  try {
    await updateAdminAnnouncement(item.id, { active: !item.active })
    await load()
  } catch (error) {
    errorText.value = error.response?.data?.message || '操作失败'
  }
}

const remove = async (item) => {
  try {
    await deleteAdminAnnouncement(item.id)
    await load()
  } catch (error) {
    errorText.value = error.response?.data?.message || '删除失败'
  }
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}

// ================= 视图层胶水（不改动上方业务逻辑） =================
// 注：原 remove() 内的 window.confirm 守卫按硬性规则迁移为下方 AlertDialog（pendingRemoveItem）确认，业务逻辑本身未改。
import Alert from '@/components/ui/alert/Alert.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import Button from '@/components/ui/button/Button.vue'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Textarea from '@/components/ui/textarea/Textarea.vue'
import Switch from '@/components/ui/switch/Switch.vue'
import Dialog from '@/components/ui/dialog/Dialog.vue'
import DialogContent from '@/components/ui/dialog/DialogContent.vue'
import DialogHeader from '@/components/ui/dialog/DialogHeader.vue'
import DialogTitle from '@/components/ui/dialog/DialogTitle.vue'
import DialogDescription from '@/components/ui/dialog/DialogDescription.vue'
import DialogFooter from '@/components/ui/dialog/DialogFooter.vue'
import AlertDialog from '@/components/ui/alert-dialog/AlertDialog.vue'
import AlertDialogContent from '@/components/ui/alert-dialog/AlertDialogContent.vue'
import AlertDialogHeader from '@/components/ui/alert-dialog/AlertDialogHeader.vue'
import AlertDialogTitle from '@/components/ui/alert-dialog/AlertDialogTitle.vue'
import AlertDialogDescription from '@/components/ui/alert-dialog/AlertDialogDescription.vue'
import AlertDialogFooter from '@/components/ui/alert-dialog/AlertDialogFooter.vue'
import AlertDialogAction from '@/components/ui/alert-dialog/AlertDialogAction.vue'
import AlertDialogCancel from '@/components/ui/alert-dialog/AlertDialogCancel.vue'
import { toast } from 'vue-sonner'
import { Megaphone, Plus } from 'lucide-vue-next'

// 公告表单弹层（新建/编辑共用）
const dialogOpen = ref(false)
const openCreateDialog = () => {
  resetForm()
  dialogOpen.value = true
}
const openEditDialog = (item) => {
  startEdit(item)
  dialogOpen.value = true
}
const closeDialog = () => {
  dialogOpen.value = false
  resetForm()
}
const handleSubmit = async () => {
  await submit()
  if (noticeText.value) {
    toast.success(noticeText.value)
    closeDialog()
  }
}

// 删除确认（替代 window.confirm）：ref 暂存待确认对象 → AlertDialogAction 里调原函数
const pendingRemoveItem = ref(null)
const confirmRemove = () => {
  const item = pendingRemoveItem.value
  pendingRemoveItem.value = null
  if (item) {
    remove(item)
  }
}

// 公告下线属危险操作：AlertDialog 确认后再调原 toggleActive；启用为安全操作，直接执行
const pendingOfflineItem = ref(null)
const confirmOffline = () => {
  const item = pendingOfflineItem.value
  pendingOfflineItem.value = null
  if (item) {
    toggleActive(item)
  }
}
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <div class="flex flex-wrap items-center justify-between gap-2.5">
          <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
            <Megaphone class="size-5 text-primary" />
            公告管理
          </h2>
          <Button @click="openCreateDialog">
            <Plus class="size-4" />
            发布新公告
          </Button>
        </div>

        <Alert v-if="errorText" variant="destructive" class="mt-3">{{ errorText }}</Alert>
        <p v-if="noticeText && !errorText" class="mt-3 mb-0 text-[13px] text-primary">{{ noticeText }}</p>

        <!-- 桌面表格 -->
        <div class="mt-4 hidden overflow-hidden rounded-xl border border-border md:block">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>标题</TableHead>
                <TableHead>内容</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>已读人数</TableHead>
                <TableHead>发布时间</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in announcements" :key="item.id">
                <TableCell class="font-medium">{{ item.title }}</TableCell>
                <TableCell class="max-w-80 break-words text-[13px] whitespace-normal text-muted-foreground">{{ item.content }}</TableCell>
                <TableCell>
                  <Badge :variant="item.active ? 'default' : 'secondary'">
                    {{ item.active ? '生效中' : '已停用' }}
                  </Badge>
                </TableCell>
                <TableCell>{{ item.readCount ?? 0 }}</TableCell>
                <TableCell class="text-muted-foreground">{{ formatTime(item.createdAt) }}</TableCell>
                <TableCell>
                  <div class="flex flex-wrap gap-1.5">
                    <Button variant="outline" size="sm" @click="openEditDialog(item)">编辑</Button>
                    <Button
                      variant="outline"
                      size="sm"
                      @click="item.active ? (pendingOfflineItem = item) : toggleActive(item)"
                    >
                      {{ item.active ? '停用' : '启用' }}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      class="text-destructive hover:text-destructive"
                      @click="pendingRemoveItem = item"
                    >
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
              <TableRow v-if="announcements.length === 0">
                <TableCell colspan="6" class="text-center text-muted-foreground">还没有公告</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        <!-- 移动卡片列表 -->
        <div class="mt-4 grid gap-2.5 md:hidden">
          <Card v-for="item in announcements" :key="item.id" class="shadow-none">
            <CardContent class="grid gap-2 p-4">
              <div class="flex items-center justify-between gap-2">
                <span class="font-semibold break-words text-foreground">{{ item.title }}</span>
                <Badge :variant="item.active ? 'default' : 'secondary'">
                  {{ item.active ? '生效中' : '已停用' }}
                </Badge>
              </div>
              <p class="m-0 text-[13px] break-words text-muted-foreground">{{ item.content }}</p>
              <p class="m-0 text-xs text-faint">已读 {{ item.readCount ?? 0 }} 人 · 发布：{{ formatTime(item.createdAt) }}</p>
              <div class="mt-1 grid grid-cols-3 gap-1.5">
                <Button variant="outline" size="sm" @click="openEditDialog(item)">编辑</Button>
                <Button
                  variant="outline"
                  size="sm"
                  @click="item.active ? (pendingOfflineItem = item) : toggleActive(item)"
                >
                  {{ item.active ? '停用' : '启用' }}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  class="text-destructive hover:text-destructive"
                  @click="pendingRemoveItem = item"
                >
                  删除
                </Button>
              </div>
            </CardContent>
          </Card>
          <p v-if="announcements.length === 0" class="py-6 text-center text-[13px] text-muted-foreground">还没有公告</p>
        </div>
      </CardContent>
    </Card>

    <!-- 新建/编辑公告弹层 -->
    <Dialog :open="dialogOpen" @update:open="(v) => { if (!v) closeDialog() }">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ editing ? '编辑公告' : '发布新公告' }}</DialogTitle>
          <DialogDescription>公告启用后对用户展示；停用或删除后用户端不再显示。</DialogDescription>
        </DialogHeader>
        <form class="grid gap-3" @submit.prevent="handleSubmit">
          <div class="grid gap-1.5">
            <Label for="announcement-title">标题</Label>
            <Input id="announcement-title" v-model="form.title" placeholder="例如：系统维护通知" />
          </div>
          <div class="grid gap-1.5">
            <Label for="announcement-content">内容</Label>
            <Textarea id="announcement-content" v-model="form.content" rows="5" placeholder="公告正文，用户确认后不再显示" />
          </div>
          <label class="flex cursor-pointer items-center gap-2.5 text-sm text-foreground">
            <Switch v-model="form.active" />
            立即启用
          </label>
          <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>
          <DialogFooter class="gap-2">
            <Button type="button" variant="ghost" @click="closeDialog">取消</Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? '保存中...' : editing ? '保存修改' : '发布公告' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- 删除确认 -->
    <AlertDialog :open="pendingRemoveItem !== null" @update:open="(v) => { if (!v) pendingRemoveItem = null }">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>删除这条公告？</AlertDialogTitle>
          <AlertDialogDescription>确定删除公告「{{ pendingRemoveItem?.title }}」？删除后不可恢复。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="confirmRemove">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 下线确认 -->
    <AlertDialog :open="pendingOfflineItem !== null" @update:open="(v) => { if (!v) pendingOfflineItem = null }">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>下线这条公告？</AlertDialogTitle>
          <AlertDialogDescription>下线「{{ pendingOfflineItem?.title }}」后用户端立即不再显示，可随时重新启用。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="confirmOffline">下线</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </section>
</template>
