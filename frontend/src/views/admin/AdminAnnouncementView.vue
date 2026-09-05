<script setup>
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import {
  getAdminAnnouncements,
  createAdminAnnouncement,
  updateAdminAnnouncement,
  deleteAdminAnnouncement
} from '../../api/admin'

const announcements = ref([])
const noticeText = ref('')
const errorText = ref('')
const saving = ref(false)

const form = reactive({
  id: '',
  title: '',
  content: '',
  active: true
})

const editing = ref(false)

const load = async () => {
  try {
    const { data } = await getAdminAnnouncements()
    announcements.value = data.data || []
  } catch (error) {
    errorText.value = error.response?.data?.message || '公告列表加载失败'
  }
}

onMounted(load)

const resetForm = () => {
  form.id = ''
  form.title = ''
  form.content = ''
  form.active = true
  editing.value = false
}

const startEdit = (item) => {
  editing.value = true
  form.id = item.id
  form.title = item.title
  form.content = item.content
  form.active = Boolean(item.active)
}

const submit = async () => {
  errorText.value = ''
  noticeText.value = ''
  if (!form.title.trim() || !form.content.trim()) {
    errorText.value = '标题和内容不能为空'
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateAdminAnnouncement(form.id, {
        title: form.title,
        content: form.content,
        active: form.active
      })
      noticeText.value = '公告已更新'
    } else {
      await createAdminAnnouncement({
        title: form.title,
        content: form.content,
        active: form.active
      })
      noticeText.value = '公告已发布'
    }
    resetForm()
    await load()
  } catch (error) {
    errorText.value = error.response?.data?.message || '保存失败'
  } finally {
    saving.value = false
  }
}

const toggleActive = async (item) => {
  try {
    await updateAdminAnnouncement(item.id, { active: !item.active })
    await load()
  } catch (error) {
    errorText.value = error.response?.data?.message || '操作失败'
  }
}

const remove = async (item) => {
  if (!window.confirm(`确定删除公告「${item.title}」？`)) {
    return
  }
  try {
    await deleteAdminAnnouncement(item.id)
    await load()
  } catch (error) {
    errorText.value = error.response?.data?.message || '删除失败'
  }
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}
</script>

<template>
  <section class="card panelShell panel announcementPanel">
    <span class="sectionLabel">Announcements</span>
    <h2 class="sectionTitle">公告管理</h2>

    <div class="formCard">
      <div class="formItem">
        <label>标题</label>
        <input v-model="form.title" placeholder="例如：系统维护通知" />
      </div>
      <div class="formItem">
        <label>内容</label>
        <textarea v-model="form.content" rows="4" placeholder="公告正文，用户确认后不再显示"></textarea>
      </div>
      <div class="formRow">
        <label class="checkboxLabel">
          <input v-model="form.active" type="checkbox" />
          立即启用
        </label>
        <div class="formActions">
          <button v-if="editing" class="ghostBtn" @click="resetForm">取消编辑</button>
          <button class="primaryBtn" :disabled="saving" @click="submit">
            {{ editing ? '保存修改' : '发布公告' }}
          </button>
        </div>
      </div>
      <p v-if="noticeText" class="okText">{{ noticeText }}</p>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
    </div>

    <div class="tableWrap">
      <table>
        <thead>
          <tr>
            <th>标题</th>
            <th>内容</th>
            <th>状态</th>
            <th>已读人数</th>
            <th>发布时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in announcements" :key="item.id">
            <td>{{ item.title }}</td>
            <td class="contentCell">{{ item.content }}</td>
            <td>
              <span class="tag" :class="item.active ? 'tagActive' : 'tagOff'">
                {{ item.active ? '生效中' : '已停用' }}
              </span>
            </td>
            <td>{{ item.readCount ?? 0 }}</td>
            <td>{{ formatTime(item.createdAt) }}</td>
            <td>
              <div class="rowActions">
                <button class="ghostBtn" @click="startEdit(item)">编辑</button>
                <button class="ghostBtn" @click="toggleActive(item)">
                  {{ item.active ? '停用' : '启用' }}
                </button>
                <button class="ghostBtn dangerAction" @click="remove(item)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="announcements.length === 0">
            <td colspan="6" class="mutedText emptyRow">还没有公告</td>
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

.announcementPanel .sectionTitle {
  margin: 14px 0 20px;
}

.formCard {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--line-soft);
  border-radius: 18px;
  margin-bottom: 18px;
}

.formItem {
  display: grid;
  gap: 6px;
}

.formItem label {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-soft);
}

.formItem input,
.formItem textarea {
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 10px 12px;
  font-size: 14px;
  background: rgba(255, 255, 255, 0.5);
  color: var(--text-main);
  font-family: inherit;
}

[data-theme='dark'] .formItem input,
[data-theme='dark'] .formItem textarea {
  background: rgba(255, 255, 255, 0.04);
}

.formRow {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.checkboxLabel {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-main);
}

.formActions {
  display: flex;
  gap: 8px;
}

.tableWrap {
  overflow-x: auto;
}

table td,
table th {
  padding-top: 10px;
  padding-bottom: 10px;
}

.contentCell {
  max-width: 320px;
  white-space: normal;
  word-break: break-word;
  font-size: 13px;
}

.tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.tagActive {
  background: rgba(63, 125, 78, 0.12);
  color: #3f7d4e;
}

.tagOff {
  background: rgba(148, 163, 184, 0.16);
  color: var(--text-soft);
}

.rowActions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.dangerAction {
  color: var(--danger, #c65d4b);
}

.emptyRow {
  text-align: center;
}

.okText {
  color: #3f7d4e;
  font-size: 13px;
  margin: 0;
}

.dangerText {
  color: var(--danger, #c65d4b);
  font-size: 13px;
  margin: 0;
}
</style>
