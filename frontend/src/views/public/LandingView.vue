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
      <router-link class="primaryBtn ctaOnly" to="/login">开启灵感</router-link>
    </div>
  </section>
</template>

<style scoped>
.landingPage {
  background:
    linear-gradient(120deg, rgba(9, 16, 33, 0.62), rgba(31, 78, 147, 0.48)),
    url('/assets/login-bg.png') center/cover;
  display: grid;
  place-items: center;
  padding: 24px;
}

.contentWrap {
  width: min(520px, 100%);
  text-align: center;
  color: #fff;
  background: rgba(11, 20, 41, 0.38);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  padding: 32px 24px;
  backdrop-filter: blur(6px);
  box-shadow: 0 14px 30px rgba(12, 20, 46, 0.35);
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

.ctaOnly {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 14px 34px;
  border-radius: 999px;
  font-size: 16px;
  letter-spacing: 1px;
  box-shadow: 0 10px 22px rgba(47, 124, 246, 0.38);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.ctaOnly:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 28px rgba(47, 124, 246, 0.35);
}

@media (max-width: 640px) {
  .contentWrap {
    padding: 26px 18px;
  }

  .themeToggle {
    right: 10px;
    top: 10px;
  }

  .ctaOnly {
    width: 100%;
  }
}
</style>
