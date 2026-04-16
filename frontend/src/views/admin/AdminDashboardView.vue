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
</script>

<template>
  <section class="dashboardPage">
    <article class="dashboardHero card panelShell">
      <div>
        <span class="sectionLabel">Overview</span>
        <h2 class="sectionTitle">后台控制台</h2>
      </div>
      <div class="heroStats mutedText">共 4 项核心指标</div>
    </article>

    <section class="dashboardGrid">
      <article v-for="item in metricCards" :key="item.label" class="metricCard card panelShell" :class="item.tone">
        <p>{{ item.label }}</p>
        <h3>{{ item.value }}</h3>
        <small>{{ item.hint }}</small>
      </article>
    </section>
  </section>
</template>

<style scoped>
.dashboardPage {
  display: grid;
  gap: 16px;
}

.dashboardHero {
  padding: 24px;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 16px;
}

.dashboardHero .sectionTitle {
  margin: 14px 0 0;
}

.heroStats {
  white-space: nowrap;
}

.dashboardGrid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.metricCard {
  padding: 22px;
  display: grid;
  gap: 10px;
}

.metricCard p,
.metricCard small {
  margin: 0;
  color: var(--text-soft);
}

.metricCard h3 {
  margin: 0;
  font-size: clamp(30px, 3vw, 42px);
  color: var(--text-title);
}

.metricCard.accent {
  background: linear-gradient(180deg, rgba(246, 249, 252, 0.96) 0%, rgba(229, 236, 244, 0.86) 100%);
}

.metricCard.warm {
  background: linear-gradient(180deg, rgba(252, 247, 240, 0.96) 0%, rgba(243, 231, 216, 0.82) 100%);
}

.metricCard.deep {
  background: linear-gradient(180deg, rgba(242, 240, 236, 0.96) 0%, rgba(229, 226, 221, 0.86) 100%);
}

@media (max-width: 1024px) {
  .dashboardGrid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .dashboardHero {
    padding: 18px;
    flex-direction: column;
    align-items: flex-start;
  }

  .dashboardGrid {
    grid-template-columns: 1fr;
  }
}
</style>
