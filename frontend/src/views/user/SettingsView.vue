<script setup>
import { reactive, ref } from 'vue'
import { changePassword } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const formState = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const errorText = ref('')

const submitChangePassword = async () => {
  errorText.value = ''
  if (!formState.oldPassword || !formState.newPassword || !formState.confirmPassword) {
    errorText.value = '请完整填写密码信息'
    return
  }

  if (formState.newPassword !== formState.confirmPassword) {
    errorText.value = '两次新密码不一致'
    return
  }

  if (formState.newPassword.length < 6) {
    errorText.value = '新密码至少 6 位'
    return
  }

  try {
    await changePassword({
      oldPassword: formState.oldPassword,
      newPassword: formState.newPassword
    })
    authStore.logoutUser()
    router.push('/login')
  } catch (error) {
    errorText.value = error.response?.data?.message || '修改密码失败'
  }
}
</script>

<template>
  <section class="card panelShell settingCard">
    <span class="sectionLabel">Security</span>
    <h2 class="sectionTitle">修改密码</h2>
    <p class="sectionIntro">提交新密码后会按原逻辑退出登录并返回登录页，这里只优化页面布局与表单呈现。</p>

    <div class="formItem">
      <label>旧密码</label>
      <input v-model="formState.oldPassword" type="password" />
    </div>
    <div class="formItem">
      <label>新密码</label>
      <input v-model="formState.newPassword" type="password" />
    </div>
    <div class="formItem">
      <label>确认新密码</label>
      <input v-model="formState.confirmPassword" type="password" />
    </div>
    <p v-if="errorText" class="dangerText errorText">{{ errorText }}</p>
    <button class="primaryBtn" @click="submitChangePassword">确认修改</button>
  </section>
</template>

<style scoped>
.settingCard {
  padding: 24px;
  max-width: 760px;
}

.settingCard .sectionTitle {
  margin-top: 14px;
}

.settingCard .sectionIntro {
  margin-bottom: 22px;
}

.errorText {
  margin: 0 0 12px;
}
</style>
