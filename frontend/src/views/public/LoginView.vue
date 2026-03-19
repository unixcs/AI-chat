<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginUser, getUserProfile } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { applyTheme } from '../../utils/theme'

const authStore = useAuthStore()
const chatStore = useChatStore()
const router = useRouter()

const formState = reactive({
  phone: '',
  password: ''
})

const errorText = ref('')
const loading = ref(false)

const submitLogin = async () => {
  errorText.value = ''
  if (!formState.phone || !formState.password) {
    errorText.value = '请输入手机号和密码'
    return
  }

  loading.value = true
  try {
    chatStore.resetChatState()
    const { data } = await loginUser(formState)
    authStore.setUserToken(data.data.token)
    applyTheme('light')
    const profileResp = await getUserProfile()
    authStore.setProfile(profileResp.data.data)
    router.push('/app/chat')
  } catch (error) {
    errorText.value = error.response?.data?.message || '登录失败，请检查账号信息'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="authPage pageWrap">
    <div class="card authCard">
      <p class="authSub">欢迎回来</p>
      <h2 class="sectionTitle">用户登录</h2>
      <div class="formItem">
        <label>手机号</label>
        <input v-model="formState.phone" maxlength="11" placeholder="请输入手机号" />
      </div>
      <div class="formItem">
        <label>密码</label>
        <input v-model="formState.password" type="password" placeholder="请输入密码" />
      </div>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
      <button class="primaryBtn fullBtn" :disabled="loading" @click="submitLogin">
        {{ loading ? '登录中...' : '登录' }}
      </button>
      <div class="authActions">
        <router-link to="/register">没有账号？去注册</router-link>
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
  width: min(450px, 100%);
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
