<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import {
  createAdminRedeemCodes,
  exportAdminRedeemCodes,
  getAdminRedeemCodesByQuery,
  voidAdminRedeemCode
} from '../../api/admin'

const codes = ref([])
const total = ref(0)
const formState = reactive({
  quantity: 5,
  durationMonths: 1
})

const queryState = reactive({
  code: '',
  status: '',
  page: 1,
  pageSize: 8
})

const noticeText = ref('')

const maxPage = computed(() => {
  return Math.max(1, Math.ceil(total.value / queryState.pageSize))
})

const loadCodes = async () => {
  const { data } = await getAdminRedeemCodesByQuery(queryState)
  codes.value = data.data.items
  total.value = data.data.total
}

const createBatchCodes = async () => {
  await createAdminRedeemCodes(formState)
  queryState.page = 1
  noticeText.value = '兑换码已生成'
  await loadCodes()
}

const voidCode = async (codeId) => {
  await voidAdminRedeemCode(codeId)
  noticeText.value = '兑换码作废成功'
  await loadCodes()
}

const nextPage = async () => {
  if (queryState.page >= maxPage.value) {
    return
  }
  queryState.page += 1
  await loadCodes()
}

const prevPage = async () => {
  if (queryState.page <= 1) {
    return
  }
  queryState.page -= 1
  await loadCodes()
}

