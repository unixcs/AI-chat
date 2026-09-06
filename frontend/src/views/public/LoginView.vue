<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginUser, getUserProfile } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { storeDraftSessionFlag, storeFreshChatFlag } from '../../utils/chat-entry'
import { applyTheme } from '../../utils/theme'
import AuthShell from '@/components/layout/AuthShell.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import { Sparkles } from 'lucide-vue-next'

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
  <AuthShell>
    <div class="mb-6 flex flex-col items-center gap-3 text-center">
      <span class="inline-flex size-12 items-center justify-center rounded-2xl bg-primary/25 text-primary-foreground">
        <Sparkles class="size-6" />
      </span>
      <h1 class="text-2xl leading-tight font-bold text-white">欢迎登陆 Thallo</h1>
      <p class="text-[13px] text-white/75">使用手机号密码登录以开始解读</p>
    </div>

    <form class="space-y-4" @submit.prevent="submitLogin">
      <div class="space-y-2">
        <Label class="text-white/85" for="login-phone">手机号</Label>
        <Input id="login-phone" v-model="formState.phone" maxlength="11" placeholder="请输入手机号" />
      </div>
      <div class="space-y-2">
        <Label class="text-white/85" for="login-password">密码</Label>
        <Input id="login-password" v-model="formState.password" type="password" placeholder="请输入密码" />
      </div>

      <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>

      <Button type="submit" class="w-full" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </Button>

      <p class="text-right text-[13px]">
        <router-link to="/register" class="font-semibold text-white/85 transition-colors hover:text-white">
          没有账号？去注册
        </router-link>
      </p>
    </form>
  </AuthShell>
</template>
