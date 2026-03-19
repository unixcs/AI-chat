const fs = require('fs')
const path = require('path')
const bcrypt = require('bcryptjs')
const { DatabaseSync } = require('node:sqlite')

const jsonPath = path.join(__dirname, 'data.json')
const sqlitePath = path.join(__dirname, 'data.sqlite')
const sqliteDb = new DatabaseSync(sqlitePath)

sqliteDb.exec('PRAGMA journal_mode = WAL;')
sqliteDb.exec('PRAGMA foreign_keys = OFF;')

const newId = () => {
  return `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

const getNow = () => {
  return new Date().toISOString()
}

const addMonths = (dateValue, months) => {
  const date = new Date(dateValue)
  date.setMonth(date.getMonth() + months)
  return date.toISOString()
}

const createDefaultData = () => {
  const now = getNow()
  const adminRoleId = 'role_admin'
  const userRoleId = 'role_user'

  return {
    users: [
      {
        id: newId(),
        phone: '15555166986',
        nickname: 'demo用户',
        passwordHash: bcrypt.hashSync('simple1270.', 10),
        avatarUrl: '',
        status: 'active',
        role: 'user',
        currentSessionId: null,
        sessionUpdatedAt: null,
        memberExpireAt: addMonths(now, 1),
        createdAt: now,
        lastLoginAt: now
      },
      {
        id: newId(),
        phone: '10000000000',
        nickname: '系统管理员',
        passwordHash: bcrypt.hashSync('admin123', 10),
        avatarUrl: '',
        status: 'active',
        role: 'admin',
        memberExpireAt: null,
        createdAt: now,
        lastLoginAt: now,
        adminUsername: 'admin'
      }
    ],
    roles: [
      { id: adminRoleId, name: 'admin', description: '系统管理员' },
      { id: userRoleId, name: 'user', description: '普通用户' }
    ],
    menus: [
      { id: newId(), name: '控制台', path: '/admin', group: '首页' },
      { id: newId(), name: '用户管理', path: '/admin/users', group: '系统管理' },
      { id: newId(), name: '角色管理', path: '/admin/roles', group: '系统管理' },
      { id: newId(), name: '菜单管理', path: '/admin/menus', group: '系统管理' },
      { id: newId(), name: '会员管理', path: '/admin/members', group: '业务管理' },
      { id: newId(), name: '兑换码管理', path: '/admin/redeem-codes', group: '业务管理' },
      { id: newId(), name: '兑换记录', path: '/admin/redeem-records', group: '业务管理' },
      { id: newId(), name: '会话管理', path: '/admin/conversations', group: '业务管理' }
    ],
    redeemCodes: [
      {
        id: newId(),
        code: 'VIP-2026-DEMO-0001',
        durationMonths: 1,
        status: 'unused',
        createdAt: now,
        usedAt: null,
        usedByUserId: null
      }
    ],
    redeemRecords: [],
    conversations: [],
    messages: []
  }
}

const initSchema = () => {
  sqliteDb.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      phone TEXT,
      nickname TEXT,
      passwordHash TEXT,
      avatarUrl TEXT,
      status TEXT,
      role TEXT,
      currentSessionId TEXT,
      sessionUpdatedAt TEXT,
      memberExpireAt TEXT,
      createdAt TEXT,
      lastLoginAt TEXT,
      adminUsername TEXT
    );

    CREATE TABLE IF NOT EXISTS roles (
      id TEXT PRIMARY KEY,
      name TEXT,
      description TEXT
    );

    CREATE TABLE IF NOT EXISTS menus (
      id TEXT PRIMARY KEY,
      name TEXT,
      path TEXT,
      menuGroup TEXT
    );

    CREATE TABLE IF NOT EXISTS redeemCodes (
      id TEXT PRIMARY KEY,
      code TEXT,
      durationMonths INTEGER,
      status TEXT,
      createdAt TEXT,
      usedAt TEXT,
      usedByUserId TEXT
    );

    CREATE TABLE IF NOT EXISTS redeemRecords (
      id TEXT PRIMARY KEY,
      userId TEXT,
      phone TEXT,
      code TEXT,
      activatedAt TEXT,
      beforeExpireAt TEXT,
      afterExpireAt TEXT
    );

    CREATE TABLE IF NOT EXISTS conversations (
      id TEXT PRIMARY KEY,
      userId TEXT,
      title TEXT,
      createdAt TEXT,
      updatedAt TEXT
    );

    CREATE TABLE IF NOT EXISTS messages (
      id TEXT PRIMARY KEY,
      conversationId TEXT,
      role TEXT,
      content TEXT,
      createdAt TEXT
    );
  `)
}

