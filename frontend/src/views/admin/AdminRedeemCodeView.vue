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
</script>

<template>
  <section class="panelWrap">
    <article class="card panel">
      <h2 class="sectionTitle">生成兑换码</h2>
      <div class="formRow">
        <div class="formItem">
          <label>数量</label>
          <input v-model.number="formState.quantity" min="1" max="200" type="number" />
        </div>
        <div class="formItem">
          <label>有效月数</label>
          <input v-model.number="formState.durationMonths" min="1" type="number" />
        </div>
      </div>
      <button class="primaryBtn" @click="createBatchCodes">批量生成</button>
    </article>

    <article class="card panel">
      <h2 class="sectionTitle">兑换码列表</h2>
      <div class="toolbarRow">
        <input v-model="queryState.code" placeholder="按兑换码筛选" />
        <select v-model="queryState.status">
          <option value="">全部状态</option>
          <option value="unused">未使用</option>
          <option value="used">已使用</option>
          <option value="void">已作废</option>
        </select>
        <button class="ghostBtn" @click="queryState.page = 1; loadCodes()">查询</button>
        <button class="ghostBtn" @click="exportAllByFilter">导出筛选结果</button>
      </div>
      <p v-if="noticeText" class="mutedText">{{ noticeText }}</p>
      <div class="tableWrap">
        <table>
          <thead>
            <tr>
              <th>兑换码</th>
              <th>时效(月)</th>
              <th>状态</th>
              <th>使用人</th>
              <th>使用时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in codes" :key="item.id">
              <td>{{ item.code }}</td>
              <td>{{ item.durationMonths }}</td>
              <td>
                <span class="tag" :class="item.status === 'unused' ? 'tagSuccess' : 'tagWarn'">
                  {{ item.status === 'unused' ? '未使用' : '已使用' }}
                </span>
              </td>
              <td>{{ item.usedByPhone || '-' }}</td>
              <td>{{ formatTime(item.usedAt) }}</td>
              <td>
                <button class="ghostBtn" :disabled="item.status !== 'unused'" @click="voidCode(item.id)">
                  作废
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagerRow">
        <button class="ghostBtn" @click="prevPage">上一页</button>
        <span class="mutedText">第 {{ queryState.page }} 页 / 共 {{ maxPage }} 页</span>
        <button class="ghostBtn" @click="nextPage">下一页</button>
      </div>
    </article>
  </section>
</template>

<style scoped>
.panelWrap {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 14px;
}

.panel {
  padding: 16px;
}

.toolbarRow {
  flex-wrap: wrap;
}

.toolbarRow input,
.toolbarRow select {
  flex: 1 1 180px;
  min-width: 0;
}

.toolbarRow .ghostBtn {
  flex: 0 0 auto;
}

.tableWrap table {
  min-width: 860px;
}

.tableWrap .ghostBtn {
  padding: 6px 10px;
  font-size: 12px;
}

.formRow {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

@media (max-width: 980px) {
  .panelWrap {
    grid-template-columns: 1fr;
  }

  .panel {
    padding: 12px;
  }

  .formRow {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .toolbarRow {
    flex-direction: column;
  }

  .toolbarRow .ghostBtn,
  .toolbarRow .primaryBtn {
    width: 100%;
  }

  .pagerRow {
    flex-direction: column;
    align-items: stretch;
    row-gap: 6px;
  }
}
</style>
