<script setup>
import { computed, reactive, ref } from 'vue'
import { changePassword } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import { toast } from 'vue-sonner'

const authStore = useAuthStore()
const router = useRouter()

// 与 ChatView/ProfileView 同一套标签映射（Plan §2.6：Settings 同步展示）
const prefLengthLabels = { concise: '精简', standard: '适中', detailed: '详细' }
const prefStyleLabels = { plain: '大白话', standard: '标准', professional: '专业', rigorous: '严谨', encouraging: '鼓励' }
const prefFormatLabels = { standard: '标准', plain: '纯文字' }
const answerPrefs = computed(() => {
  const p = authStore.profile || {}
  return [
    { name: '回答长度', label: prefLengthLabels[p.answerLength] || '适中' },
    { name: '回答风格', label: prefStyleLabels[p.answerStyle] || '标准' },
    { name: '输出格式', label: prefFormatLabels[p.answerFormat] || '标准' }
  ]
})

const formState = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const errorText = ref('')

const submitChangePassword = async () => {
  errorText.value = ''
  if (!formState.oldPassword || !formState.newPassword || !formState.confirmPassword) {
    errorText.value = '请完整填写密码信息'
    return
  }

  if (formState.newPassword !== formState.confirmPassword) {
    errorText.value = '两次新密码不一致'
    return
  }

  if (formState.newPassword.length < 6) {
    errorText.value = '新密码至少 6 位'
    return
  }

  try {
    await changePassword({
      oldPassword: formState.oldPassword,
      newPassword: formState.newPassword
    })
    authStore.logoutUser()
    router.push('/login')
  } catch (error) {
    errorText.value = error.response?.data?.message || '修改密码失败'
    toast.error(errorText.value)
  }
}
</script>

<template>
  <section class="p-4 md:p-6">
    <Card class="mx-auto max-w-3xl">
      <CardContent class="p-5 sm:p-6">
        <h2 class="text-lg font-semibold text-card-foreground">修改密码</h2>
        <p class="mt-1 mb-5 text-sm text-muted-foreground">提交新密码后会按原逻辑退出登录并返回登录页</p>

        <form class="space-y-4" @submit.prevent="submitChangePassword">
          <div class="space-y-2">
            <Label for="settings-old">旧密码</Label>
            <Input id="settings-old" v-model="formState.oldPassword" type="password" />
          </div>
          <div class="space-y-2">
            <Label for="settings-new">新密码</Label>
            <Input id="settings-new" v-model="formState.newPassword" type="password" />
          </div>
          <div class="space-y-2">
            <Label for="settings-confirm">确认新密码</Label>
            <Input id="settings-confirm" v-model="formState.confirmPassword" type="password" />
          </div>

          <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>

          <Button type="submit">确认修改</Button>
        </form>

        <div class="mt-6 border-t border-border pt-4">
          <h3 class="text-[15px] font-semibold text-card-foreground">回答偏好</h3>
          <div class="mt-3 flex flex-wrap gap-2.5">
            <span v-for="pref in answerPrefs" :key="pref.name" class="flex min-w-21 flex-col gap-0.5 rounded-lg border border-border bg-muted/50 px-3.5 py-2">
              <small class="text-xs text-muted-foreground">{{ pref.name }}</small>
              <b class="text-sm font-semibold text-foreground">{{ pref.label }}</b>
            </span>
          </div>
          <p class="mt-2.5 text-xs text-faint">在聊天界面的偏好弹层中调整，对所有新对话生效。</p>
        </div>
      </CardContent>
    </Card>
  </section>
</template>
