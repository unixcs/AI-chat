<script setup>
import { computed, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { useAuthStore } from '../../stores/auth'
import { updateProfile, getUserProfile } from '../../api/auth'
import { redeemCode } from '../../api/redeem'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import { toast } from 'vue-sonner'

const authStore = useAuthStore()
// 两张卡片各自独立提示：共用状态会导致“资料更新成功”漏进兑换卡（反之亦然）
const profileError = ref('')
const profileNotice = ref('')
const redeemError = ref('')
const redeemNotice = ref('')

// 与 ChatView 偏好弹层保持同一套标签映射（Plan §2.6：Profile 同步展示）
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

const profileForm = reactive({
  nickname: authStore.profile?.nickname || '',
  avatarUrl: authStore.profile?.avatarUrl || ''
})

const redeemForm = reactive({
  code: ''
})

const saveProfile = async () => {
  profileError.value = ''
  profileNotice.value = ''
  try {
    await updateProfile(profileForm)
    const { data } = await getUserProfile()
    authStore.setProfile(data.data)
    profileNotice.value = '资料更新成功'
    toast.success('资料更新成功')
  } catch (error) {
    profileError.value = error.response?.data?.message || '资料更新失败'
    toast.error(profileError.value)
  }
}

const submitRedeem = async () => {
  redeemError.value = ''
  redeemNotice.value = ''
  if (!redeemForm.code) {
    redeemError.value = '请输入兑换码'
    return
  }

  try {
    const { data } = await redeemCode(redeemForm.code)
    authStore.setProfile(data.data.profile)
    redeemForm.code = ''
    redeemNotice.value = `兑换成功，会员到期：${dayjs(data.data.profile.memberExpireAt).format('YYYY-MM-DD HH:mm')}`
    toast.success('兑换成功')
  } catch (error) {
    redeemError.value = error.response?.data?.message || '兑换失败'
    toast.error(redeemError.value)
  }
}
</script>

<template>
  <section class="grid gap-4 p-4 md:grid-cols-2 md:gap-5 md:p-6">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="text-lg font-semibold text-card-foreground">个人资料</h2>
        <p class="mt-1 mb-5 text-sm text-muted-foreground">昵称与头像会展示在对话界面</p>

        <form class="space-y-4" @submit.prevent="saveProfile">
          <div class="space-y-2">
            <Label for="profile-nickname">昵称</Label>
            <Input id="profile-nickname" v-model="profileForm.nickname" placeholder="请输入昵称" />
          </div>
          <div class="space-y-2">
            <Label for="profile-avatar">头像地址</Label>
            <Input id="profile-avatar" v-model="profileForm.avatarUrl" placeholder="请输入头像 URL" />
          </div>

          <Alert v-if="profileError" variant="destructive">{{ profileError }}</Alert>
          <p v-if="profileNotice" class="text-sm text-primary">{{ profileNotice }}</p>

          <Button type="submit">保存资料</Button>
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

    <Card class="border-primary/20">
      <CardContent class="p-5 sm:p-6">
        <div class="flex items-center justify-between gap-2">
          <h2 class="text-lg font-semibold text-card-foreground">卡密充值</h2>
          <Badge>Membership</Badge>
        </div>
        <p class="mt-1 mb-5 text-sm text-muted-foreground">邀请码或卡密固定兑换 30 天会员</p>

        <form class="space-y-4" @submit.prevent="submitRedeem">
          <div class="space-y-2">
            <Label for="redeem-code">兑换码</Label>
            <Input id="redeem-code" v-model="redeemForm.code" placeholder="请输入兑换码" />
          </div>

          <Alert v-if="redeemError" variant="destructive">{{ redeemError }}</Alert>
          <p v-if="redeemNotice" class="text-sm text-primary">{{ redeemNotice }}</p>

          <Button type="submit">立即兑换</Button>
        </form>
      </CardContent>
    </Card>
  </section>
</template>
