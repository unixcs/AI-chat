import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const readProjectFile = (relativePath) => {
  return readFileSync(resolve(root, relativePath), 'utf8')
}

// 文案回归锁：换装后用户可见的既有文案保持不变（不做营销话术膨胀）

test('landing page keeps the original single CTA copy without extra marketing text', () => {
  const vue = readProjectFile('src/views/public/LandingView.vue')

  assert.match(vue, /开启灵感/)
  assert.doesNotMatch(vue, /高级、安静、现代的 AI 对话入口/)
})

test('login page keeps the original login copy', () => {
  const vue = readProjectFile('src/views/public/LoginView.vue')

  assert.match(vue, /欢迎登陆 Thallo/)
  assert.match(vue, /使用手机号密码登录以开始解读/)
  assert.doesNotMatch(vue, /把每一次提问，都放进更安静的界面里/)
})

test('register page keeps the original registration copy', () => {
  const vue = readProjectFile('src/views/public/RegisterView.vue')

  assert.match(vue, /欢迎加入/)
  assert.match(vue, /注册账号/)
  assert.doesNotMatch(vue, /创建你的专属入口/)
})

test('chat page keeps empty-state, placeholder and streaming copy', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /告诉我你有什么想法/)
  assert.match(vue, /开始新的对话吧/)
  assert.match(vue, /把你此刻最想问的内容写下来\.\.\./)
  assert.match(vue, /灵感开启中，请稍等片刻\.\.\./)
  assert.doesNotMatch(vue, /在安静而清晰的空间里继续对话/)
})

test('chat copy button and history/announcement copy stay intact', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  // 复制按钮两态
  assert.match(vue, /\?\s*'已复制'\s*:\s*'复制'/)
  assert.match(vue, /'已复制'\s*:\s*'复制消息'/)
  // 历史与公告
  assert.match(vue, /历史对话/)
  assert.match(vue, /我知道了/)
  // 删除确认（新 AlertDialog 文案）
  assert.match(vue, /删除这段对话？/)
  assert.match(vue, /删除后对话与其中全部消息不可恢复。/)
})

test('script-owned user-facing error copy is untouched', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /最多只能占卜 1,000 字以内哦~/)
  assert.match(vue, /会员过期，请续费后使用。/)
  assert.match(vue, /复制失败，请长按文本手动复制/)
})

test('admin prompt page keeps hint copy explaining the assembly chain', () => {
  const vue = readProjectFile('src/views/admin/AdminPromptView.vue')

  assert.match(vue, /修改保存后无需重启，下一个新聊天立即使用最新提示词/)
  assert.match(vue, /基础内置 Prompt \+ 对应的三张卡片文案/)
  assert.match(vue, /已预置默认文案/)
})
