import http from './http'

export const getConversations = () => {
  return http.get('/chat/conversations', { headers: { needUserAuth: true } })
}

export const createConversation = () => {
  return http.post('/chat/conversations', {}, { headers: { needUserAuth: true } })
}

export const getMessages = (conversationId) => {
  return http.get(`/chat/conversations/${conversationId}/messages`, { headers: { needUserAuth: true } })
}

export const sendMessage = (conversationId, content) => {
  return http.post(
    `/chat/conversations/${conversationId}/messages`,
    { content },
    { headers: { needUserAuth: true } }
  )
}

export const streamMessage = (conversationId, content, callbacks = {}) => {
  const token = localStorage.getItem('userToken') || ''
  const query = new URLSearchParams({
    content,
    token
  }).toString()

  const source = new EventSource(`/api/chat/conversations/${conversationId}/stream?${query}`)
  let streamClosed = false

  const closeStream = () => {
    if (streamClosed) {
      return
    }
    streamClosed = true
    source.close()
  }

  source.onmessage = (event) => {
    const text = event.data || ''
    if (text === '[DONE]') {
      closeStream()
      callbacks.onDone?.()
      return
    }

    try {
      const payload = JSON.parse(text)
      if (payload.error) {
        closeStream()
        const isKicked = payload.code === 'SESSION_KICKED' || payload.error.includes('账号已在其他设备登录')
        callbacks.onError?.(new Error(isKicked ? '账号已在其他设备登录' : '请联系管理员 ⚠️E1'))
        return
      }
      if (payload.delta) {
        callbacks.onDelta?.(payload.delta)
      }
    } catch (error) {
      closeStream()
      callbacks.onError?.(new Error('请联系管理员 ⚠️E2'))
    }
  }

  source.onerror = async () => {
    if (streamClosed) {
      return
    }

    closeStream()

    try {
      const sessionCheckResp = await fetch('/api/user/session-check', {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      if (!sessionCheckResp.ok) {
        const payload = await sessionCheckResp.json().catch(() => ({}))
        const isKicked =
          payload.code === 'SESSION_KICKED' || String(payload.message || '').includes('账号已在其他设备登录')
        callbacks.onError?.(new Error(isKicked ? '账号已在其他设备登录' : '请联系管理员 ⚠️E3'))
        return
      }
    } catch (error) {
    }

    callbacks.onError?.(new Error('请联系管理员 ⚠️E3'))
  }

  return source
}
