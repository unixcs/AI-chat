<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminUsersByQuery,
  resetAdminUserPassword,
  updateAdminUser,
  updateAdminUserMemberExpireAt,
  updateAdminUserStatus
} from '../../api/admin'
import {
  getMemberExpireDisplayMeta,
  formatMemberExpireAtForInput,
  formatNowForMemberExpireInput
} from '../../utils/admin-member-expire'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Select from '@/components/ui/select/Select.vue'
import SelectTrigger from '@/components/ui/select/SelectTrigger.vue'
import SelectValue from '@/components/ui/select/SelectValue.vue'
import SelectContent from '@/components/ui/select/SelectContent.vue'
import SelectItem from '@/components/ui/select/SelectItem.vue'
import Dialog from '@/components/ui/dialog/Dialog.vue'
import DialogContent from '@/components/ui/dialog/DialogContent.vue'
import DialogHeader from '@/components/ui/dialog/DialogHeader.vue'
import DialogTitle from '@/components/ui/dialog/DialogTitle.vue'
import DialogDescription from '@/components/ui/dialog/DialogDescription.vue'
import DialogFooter from '@/components/ui/dialog/DialogFooter.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import { toast } from 'vue-sonner'
import { ChevronLeft, ChevronRight, Search, Users } from 'lucide-vue-next'

const users = ref([])
const total = ref(0)
const loading = ref(false)
const noticeText = ref('')
const errorText = ref('')

const queryState = reactive({
  phone: '',
  status: '',
  page: 1,
  pageSize: 8
})

const editState = reactive({
  id: '',
  phone: '',
  nickname: ''
})

const passwordState = reactive({
  userId: '',
  newPassword: ''
})

const memberExpireState = reactive({
  userId: '',
  phone: '',
  nickname: '',
  value: '',
  submitting: false
})

const loadUsers = async () => {
  loading.value = true
  noticeText.value = ''
  try {
    const { data } = await getAdminUsersByQuery(queryState)
    users.value = data.data.items
    total.value = data.data.total
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadUsers()
})

