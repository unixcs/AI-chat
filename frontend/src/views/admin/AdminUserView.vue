<script setup>
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminUsersByQuery,
  resetAdminUserPassword,
  updateAdminUser,
  updateAdminUserMemberExpireAt,
  updateAdminUserStatus
} from '../../api/admin'
import {
  getMemberExpireDisplayMeta,
  formatMemberExpireAtForInput,
  formatNowForMemberExpireInput
} from '../../utils/admin-member-expire'

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

const memberExpireState = reactive({
  userId: '',
  phone: '',
  nickname: '',
  value: '',
  submitting: false
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
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const getMemberExpireMeta = (time) => {
  return getMemberExpireDisplayMeta(time)
}

const closeMemberExpireDialog = () => {
  memberExpireState.userId = ''
  memberExpireState.phone = ''
  memberExpireState.nickname = ''
  memberExpireState.value = ''
  memberExpireState.submitting = false
}

const openMemberExpireDialog = (user) => {
  memberExpireState.userId = user.id
  memberExpireState.phone = user.phone
  memberExpireState.nickname = user.nickname
  memberExpireState.value = formatMemberExpireAtForInput(user.memberExpireAt)
}

const fillCurrentMemberExpireAt = () => {
  memberExpireState.value = formatNowForMemberExpireInput()
}

const clearMemberExpireAt = () => {
  memberExpireState.value = ''
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

const submitMemberExpireAt = async () => {
  if (!memberExpireState.userId || memberExpireState.submitting) {
    return
  }

  memberExpireState.submitting = true
  try {
    const value = memberExpireState.value || null
    await updateAdminUserMemberExpireAt(memberExpireState.userId, value)
    noticeText.value = '会员到期时间更新成功'
    closeMemberExpireDialog()
    await loadUsers()
  } finally {
    memberExpireState.submitting = false
  }
}
</script>

<template>
  <section class="card panelShell panel userPanel">
    <span class="sectionLabel">Users</span>
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
            <td>{{ formatTime(item.createdAt) }}</td>
            <td>
              <div class="actionGroup">
                <button class="ghostBtn" @click="startEdit(item)">编辑</button>
                <button class="ghostBtn" @click="openMemberExpireDialog(item)">设置会员</button>
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

    <div
      v-if="memberExpireState.userId"
      class="memberExpireDialogMask"
      @click.self="closeMemberExpireDialog"
    >
      <section class="memberExpireDialog card">
        <div class="memberExpireDialogHead">
          <div>
            <h3>设置会员到期时间</h3>
            <p class="mutedText">
              {{ memberExpireState.nickname || '未设置昵称' }} / {{ memberExpireState.phone }}
            </p>
          </div>
          <button class="ghostBtn" @click="closeMemberExpireDialog">关闭</button>
        </div>

        <div class="formItem memberExpireField">
          <label for="member-expire-at-input">会员到期时间</label>
          <input
            id="member-expire-at-input"
            v-model="memberExpireState.value"
            type="datetime-local"
            step="1"
          />
          <p class="mutedText memberExpireHelp">支持精确到秒，留空表示清除会员时间</p>
        </div>

        <div class="memberExpireDialogActions">
          <button class="ghostBtn" @click="fillCurrentMemberExpireAt">此刻</button>
          <button class="ghostBtn" @click="clearMemberExpireAt">清除</button>
          <span class="memberExpireDialogSpacer"></span>
          <button class="ghostBtn" @click="closeMemberExpireDialog">取消</button>
          <button
            class="primaryBtn"
            :disabled="memberExpireState.submitting"
            @click="submitMemberExpireAt"
          >
            {{ memberExpireState.submitting ? '保存中...' : '确定' }}
          </button>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 20px;
}

.userPanel .sectionTitle {
  margin: 14px 0 18px;
}

.tableWrap table {
  width: 100%;
}

.tableWrap :deep(th),
.tableWrap :deep(td) {
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
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

.memberExpireDialogMask {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(3, 8, 20, 0.72);
  backdrop-filter: blur(10px);
}

.memberExpireDialog {
  width: min(560px, 100%);
  padding: 20px;
  border: 1px solid rgba(110, 143, 191, 0.22);
  background:
    linear-gradient(180deg, rgba(23, 34, 51, 0.96) 0%, rgba(13, 21, 35, 0.98) 100%),
    var(--bg-panel);
  box-shadow: 0 24px 80px rgba(2, 6, 23, 0.55);
}

.memberExpireDialogHead {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.memberExpireDialogHead h3 {
  margin: 0 0 6px;
  color: #f8fbff;
}

.memberExpireDialogHead .mutedText {
  margin: 0;
  color: #97abc8;
}

.memberExpireField {
  margin-bottom: 18px;
}

.memberExpireField label {
  color: #d8e4f6;
}

.memberExpireField input {
  color-scheme: dark;
  border-color: rgba(114, 140, 180, 0.35);
  background: rgba(13, 20, 33, 0.96);
  color: #f8fbff;
}

.memberExpireField input:focus {
  outline: 2px solid rgba(94, 160, 255, 0.24);
  border-color: #5ea0ff;
}

.memberExpireHelp {
  margin: 8px 0 0;
  color: #8fa6c5;
}

.memberExpireDialogActions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.memberExpireDialogActions .ghostBtn {
  border-color: rgba(114, 140, 180, 0.35);
  background: rgba(14, 22, 37, 0.88);
  color: #e5ecf6;
}

.memberExpireDialogActions .ghostBtn:hover {
  border-color: rgba(148, 176, 221, 0.55);
  background: rgba(24, 36, 57, 0.98);
}

.memberExpireDialogSpacer {
  flex: 1;
}

@media (max-width: 980px) {
  .panel {
    padding: 12px;
  }

  .tableWrap :deep(th),
  .tableWrap :deep(td) {
    padding: 11px 10px;
    font-size: 13px;
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

  .memberExpireDialogMask {
    padding: 12px;
    align-items: flex-end;
  }

  .memberExpireDialog {
    width: 100%;
    padding: 16px;
    border-radius: 16px 16px 12px 12px;
  }

  .memberExpireDialogHead,
  .memberExpireDialogActions {
    flex-direction: column;
    align-items: stretch;
  }

  .memberExpireDialogSpacer {
    display: none;
  }
}

@media (max-width: 640px) {
  .actionGroup {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
