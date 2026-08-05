const test = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const os = require('node:os')

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-chat-db-test-'))
process.env.SQLITE_PATH = path.join(tempDir, 'test.sqlite')

const {
  readDb,
  writeDb,
  newId,
  getNow,
  insertMessage,
  getMessagesByConversation,
  updateConversationTimestamp,
  getUserById,
  findConversationIdsByMessageContent
} = require('../db')

test.after(() => {
  fs.rmSync(tempDir, { recursive: true, force: true })
})

test('readDb seeds admin user and returns empty messages array', () => {
  const db = readDb()
  assert.equal(Array.isArray(db.messages), true)
  assert.equal(db.messages.length, 0)
  assert.equal(db.users.some((item) => item.role === 'admin'), true)
})

test('writeDb preserves messages table (incremental separation)', () => {
  const conversationId = newId()

  const db = readDb()
  db.conversations.push({
    id: conversationId,
    userId: 'user_demo',
    title: '测试会话',
    createdAt: getNow(),
    updatedAt: getNow()
  })
  writeDb(db)

  insertMessage({
    id: newId(),
    conversationId,
    role: 'user',
    content: '你好，这是第一条消息',
    createdAt: getNow()
  })
  insertMessage({
    id: newId(),
    conversationId,
    role: 'assistant',
    content: '你好，这是回复',
    createdAt: getNow()
  })

  const before = getMessagesByConversation(conversationId)
  assert.equal(before.length, 2)

  const dbAfter = readDb()
  dbAfter.users[0].nickname = '改名后的管理员'
  writeDb(dbAfter)

  const after = getMessagesByConversation(conversationId)
  assert.equal(after.length, 2)
  assert.equal(after[0].content, '你好，这是第一条消息')
  assert.equal(after[1].role, 'assistant')
})

test('updateConversationTimestamp updates only the target row', () => {
  const db = readDb()
  const conversation = db.conversations[0]
  const future = '2099-01-01T00:00:00.000Z'

  updateConversationTimestamp(conversation.id, future)

  const updated = readDb().conversations.find((item) => item.id === conversation.id)
  assert.equal(updated.updatedAt, future)
})

test('getMessagesByConversation returns messages ordered by createdAt', () => {
  const db = readDb()
  const conversation = db.conversations[0]
  const messages = getMessagesByConversation(conversation.id)

  for (let i = 1; i < messages.length; i += 1) {
    assert.ok(messages[i - 1].createdAt <= messages[i].createdAt)
  }
})

test('findConversationIdsByMessageContent matches content and rejects misses', () => {
  const db = readDb()
  const conversation = db.conversations[0]

  const hits = findConversationIdsByMessageContent('第一条消息')
  assert.equal(hits.has(conversation.id), true)

  const misses = findConversationIdsByMessageContent('完全不存在的关键词')
  assert.equal(misses.has(conversation.id), false)
})

test('getUserById returns normalized user or null', () => {
  const db = readDb()
  const admin = db.users.find((item) => item.role === 'admin')

  const found = getUserById(admin.id)
  assert.equal(found.id, admin.id)
  assert.equal(found.role, 'admin')

  assert.equal(getUserById('not-exist-id'), null)
})
