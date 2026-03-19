import http from '../api/http'
import router from '../router'
import { useChatStore } from '../stores/chat'

let sessionTimer = null

const handleKicked = () => {
  const chatStore = useChatStore()
  chatStore.resetChatState()
  localStorage.removeItem('userToken')
  if (router.currentRoute.value.path !== '/login') {
    router.push('/login')
  }
}

export const checkSession = async () => {
  const token = localStorage.getItem('userToken')
  if (!token) {
    return
  }

  try {
    await http.get('/user/session-check', { headers: { needUserAuth: true } })
  } catch (error) {
    const code = error.response?.data?.code || ''
    const message = error.response?.data?.message || ''
    if (code === 'SESSION_KICKED' || message.includes('账号已在其他设备登录')) {
      handleKicked()
    }
  }
}

export const startSessionWatch = () => {
  if (sessionTimer) {
    window.clearInterval(sessionTimer)
  }

  sessionTimer = window.setInterval(() => {
    checkSession()
  }, 30000)
}

export const stopSessionWatch = () => {
  if (!sessionTimer) {
    return
  }
  window.clearInterval(sessionTimer)
  sessionTimer = null
}
