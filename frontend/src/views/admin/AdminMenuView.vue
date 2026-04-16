<script setup>
import { onMounted, ref } from 'vue'
import { getAdminMenus } from '../../api/admin'

const menus = ref([])

onMounted(async () => {
  const { data } = await getAdminMenus()
  menus.value = data.data
})
</script>

<template>
  <section class="card panelShell panel menuPanel">
    <span class="sectionLabel">Navigation</span>
    <h2 class="sectionTitle">菜单管理</h2>
    <div class="tableWrap">
      <table>
        <thead>
          <tr>
            <th>菜单名称</th>
            <th>路径</th>
            <th>分组</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in menus" :key="item.id">
            <td>{{ item.name }}</td>
            <td>{{ item.path }}</td>
            <td>{{ item.group }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 20px;
}

.menuPanel .sectionTitle {
  margin: 14px 0 20px;
}

table td,
table th {
  padding-top: 10px;
  padding-bottom: 10px;
}
</style>
