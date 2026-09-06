import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const readProjectFile = (relativePath) => {
  return readFileSync(resolve(root, relativePath), 'utf8')
}

test('landing CTA layout is explicitly centered', () => {
  const vue = readProjectFile('src/views/public/LandingView.vue')

  assert.match(vue, /\.landingActions\s*\{[\s\S]*justify-content:\s*center;/)
  assert.match(vue, /\.landingPanel\s*\{[\s\S]*text-align:\s*center;/)
})

test('login page no longer keeps the empty left hero column and centers auth card', () => {
  const vue = readProjectFile('src/views/public/LoginView.vue')

  assert.doesNotMatch(vue, /<article class="authHeroPanel"><\/article>/)
  assert.match(vue, /\.authGrid\s*\{[\s\S]*display:\s*flex;[\s\S]*justify-content:\s*center;/)
  assert.match(vue, /\.authSurface\s*\{[\s\S]*margin:\s*0 auto;/)
})

test('register page matches the centered auth layout without left hero column', () => {
  const vue = readProjectFile('src/views/public/RegisterView.vue')

  assert.doesNotMatch(vue, /<article class="authHeroPanel"><\/article>/)
  assert.match(vue, /\.authGrid\s*\{[\s\S]*display:\s*flex;[\s\S]*justify-content:\s*center;/)
  assert.match(vue, /\.authSurface\s*\{[\s\S]*margin:\s*0 auto;/)
})

test('chat page keeps the composer pinned by using a full-height shell and scrollable viewport', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /\.chatStage\s*\{[\s\S]*height:\s*calc\(100dvh - 44px\);/)
  assert.match(vue, /\.chatPanel\s*\{[\s\S]*height:\s*100%;[\s\S]*min-height:\s*0;/)
  assert.match(vue, /\.messageViewport\s*\{[\s\S]*min-height:\s*0;[\s\S]*overflow:\s*auto;/)
  assert.match(vue, /\.composerShell\s*\{[\s\S]*flex:\s*0 0 auto;/)
})

test('user mobile sidebar is capped to roughly one third of the screen width', () => {
  const vue = readProjectFile('src/views/user/UserLayout.vue')

  assert.match(vue, /@media \(max-width:\s*960px\)\s*\{[\s\S]*\.userSidebar\s*\{[\s\S]*width:\s*clamp\(96px,\s*33\.333vw,\s*140px\);/)
  assert.doesNotMatch(vue, /@media \(max-width:\s*960px\)\s*\{[\s\S]*\.userSidebar\s*\{[\s\S]*width:\s*min\(320px,\s*calc\(100vw - 28px\)\);/)
})

test('user mobile sidebar menu removes descriptions and uses compact typography', () => {
  const vue = readProjectFile('src/views/user/UserLayout.vue')

  assert.doesNotMatch(vue, /<small>\{\{ item\.desc \}\}<\/small>/)
  assert.match(vue, /@media \(max-width:\s*960px\)\s*\{[\s\S]*\.sidebarNav\s*\{[\s\S]*gap:\s*8px;/)
  assert.match(vue, /@media \(max-width:\s*960px\)\s*\{[\s\S]*\.menuBtn\s*\{[\s\S]*padding:\s*12px 10px;[\s\S]*border-radius:\s*16px;/)
  assert.match(vue, /@media \(max-width:\s*960px\)\s*\{[\s\S]*\.menuBtnTitle\s*\{[\s\S]*font-size:\s*14px;/)
})

test('user sidebar logout button splits its label across two lines', () => {
  const vue = readProjectFile('src/views/user/UserLayout.vue')

  assert.match(vue, /<button class="ghostBtn logoutBtn" @click="logout">\s*退出<br ?\/?>登录\s*<\/button>/)
})

test('admin data tables avoid oversized fixed minimum widths on mobile-critical pages', () => {
  const userVue = readProjectFile('src/views/admin/AdminUserView.vue')
  const redeemCodeVue = readProjectFile('src/views/admin/AdminRedeemCodeView.vue')
  const redeemRecordVue = readProjectFile('src/views/admin/AdminRedeemRecordView.vue')

  assert.doesNotMatch(userVue, /min-width:\s*980px;/)
  assert.doesNotMatch(redeemCodeVue, /min-width:\s*860px;/)
  assert.doesNotMatch(redeemRecordVue, /min-width:\s*860px;/)
})

test('admin conversation page keeps a left-list/right-detail split on small screens', () => {
  const vue = readProjectFile('src/views/admin/AdminConversationView.vue')

  // 2026-09 验收反馈：手机端不再竖向堆叠，改为左列表(5fr)右详情(7fr)分栏
  assert.match(vue, /@media \(max-width:\s*980px\)\s*\{[\s\S]*\.conversationPanel\s*\{[\s\S]*grid-template-columns:\s*minmax\(0,\s*5fr\)\s*minmax\(0,\s*7fr\);/)
  assert.match(vue, /\.conversationHistoryCard,[\s\S]*\.transcriptPanel\s*\{[\s\S]*min-width:\s*0;/)
})

test('admin user and announcement tables switch to labelled cards on phones', () => {
  const userVue = readProjectFile('src/views/admin/AdminUserView.vue')
  const announcementVue = readProjectFile('src/views/admin/AdminAnnouncementView.vue')

  for (const vue of [userVue, announcementVue]) {
    assert.match(vue, /@media \(max-width:\s*640px\)\s*\{[\s\S]*\.tableWrap thead\s*\{[\s\S]*display:\s*none;/)
    assert.match(vue, /\.tableWrap td::before\s*\{[\s\S]*content:\s*attr\(data-label\);/)
  }
  assert.match(userVue, /data-label="手机号"/)
  assert.match(announcementVue, /data-label="标题"/)
})

test('member expire dialog surfaces backend errors instead of dying silently', () => {
  const vue = readProjectFile('src/views/admin/AdminUserView.vue')

  assert.match(vue, /catch \(error\)\s*\{[\s\S]*?errorText\.value = error\.response\?\.data\?\.message/)
  assert.match(vue, /v-if="errorText" class="dangerText memberExpireError"/)
})
