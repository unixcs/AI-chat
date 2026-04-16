<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginUser, getUserProfile } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { storeDraftSessionFlag, storeFreshChatFlag } from '../../utils/chat-entry'
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
    storeFreshChatFlag()
    storeDraftSessionFlag()
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
  <section class="authShell pageWrap">
    <div class="authBackdrop"></div>

    <div class="authGrid contentContainer">
      <article class="authSurface card" :class="'panelShell'">
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

        <p v-if="errorText" class="dangerText authError">{{ errorText }}</p>

        <button class="primaryBtn fullBtn" :disabled="loading" @click="submitLogin">
          {{ loading ? '登录中...' : '登录' }}
        </button>

        <div class="authActions">
          <router-link to="/register">没有账号？去注册</router-link>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.authShell {
  position: relative;
  display: grid;
  place-items: center;
  padding: 28px;
  overflow: hidden;
}

.authBackdrop {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(135deg, rgba(31, 38, 48, 0.56), rgba(82, 91, 105, 0.34)),
    url('/assets/login-bg.png') center / cover no-repeat;
  filter: blur(1px) saturate(0.88);
  transform: scale(1.03);
}

.authBackdrop::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.28), transparent 30%),
    rgba(19, 24, 31, 0.18);
  backdrop-filter: blur(10px);
}

.authGrid {
  position: relative;
  z-index: 1;
  width: min(1220px, 100%);
  display: flex;
  justify-content: center;
  align-items: center;
}

.authSurface {
  padding: 28px;
  width: min(440px, 100%);
  margin: 0 auto;
}

.authTitle {
  text-align: center;
  margin-top: 8px;
  margin-bottom: 6px;
}

.authDesc {
  margin: 0 0 22px;
  color: var(--text-soft);
  font-size: 13px;
  line-height: 1.65;
  text-align: center;
}

.authError {
  margin: 0 0 12px;
}

.fullBtn {
  width: 100%;
}

.authActions {
  margin-top: 16px;
  text-align: right;
}

.authActions a {
  color: var(--accent-strong);
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 960px) {
  .authGrid {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .authShell {
    padding: 14px;
  }

  .authSurface {
    padding: 20px;
  }
}
</style>
