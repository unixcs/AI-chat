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
</script>

<template>
  <section class="card panel">
    <h2 class="sectionTitle">兑换记录</h2>
    <div class="toolbarRow">
      <input v-model="queryState.phone" placeholder="按手机号筛选" />
      <input v-model="queryState.start" type="date" />
      <input v-model="queryState.end" type="date" />
      <button class="ghostBtn" @click="queryState.page = 1; loadRecords()">查询</button>
    </div>
    <div class="tableWrap">
      <table>
        <thead>
          <tr>
            <th>手机号</th>
            <th>兑换码</th>
            <th>激活时间</th>
            <th>激活前到期</th>
            <th>激活后到期</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in records" :key="item.id">
            <td>{{ item.phone }}</td>
            <td>{{ item.code }}</td>
            <td>{{ formatTime(item.activatedAt) }}</td>
            <td>{{ item.beforeExpireAt ? formatTime(item.beforeExpireAt) : '-' }}</td>
            <td>{{ formatTime(item.afterExpireAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagerRow">
      <button class="ghostBtn" @click="prevPage">上一页</button>
      <span class="mutedText">第 {{ queryState.page }} 页 / 共 {{ maxPage }} 页</span>
      <button class="ghostBtn" @click="nextPage">下一页</button>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 16px;
}
</style>
