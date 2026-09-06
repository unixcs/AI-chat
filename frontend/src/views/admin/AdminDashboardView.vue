<script setup>
import { computed, onMounted, ref } from 'vue'
import { getAdminDashboard } from '../../api/admin'

const metrics = ref({
  userCount: 0,
  memberCount: 0,
  todayRedeemCount: 0,
  conversationCount: 0
})

const metricCards = computed(() => [
  { label: '用户总数', value: metrics.value.userCount, hint: '平台已注册用户', tone: 'soft' },
  { label: '会员用户', value: metrics.value.memberCount, hint: '当前具有会员身份', tone: 'accent' },
  { label: '今日兑换', value: metrics.value.todayRedeemCount, hint: '今天已完成兑换', tone: 'warm' },
  { label: '会话总数', value: metrics.value.conversationCount, hint: '累计沉淀会话数据', tone: 'deep' }
])

onMounted(async () => {
  const { data } = await getAdminDashboard()
  metrics.value = data.data
})

// ---------- 视图层胶水（仅重写模板新增，业务逻辑未动） ----------
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import { Crown, Gift, LayoutDashboard, MessagesSquare, Users } from 'lucide-vue-next'

// 旧版按 tone 区分底色，映射为 tailwind 语义类（视图层代理）
const toneMeta = {
  soft: { icon: Users, cardClass: '', iconClass: 'bg-primary/10 text-primary' },
  accent: { icon: Crown, cardClass: 'border-primary/25 bg-primary/5', iconClass: 'bg-primary/15 text-primary' },
  warm: { icon: Gift, cardClass: 'border-warning/30 bg-warning/10', iconClass: 'bg-warning/15 text-warning' },
  deep: { icon: MessagesSquare, cardClass: 'border-info/30 bg-info/10', iconClass: 'bg-info/15 text-info' }
}
const toneOf = (tone) => toneMeta[tone] || toneMeta.soft
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-4">
    <Card>
      <CardContent class="flex flex-wrap items-end justify-between gap-3 p-5 sm:p-6">
        <div>
          <small class="text-xs font-medium uppercase tracking-widest text-faint">Overview</small>
          <h2 class="mt-1.5 flex items-center gap-2 text-lg font-semibold text-card-foreground">
            <LayoutDashboard class="size-5 text-primary" />
            后台控制台
          </h2>
        </div>
        <p class="m-0 whitespace-nowrap text-[13px] text-muted-foreground">共 4 项核心指标</p>
      </CardContent>
    </Card>

    <div class="grid gap-3.5 sm:grid-cols-2 xl:grid-cols-4">
      <Card v-for="item in metricCards" :key="item.label" :class="toneOf(item.tone).cardClass">
        <CardContent class="grid gap-1.5 p-5">
          <div class="flex items-center justify-between gap-2">
            <p class="m-0 text-[13px] text-muted-foreground">{{ item.label }}</p>
            <span class="flex size-8 items-center justify-center rounded-lg" :class="toneOf(item.tone).iconClass">
              <component :is="toneOf(item.tone).icon" class="size-4" />
            </span>
          </div>
          <h3 class="m-0 text-3xl font-bold tabular-nums text-card-foreground sm:text-4xl">{{ item.value }}</h3>
          <small class="text-xs text-faint">{{ item.hint }}</small>
        </CardContent>
      </Card>
    </div>
  </section>
</template>
