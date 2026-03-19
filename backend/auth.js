const jwt = require('jsonwebtoken')
const { readDb } = require('./db')

const jwtSecret = 'demo_secret_key_2026'

const signToken = (payload, expiresIn = '7d') => {
  return jwt.sign(payload, jwtSecret, { expiresIn })
}

const verifyToken = (token) => {
  return jwt.verify(token, jwtSecret)
}

const parseBearerToken = (authHeader = '') => {
  if (!authHeader.startsWith('Bearer ')) {
    return ''
  }
  return authHeader.replace('Bearer ', '').trim()
}

const validateUserSession = (payload) => {
  const db = readDb()
  const user = db.users.find((item) => item.id === payload.userId && item.role === 'user')

  if (!user) {
    return { ok: false, status: 401, code: 'USER_INVALID', message: '用户不存在或已失效' }
  }

  if (user.status !== 'active') {
    return { ok: false, status: 403, code: 'USER_DISABLED', message: '账号已被禁用' }
  }

  if (!payload.sessionId || !user.currentSessionId || payload.sessionId !== user.currentSessionId) {
    return { ok: false, status: 401, code: 'SESSION_KICKED', message: '账号已在其他设备登录' }
  }

  return { ok: true, user }
}

const userAuth = (req, res, next) => {
  try {
    const token = parseBearerToken(req.headers.authorization)
    if (!token) {
      return res.status(401).json({ message: '未登录或登录已过期' })
    }
    const payload = verifyToken(token)
    if (payload.type !== 'user') {
      return res.status(403).json({ message: '无效用户令牌' })
    }
    const sessionCheck = validateUserSession(payload)
    if (!sessionCheck.ok) {
      return res.status(sessionCheck.status).json({ code: sessionCheck.code, message: sessionCheck.message })
    }
    req.user = payload
    return next()
  } catch (error) {
    return res.status(401).json({ message: '用户鉴权失败' })
  }
}

const adminAuth = (req, res, next) => {
  try {
    const token = parseBearerToken(req.headers.authorization)
    if (!token) {
      return res.status(401).json({ message: '管理员未登录' })
    }
    const payload = verifyToken(token)
    if (payload.type !== 'admin') {
      return res.status(403).json({ message: '无效管理员令牌' })
    }
    req.admin = payload
    return next()
  } catch (error) {
    return res.status(401).json({ message: '管理员鉴权失败' })
  }
}

module.exports = {
  signToken,
  verifyToken,
  validateUserSession,
  userAuth,
  adminAuth
}
