<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginAdmin } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const errorText = ref('')
const loading = ref(false)

const formState = reactive({
  username: '',
  password: ''
})

const submitLogin = async () => {
  errorText.value = ''
  if (!formState.username || !formState.password) {
    errorText.value = '请输入管理员账号和密码'
    return
  }

  loading.value = true
  try {
    const { data } = await loginAdmin(formState)
    authStore.setAdminToken(data.data.token)
    router.push('/admin')
  } catch (error) {
    errorText.value = error.response?.data?.message || '管理员登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="adminAuthShell pageWrap">
    <div class="adminBackdrop"></div>

    <div class="adminAuthWrap contentContainer">
      <article class="adminAuthIntro">
        <span class="sectionLabel">Admin Console</span>
        <h1>后台管理也应当属于同一套审美系统。</h1>
        <p>保留原有管理权限、筛选、表格和数据操作，仅统一外观层次、布局与控件细节。</p>
      </article>

      <article class="authSurface card panelShell">
        <span class="sectionLabel">System Access</span>
        <h2 class="sectionTitle authTitle">管理后台登录</h2>
        <p class="authDesc">输入管理员账号与密码后进入后台，原登录与权限逻辑保持不变。</p>

        <div class="formItem">
          <label>管理员账号</label>
          <input v-model="formState.username" placeholder="请输入管理员账号" />
        </div>
        <div class="formItem">
          <label>密码</label>
          <input v-model="formState.password" type="password" placeholder="请输入密码" />
        </div>

        <p v-if="errorText" class="dangerText authError">{{ errorText }}</p>

        <button class="primaryBtn fullBtn" :disabled="loading" @click="submitLogin">
          {{ loading ? '登录中...' : '登录后台' }}
        </button>
      </article>
    </div>
  </section>
</template>

<style scoped>
.adminAuthShell {
  position: relative;
  display: grid;
  place-items: center;
  min-height: 100vh;
  padding: 28px;
  overflow: hidden;
}

.adminBackdrop {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.06), transparent 26%),
    linear-gradient(180deg, rgba(22, 27, 33, 0.94) 0%, rgba(17, 20, 25, 0.98) 100%);
}

.adminAuthWrap {
  position: relative;
  z-index: 1;
  width: min(1120px, 100%);
  display: grid;
  grid-template-columns: 1fr minmax(360px, 420px);
  gap: 24px;
  align-items: stretch;
}

.adminAuthIntro {
  padding: 34px;
  border-radius: var(--radius-xl);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: var(--shadow-float);
  backdrop-filter: blur(14px);
  color: #f5f7fa;
  display: grid;
  align-content: end;
  gap: 18px;
}

.adminAuthIntro :deep(.sectionLabel) {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
}

.adminAuthIntro h1 {
  margin: 0;
  font-size: clamp(34px, 4.5vw, 56px);
  line-height: 1.04;
  letter-spacing: -0.03em;
}

.adminAuthIntro p {
  margin: 0;
  color: rgba(255, 255, 255, 0.72);
  line-height: 1.75;
}

.authSurface {
  padding: 28px;
  align-self: center;
}

.authTitle {
  margin-top: 14px;
}

.authDesc {
  margin: 0 0 22px;
  color: var(--text-soft);
  line-height: 1.65;
}

.authError {
  margin: 0 0 12px;
}

.fullBtn {
  width: 100%;
}

@media (max-width: 960px) {
  .adminAuthWrap {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .adminAuthShell {
    padding: 14px;
  }

  .adminAuthIntro,
  .authSurface {
    padding: 20px;
  }
}
</style>