const exportCurrentPage = (exportList) => {
  const header = ['兑换码', '时效(月)', '状态', '使用人', '使用时间']
  const lines = exportList.map((item) => {
    return [
      item.code,
      item.durationMonths,
      item.status,
      item.usedByPhone || '',
      item.usedAt ? dayjs(item.usedAt).format('YYYY-MM-DD HH:mm') : ''
    ].join(',')
  })

  const csvContent = [header.join(','), ...lines].join('\n')
  const blob = new Blob([`\uFEFF${csvContent}`], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `兑换码列表_筛选结果_${dayjs().format('YYYYMMDD_HHmmss')}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const exportAllByFilter = async () => {
  const { data } = await exportAdminRedeemCodes({
    code: queryState.code,
    status: queryState.status
  })
  exportCurrentPage(data.data)
}

onMounted(async () => {
  await loadCodes()
})

const formatTime = (time) => {
  if (!time) {
    return '-'
  }
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

// ================= 视图层胶水（不改动上方业务逻辑） =================
import Badge from '@/components/ui/badge/Badge.vue'
import Button from '@/components/ui/button/Button.vue'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Select from '@/components/ui/select/Select.vue'
import SelectTrigger from '@/components/ui/select/SelectTrigger.vue'
import SelectValue from '@/components/ui/select/SelectValue.vue'
import SelectContent from '@/components/ui/select/SelectContent.vue'
import SelectItem from '@/components/ui/select/SelectItem.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import AlertDialog from '@/components/ui/alert-dialog/AlertDialog.vue'
import AlertDialogContent from '@/components/ui/alert-dialog/AlertDialogContent.vue'
import AlertDialogHeader from '@/components/ui/alert-dialog/AlertDialogHeader.vue'
import AlertDialogTitle from '@/components/ui/alert-dialog/AlertDialogTitle.vue'
import AlertDialogDescription from '@/components/ui/alert-dialog/AlertDialogDescription.vue'
import AlertDialogFooter from '@/components/ui/alert-dialog/AlertDialogFooter.vue'
import AlertDialogAction from '@/components/ui/alert-dialog/AlertDialogAction.vue'
import AlertDialogCancel from '@/components/ui/alert-dialog/AlertDialogCancel.vue'
import { toast } from 'vue-sonner'
import { Ban, ChevronLeft, ChevronRight, Download, Search, Ticket, TicketPlus } from 'lucide-vue-next'

// reka-ui SelectItem 禁止空字符串 value：用哨兵值映射"全部状态"（view 层代理，不改业务逻辑）
const STATUS_ALL = '__all__'
const statusFilter = computed({
  get: () => (queryState.status === '' ? STATUS_ALL : queryState.status),
  set: (v) => { queryState.status = v === STATUS_ALL ? '' : v },
})

// 旧模板 tagSuccess/tagWarn → Badge variant（unused=绿，其余=中性黄）
const statusBadgeVariant = (status) => (status === 'unused' ? 'default' : 'secondary')

// 生成成功反馈（失败时与原行为一致：无页内错误提示）
const handleCreate = async () => {
  await createBatchCodes()
  toast.success('兑换码已生成')
}

// 作废是不可逆操作：AlertDialog 确认后再调用原 voidCode
const pendingVoidItem = ref(null)
const confirmVoidCode = async () => {
  const item = pendingVoidItem.value
  pendingVoidItem.value = null
  if (!item) {
    return
  }
  await voidCode(item.id)
  toast.success('兑换码作废成功')
}
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <div class="grid items-start gap-4 lg:grid-cols-[320px_minmax(0,1fr)]">
      <!-- 生成兑换码 -->
      <Card>
        <CardContent class="p-5">
          <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
            <TicketPlus class="size-5 text-primary" />
            生成兑换码
          </h2>
          <form class="mt-4 grid gap-2.5" @submit.prevent="handleCreate">
            <div class="grid gap-1.5">
              <Label for="redeem-quantity">数量</Label>
              <Input
                id="redeem-quantity"
                v-model.number="formState.quantity"
                type="number"
                min="1"
                max="200"
              />
            </div>
            <Button type="submit">批量生成</Button>
          </form>
        </CardContent>
      </Card>

      <!-- 兑换码列表 -->
      <Card>
        <CardContent class="p-5 sm:p-6">
          <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
            <Ticket class="size-5 text-primary" />
            兑换码列表
          </h2>

          <form class="mt-4 flex flex-col gap-2.5 sm:flex-row sm:flex-wrap" @submit.prevent="queryState.page = 1; loadCodes()">
            <Input v-model="queryState.code" placeholder="按兑换码筛选" class="sm:max-w-60 sm:flex-1" />
            <Select v-model="statusFilter">
              <SelectTrigger class="sm:w-36">
                <SelectValue placeholder="全部状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">全部状态</SelectItem>
                <SelectItem value="unused">未使用</SelectItem>
                <SelectItem value="used">已使用</SelectItem>
                <SelectItem value="void">已作废</SelectItem>
              </SelectContent>
            </Select>
            <Button type="submit" variant="outline">
              <Search class="size-4" />
              查询
            </Button>
            <Button type="button" variant="outline" @click="exportAllByFilter">
              <Download class="size-4" />
              导出筛选结果
            </Button>
          </form>

          <p v-if="noticeText" class="mt-3 mb-0 text-[13px] text-primary">{{ noticeText }}</p>

          <!-- 桌面表格 -->
          <div class="mt-4 hidden overflow-hidden rounded-xl border border-border md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>兑换码</TableHead>
                  <TableHead>时效(月)</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>使用人</TableHead>
                  <TableHead>使用时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in codes" :key="item.id">
                  <TableCell class="font-mono break-all">{{ item.code }}</TableCell>
                  <TableCell>{{ item.durationMonths }}</TableCell>
                  <TableCell>
                    <Badge :variant="statusBadgeVariant(item.status)">
                      {{ item.status === 'unused' ? '未使用' : '已使用' }}
                    </Badge>
                  </TableCell>
                  <TableCell>{{ item.usedByPhone || '-' }}</TableCell>
                  <TableCell class="text-muted-foreground">{{ formatTime(item.usedAt) }}</TableCell>
                  <TableCell>
                    <Button
                      variant="outline"
                      size="sm"
                      :disabled="item.status !== 'unused'"
                      @click="pendingVoidItem = item"
                    >
                      <Ban class="size-3.5" />
                      作废
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <!-- 移动卡片列表 -->
          <div class="mt-4 grid gap-2.5 md:hidden">
            <Card v-for="item in codes" :key="item.id" class="shadow-none">
              <CardContent class="grid gap-2 p-4">
                <div class="flex items-center justify-between gap-2">
                  <span class="font-mono text-sm font-semibold break-all text-foreground">{{ item.code }}</span>
                  <Badge :variant="statusBadgeVariant(item.status)">
                    {{ item.status === 'unused' ? '未使用' : '已使用' }}
                  </Badge>
                </div>
                <p class="m-0 text-sm text-muted-foreground">时效 {{ item.durationMonths }} 个月</p>
                <p class="m-0 text-sm text-muted-foreground">使用人：{{ item.usedByPhone || '-' }}</p>
                <p class="m-0 text-xs text-faint">使用时间：{{ formatTime(item.usedAt) }}</p>
                <Button
                  variant="outline"
                  size="sm"
                  class="mt-1"
                  :disabled="item.status !== 'unused'"
                  @click="pendingVoidItem = item"
                >
                  <Ban class="size-3.5" />
                  作废
                </Button>
              </CardContent>
            </Card>
          </div>

          <div class="mt-4 flex flex-wrap items-center justify-end gap-2.5">
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
    </div>

    <!-- 作废确认 -->
    <AlertDialog :open="pendingVoidItem !== null" @update:open="(v) => { if (!v) pendingVoidItem = null }">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>作废这个兑换码？</AlertDialogTitle>
          <AlertDialogDescription>
            「{{ pendingVoidItem?.code }}」作废后将无法被使用，操作不可撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="confirmVoidCode">作废</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </section>
</template>
