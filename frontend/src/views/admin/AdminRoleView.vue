<script setup>
import { onMounted, reactive, ref } from 'vue'
import { getAdminRoles, updateAdminPassword } from '../../api/admin'

const roles = ref([])
const noticeText = ref('')
const errorText = ref('')

const passwordForm = reactive({
  newPassword: '',
  confirmPassword: ''
})

onMounted(async () => {
  const { data } = await getAdminRoles()
  roles.value = data.data
})

const submitAdminPassword = async () => {
  noticeText.value = ''
  errorText.value = ''
  if (!passwordForm.newPassword || !passwordForm.confirmPassword) {
    errorText.value = '请完整填写密码'
    return
  }
  if (passwordForm.newPassword.length < 6) {
    errorText.value = '新密码至少 6 位'
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    errorText.value = '两次密码不一致'
    return
  }

  await updateAdminPassword(passwordForm.newPassword)
  noticeText.value = '管理员密码修改成功'
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
}
</script>

<template>
  <section class="rolePanelWrap" :class="'rolePage'">
    <article class="roleTablePanel card panelShell">
      <span class="sectionLabel">Roles</span>
      <h2 class="sectionTitle">角色管理</h2>

      <div class="tableWrap">
        <table>
          <thead>
            <tr>
              <th>角色名称</th>
              <th>描述</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in roles" :key="item.id">
              <td>{{ item.name }}</td>
              <td>{{ item.description }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <article class="passwordPanel card panelShell">
      <span class="sectionLabel">Security</span>
      <h2 class="sectionTitle">管理员密码</h2>
      <div class="formItem">
        <label>新密码</label>
        <input v-model="passwordForm.newPassword" type="password" placeholder="请输入新密码" />
      </div>
      <div class="formItem">
        <label>确认密码</label>
        <input v-model="passwordForm.confirmPassword" type="password" placeholder="请再次输入新密码" />
      </div>
      <p v-if="noticeText" class="mutedText successText">{{ noticeText }}</p>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
      <button class="primaryBtn" @click="submitAdminPassword">修改密码</button>
    </article>
  </section>
</template>

<style scoped>
.rolePage,
.rolePanelWrap {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 14px;
}

.roleTablePanel,
.passwordPanel {
  padding: 20px;
}

.roleTablePanel .sectionTitle,
.passwordPanel .sectionTitle {
  margin: 14px 0 20px;
}

.successText {
  color: var(--success);
}

@media (max-width: 980px) {
  .rolePage,
  .rolePanelWrap {
    grid-template-columns: 1fr;
  }
}
</style>
