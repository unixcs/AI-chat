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
  <section class="authShell pageWrap">
    <div class="authBackdrop"></div>

    <div class="authGrid contentContainer">
      <article class="authSurface card" :class="'panelShell'">
        <p class="authSub">欢迎加入</p>
        <h2 class="sectionTitle authTitle">注册账号</h2>

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

        <p v-if="errorText" class="dangerText authError">{{ errorText }}</p>

        <button class="primaryBtn fullBtn" :disabled="loading" @click="submitRegister">
          {{ loading ? '注册中...' : '注册' }}
        </button>

        <div class="authActions">
          <router-link to="/login">已有账号？返回登录</router-link>
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
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.26), transparent 30%),
    rgba(19, 24, 31, 0.2);
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
  width: min(450px, 100%);
  margin: 0 auto;
}

.authSub {
  margin: 0;
  color: var(--text-soft);
  font-size: 13px;
  text-align: center;
}

.authTitle {
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
