<script setup>
import { reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { useAuthStore } from '../../stores/auth'
import { updateProfile, getUserProfile } from '../../api/auth'
import { redeemCode } from '../../api/redeem'

const authStore = useAuthStore()
const noticeText = ref('')
const errorText = ref('')

const profileForm = reactive({
  nickname: authStore.profile?.nickname || '',
  avatarUrl: authStore.profile?.avatarUrl || ''
})

const redeemForm = reactive({
  code: ''
})

const saveProfile = async () => {
  errorText.value = ''
  noticeText.value = ''
  try {
    await updateProfile(profileForm)
    const { data } = await getUserProfile()
    authStore.setProfile(data.data)
    noticeText.value = '资料更新成功'
  } catch (error) {
    errorText.value = error.response?.data?.message || '资料更新失败'
  }
}

const submitRedeem = async () => {
  errorText.value = ''
  noticeText.value = ''
  if (!redeemForm.code) {
    errorText.value = '请输入兑换码'
    return
  }

  try {
    const { data } = await redeemCode(redeemForm.code)
    authStore.setProfile(data.data.profile)
    redeemForm.code = ''
    noticeText.value = `兑换成功，会员到期：${dayjs(data.data.profile.memberExpireAt).format('YYYY-MM-DD HH:mm')}`
  } catch (error) {
    errorText.value = error.response?.data?.message || '兑换失败'
  }
}
</script>

<template>
  <section class="profilePage">
    <article class="card panel">
      <h2 class="sectionTitle">个人中心</h2>
      <div class="formItem">
        <label>昵称</label>
        <input v-model="profileForm.nickname" placeholder="请输入昵称" />
      </div>
      <div class="formItem">
        <label>头像地址</label>
        <input v-model="profileForm.avatarUrl" placeholder="请输入头像 URL" />
      </div>
      <button class="primaryBtn" @click="saveProfile">保存资料</button>
    </article>

    <article class="card panel">
      <h2 class="sectionTitle">卡密充值</h2>
      <div class="formItem">
        <label>兑换码</label>
        <input v-model="redeemForm.code" placeholder="请输入兑换码" />
      </div>
      <button class="primaryBtn" @click="submitRedeem">立即兑换</button>
      <p v-if="noticeText" class="mutedText successText">{{ noticeText }}</p>
      <p v-if="errorText" class="dangerText">{{ errorText }}</p>
    </article>
  </section>
</template>

<style scoped>
.profilePage {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.panel {
  padding: 18px;
}

.successText {
  color: var(--success);
}

@media (max-width: 960px) {
  .profilePage {
    grid-template-columns: 1fr;
  }
}
</style>
