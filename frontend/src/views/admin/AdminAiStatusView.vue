<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { getAdminAiStatus } from '../../api/admin'

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
  <section class="card panelShell panel aiStatusPanel">
    <span class="sectionLabel">AI Status</span>
    <h2 class="sectionTitle">AI 状态</h2>

    <template v-if="status">
      <div class="statsGrid">
        <div class="statCard">
          <span class="statLabel">运行模式</span>
          <strong>{{ modeLabel(status.mode) }}</strong>
        </div>
        <div class="statCard">
          <span class="statLabel">总请求</span>
          <strong>{{ status.stats.totalRequests }}</strong>
        </div>
        <div class="statCard">
          <span class="statLabel">免费池成功</span>
          <strong>{{ status.stats.gatewaySuccess }}</strong>
        </div>
        <div class="statCard">
          <span class="statLabel">官方兜底次数</span>
          <strong>{{ status.stats.officialFallback }}</strong>
        </div>
        <div class="statCard">
          <span class="statLabel">自动切换次数</span>
          <strong>{{ status.stats.switches }}</strong>
        </div>
      </div>

      <div class="tableWrap">
        <table>
          <thead>
            <tr>
              <th>模型入口</th>
              <th>模型</th>
              <th>状态</th>
              <th>成功</th>
              <th>失败</th>
              <th>超时</th>
              <th>最近首字</th>
              <th>冷却剩余</th>
              <th>最近错误</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in status.models" :key="m.name">
              <td>{{ m.name }}</td>
              <td>{{ m.model }}</td>
              <td>
                <span class="tag" :class="m.available ? 'tagOk' : 'tagCool'">
                  {{ m.available ? '可用' : `冷却中 ${m.cooldownRemainingSecs}s` }}
                </span>
              </td>
              <td>{{ m.successes }}</td>
              <td>{{ m.failures }}</td>
              <td>{{ m.timeouts }}</td>
              <td>{{ m.lastFirstTokenMs ? `${m.lastFirstTokenMs}ms` : '-' }}</td>
              <td>{{ m.cooldownRemainingSecs ? `${m.cooldownRemainingSecs}s` : '-' }}</td>
              <td class="errCell">{{ m.lastError || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
    <p v-if="errorText" class="dangerText">{{ errorText }}</p>
  </section>
</template>

<style scoped>
.panel {
  padding: 20px;
}

.aiStatusPanel .sectionTitle {
  margin: 14px 0 20px;
}

.statsGrid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin-bottom: 20px;
}

.statCard {
  display: grid;
  gap: 6px;
  padding: 14px 16px;
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.4);
}

[data-theme='dark'] .statCard {
  background: rgba(255, 255, 255, 0.03);
}

.statLabel {
  font-size: 12px;
  color: var(--text-soft);
}

.tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.tagOk {
  background: rgba(63, 125, 78, 0.12);
  color: #3f7d4e;
}

.tagCool {
  background: rgba(198, 93, 75, 0.1);
  color: var(--danger, #c65d4b);
}

.errCell {
  max-width: 220px;
  word-break: break-all;
  font-size: 12px;
}

.dangerText {
  color: var(--danger, #c65d4b);
  font-size: 13px;
}
</style>
