<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { registerUser } from '../../api/auth'

const router = useRouter()
const loading = ref(false)
const errorText = ref('')

const formState = reactive({
  phone: '',
  nickname: '',
  password: '',
  confirmPassword: ''
})

const submitRegister = async () => {
  errorText.value = ''
  if (!/^1\d{10}$/.test(formState.phone)) {
    errorText.value = '请输入正确的 11 位手机号'
    return
  }
  if (!formState.nickname) {
    errorText.value = '请输入昵称'
    return
  }
  if (formState.password.length < 6) {
    errorText.value = '密码至少 6 位'
    return
  }
  if (formState.password !== formState.confirmPassword) {
    errorText.value = '两次密码不一致'
    return
  }

  loading.value = true
  try {
    await registerUser({
      phone: formState.phone,
      nickname: formState.nickname,
      password: formState.password
    })
    router.push('/login')
  } catch (error) {
    errorText.value = error.response?.data?.message || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="authPage pageWrap">
    <div class="card authCard">
      <p class="authSub">欢迎加入</p>
      <h2 class="sectionTitle">注册账号</h2>
      <div class="formItem">
        <label>手机号</label>
        <input v-model="formState.phone" maxlength="11" placeholder="请输入手机号" />
      </div>
      <div class="formItem">
        <label>昵称</label>
        <input v-model="formState.nickname" placeholder="请输入昵称" />
      </div>
      <div class="formItem">
        <label>密码</label>
        <input v-model="formState.password" type="password" placeholder="请输入密码" />
      </div>
      <div class="formItem">
        <label>确认密码</label>
        <input v-model="formState.confirmPassword" type="password" placeholder="请再次输入密码" />
      </div>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
      <button class="primaryBtn fullBtn" :disabled="loading" @click="submitRegister">
        {{ loading ? '注册中...' : '注册' }}
      </button>
      <div class="authActions">
        <router-link to="/login">已有账号？返回登录</router-link>
      </div>
    </div>
  </section>
</template>

<style scoped>
.authPage {
  display: grid;
  place-items: center;
  padding: 30px 24px;
}

.authCard {
  width: min(460px, 100%);
  padding: 28px;
  background: linear-gradient(180deg, var(--bg-panel) 0%, var(--bg-soft) 100%);
}

.authSub {
  margin: 0;
  color: var(--text-soft);
  font-size: 13px;
}

.fullBtn {
  width: 100%;
}

.authActions {
  margin-top: 14px;
  text-align: right;
}

.authActions a {
  color: var(--bg-accent);
  font-size: 13px;
}

@media (max-width: 640px) {
  .authPage {
    padding: 14px;
  }

  .authCard {
    padding: 18px;
  }
}
</style>
