const fs = require('fs')
const path = require('path')
const express = require('express')
const cors = require('cors')
const bcrypt = require('bcryptjs')
const dayjs = require('dayjs')
const dotenv = require('dotenv')
const { readDb, writeDb, newId, getNow, addMonths } = require('./db')
const { signToken, userAuth, adminAuth, verifyToken, validateUserSession } = require('./auth')
const { normalizeAdminMemberExpireAtInput } = require('./admin-member-expire')

dotenv.config()

const app = express()
const port = 3001
const deepseekApiKey = process.env.DEEPSEEK_API_KEY || ''
const deepseekModel = process.env.DEEPSEEK_MODEL || 'deepseek-chat'
const deepseekBaseUrl = process.env.DEEPSEEK_BASE_URL || 'https://api.deepseek.com/v1'
const deepseekPromptFile = process.env.DEEPSEEK_SYSTEM_PROMPT_FILE || ''
const deepseekSystemPrompt = (() => {
  if (!deepseekPromptFile) {
    return '你是一个专业、简洁、友好的中文 AI 助手。'
  }
  try {
    const resolvedPath = path.isAbsolute(deepseekPromptFile)
      ? deepseekPromptFile
      : path.join(__dirname, deepseekPromptFile)
    return fs.readFileSync(resolvedPath, 'utf8').trim()
  } catch (error) {
    return '你是一个专业、简洁、友好的中文 AI 助手。'
  }
})()

function toPositiveInt(value, fallback) {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue) || numberValue <= 0) {
    return fallback
  }
  return Math.floor(numberValue)
}

const modelConcurrency = Math.max(1, toPositiveInt(process.env.MODEL_CONCURRENCY || 6, 6))
const modelQueueMax = Math.max(1, toPositiveInt(process.env.MODEL_QUEUE_MAX || 50, 50))

app.use(cors())
app.use(express.json())

app.get('/api/health', (req, res) => {
  return res.json({ ok: true, service: 'backend', time: getNow() })
})

const sendOk = (res, data) => {
  return res.json({ code: 0, data })
}

const findUserById = (db, userId) => {
  return db.users.find((user) => user.id === userId)
}

const toSafeUser = (user) => {
  return {
    id: user.id,
    phone: user.phone,
    nickname: user.nickname,
    avatarUrl: user.avatarUrl,
    status: user.status,
    role: user.role,
    memberExpireAt: user.memberExpireAt,
    createdAt: user.createdAt
  }
}

const ensureMembershipValid = (user) => {
  if (!user.memberExpireAt) {
    return false
  }
  return dayjs(user.memberExpireAt).isAfter(dayjs())
}

const parseBearerToken = (authHeader = '') => {
  if (!authHeader.startsWith('Bearer ')) {
    return ''
  }
  return authHeader.replace('Bearer ', '').trim()
}

const createSessionId = () => {
  return `${Date.now()}_${Math.random().toString(36).slice(2, 12)}`
}

let modelInFlight = 0
const modelQueue = []

const dequeueModelTask = () => {
  if (modelInFlight >= modelConcurrency) {
    return
  }

  const task = modelQueue.shift()
  if (!task) {
    return
  }

  modelInFlight += 1
  task.run()
}

const withModelSlot = (executor) => {
  return new Promise((resolve, reject) => {
    const run = async () => {
      try {
        const result = await executor()
        resolve(result)
      } catch (error) {
        reject(error)
      } finally {
        modelInFlight = Math.max(0, modelInFlight - 1)
        dequeueModelTask()
      }
    }

    if (modelInFlight < modelConcurrency) {
      modelInFlight += 1
      run()
      return
    }

    if (modelQueue.length >= modelQueueMax) {
      reject(new Error('MODEL_QUEUE_OVERFLOW'))
      return
    }

    modelQueue.push({ run })
  })
}

