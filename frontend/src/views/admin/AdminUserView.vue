<script setup>
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminUsersByQuery,
  resetAdminUserPassword,
  updateAdminUser,
  updateAdminUserStatus
} from '../../api/admin'

const users = ref([])
const total = ref(0)
const loading = ref(false)
const noticeText = ref('')

const queryState = reactive({
  phone: '',
  status: '',
  page: 1,
  pageSize: 8
})

const editState = reactive({
  id: '',
  phone: '',
  nickname: ''
})

const passwordState = reactive({
  userId: '',
  newPassword: ''
})

const loadUsers = async () => {
  loading.value = true
  noticeText.value = ''
  try {
    const { data } = await getAdminUsersByQuery(queryState)
    users.value = data.data.items
    total.value = data.data.total
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadUsers()
})

const formatTime = (time) => {
  if (!time) {
    return '-'
  }
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

const nextPage = async () => {
  const maxPage = Math.ceil(total.value / queryState.pageSize)
  if (queryState.page >= maxPage) {
    return
  }
  queryState.page += 1
  await loadUsers()
}

const prevPage = async () => {
  if (queryState.page <= 1) {
    return
  }
  queryState.page -= 1
  await loadUsers()
}

const updateStatus = async (user, status) => {
  await updateAdminUserStatus(user.id, status)
  noticeText.value = '状态更新成功'
  await loadUsers()
}

const startEdit = (user) => {
  editState.id = user.id
  editState.phone = user.phone
  editState.nickname = user.nickname
}

const submitEdit = async () => {
  if (!editState.id) {
    return
  }
  await updateAdminUser(editState.id, {
    phone: editState.phone,
    nickname: editState.nickname
  })
  noticeText.value = '用户信息更新成功'
  editState.id = ''
  editState.phone = ''
  editState.nickname = ''
  await loadUsers()
}

const resetPassword = async (userId) => {
  if (!passwordState.newPassword || passwordState.newPassword.length < 6) {
    noticeText.value = '重置密码至少 6 位'
    return
  }
  await resetAdminUserPassword(userId, passwordState.newPassword)
  noticeText.value = '密码重置成功'
  passwordState.userId = ''
  passwordState.newPassword = ''
}
</script>

<template>
  <section class="card panel">
    <h2 class="sectionTitle">用户管理</h2>
    <div class="toolbarRow">
      <input v-model="queryState.phone" placeholder="按手机号筛选" />
      <select v-model="queryState.status">
        <option value="">全部状态</option>
        <option value="active">正常</option>
        <option value="disabled">禁用</option>
      </select>
      <button class="ghostBtn" @click="queryState.page = 1; loadUsers()">查询</button>
    </div>
    <p v-if="noticeText" class="mutedText">{{ noticeText }}</p>
    <div class="tableWrap">
      <table>
        <thead>
          <tr>
            <th>手机号</th>
            <th>昵称</th>
            <th>角色</th>
            <th>状态</th>
            <th>会员到期</th>
            <th>注册时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" class="mutedText">加载中...</td>
          </tr>
          <tr v-for="item in users" :key="item.id">
            <td>{{ item.phone }}</td>
            <td>{{ item.nickname }}</td>
            <td>{{ item.role }}</td>
            <td>
              <span class="tag" :class="item.status === 'active' ? 'tagSuccess' : 'tagDanger'">
                {{ item.status === 'active' ? '正常' : '禁用' }}
              </span>
            </td>
            <td>{{ formatTime(item.memberExpireAt) }}</td>
            <td>{{ formatTime(item.createdAt) }}</td>
            <td>
              <div class="actionGroup">
                <button class="ghostBtn" @click="startEdit(item)">编辑</button>
                <button
                  class="ghostBtn"
                  @click="updateStatus(item, item.status === 'active' ? 'disabled' : 'active')"
                >
                  {{ item.status === 'active' ? '禁用' : '启用' }}
                </button>
                <button class="ghostBtn" @click="passwordState.userId = item.id">重置密码</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagerRow">
      <button class="ghostBtn" @click="prevPage">上一页</button>
      <span class="mutedText">第 {{ queryState.page }} 页 / 共 {{ Math.max(1, Math.ceil(total / queryState.pageSize)) }} 页</span>
      <button class="ghostBtn" @click="nextPage">下一页</button>
    </div>

    <section v-if="editState.id" class="editBox card">
      <h3>编辑用户</h3>
      <div class="toolbarRow">
        <input v-model="editState.phone" placeholder="手机号" />
        <input v-model="editState.nickname" placeholder="昵称" />
        <button class="primaryBtn" @click="submitEdit">保存</button>
      </div>
    </section>

    <section v-if="passwordState.userId" class="editBox card">
      <h3>重置密码</h3>
      <div class="toolbarRow">
        <input v-model="passwordState.newPassword" type="password" placeholder="输入新密码" />
        <button class="primaryBtn" @click="resetPassword(passwordState.userId)">确认重置</button>
        <button class="ghostBtn" @click="passwordState.userId = ''">取消</button>
      </div>
    </section>
  </section>
</template>

<style scoped>
.panel {
  padding: 16px;
}

.tableWrap table {
  min-width: 980px;
}

.actionGroup {
  display: flex;
  gap: 8px;
}

.actionGroup .ghostBtn {
  padding: 6px 10px;
  font-size: 12px;
}

.editBox {
  margin-top: 12px;
  padding: 12px;
}

.editBox h3 {
  margin: 0 0 10px;
}

@media (max-width: 980px) {
  .panel {
    padding: 12px;
  }

  .toolbarRow {
    flex-direction: column;
  }

  .toolbarRow .ghostBtn,
  .toolbarRow .primaryBtn {
    width: 100%;
  }

  .actionGroup {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px;
  }

  .actionGroup .ghostBtn {
    width: 100%;
    padding: 7px 6px;
  }

  .pagerRow {
    flex-direction: column;
    align-items: stretch;
    row-gap: 6px;
  }

  .editBox {
    padding: 10px;
  }
}
</style>
