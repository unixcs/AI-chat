<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { checkSession, startSessionWatch, stopSessionWatch } from './utils/session-watch'

const backendOffline = ref(false)
let retryTimer = null

const setOnline = () => {
  backendOffline.value = false
}

const setOffline = () => {
  backendOffline.value = true
}

const probeApiByAuth = async () => {
  const resp = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone: '', password: '' }),
    cache: 'no-store'
  })

  if (resp.status === 400 || resp.status === 401 || resp.status === 403 || resp.status === 200) {
    return true
  }

  return false
}

const checkBackendHealth = async () => {
  try {
    const resp = await fetch('/api/health', { cache: 'no-store' })
    if (resp.ok) {
      setOnline()
      return
    }

    const fallbackOk = await probeApiByAuth()
    if (fallbackOk) {
      setOnline()
    } else {
      setOffline()
    }
  } catch (error) {
    try {
      const fallbackOk = await probeApiByAuth()
      if (fallbackOk) {
        setOnline()
        return
      }
    } catch (fallbackError) {
    }

    setOffline()
  }
}

const startAutoCheck = () => {
  retryTimer = window.setInterval(() => {
    checkBackendHealth()
  }, 8000)
}

const handleVisibilityChange = async () => {
  if (document.visibilityState === 'hidden') {
    stopSessionWatch()
    return
  }

  await checkSession()
  startSessionWatch()
}

const handleWindowFocus = async () => {
  await checkBackendHealth()
  await checkSession()
}

onMounted(() => {
  checkBackendHealth()
  startAutoCheck()
  if (document.visibilityState === 'visible') {
    startSessionWatch()
  }
  document.addEventListener('visibilitychange', handleVisibilityChange)
  window.addEventListener('focus', handleWindowFocus)
})

onBeforeUnmount(() => {
  if (retryTimer) {
    window.clearInterval(retryTimer)
    retryTimer = null
  }
  stopSessionWatch()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  window.removeEventListener('focus', handleWindowFocus)
})
</script>

<template>
  <div v-if="backendOffline" class="backendAlert">
    后端服务未启动（3001），当前仅显示页面外壳。请联系管理员。
  </div>
  <router-view />
</template>

<style scoped>
.backendAlert {
  position: fixed;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  padding: 8px 14px;
  border-radius: 8px;
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fecaca;
  box-shadow: 0 8px 16px rgba(127, 29, 29, 0.12);
  font-size: 13px;
}
</style>
