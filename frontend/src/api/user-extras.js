import http from './http'

// 用户回答模式（会话层生效，不改写用户消息）
export const updatePreferences = (payload) => {
  return http.put('/user/preferences', payload, { headers: { needUserAuth: true } })
}

// 一次性公告：获取当前未读 → 确认已读
export const getCurrentAnnouncement = () => {
  return http.get('/announcements/current', { headers: { needUserAuth: true } })
}

export const ackAnnouncement = (id) => {
  return http.post(`/announcements/${id}/ack`, {}, { headers: { needUserAuth: true } })
}
