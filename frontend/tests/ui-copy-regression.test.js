import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const readProjectFile = (relativePath) => {
  return readFileSync(resolve(root, relativePath), 'utf8')
}

test('landing page keeps the original single CTA copy without extra marketing text', () => {
  const vue = readProjectFile('src/views/public/LandingView.vue')

  assert.match(vue, />开启灵感</)
  assert.doesNotMatch(vue, /高级、安静、现代的 AI 对话入口/)
})

test('login page keeps the original login copy only', () => {
  const vue = readProjectFile('src/views/public/LoginView.vue')

  assert.match(vue, /欢迎登陆Thallo 🔮/)
  assert.match(vue, /使用手机号密码登录以开始解读/)
  assert.doesNotMatch(vue, /把每一次提问，都放进更安静的界面里/)
})

test('register page keeps the original registration copy only', () => {
  const vue = readProjectFile('src/views/public/RegisterView.vue')

  assert.match(vue, /欢迎加入/)
  assert.match(vue, /注册账号/)
  assert.doesNotMatch(vue, /创建你的专属入口/)
})

test('chat page does not contain the added hero copy', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /开始新的对话吧/)
  assert.doesNotMatch(vue, /在安静而清晰的空间里继续对话/)
  assert.doesNotMatch(vue, /保留原有聊天逻辑与历史能力/)
})
