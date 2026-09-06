<script setup>
import { computed, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { useAuthStore } from '../../stores/auth'
import { updateProfile, getUserProfile } from '../../api/auth'
import { redeemCode } from '../../api/redeem'

const authStore = useAuthStore()
const noticeText = ref('')
const errorText = ref('')

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
    <article class="card panelShell profilePanel">
      <span class="sectionLabel">Profile</span>
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

      <div class="prefDisplay">
        <span class="sectionLabel">Preferences</span>
        <h3 class="prefTitle">回答偏好</h3>
        <div class="prefRow">
          <span v-for="pref in answerPrefs" :key="pref.name" class="prefItem">
            <small class="prefName">{{ pref.name }}</small>
            <b class="prefValue">{{ pref.label }}</b>
          </span>
        </div>
        <p class="mutedText prefHint">在聊天界面的偏好弹层中调整，对所有新对话生效。</p>
      </div>
    </article>

    <article class="card panelShell profilePanel accentPanel">
      <span class="sectionLabel">Membership</span>
      <h2 class="sectionTitle">卡密充值</h2>

      <div class="formItem">
        <label>兑换码</label>
        <input v-model="redeemForm.code" placeholder="请输入兑换码" />
      </div>
      <button class="primaryBtn" @click="submitRedeem">立即兑换</button>
      <p v-if="noticeText" class="mutedText successText">{{ noticeText }}</p>
      <p v-if="errorText" class="dangerText feedbackText">{{ errorText }}</p>
    </article>
  </section>
</template>

<style scoped>
.profilePage {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.profilePanel {
  padding: 24px;
}

.profilePanel .sectionTitle {
  margin: 14px 0 22px;
}

.accentPanel {
  background: linear-gradient(180deg, rgba(255, 251, 247, 0.92) 0%, rgba(243, 238, 231, 0.8) 100%);
}

[data-theme='dark'] .accentPanel {
  background: linear-gradient(180deg, rgba(39, 45, 54, 0.92) 0%, rgba(29, 34, 41, 0.9) 100%);
}

.successText {
  margin-top: 14px;
  color: var(--success);
}

.feedbackText {
  margin-top: 8px;
}

.prefDisplay {
  margin-top: 22px;
  padding-top: 16px;
  border-top: 1px solid var(--line-soft);
}

.prefTitle {
  margin: 10px 0 12px;
  font-size: 15px;
  color: var(--text-title);
}

.prefRow {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.prefItem {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 14px;
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.45);
  min-width: 84px;
}

[data-theme='dark'] .prefItem {
  background: rgba(255, 255, 255, 0.04);
}

.prefName {
  color: var(--text-soft);
  font-size: 12px;
}

.prefValue {
  color: var(--text-title);
  font-size: 14px;
}

.prefHint {
  margin: 10px 0 0;
  font-size: 12px;
}

@media (max-width: 960px) {
  .profilePage {
    grid-template-columns: 1fr;
  }
}
</style>
