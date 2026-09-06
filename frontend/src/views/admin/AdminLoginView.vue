<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginAdmin } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import AuthShell from '@/components/layout/AuthShell.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import { ShieldCheck } from 'lucide-vue-next'

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
  <AuthShell>
    <div class="mb-6 flex flex-col items-center gap-3 text-center">
      <span class="inline-flex size-12 items-center justify-center rounded-2xl bg-primary/25 text-primary-foreground">
        <ShieldCheck class="size-6" />
      </span>
      <h1 class="text-2xl leading-tight font-bold text-white">管理后台登录</h1>
      <p class="text-[13px] text-white/75">输入管理员账号与密码后进入后台，权限逻辑保持不变</p>
    </div>

    <form class="space-y-4" @submit.prevent="submitLogin">
      <div class="space-y-2">
        <Label class="text-white/85" for="admin-username">管理员账号</Label>
        <Input id="admin-username" v-model="formState.username" placeholder="请输入管理员账号" />
      </div>
      <div class="space-y-2">
        <Label class="text-white/85" for="admin-password">密码</Label>
        <Input id="admin-password" v-model="formState.password" type="password" placeholder="请输入密码" />
      </div>

      <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>

      <Button type="submit" class="w-full" :disabled="loading">
        {{ loading ? '登录中...' : '登录后台' }}
      </Button>
    </form>
  </AuthShell>
</template>
