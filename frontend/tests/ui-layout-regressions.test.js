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

test('admin data tables avoid oversized fixed minimum widths on mobile-critical pages', () => {
  const userVue = readProjectFile('src/views/admin/AdminUserView.vue')
  const redeemCodeVue = readProjectFile('src/views/admin/AdminRedeemCodeView.vue')
  const redeemRecordVue = readProjectFile('src/views/admin/AdminRedeemRecordView.vue')

  assert.doesNotMatch(userVue, /min-width:\s*980px;/)
  assert.doesNotMatch(redeemCodeVue, /min-width:\s*860px;/)
  assert.doesNotMatch(redeemRecordVue, /min-width:\s*860px;/)
})

test('admin conversation page collapses to a single narrow-safe column on small screens', () => {
  const vue = readProjectFile('src/views/admin/AdminConversationView.vue')

  assert.match(vue, /@media \(max-width:\s*980px\)\s*\{[\s\S]*\.conversationPanel\s*\{[\s\S]*grid-template-columns:\s*1fr;/)
  assert.match(vue, /\.conversationHistoryCard,[\s\S]*\.transcriptPanel\s*\{[\s\S]*min-width:\s*0;/)
})
