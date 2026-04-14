import http from './http'

export const getAdminDashboard = () => {
  return http.get('/admin/dashboard', { headers: { needAdminAuth: true } })
}

export const getAdminUsers = () => {
  return http.get('/admin/users', { headers: { needAdminAuth: true } })
}

export const getAdminUsersByQuery = (params) => {
  return http.get('/admin/users', { params, headers: { needAdminAuth: true } })
}

export const updateAdminUser = (userId, payload) => {
  return http.put(`/admin/users/${userId}`, payload, { headers: { needAdminAuth: true } })
}

export const updateAdminUserStatus = (userId, status) => {
  return http.put(`/admin/users/${userId}/status`, { status }, { headers: { needAdminAuth: true } })
}

export const resetAdminUserPassword = (userId, newPassword) => {
  return http.put(
    `/admin/users/${userId}/reset-password`,
    { newPassword },
    { headers: { needAdminAuth: true } }
  )
}

export const updateAdminUserMemberExpireAt = (userId, memberExpireAt) => {
  return http.put(
    `/admin/users/${userId}/member-expire-at`,
    { memberExpireAt },
    { headers: { needAdminAuth: true } }
  )
}

export const getAdminRoles = () => {
  return http.get('/admin/roles', { headers: { needAdminAuth: true } })
}

export const updateAdminPassword = (newPassword) => {
  return http.put('/admin/auth/password', { newPassword }, { headers: { needAdminAuth: true } })
}

export const getAdminMenus = () => {
  return http.get('/admin/menus', { headers: { needAdminAuth: true } })
}

export const getAdminMembers = () => {
  return http.get('/admin/members', { headers: { needAdminAuth: true } })
}

export const getAdminRedeemCodes = () => {
  return http.get('/admin/redeem-codes', { headers: { needAdminAuth: true } })
}

export const getAdminRedeemCodesByQuery = (params) => {
  return http.get('/admin/redeem-codes', { params, headers: { needAdminAuth: true } })
}

export const createAdminRedeemCodes = (payload) => {
  return http.post('/admin/redeem-codes/batch', payload, { headers: { needAdminAuth: true } })
}

export const voidAdminRedeemCode = (codeId) => {
  return http.put(`/admin/redeem-codes/${codeId}/void`, {}, { headers: { needAdminAuth: true } })
}

export const exportAdminRedeemCodes = (params) => {
  return http.get('/admin/redeem-codes/export', { params, headers: { needAdminAuth: true } })
}

export const getAdminRedeemRecords = () => {
  return http.get('/admin/redeem-records', { headers: { needAdminAuth: true } })
}

export const getAdminRedeemRecordsByQuery = (params) => {
  return http.get('/admin/redeem-records', { params, headers: { needAdminAuth: true } })
}

export const getAdminConversations = () => {
  return http.get('/admin/conversations', { headers: { needAdminAuth: true } })
}

export const getAdminConversationsByQuery = (params) => {
  return http.get('/admin/conversations', { params, headers: { needAdminAuth: true } })
}

export const getAdminConversationMessages = (conversationId) => {
  return http.get(`/admin/conversations/${conversationId}/messages`, {
    headers: { needAdminAuth: true }
  })
}
