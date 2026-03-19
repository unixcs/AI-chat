<script setup>
import { computed, ref } from 'vue'

const currentTheme = ref(document.documentElement.dataset.theme || 'light')

const toggleLabel = computed(() => {
  return currentTheme.value === 'dark' ? '切换浅色' : '切换深色'
})

const toggleTheme = () => {
  const nextTheme = currentTheme.value === 'dark' ? 'light' : 'dark'
  currentTheme.value = nextTheme
  document.documentElement.dataset.theme = nextTheme
  localStorage.setItem('theme', nextTheme)
}
</script>

<template>
  <section class="landingPage pageWrap">
    <div class="contentWrap">
      <button class="themeToggle" @click="toggleTheme">{{ toggleLabel }}</button>
      <div class="brandLine">AI MEMBERSHIP CHAT</div>
      <h1>智能会员对话系统</h1>
      <p>支持会员兑换、历史会话、后台管理与移动端访问</p>
      <div class="ctaRow">
        <router-link class="primaryBtn" to="/login">进入登录</router-link>
        <router-link class="ghostBtn" to="/register">快速注册</router-link>
      </div>
    </div>
  </section>
</template>

<style scoped>
.landingPage {
  background:
    linear-gradient(120deg, rgba(9, 16, 33, 0.65), rgba(25, 93, 188, 0.45)),
    url('https://images.unsplash.com/photo-1526379095098-d400fd0bf935?auto=format&fit=crop&w=1600&q=80') center/cover;
  display: grid;
  place-items: center;
  padding: 24px;
}

.contentWrap {
  width: min(760px, 100%);
  text-align: center;
  color: #fff;
  background: rgba(11, 20, 41, 0.42);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  padding: 44px 26px;
  backdrop-filter: blur(5px);
  box-shadow: 0 12px 34px rgba(12, 20, 46, 0.4);
  position: relative;
}

.themeToggle {
  position: absolute;
  right: 14px;
  top: 14px;
  border: 1px solid rgba(255, 255, 255, 0.35);
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  border-radius: 8px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 12px;
}

.brandLine {
  font-size: 12px;
  letter-spacing: 2px;
  opacity: 0.85;
  margin-bottom: 12px;
}

h1 {
  margin: 0 0 10px;
  font-size: clamp(30px, 4vw, 42px);
}

p {
  margin: 0 0 24px;
  opacity: 0.92;
  font-size: 15px;
}

.ctaRow {
  display: flex;
  justify-content: center;
  gap: 10px;
}

.ctaRow .ghostBtn {
  background: rgba(255, 255, 255, 0.9);
}

@media (max-width: 640px) {
  .contentWrap {
    padding: 30px 18px;
  }

  .themeToggle {
    right: 10px;
    top: 10px;
  }

  .ctaRow {
    flex-direction: column;
  }
}
</style>
