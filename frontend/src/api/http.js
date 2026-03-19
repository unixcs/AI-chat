import axios from 'axios'
import router from '../router'
import { useChatStore } from '../stores/chat'

const handleAuthExpired = ({ code = '', message = '' } = {}) => {
  if (code === 'SESSION_KICKED' || message.includes('账号已在其他设备登录')) {
    const chatStore = useChatStore()
    chatStore.resetChatState()
    localStorage.removeItem('userToken')
    localStorage.removeItem('adminToken')
    if (router.currentRoute.value.path !== '/login') {
      router.push('/login')
    }
  }
}

const http = axios.create({
  baseURL: '/api',
  timeout: 15000
})

http.interceptors.request.use((config) => {
  const userToken = localStorage.getItem('userToken')
  const adminToken = localStorage.getItem('adminToken')

  if (config.headers.needUserAuth && userToken) {
    config.headers.Authorization = `Bearer ${userToken}`
  }

  if (config.headers.needAdminAuth && adminToken) {
    config.headers.Authorization = `Bearer ${adminToken}`
  }

  delete config.headers.needUserAuth
  delete config.headers.needAdminAuth
  return config
})

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const code = error.response?.data?.code || ''
    const message = error.response?.data?.message || ''
    handleAuthExpired({ code, message })
    return Promise.reject(error)
  }
)

export default http
