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
  <section class="adminAuth pageWrap">
    <div class="card authCard">
      <p class="authSub">System Access</p>
      <h2 class="sectionTitle">管理后台登录</h2>
      <div class="formItem">
        <label>管理员账号</label>
        <input v-model="formState.username" placeholder="请输入管理员账号" />
      </div>
      <div class="formItem">
        <label>密码</label>
        <input v-model="formState.password" type="password" placeholder="请输入密码" />
      </div>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
      <button class="primaryBtn fullBtn" :disabled="loading" @click="submitLogin">
        {{ loading ? '登录中...' : '登录后台' }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.adminAuth {
  display: grid;
  place-items: center;
  padding: 30px 24px;
}

.authCard {
  width: min(430px, 100%);
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

@media (max-width: 640px) {
  .adminAuth {
    padding: 14px;
  }

  .authCard {
    padding: 18px;
  }
}
</style>
