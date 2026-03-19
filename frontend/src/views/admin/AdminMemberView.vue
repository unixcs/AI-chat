<script setup>
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { getAdminMembers } from '../../api/admin'

const members = ref([])

onMounted(async () => {
  const { data } = await getAdminMembers()
  members.value = data.data
})

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}
</script>

<template>
  <section class="card panel">
    <h2 class="sectionTitle">会员管理</h2>
    <div class="tableWrap">
      <table>
        <thead>
          <tr>
            <th>手机号</th>
            <th>昵称</th>
            <th>会员到期</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in members" :key="item.id">
            <td>{{ item.phone }}</td>
            <td>{{ item.nickname }}</td>
            <td>{{ formatTime(item.memberExpireAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 16px;
}

table td,
table th {
  padding-top: 10px;
  padding-bottom: 10px;
}
</style>