const formatTime = (time) => {
  if (!time) {
    return '-'
  }
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const getMemberExpireMeta = (time) => {
  return getMemberExpireDisplayMeta(time)
}

// 旧工具函数返回 tagClass（tagSuccess/tagWarn/tagDanger），映射到 Badge variant
const memberBadgeVariant = (time) => {
  const tagClass = getMemberExpireMeta(time).tagClass
  if (tagClass === 'tagSuccess') return 'default'
  if (tagClass === 'tagDanger') return 'destructive'
  return 'secondary'
}

const closeMemberExpireDialog = () => {
  memberExpireState.userId = ''
  memberExpireState.phone = ''
  memberExpireState.nickname = ''
  memberExpireState.value = ''
  memberExpireState.submitting = false
}

const openMemberExpireDialog = (user) => {
  memberExpireState.userId = user.id
  memberExpireState.phone = user.phone
  memberExpireState.nickname = user.nickname
  memberExpireState.value = formatMemberExpireAtForInput(user.memberExpireAt)
  errorText.value = ''
}

const fillCurrentMemberExpireAt = () => {
  memberExpireState.value = formatNowForMemberExpireInput()
}

const clearMemberExpireAt = () => {
  memberExpireState.value = ''
}

const nextPage = async () => {
  const maxPage = Math.ceil(total.value / queryState.pageSize)
  if (queryState.page >= maxPage) {
    return
  }
  queryState.page += 1
  await loadUsers()
}

const prevPage = async () => {
  if (queryState.page <= 1) {
    return
  }
  queryState.page -= 1
  await loadUsers()
}

const updateStatus = async (user, status) => {
  await updateAdminUserStatus(user.id, status)
  noticeText.value = '状态更新成功'
  toast.success('状态更新成功')
  await loadUsers()
}

const startEdit = (user) => {
  editState.id = user.id
  editState.phone = user.phone
  editState.nickname = user.nickname
}

const submitEdit = async () => {
  if (!editState.id) {
    return
  }
  await updateAdminUser(editState.id, {
    phone: editState.phone,
    nickname: editState.nickname
  })
  noticeText.value = '用户信息更新成功'
  toast.success('用户信息更新成功')
  editState.id = ''
  editState.phone = ''
  editState.nickname = ''
  await loadUsers()
}

const resetPassword = async (userId) => {
  if (!passwordState.newPassword || passwordState.newPassword.length < 6) {
    noticeText.value = '重置密码至少 6 位'
    return
  }
  await resetAdminUserPassword(userId, passwordState.newPassword)
  noticeText.value = '密码重置成功'
  toast.success('密码重置成功')
  passwordState.userId = ''
  passwordState.newPassword = ''
}

const submitMemberExpireAt = async () => {
  if (!memberExpireState.userId || memberExpireState.submitting) {
    return
  }

  memberExpireState.submitting = true
  errorText.value = ''
  try {
    const value = memberExpireState.value || null
    await updateAdminUserMemberExpireAt(memberExpireState.userId, value)
    noticeText.value = '会员到期时间更新成功'
    toast.success('会员到期时间更新成功')
    closeMemberExpireDialog()
    await loadUsers()
  } catch (error) {
    errorText.value = error.response?.data?.message || '会员到期时间保存失败'
  } finally {
    memberExpireState.submitting = false
  }
}

const editUser = () => users.value.find((u) => u.id === editState.id)
const passwordUser = () => users.value.find((u) => u.id === passwordState.userId)

// reka-ui SelectItem 禁止空字符串 value：用哨兵值映射查询状态（view 层代理，不改业务逻辑）
const STATUS_ALL = '__all__'
const statusFilter = computed({
  get: () => (queryState.status === '' ? STATUS_ALL : queryState.status),
  set: (v) => { queryState.status = v === STATUS_ALL ? '' : v },
})
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <Users class="size-5 text-primary" />
          用户管理
        </h2>

        <form class="mt-4 flex flex-col gap-2.5 sm:flex-row" @submit.prevent="queryState.page = 1; loadUsers()">
          <Input v-model="queryState.phone" placeholder="按手机号筛选" class="sm:max-w-xs" />
          <Select v-model="statusFilter">
            <SelectTrigger class="sm:w-40">
              <SelectValue placeholder="全部状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__all__">全部状态</SelectItem>
              <SelectItem value="active">正常</SelectItem>
              <SelectItem value="disabled">禁用</SelectItem>
            </SelectContent>
          </Select>
          <Button type="submit" variant="outline" class="sm:w-28">
            <Search class="size-4" />
            查询
          </Button>
        </form>

        <Alert v-if="errorText" variant="destructive" class="mt-3">{{ errorText }}</Alert>
        <p v-if="noticeText && !errorText" class="mt-3 mb-0 text-[13px] text-primary">{{ noticeText }}</p>

        <!-- 桌面表格 -->
        <div class="mt-4 hidden overflow-hidden rounded-xl border border-border md:block">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>手机号</TableHead>
                <TableHead>昵称</TableHead>
                <TableHead>角色</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>会员到期</TableHead>
                <TableHead>注册时间</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="loading">
                <TableCell colspan="7" class="text-muted-foreground">加载中...</TableCell>
              </TableRow>
              <TableRow v-for="item in users" :key="item.id">
                <TableCell class="font-medium">{{ item.phone }}</TableCell>
                <TableCell>{{ item.nickname }}</TableCell>
                <TableCell>{{ item.role }}</TableCell>
                <TableCell>
                  <Badge :variant="item.status === 'active' ? 'default' : 'destructive'">
                    {{ item.status === 'active' ? '正常' : '禁用' }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div class="flex flex-col items-start gap-1">
                    <Badge :variant="memberBadgeVariant(item.memberExpireAt)">
                      {{ getMemberExpireMeta(item.memberExpireAt).label }}
                    </Badge>
                    <small class="text-xs text-muted-foreground">{{ getMemberExpireMeta(item.memberExpireAt).formattedTime }}</small>
                  </div>
                </TableCell>
                <TableCell class="text-muted-foreground">{{ formatTime(item.createdAt) }}</TableCell>
                <TableCell>
                  <div class="flex flex-wrap gap-1.5">
                    <Button variant="outline" size="sm" @click="startEdit(item)">编辑</Button>
                    <Button variant="outline" size="sm" @click="openMemberExpireDialog(item)">设置会员</Button>
                    <Button
                      variant="outline"
                      size="sm"
                      @click="updateStatus(item, item.status === 'active' ? 'disabled' : 'active')"
                    >
                      {{ item.status === 'active' ? '禁用' : '启用' }}
                    </Button>
                    <Button variant="outline" size="sm" @click="passwordState.userId = item.id">重置密码</Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        <!-- 移动卡片列表 -->
        <div class="mt-4 grid gap-2.5 md:hidden">
          <p v-if="loading" class="text-sm text-muted-foreground">加载中...</p>
          <Card v-for="item in users" :key="item.id" class="shadow-none">
            <CardContent class="grid gap-2 p-4">
              <div class="flex items-center justify-between gap-2">
                <span class="font-semibold text-foreground">{{ item.phone }}</span>
                <Badge :variant="item.status === 'active' ? 'default' : 'destructive'">
                  {{ item.status === 'active' ? '正常' : '禁用' }}
                </Badge>
              </div>
              <p class="m-0 text-sm text-muted-foreground">{{ item.nickname }} · {{ item.role }}</p>
              <div class="flex items-center gap-2">
                <Badge :variant="memberBadgeVariant(item.memberExpireAt)">
                  会员{{ getMemberExpireMeta(item.memberExpireAt).label }}
                </Badge>
                <small class="text-xs text-muted-foreground">{{ getMemberExpireMeta(item.memberExpireAt).formattedTime }}</small>
              </div>
              <p class="m-0 text-xs text-faint">注册：{{ formatTime(item.createdAt) }}</p>
              <div class="mt-1 grid grid-cols-2 gap-1.5">
                <Button variant="outline" size="sm" @click="startEdit(item)">编辑</Button>
                <Button variant="outline" size="sm" @click="openMemberExpireDialog(item)">设置会员</Button>
                <Button
                  variant="outline"
                  size="sm"
                  @click="updateStatus(item, item.status === 'active' ? 'disabled' : 'active')"
                >
                  {{ item.status === 'active' ? '禁用' : '启用' }}
                </Button>
                <Button variant="outline" size="sm" @click="passwordState.userId = item.id">重置密码</Button>
              </div>
            </CardContent>
          </Card>
        </div>

        <div class="mt-4 flex flex-wrap items-center justify-end gap-2.5">
          <Button variant="outline" size="sm" :disabled="queryState.page <= 1" @click="prevPage">
            <ChevronLeft class="size-4" />
            上一页
          </Button>
          <span class="text-[13px] text-muted-foreground">第 {{ queryState.page }} 页 / 共 {{ Math.max(1, Math.ceil(total / queryState.pageSize)) }} 页</span>
          <Button variant="outline" size="sm" @click="nextPage">
            下一页
            <ChevronRight class="size-4" />
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- 编辑用户 -->
    <Card v-if="editState.id">
      <CardContent class="p-5">
        <h3 class="text-[15px] font-semibold text-card-foreground">编辑用户{{ editUser() ? ` · ${editUser().phone}` : '' }}</h3>
        <form class="mt-3 flex flex-col gap-2.5 sm:flex-row" @submit.prevent="submitEdit">
          <Input v-model="editState.phone" placeholder="手机号" />
          <Input v-model="editState.nickname" placeholder="昵称" />
          <Button type="submit" class="sm:w-24">保存</Button>
        </form>
      </CardContent>
    </Card>

    <!-- 重置密码 -->
    <Card v-if="passwordState.userId">
      <CardContent class="p-5">
        <h3 class="text-[15px] font-semibold text-card-foreground">重置密码{{ passwordUser() ? ` · ${passwordUser().phone}` : '' }}</h3>
        <form class="mt-3 flex flex-col gap-2.5 sm:flex-row" @submit.prevent="resetPassword(passwordState.userId)">
          <Input v-model="passwordState.newPassword" type="password" placeholder="输入新密码" />
          <Button type="submit" class="sm:w-28">确认重置</Button>
          <Button variant="ghost" class="sm:w-20" @click="passwordState.userId = ''">取消</Button>
        </form>
      </CardContent>
    </Card>

    <!-- 设置会员到期时间 -->
    <Dialog :open="!!memberExpireState.userId" @update:open="(v) => { if (!v) closeMemberExpireDialog() }">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>设置会员到期时间</DialogTitle>
          <DialogDescription>{{ memberExpireState.nickname || '未设置昵称' }} / {{ memberExpireState.phone }}</DialogDescription>
        </DialogHeader>

        <div class="grid gap-2">
          <Label for="member-expire-at-input">会员到期时间</Label>
          <input
            id="member-expire-at-input"
            v-model="memberExpireState.value"
            type="datetime-local"
            step="1"
            class="flex h-11 w-full rounded-md border border-input bg-card/70 px-3.5 py-2 text-[15px] text-foreground shadow-sm transition-colors focus-visible:border-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
          />
          <p class="m-0 text-xs text-muted-foreground">支持精确到秒，留空表示清除会员时间</p>
          <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>
        </div>

        <DialogFooter class="gap-2 sm:justify-between">
          <div class="flex gap-2">
            <Button variant="outline" size="sm" @click="fillCurrentMemberExpireAt">此刻</Button>
            <Button variant="outline" size="sm" @click="clearMemberExpireAt">清除</Button>
          </div>
          <div class="flex gap-2">
            <Button variant="ghost" size="sm" @click="closeMemberExpireDialog">取消</Button>
            <Button :disabled="memberExpireState.submitting" @click="submitMemberExpireAt">
              {{ memberExpireState.submitting ? '保存中...' : '确定' }}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
</template>
