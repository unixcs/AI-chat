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
  <section class="card settingCard">
    <h2 class="sectionTitle">修改密码</h2>
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
    <p v-if="errorText" class="dangerText">{{ errorText }}</p>
    <button class="primaryBtn" @click="submitChangePassword">确认修改</button>
  </section>
</template>

<style scoped>
.settingCard {
  padding: 20px;
}
</style>
