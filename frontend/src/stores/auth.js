import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    userToken: localStorage.getItem('userToken') || '',
    adminToken: localStorage.getItem('adminToken') || '',
    profile: null
  }),
  actions: {
    setUserToken(token) {
      this.userToken = token
      if (token) {
        localStorage.setItem('userToken', token)
      } else {
        localStorage.removeItem('userToken')
      }
    },
    setAdminToken(token) {
      this.adminToken = token
      if (token) {
        localStorage.setItem('adminToken', token)
      } else {
        localStorage.removeItem('adminToken')
      }
    },
    setProfile(profile) {
      this.profile = profile
    },
    logoutUser() {
      this.setUserToken('')
      this.profile = null
    },
    logoutAdmin() {
      this.setAdminToken('')
    }
  }
})
