<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { getAdminAiStatus } from '../../api/admin'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import { Activity, Snowflake, Zap } from 'lucide-vue-next'

const status = ref(null)
const errorText = ref('')
let timer = null

const load = async () => {
  try {
    const { data } = await getAdminAiStatus()
    status.value = data.data
    errorText.value = ''
  } catch (error) {
    errorText.value = error.response?.data?.message || 'AI 状态加载失败'
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 15000)
})

onBeforeUnmount(() => {
  clearInterval(timer)
})

const modeLabel = (mode) => {
  return mode === 'gateway' ? 'Gateway 模式（免费池 → 官方兜底）' : 'Official 模式（仅官方 API）'
}
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <Activity class="size-5 text-primary" />
          AI 状态
        </h2>

        <div v-if="status" class="mt-4">
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
            <div class="rounded-xl border border-primary/25 bg-accent p-4">
              <span class="text-xs text-muted-foreground">运行模式</span>
              <strong class="mt-1 block text-sm font-semibold leading-snug text-foreground">{{ modeLabel(status.mode) }}</strong>
            </div>
            <div class="rounded-xl border border-border bg-muted/40 p-4">
              <span class="text-xs text-muted-foreground">总请求</span>
              <strong class="mt-1 block text-2xl font-bold tabular-nums text-foreground">{{ status.stats.totalRequests }}</strong>
            </div>
            <div class="rounded-xl border border-border bg-muted/40 p-4">
              <span class="text-xs text-muted-foreground">免费池成功</span>
              <strong class="mt-1 block text-2xl font-bold tabular-nums text-foreground">{{ status.stats.gatewaySuccess }}</strong>
            </div>
            <div class="rounded-xl border border-warning/30 bg-warning/10 p-4">
              <span class="text-xs text-muted-foreground">官方兜底次数</span>
              <strong class="mt-1 block text-2xl font-bold tabular-nums text-warning">{{ status.stats.officialFallback }}</strong>
            </div>
            <div class="rounded-xl border border-border bg-muted/40 p-4">
              <span class="text-xs text-muted-foreground">自动切换次数</span>
              <strong class="mt-1 block text-2xl font-bold tabular-nums text-foreground">{{ status.stats.switches }}</strong>
            </div>
          </div>

          <div class="mt-5 overflow-hidden rounded-xl border border-border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>模型入口</TableHead>
                  <TableHead>模型</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>成功</TableHead>
                  <TableHead>失败</TableHead>
                  <TableHead>超时</TableHead>
                  <TableHead>最近首字</TableHead>
                  <TableHead>冷却剩余</TableHead>
                  <TableHead>最近错误</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="m in status.models" :key="m.name">
                  <TableCell class="font-medium">{{ m.name }}</TableCell>
                  <TableCell class="text-muted-foreground">{{ m.model }}</TableCell>
                  <TableCell>
                    <Badge :variant="m.available ? 'default' : 'destructive'" class="gap-1">
                      <Zap v-if="m.available" class="size-3" />
                      <Snowflake v-else class="size-3" />
                      {{ m.available ? '可用' : `冷却中 ${m.cooldownRemainingSecs}s` }}
                    </Badge>
                  </TableCell>
                  <TableCell class="tabular-nums">{{ m.successes }}</TableCell>
                  <TableCell class="tabular-nums">{{ m.failures }}</TableCell>
                  <TableCell class="tabular-nums">{{ m.timeouts }}</TableCell>
                  <TableCell class="tabular-nums">{{ m.lastFirstTokenMs ? `${m.lastFirstTokenMs}ms` : '-' }}</TableCell>
                  <TableCell class="tabular-nums">{{ m.cooldownRemainingSecs ? `${m.cooldownRemainingSecs}s` : '-' }}</TableCell>
                  <TableCell class="max-w-56 text-xs break-all whitespace-normal text-muted-foreground">{{ m.lastError || '-' }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </div>

        <Alert v-if="errorText" variant="destructive" class="mt-3">{{ errorText }}</Alert>
      </CardContent>
    </Card>
  </section>
</template>
