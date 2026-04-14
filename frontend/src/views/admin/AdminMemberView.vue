<script setup>
import { onMounted, ref } from 'vue'
import { getAdminMembers } from '../../api/admin'
import { getMemberExpireDisplayMeta } from '../../utils/admin-member-expire'

const members = ref([])

onMounted(async () => {
  const { data } = await getAdminMembers()
  members.value = data.data
})

const getMemberExpireMeta = (time) => {
  return getMemberExpireDisplayMeta(time)
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
            <td>
              <div class="memberExpireCell">
                <span class="tag" :class="getMemberExpireMeta(item.memberExpireAt).tagClass">
                  {{ getMemberExpireMeta(item.memberExpireAt).label }}
                </span>
                <small class="memberExpireText mutedText">
                  {{ getMemberExpireMeta(item.memberExpireAt).formattedTime }}
                </small>
              </div>
            </td>
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

.memberExpireCell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.memberExpireText {
  font-size: 12px;
  line-height: 1.4;
}
</style>