const insertAll = (data) => {
  const insertUser = sqliteDb.prepare(`
    INSERT INTO users (
      id, phone, nickname, passwordHash, avatarUrl, status, role,
      currentSessionId, sessionUpdatedAt, memberExpireAt,
      createdAt, lastLoginAt, adminUsername
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `)

  const insertRole = sqliteDb.prepare('INSERT INTO roles (id, name, description) VALUES (?, ?, ?)')
  const insertMenu = sqliteDb.prepare('INSERT INTO menus (id, name, path, menuGroup) VALUES (?, ?, ?, ?)')
  const insertRedeemCode = sqliteDb.prepare(`
    INSERT INTO redeemCodes (id, code, durationMonths, status, createdAt, usedAt, usedByUserId)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `)
  const insertRedeemRecord = sqliteDb.prepare(`
    INSERT INTO redeemRecords (id, userId, phone, code, activatedAt, beforeExpireAt, afterExpireAt)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `)
  const insertConversation = sqliteDb.prepare(`
    INSERT INTO conversations (id, userId, title, createdAt, updatedAt)
    VALUES (?, ?, ?, ?, ?)
  `)
  const insertMessage = sqliteDb.prepare(`
    INSERT INTO messages (id, conversationId, role, content, createdAt)
    VALUES (?, ?, ?, ?, ?)
  `)

  for (const user of data.users || []) {
    insertUser.run(
      user.id,
      user.phone || '',
      user.nickname || '',
      user.passwordHash || '',
      user.avatarUrl || '',
      user.status || 'active',
      user.role || 'user',
      user.currentSessionId || null,
      user.sessionUpdatedAt || null,
      user.memberExpireAt || null,
      user.createdAt || getNow(),
      user.lastLoginAt || getNow(),
      user.adminUsername || null
    )
  }

  for (const role of data.roles || []) {
    insertRole.run(role.id, role.name || '', role.description || '')
  }

  for (const menu of data.menus || []) {
    insertMenu.run(menu.id, menu.name || '', menu.path || '', menu.group || '')
  }

  for (const item of data.redeemCodes || []) {
    insertRedeemCode.run(
      item.id,
      item.code || '',
      Number(item.durationMonths || 1),
      item.status || 'unused',
      item.createdAt || getNow(),
      item.usedAt || null,
      item.usedByUserId || null
    )
  }

  for (const item of data.redeemRecords || []) {
    insertRedeemRecord.run(
      item.id,
      item.userId || null,
      item.phone || '',
      item.code || '',
      item.activatedAt || getNow(),
      item.beforeExpireAt || null,
      item.afterExpireAt || null
    )
  }

  for (const item of data.conversations || []) {
    insertConversation.run(
      item.id,
      item.userId || null,
      item.title || '',
      item.createdAt || getNow(),
      item.updatedAt || getNow()
    )
  }

  for (const item of data.messages || []) {
    insertMessage.run(
      item.id,
      item.conversationId || null,
      item.role || 'assistant',
      item.content || '',
      item.createdAt || getNow()
    )
  }
}

const clearAll = () => {
  sqliteDb.exec(`
    DELETE FROM users;
    DELETE FROM roles;
    DELETE FROM menus;
    DELETE FROM redeemCodes;
    DELETE FROM redeemRecords;
    DELETE FROM conversations;
    DELETE FROM messages;
  `)
}

const runInTransaction = (task) => {
  sqliteDb.exec('BEGIN')
  try {
    task()
    sqliteDb.exec('COMMIT')
  } catch (error) {
    try {
      sqliteDb.exec('ROLLBACK')
    } catch (rollbackError) {
    }
    throw error
  }
}

const ensureSeedData = () => {
  const countRow = sqliteDb.prepare('SELECT COUNT(*) as count FROM users').get()
  if (countRow.count > 0) {
    return
  }

  let initData = null
  if (fs.existsSync(jsonPath)) {
    try {
      initData = JSON.parse(fs.readFileSync(jsonPath, 'utf8'))
    } catch (error) {
      initData = null
    }
  }

  if (!initData) {
    initData = createDefaultData()
  }

  runInTransaction(() => {
    clearAll()
    insertAll(initData)
  })
}

const readDb = () => {
  const users = sqliteDb.prepare('SELECT * FROM users').all().map((item) => ({
    ...item,
    currentSessionId: item.currentSessionId || null,
    sessionUpdatedAt: item.sessionUpdatedAt || null,
    memberExpireAt: item.memberExpireAt || null,
    adminUsername: item.adminUsername || null
  }))

  const roles = sqliteDb.prepare('SELECT * FROM roles').all()
  const menus = sqliteDb.prepare('SELECT id, name, path, menuGroup as "group" FROM menus').all()
  const redeemCodes = sqliteDb.prepare('SELECT * FROM redeemCodes').all()
  const redeemRecords = sqliteDb.prepare('SELECT * FROM redeemRecords').all()
  const conversations = sqliteDb.prepare('SELECT * FROM conversations').all()
  const messages = sqliteDb.prepare('SELECT * FROM messages').all()

  return {
    users,
    roles,
    menus,
    redeemCodes,
    redeemRecords,
    conversations,
    messages
  }
}

const writeDb = (data) => {
  runInTransaction(() => {
    clearAll()
    insertAll(data)
  })
}

initSchema()
ensureSeedData()

module.exports = {
  readDb,
  writeDb,
  newId,
  getNow,
  addMonths
}
