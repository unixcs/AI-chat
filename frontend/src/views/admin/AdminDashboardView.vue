<script setup>
import { onMounted, ref } from 'vue'
import { getAdminDashboard } from '../../api/admin'

const metrics = ref({
  userCount: 0,
  memberCount: 0,
  todayRedeemCount: 0,
  conversationCount: 0
})

onMounted(async () => {
  const { data } = await getAdminDashboard()
  metrics.value = data.data
})
</script>

<template>
  <section class="dashboardGrid">
    <article class="card metricCard">
      <p>用户总数</p>
      <h3>{{ metrics.userCount }}</h3>
    </article>
    <article class="card metricCard">
      <p>会员用户</p>
      <h3>{{ metrics.memberCount }}</h3>
    </article>
    <article class="card metricCard">
      <p>今日兑换</p>
      <h3>{{ metrics.todayRedeemCount }}</h3>
    </article>
    <article class="card metricCard">
      <p>会话总数</p>
      <h3>{{ metrics.conversationCount }}</h3>
    </article>
  </section>
</template>

<style scoped>
.dashboardGrid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metricCard {
  padding: 20px;
}

.metricCard p {
  margin: 0;
  color: var(--text-soft);
}

.metricCard h3 {
  margin: 12px 0 0;
  font-size: 32px;
}

@media (max-width: 1024px) {
  .dashboardGrid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .dashboardGrid {
    grid-template-columns: 1fr;
  }
}
</style>
