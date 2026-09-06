import http from './http'

// 用户回答模式（会话层生效，不改写用户消息）
export const updatePreferences = (payload) => {
  return http.put('/user/preferences', payload, { headers: { needUserAuth: true } })
}

// 一次性公告：获取当前未读 → 确认已读
export const getCurrentAnnouncement = () => {
  return http.get('/announcements/current', { headers: { needUserAuth: true } })
}

// asOf = 用户看到公告时的 updatedAt：若管理员在用户阅读期间改版，
// 后端丢弃过期 ack，新内容仍视为未读重新提示
export const ackAnnouncement = (id, asOf) => {
  return http.post(`/announcements/${id}/ack`, asOf ? { asOf } : {}, { headers: { needUserAuth: true } })
}