const wait = (ms) => {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

const createModelError = (code, message) => {
  const error = new Error(message)
  error.modelCode = code
  return error
}

const callDeepseekWithRetry = async ({ content, contextMessages, signal }) => {
  const maxRetries = 2
  let retryDelay = 500
  let lastError = null

  for (let attempt = 0; attempt <= maxRetries; attempt += 1) {
    try {
      const upstreamResp = await fetch(`${deepseekBaseUrl}/chat/completions`, {
        method: 'POST',
        signal,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${deepseekApiKey}`
        },
        body: JSON.stringify({
          model: deepseekModel,
          stream: true,
          messages: [
            { role: 'system', content: deepseekSystemPrompt },
            ...contextMessages,
            { role: 'user', content }
          ]
        })
      })

      if (upstreamResp.ok && upstreamResp.body) {
        return upstreamResp
      }

      const detail = await upstreamResp.text()
      if (upstreamResp.status === 402) {
        throw createModelError('MODEL_INSUFFICIENT_BALANCE', '模型余额不足，请联系管理员')
      }
      if (upstreamResp.status === 429) {
        lastError = createModelError('MODEL_RATE_LIMIT', '模型请求过于频繁，请稍后重试')
      } else if (upstreamResp.status >= 500) {
        lastError = createModelError('MODEL_UPSTREAM_ERROR', `模型服务异常: ${upstreamResp.status}`)
      } else {
        throw createModelError('MODEL_UPSTREAM_ERROR', `模型请求失败: ${upstreamResp.status} ${detail}`)
      }
    } catch (error) {
      if (error.name === 'AbortError') {
        throw createModelError('MODEL_TIMEOUT', '模型响应超时或连接中断，请重试')
      }

      if (error.modelCode) {
        if (
          ['MODEL_RATE_LIMIT', 'MODEL_UPSTREAM_ERROR'].includes(error.modelCode) &&
          attempt < maxRetries
        ) {
          await wait(retryDelay)
          retryDelay *= 3
          continue
        }
        throw error
      }

      lastError = createModelError('MODEL_UPSTREAM_ERROR', error.message || '模型调用失败')
    }

    if (lastError && attempt < maxRetries) {
      await wait(retryDelay)
      retryDelay *= 3
      continue
    }
  }

  throw lastError || createModelError('MODEL_UPSTREAM_ERROR', '模型调用失败')
}

const paginateList = (list, page, pageSize) => {
  const start = (page - 1) * pageSize
  const end = start + pageSize
  const items = list.slice(start, end)
  return {
    items,
    total: list.length,
    page,
    pageSize
  }
}

const buildRedeemCodeList = (db, { status = '', codeKeyword = '' }) => {
  let list = db.redeemCodes.map((item) => {
    const user = item.usedByUserId ? db.users.find((x) => x.id === item.usedByUserId) : null
    return {
      ...item,
      usedByPhone: user?.phone || ''
    }
  })

  if (status) {
    list = list.filter((item) => item.status === status)
  }

  if (codeKeyword) {
    list = list.filter((item) => item.code.toUpperCase().includes(codeKeyword))
  }

  return list.sort((a, b) => dayjs(b.createdAt).valueOf() - dayjs(a.createdAt).valueOf())
}

const buildContextByRounds = (historyMessages, roundCount = 6) => {
  const ordered = [...historyMessages].sort(
    (a, b) => dayjs(a.createdAt).valueOf() - dayjs(b.createdAt).valueOf()
  )

  const rounds = []
  let pendingUser = null

  for (const msg of ordered) {
    if (!['user', 'assistant'].includes(msg.role)) {
      continue
    }

    if (msg.role === 'user') {
      if (pendingUser) {
        rounds.push([pendingUser])
      }
      pendingUser = msg
      continue
    }

    if (pendingUser) {
      rounds.push([pendingUser, msg])
      pendingUser = null
    }
  }

  if (pendingUser) {
    rounds.push([pendingUser])
  }

  return rounds
    .slice(-roundCount)
    .flat()
    .map((msg) => ({ role: msg.role, content: msg.content }))
}

app.post('/api/auth/register', (req, res) => {
  const { phone, nickname, password } = req.body
  if (!/^1\d{10}$/.test(phone || '')) {
    return res.status(400).json({ message: '手机号格式不正确' })
  }
  if (!nickname || String(nickname).trim().length === 0) {
    return res.status(400).json({ message: '昵称不能为空' })
  }
  if (!password || password.length < 6) {
    return res.status(400).json({ message: '密码至少 6 位' })
  }

  const db = readDb()
  const existed = db.users.find((item) => item.phone === phone)
  if (existed) {
    return res.status(400).json({ message: '手机号已注册' })
  }

  const now = getNow()
  db.users.push({
    id: newId(),
    phone,
    nickname,
    passwordHash: bcrypt.hashSync(password, 10),
    avatarUrl: '',
    status: 'active',
    role: 'user',
    currentSessionId: null,
    sessionUpdatedAt: null,
    memberExpireAt: null,
    createdAt: now,
    lastLoginAt: now
  })
  writeDb(db)
  return sendOk(res, true)
})

app.post('/api/auth/login', (req, res) => {
  const { phone, password } = req.body
  const db = readDb()
  const user = db.users.find((item) => item.phone === phone && item.role === 'user')

  if (!user) {
    return res.status(400).json({ message: '账号不存在' })
  }
  if (user.status !== 'active') {
    return res.status(403).json({ message: '账号已被禁用' })
  }
  if (!bcrypt.compareSync(password, user.passwordHash)) {
    return res.status(400).json({ message: '密码错误' })
  }

  user.lastLoginAt = getNow()
  user.currentSessionId = createSessionId()
  user.sessionUpdatedAt = getNow()
  writeDb(db)

  const token = signToken({ userId: user.id, sessionId: user.currentSessionId, type: 'user' })
  return sendOk(res, { token })
})

app.post('/api/auth/logout', userAuth, (req, res) => {
  const db = readDb()
  const user = findUserById(db, req.user.userId)
  if (!user) {
    return res.status(404).json({ message: '用户不存在' })
  }

  user.currentSessionId = null
  user.sessionUpdatedAt = getNow()
  writeDb(db)
  return sendOk(res, true)
})

app.get('/api/user/session-check', userAuth, (req, res) => {
  return sendOk(res, { ok: true })
})

app.get('/api/user/profile', userAuth, (req, res) => {
  const db = readDb()
  const user = findUserById(db, req.user.userId)
  return sendOk(res, toSafeUser(user))
})

app.put('/api/user/profile', userAuth, (req, res) => {
  const { nickname, avatarUrl } = req.body
  const db = readDb()
  const user = findUserById(db, req.user.userId)
  user.nickname = nickname || user.nickname
  user.avatarUrl = avatarUrl || ''
  writeDb(db)
  return sendOk(res, toSafeUser(user))
})

app.put('/api/user/password', userAuth, (req, res) => {
  const { oldPassword, newPassword } = req.body
  const db = readDb()
  const user = findUserById(db, req.user.userId)
  if (!bcrypt.compareSync(oldPassword, user.passwordHash)) {
    return res.status(400).json({ message: '旧密码错误' })
  }
  user.passwordHash = bcrypt.hashSync(newPassword, 10)
  writeDb(db)
  return sendOk(res, true)
})

app.post('/api/user/redeem', userAuth, (req, res) => {
  const { code } = req.body
  const db = readDb()
  const user = findUserById(db, req.user.userId)
  const redeemCode = db.redeemCodes.find((item) => item.code === code)

  if (!redeemCode) {
    return res.status(400).json({ message: '兑换码不存在' })
  }
  if (redeemCode.status !== 'unused') {
    return res.status(400).json({ message: '兑换码已失效或已使用' })
  }

  const activatedAt = getNow()
  const beforeExpireAt = user.memberExpireAt
  const afterExpireAt = addMonths(activatedAt, redeemCode.durationMonths || 1)

  user.memberExpireAt = afterExpireAt
  redeemCode.status = 'used'
  redeemCode.usedAt = activatedAt
  redeemCode.usedByUserId = user.id

  db.redeemRecords.push({
    id: newId(),
    userId: user.id,
    phone: user.phone,
    code: redeemCode.code,
    activatedAt,
    beforeExpireAt,
    afterExpireAt
  })

  writeDb(db)
  return sendOk(res, { profile: toSafeUser(user) })
})

app.get('/api/chat/conversations', userAuth, (req, res) => {
  const db = readDb()
  const list = db.conversations
    .filter((item) => item.userId === req.user.userId)
    .sort((a, b) => dayjs(b.updatedAt).valueOf() - dayjs(a.updatedAt).valueOf())
  return sendOk(res, list)
})

app.post('/api/chat/conversations', userAuth, (req, res) => {
  const db = readDb()
  const now = getNow()
  const conversation = {
    id: newId(),
    userId: req.user.userId,
    title: `新对话 ${dayjs(now).format('MM-DD HH:mm')}`,
    createdAt: now,
    updatedAt: now
  }
  db.conversations.push(conversation)
  writeDb(db)
  return sendOk(res, conversation)
})

app.get('/api/chat/conversations/:id/messages', userAuth, (req, res) => {
  const db = readDb()
  const conversation = db.conversations.find(
    (item) => item.id === req.params.id && item.userId === req.user.userId
  )
  if (!conversation) {
    return res.status(404).json({ message: '会话不存在' })
  }

  const messages = db.messages
    .filter((item) => item.conversationId === conversation.id)
    .sort((a, b) => dayjs(a.createdAt).valueOf() - dayjs(b.createdAt).valueOf())
  return sendOk(res, messages)
})

app.post('/api/chat/conversations/:id/messages', userAuth, (req, res) => {
  const { content } = req.body
  const db = readDb()
  const user = findUserById(db, req.user.userId)
  const conversation = db.conversations.find(
    (item) => item.id === req.params.id && item.userId === req.user.userId
  )

  if (!conversation) {
    return res.status(404).json({ message: '会话不存在' })
  }

  if (!ensureMembershipValid(user)) {
    return res.status(403).json({ message: '会员过期，请续费后使用' })
  }

  const now = getNow()
  db.messages.push({
    id: newId(),
    conversationId: conversation.id,
    role: 'user',
    content,
    createdAt: now
  })

  const replyMessage = {
    id: newId(),
    conversationId: conversation.id,
    role: 'assistant',
    content: `已收到你的问题：${content}\n这是演示回复，后续可替换真实模型 API。`,
    createdAt: getNow()
  }
  db.messages.push(replyMessage)

  conversation.updatedAt = getNow()
  writeDb(db)
  return sendOk(res, replyMessage)
})

app.get('/api/chat/conversations/:id/stream', async (req, res) => {
  const token = parseBearerToken(req.headers.authorization || '') || String(req.query.token || '')
  const conversationId = req.params.id
  const content = String(req.query.content || '').trim()

  if (!token) {
    return res.status(401).json({ message: '未登录或登录已过期' })
  }

  let payload
  try {
    payload = verifyToken(token)
    if (payload.type !== 'user') {
      return res.status(403).json({ code: 'INVALID_USER_TOKEN', message: '无效用户令牌' })
    }
    const sessionCheck = validateUserSession(payload)
    if (!sessionCheck.ok) {
      return res.status(sessionCheck.status).json({ code: sessionCheck.code, message: sessionCheck.message })
    }
  } catch (error) {
    return res.status(401).json({ code: 'USER_AUTH_FAILED', message: '用户鉴权失败' })
  }

  if (!content) {
    return res.status(400).json({ message: '消息内容不能为空' })
  }

  const db = readDb()
  const user = findUserById(db, payload.userId)
  const conversation = db.conversations.find(
    (item) => item.id === conversationId && item.userId === payload.userId
  )

  if (!conversation) {
    return res.status(404).json({ message: '会话不存在' })
  }

  if (!ensureMembershipValid(user)) {
    return res.status(403).json({ message: '会员过期，请续费后使用' })
  }

  if (!deepseekApiKey) {
    return res.status(500).json({ message: '服务器未配置 DEEPSEEK_API_KEY' })
  }

  const historyMessages = db.messages.filter(
    (item) => item.conversationId === conversation.id && ['user', 'assistant'].includes(item.role)
  )
  const contextMessages = buildContextByRounds(historyMessages, 6)

  const now = getNow()
  db.messages.push({
    id: newId(),
    conversationId: conversation.id,
    role: 'user',
    content,
    createdAt: now
  })
  conversation.updatedAt = now
  writeDb(db)

  res.setHeader('Content-Type', 'text/event-stream; charset=utf-8')
  res.setHeader('Cache-Control', 'no-cache, no-transform')
  res.setHeader('Connection', 'keep-alive')
  res.setHeader('X-Accel-Buffering', 'no')
  res.flushHeaders()

  const encoder = new TextEncoder()
  const decoder = new TextDecoder('utf-8')
  let assistantText = ''
  const timeoutMs = toPositiveInt(process.env.DEEPSEEK_TIMEOUT_MS, 90000)
  const controller = new AbortController()
  const timeoutTimer = setTimeout(() => {
    controller.abort(new Error('DeepSeek 请求超时'))
  }, timeoutMs)
  let clientDisconnected = false

  req.on('close', () => {
    clientDisconnected = true
    controller.abort(new Error('客户端连接已关闭'))
  })

  try {
    const upstreamResp = await withModelSlot(() =>
      callDeepseekWithRetry({
        content,
        contextMessages,
        signal: controller.signal
      })
    )

    const reader = upstreamResp.body.getReader()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        break
      }

      if (clientDisconnected) {
        break
      }

      buffer += decoder.decode(value || encoder.encode(''), { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const rawLine of lines) {
        const line = rawLine.trim()
        if (!line || !line.startsWith('data:')) {
          continue
        }
        const dataPart = line.replace(/^data:\s*/, '')
        if (dataPart === '[DONE]') {
          continue
        }

        try {
          const parsed = JSON.parse(dataPart)
          const delta = parsed.choices?.[0]?.delta?.content || ''
          if (!delta) {
            continue
          }
          assistantText += delta
          res.write(`data: ${JSON.stringify({ delta })}\n\n`)
        } catch (error) {
          continue
        }
      }
    }

    const finalDb = readDb()
    if (assistantText || clientDisconnected) {
      finalDb.messages.push({
        id: newId(),
        conversationId: conversation.id,
        role: 'assistant',
        content: assistantText || '模型输出已中断。',
        createdAt: getNow()
      })
      const finalConversation = finalDb.conversations.find((item) => item.id === conversation.id)
      if (finalConversation) {
        finalConversation.updatedAt = getNow()
      }
      writeDb(finalDb)
    }

    clearTimeout(timeoutTimer)

    if (clientDisconnected) {
      return res.end()
    }

    res.write('data: [DONE]\n\n')
    return res.end()
  } catch (error) {
    clearTimeout(timeoutTimer)
    if (clientDisconnected) {
      return res.end()
    }

    const errorCode =
      error?.message?.includes('账号已在其他设备登录')
        ? 'SESSION_KICKED'
        : error.modelCode || 'STREAM_ERROR'
    const errorMessage =
      errorCode === 'SESSION_KICKED'
        ? '账号已在其他设备登录'
        : errorCode === 'MODEL_QUEUE_OVERFLOW'
          ? '当前请求较多，请稍后再试'
          : errorCode === 'MODEL_RATE_LIMIT'
            ? '当前请求较多，请稍后再试'
            : errorCode === 'MODEL_TIMEOUT'
              ? '响应超时，请重试'
              : errorCode === 'MODEL_INSUFFICIENT_BALANCE'
                ? '服务额度不足，请联系管理员'
        : error?.name === 'AbortError'
          ? '模型响应超时或连接中断，请重试'
          : error.message || '模型调用失败'
    res.write(`data: ${JSON.stringify({ code: errorCode, error: errorMessage })}\n\n`)
    res.write('data: [DONE]\n\n')
    return res.end()
  }
})

app.post('/api/admin/auth/login', (req, res) => {
  const { username, password } = req.body
  const db = readDb()
  const admin = db.users.find((item) => item.role === 'admin' && item.adminUsername === username)
  if (!admin) {
    return res.status(400).json({ message: '管理员账号不存在' })
  }
  if (!bcrypt.compareSync(password, admin.passwordHash)) {
    return res.status(400).json({ message: '管理员密码错误' })
  }

  admin.lastLoginAt = getNow()
  writeDb(db)

  const token = signToken({ adminId: admin.id, type: 'admin' })
  return sendOk(res, { token })
})

app.put('/api/admin/auth/password', adminAuth, (req, res) => {
  const { newPassword } = req.body
  if (!newPassword || String(newPassword).length < 6) {
    return res.status(400).json({ message: '新密码至少 6 位' })
  }

  const db = readDb()
  const admin = db.users.find((item) => item.id === req.admin.adminId && item.role === 'admin')
  if (!admin) {
    return res.status(404).json({ message: '管理员不存在' })
  }

  admin.passwordHash = bcrypt.hashSync(String(newPassword), 10)
  admin.lastLoginAt = getNow()
  writeDb(db)
  return sendOk(res, true)
})

app.get('/api/admin/dashboard', adminAuth, (req, res) => {
  const db = readDb()
  const today = dayjs().format('YYYY-MM-DD')
  const userCount = db.users.filter((item) => item.role === 'user').length
  const memberCount = db.users.filter(
    (item) => item.role === 'user' && item.memberExpireAt && dayjs(item.memberExpireAt).isAfter(dayjs())
  ).length
  const todayRedeemCount = db.redeemRecords.filter((item) => dayjs(item.activatedAt).format('YYYY-MM-DD') === today).length
  const conversationCount = db.conversations.length

  return sendOk(res, {
    userCount,
    memberCount,
    todayRedeemCount,
    conversationCount
  })
})

app.get('/api/admin/users', adminAuth, (req, res) => {
  const page = toPositiveInt(req.query.page, 1)
  const pageSize = Math.min(toPositiveInt(req.query.pageSize, 10), 100)
  const phoneKeyword = String(req.query.phone || '').trim()
  const status = String(req.query.status || '').trim()
  const db = readDb()
  let users = db.users
    .filter((item) => item.role === 'user')
    .map((item) => ({
      id: item.id,
      phone: item.phone,
      nickname: item.nickname,
      role: item.role,
      status: item.status,
      memberExpireAt: item.memberExpireAt,
      createdAt: item.createdAt
    }))

  if (phoneKeyword) {
    users = users.filter((item) => item.phone.includes(phoneKeyword))
  }

  if (status) {
    users = users.filter((item) => item.status === status)
  }

  users = users.sort((a, b) => dayjs(b.createdAt).valueOf() - dayjs(a.createdAt).valueOf())
  return sendOk(res, paginateList(users, page, pageSize))
})

app.put('/api/admin/users/:id', adminAuth, (req, res) => {
  const { nickname, phone } = req.body
  const db = readDb()
  const user = db.users.find((item) => item.id === req.params.id && item.role === 'user')

  if (!user) {
    return res.status(404).json({ message: '用户不存在' })
  }

  if (phone && !/^1\d{10}$/.test(phone)) {
    return res.status(400).json({ message: '手机号格式不正确' })
  }

  if (phone && phone !== user.phone) {
    const existed = db.users.find((item) => item.phone === phone)
    if (existed) {
      return res.status(400).json({ message: '手机号已存在' })
    }
    user.phone = phone
  }

  if (nickname && String(nickname).trim()) {
    user.nickname = String(nickname).trim()
  }

  writeDb(db)
  return sendOk(res, true)
})

app.put('/api/admin/users/:id/status', adminAuth, (req, res) => {
  const { status } = req.body
  if (!['active', 'disabled'].includes(status)) {
    return res.status(400).json({ message: '状态参数错误' })
  }

  const db = readDb()
  const user = db.users.find((item) => item.id === req.params.id && item.role === 'user')
  if (!user) {
    return res.status(404).json({ message: '用户不存在' })
  }

  user.status = status
  writeDb(db)
  return sendOk(res, true)
})

app.put('/api/admin/users/:id/reset-password', adminAuth, (req, res) => {
  const { newPassword } = req.body
  if (!newPassword || String(newPassword).length < 6) {
    return res.status(400).json({ message: '新密码至少 6 位' })
  }

  const db = readDb()
  const user = db.users.find((item) => item.id === req.params.id && item.role === 'user')
  if (!user) {
    return res.status(404).json({ message: '用户不存在' })
  }

  user.passwordHash = bcrypt.hashSync(String(newPassword), 10)
  writeDb(db)
  return sendOk(res, true)
})

app.put('/api/admin/users/:id/member-expire-at', adminAuth, (req, res) => {
  const db = readDb()
  const user = db.users.find((item) => item.id === req.params.id && item.role === 'user')

  if (!user) {
    return res.status(404).json({ message: '用户不存在' })
  }

  try {
    user.memberExpireAt = normalizeAdminMemberExpireAtInput(req.body.memberExpireAt)
  } catch (error) {
    return res.status(400).json({ message: error.message || '会员到期时间格式不正确' })
  }

  writeDb(db)
  return sendOk(res, toSafeUser(user))
})

app.get('/api/admin/roles', adminAuth, (req, res) => {
  const db = readDb()
  return sendOk(res, db.roles)
})

app.get('/api/admin/menus', adminAuth, (req, res) => {
  const db = readDb()
  return sendOk(res, db.menus)
})

app.get('/api/admin/members', adminAuth, (req, res) => {
  const db = readDb()
  const members = db.users
    .filter((item) => item.role === 'user' && item.memberExpireAt)
    .map((item) => ({
      id: item.id,
      phone: item.phone,
      nickname: item.nickname,
      memberExpireAt: item.memberExpireAt
    }))
  return sendOk(res, members)
})

app.get('/api/admin/redeem-codes', adminAuth, (req, res) => {
  const page = toPositiveInt(req.query.page, 1)
  const pageSize = Math.min(toPositiveInt(req.query.pageSize, 10), 100)
  const status = String(req.query.status || '').trim()
  const codeKeyword = String(req.query.code || '').trim().toUpperCase()
  const db = readDb()
  const list = buildRedeemCodeList(db, { status, codeKeyword })
  return sendOk(res, paginateList(list, page, pageSize))
})

app.get('/api/admin/redeem-codes/export', adminAuth, (req, res) => {
  const status = String(req.query.status || '').trim()
  const codeKeyword = String(req.query.code || '').trim().toUpperCase()
  const db = readDb()
  const list = buildRedeemCodeList(db, { status, codeKeyword })
  return sendOk(res, list)
})

app.post('/api/admin/redeem-codes/batch', adminAuth, (req, res) => {
  const { quantity = 5, durationMonths = 1 } = req.body
  const db = readDb()
  const maxCount = Math.min(Number(quantity) || 1, 200)
  const duration = Number(durationMonths) || 1
  const now = getNow()

  const created = []
  for (let i = 0; i < maxCount; i += 1) {
    const code = `VIP-${dayjs().format('YYYYMMDD')}-${Math.random().toString(36).slice(2, 8).toUpperCase()}`
    const item = {
      id: newId(),
      code,
      durationMonths: duration,
      status: 'unused',
      createdAt: now,
      usedAt: null,
      usedByUserId: null
    }
    db.redeemCodes.push(item)
    created.push(item)
  }

  writeDb(db)
  return sendOk(res, created)
})

app.put('/api/admin/redeem-codes/:id/void', adminAuth, (req, res) => {
  const db = readDb()
  const redeemCode = db.redeemCodes.find((item) => item.id === req.params.id)
  if (!redeemCode) {
    return res.status(404).json({ message: '兑换码不存在' })
  }

  if (redeemCode.status !== 'unused') {
    return res.status(400).json({ message: '仅未使用的兑换码可作废' })
  }

  redeemCode.status = 'void'
  writeDb(db)
  return sendOk(res, true)
})

app.get('/api/admin/redeem-records', adminAuth, (req, res) => {
  const page = toPositiveInt(req.query.page, 1)
  const pageSize = Math.min(toPositiveInt(req.query.pageSize, 10), 100)
  const phoneKeyword = String(req.query.phone || '').trim()
  const start = String(req.query.start || '').trim()
  const end = String(req.query.end || '').trim()
  const db = readDb()
  let records = [...db.redeemRecords]

  if (phoneKeyword) {
    records = records.filter((item) => item.phone.includes(phoneKeyword))
  }

  if (start) {
    records = records.filter((item) => dayjs(item.activatedAt).isAfter(dayjs(start).subtract(1, 'day')))
  }

  if (end) {
    records = records.filter((item) => dayjs(item.activatedAt).isBefore(dayjs(end).add(1, 'day')))
  }

  records = records.sort((a, b) => dayjs(b.activatedAt).valueOf() - dayjs(a.activatedAt).valueOf())
  return sendOk(res, paginateList(records, page, pageSize))
})

app.get('/api/admin/conversations', adminAuth, (req, res) => {
  const page = toPositiveInt(req.query.page, 1)
  const pageSize = Math.min(toPositiveInt(req.query.pageSize, 10), 100)
  const phoneKeyword = String(req.query.phone || '').trim()
  const keyword = String(req.query.keyword || '').trim()
  const search = String(req.query.search || '').trim()
  const db = readDb()
  let list = db.conversations.map((item) => {
    const user = db.users.find((u) => u.id === item.userId)
    return {
      ...item,
      userPhone: user?.phone || ''
    }
  })

  if (phoneKeyword) {
    list = list.filter((item) => item.userPhone.includes(phoneKeyword))
  }

  if (keyword) {
    list = list.filter((item) => item.title.includes(keyword))
  }

  if (search) {
    const lowered = search.toLowerCase()
    list = list.filter((item) => {
      const titleHit = String(item.title || '').toLowerCase().includes(lowered)
      const phoneHit = String(item.userPhone || '').toLowerCase().includes(lowered)
      if (titleHit || phoneHit) {
        return true
      }

      return db.messages.some(
        (msg) =>
          msg.conversationId === item.id &&
          String(msg.content || '').toLowerCase().includes(lowered)
      )
    })
  }

  list = list.sort((a, b) => dayjs(b.updatedAt).valueOf() - dayjs(a.updatedAt).valueOf())
  return sendOk(res, paginateList(list, page, pageSize))
})

app.get('/api/admin/conversations/:id/messages', adminAuth, (req, res) => {
  const db = readDb()
  const messages = db.messages.filter((item) => item.conversationId === req.params.id)
  return sendOk(res, messages)
})

app.listen(port, () => {
  console.log(`backend server running at http://localhost:${port}`)
})
