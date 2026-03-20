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
    <div class="authBackdrop"></div>
    <div class="card authCard">
      <p class="authSub">欢迎登陆Thallo 🔮</p>
      <h2 class="sectionTitle authTitle">欢迎登陆Thallo 🔮</h2>
      <p class="authDesc">使用手机号密码登录以开始解读</p>
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
  position: relative;
  display: grid;
  place-items: center;
  padding: 30px 24px;
  overflow: hidden;
}

.authBackdrop {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(135deg, rgba(16, 24, 38, 0.46), rgba(23, 32, 51, 0.32)),
    url('/assets/login-bg.png') center / cover no-repeat;
  filter: blur(1px) saturate(1.08);
  transform: scale(1.03);
}

.authBackdrop::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(6px);
}

.authCard {
  position: relative;
  z-index: 1;
  width: min(450px, 100%);
  padding: 28px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96) 0%, rgba(247, 249, 253, 0.94) 100%);
  backdrop-filter: blur(8px);
}

.authSub {
  margin: 0;
  color: var(--text-soft);
  font-size: 13px;
  text-align: center;
}

.authTitle {
  text-align: center;
  margin-top: 8px;
  margin-bottom: 6px;
}

.authDesc {
  margin: 0 0 16px;
  color: var(--text-soft);
  font-size: 13px;
  text-align: center;
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
