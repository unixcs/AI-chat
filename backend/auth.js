const jwt = require('jsonwebtoken')
const { getUserById } = require('./db')

const FALLBACK_JWT_SECRET = 'demo_secret_key_2026'
const jwtSecret = process.env.JWT_SECRET || FALLBACK_JWT_SECRET
if (!process.env.JWT_SECRET) {
  console.warn('[auth] JWT_SECRET 未设置,正在使用内置默认密钥(不安全,生产环境必须在 .env 中配置)')
}

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
  const user = getUserById(payload.userId)

  if (!user || user.role !== 'user') {
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
