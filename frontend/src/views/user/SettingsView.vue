<script setup>
import { computed, reactive, ref } from 'vue'
import { changePassword } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

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
  }
}
</script>

<template>
  <section class="card panelShell settingCard">
    <span class="sectionLabel">Security</span>
    <h2 class="sectionTitle">修改密码</h2>
    <p class="sectionIntro">提交新密码后会按原逻辑退出登录并返回登录页，这里只优化页面布局与表单呈现。</p>

    <div class="formItem">
      <label>旧密码</label>
      <input v-model="formState.oldPassword" type="password" />
    </div>
    <div class="formItem">
      <label>新密码</label>
      <input v-model="formState.newPassword" type="password" />
    </div>
    <div class="formItem">
      <label>确认新密码</label>
      <input v-model="formState.confirmPassword" type="password" />
    </div>
    <p v-if="errorText" class="dangerText errorText">{{ errorText }}</p>
    <button class="primaryBtn" @click="submitChangePassword">确认修改</button>

    <div class="prefDisplay">
      <h3 class="prefTitle">回答偏好</h3>
      <div class="prefRow">
        <span v-for="pref in answerPrefs" :key="pref.name" class="prefItem">
          <small class="prefName">{{ pref.name }}</small>
          <b class="prefValue">{{ pref.label }}</b>
        </span>
      </div>
      <p class="prefHint">在聊天界面的偏好弹层中调整，对所有新对话生效。</p>
    </div>
  </section>
</template>

<style scoped>
.settingCard {
  padding: 24px;
  max-width: 760px;
}

.settingCard .sectionTitle {
  margin-top: 14px;
}

.settingCard .sectionIntro {
  margin-bottom: 22px;
}

.errorText {
  margin: 0 0 12px;
}

.prefDisplay {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--line-soft);
}

.prefTitle {
  margin: 0 0 12px;
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
  color: var(--text-soft);
}
</style>
