<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { getAdminRedeemRecordsByQuery } from '../../api/admin'

const records = ref([])
const total = ref(0)

const queryState = reactive({
  phone: '',
  start: '',
  end: '',
  page: 1,
  pageSize: 8
})

const maxPage = computed(() => {
  return Math.max(1, Math.ceil(total.value / queryState.pageSize))
})

const loadRecords = async () => {
  const { data } = await getAdminRedeemRecordsByQuery(queryState)
  records.value = data.data.items
  total.value = data.data.total
}

onMounted(async () => {
  await loadRecords()
})

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

const nextPage = async () => {
  if (queryState.page >= maxPage.value) {
    return
  }
  queryState.page += 1
  await loadRecords()
}

const prevPage = async () => {
  if (queryState.page <= 1) {
    return
  }
  queryState.page -= 1
  await loadRecords()
}

// ================= 视图层胶水（不改动上方业务逻辑） =================
import Button from '@/components/ui/button/Button.vue'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import { ChevronLeft, ChevronRight, ReceiptText, Search } from 'lucide-vue-next'
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <ReceiptText class="size-5 text-primary" />
          兑换记录
        </h2>

        <form class="mt-4 flex flex-col gap-2.5 sm:flex-row sm:flex-wrap sm:items-end" @submit.prevent="queryState.page = 1; loadRecords()">
          <div class="grid flex-1 gap-1.5 sm:max-w-52">
            <Label for="record-phone">手机号</Label>
            <Input id="record-phone" v-model="queryState.phone" placeholder="按手机号筛选" />
          </div>
          <div class="grid flex-1 gap-1.5 sm:max-w-44">
            <Label for="record-start">开始日期</Label>
            <Input id="record-start" v-model="queryState.start" type="date" />
          </div>
          <div class="grid flex-1 gap-1.5 sm:max-w-44">
            <Label for="record-end">结束日期</Label>
            <Input id="record-end" v-model="queryState.end" type="date" />
          </div>
          <Button type="submit" variant="outline" class="sm:w-28">
            <Search class="size-4" />
            查询
          </Button>
        </form>

        <!-- 桌面表格 -->
        <div class="mt-4 hidden overflow-hidden rounded-xl border border-border md:block">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>手机号</TableHead>
                <TableHead>兑换码</TableHead>
                <TableHead>激活时间</TableHead>
                <TableHead>激活前到期</TableHead>
                <TableHead>激活后到期</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in records" :key="item.id">
                <TableCell class="font-medium">{{ item.phone }}</TableCell>
                <TableCell class="font-mono break-all">{{ item.code }}</TableCell>
                <TableCell class="text-muted-foreground">{{ formatTime(item.activatedAt) }}</TableCell>
                <TableCell class="text-muted-foreground">{{ item.beforeExpireAt ? formatTime(item.beforeExpireAt) : '-' }}</TableCell>
                <TableCell>{{ formatTime(item.afterExpireAt) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        <!-- 移动卡片列表 -->
        <div class="mt-4 grid gap-2.5 md:hidden">
          <Card v-for="item in records" :key="item.id" class="shadow-none">
            <CardContent class="grid gap-2 p-4">
              <div class="flex items-center justify-between gap-2">
                <span class="font-semibold text-foreground">{{ item.phone }}</span>
                <span class="font-mono text-[13px] break-all text-muted-foreground">{{ item.code }}</span>
              </div>
              <div class="grid gap-1 text-[13px]">
                <div class="flex items-center justify-between gap-2">
                  <span class="text-muted-foreground">激活时间</span>
                  <span>{{ formatTime(item.activatedAt) }}</span>
                </div>
                <div class="flex items-center justify-between gap-2">
                  <span class="text-muted-foreground">激活前到期</span>
                  <span>{{ item.beforeExpireAt ? formatTime(item.beforeExpireAt) : '-' }}</span>
                </div>
                <div class="flex items-center justify-between gap-2">
                  <span class="text-muted-foreground">激活后到期</span>
                  <span>{{ formatTime(item.afterExpireAt) }}</span>
                </div>
              </div>
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
  </section>
</template>
