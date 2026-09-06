<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { registerUser } from '../../api/auth'
import AuthShell from '@/components/layout/AuthShell.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'

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
  <AuthShell max-width="max-w-[460px]">
    <div class="mb-6 flex flex-col items-center gap-3 text-center">
      <p class="text-[13px] text-white/75">欢迎加入</p>
      <h1 class="text-2xl leading-tight font-bold text-white">注册账号</h1>
    </div>

    <form class="space-y-4" @submit.prevent="submitRegister">
      <div class="space-y-2">
        <Label class="text-white/85" for="reg-phone">手机号</Label>
        <Input id="reg-phone" v-model="formState.phone" maxlength="11" placeholder="请输入手机号" />
      </div>
      <div class="space-y-2">
        <Label class="text-white/85" for="reg-nickname">昵称</Label>
        <Input id="reg-nickname" v-model="formState.nickname" placeholder="请输入昵称" />
      </div>
      <div class="space-y-2">
        <Label class="text-white/85" for="reg-password">密码</Label>
        <Input id="reg-password" v-model="formState.password" type="password" placeholder="请输入密码" />
      </div>
      <div class="space-y-2">
        <Label class="text-white/85" for="reg-confirm">确认密码</Label>
        <Input id="reg-confirm" v-model="formState.confirmPassword" type="password" placeholder="请再次输入密码" />
      </div>

      <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>

      <Button type="submit" class="w-full" :disabled="loading">
        {{ loading ? '注册中...' : '注册' }}
      </Button>

      <p class="text-right text-[13px]">
        <router-link to="/login" class="font-semibold text-white/85 transition-colors hover:text-white">
          已有账号？返回登录
        </router-link>
      </p>
    </form>
  </AuthShell>
</template>
