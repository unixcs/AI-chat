import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const readProjectFile = (relativePath) => {
  return readFileSync(resolve(root, relativePath), 'utf8')
}

// Part D 布局回归锁：换装后这些交互/布局契约必须继续成立

test('landing CTA stays centered on the hero card', () => {
  const vue = readProjectFile('src/views/public/LandingView.vue')

  assert.match(vue, /place-items-center/)
  assert.match(vue, /text-center/)
  assert.match(vue, /mx-auto|items-center/)
})

test('auth pages keep a single centered card via AuthShell', () => {
  for (const f of ['src/views/public/LoginView.vue', 'src/views/public/RegisterView.vue']) {
    const vue = readProjectFile(f)
    assert.match(vue, /AuthShell/, f)
    assert.doesNotMatch(vue, /authHeroPanel/, f)
  }
  const shell = readProjectFile('src/components/layout/AuthShell.vue')
  assert.match(shell, /place-items-center/)
  assert.match(shell, /max-w-\[440px\]|maxWidth/)
})

test('chat page keeps the composer pinned: full-height shell, scrollable viewport, non-shrinking composer', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /class="[^"]*flex h-full min-h-0 flex-col/)
  assert.match(vue, /min-h-0 flex-1 overflow-y-auto/)
  assert.match(vue, /@scroll="onMessageScroll"/)
  // composer 区不参与压缩
  assert.match(vue, /shrink-0 border-t border-border bg-card/)
})

test('chat composer keeps auto-height JS contract and mobile caps', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  // syncComposerHeight 的行为前提：JS 直接改 style.height，CSS 不得用 !important 覆盖
  assert.match(vue, /el\.style\.height = 'auto'/)
  assert.match(vue, /mobileMaxHeight = 96/)
  assert.match(vue, /resize: none;/)
})

test('chat composer respects iOS safe area at the bottom', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /env\(safe-area-inset-bottom\)/)
})

test('message bubbles keep selection contrast via role-scoped highlight variables', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')
  const css = readProjectFile('src/style.css')

  // 两类气泡的 ::selection 规则存在（此前缺失导致绿底白字选中不可见）
  assert.match(vue, /\.chatUserBubble::selection,/)
  assert.match(vue, /\.chatUserBubble \*::selection \{/)
  assert.match(vue, /\.chatBotBubble::selection,/)
  assert.match(vue, /\.chatBotBubble \*::selection \{/)
  // 取值走 chat-*-selection-* 变量
  assert.match(vue, /background: var\(--chat-user-selection-bg\)/)
  assert.match(vue, /background: var\(--chat-bot-selection-bg\)/)
  // 浅色取值 + .dark 覆盖（此前 .dark 无 chat-* 覆盖，AI 气泡深色下过亮）
  assert.match(css, /--chat-user-selection-bg: #f8fafc/)
  assert.match(css, /--chat-bot-selection-bg: #2f6340/)
  assert.match(css, /--chat-user-bg: linear-gradient\(135deg, #356a42/)
  assert.match(css, /--chat-bot-bg: rgba\(32, 39, 34, 0\.88\)/)
})

test('profile and redeem forms keep independent notice states (no cross-card leaks)', () => {
  const vue = readProjectFile('src/views/user/ProfileView.vue')

  // 共用的 noticeText/errorText 已拆为每卡独立状态
  assert.match(vue, /profileError/)
  assert.match(vue, /profileNotice/)
  assert.match(vue, /redeemError/)
  assert.match(vue, /redeemNotice/)
  assert.doesNotMatch(vue, /noticeText|errorText/)
})

test('mobile navigation uses left Sheet in both shells', () => {
  for (const f of ['src/views/user/UserLayout.vue', 'src/views/admin/AdminLayout.vue']) {
    const vue = readProjectFile(f)
    assert.match(vue, /SheetContent side="left"/, f)
    assert.match(vue, /md:hidden/, f)
  }
})

test('desktop navigation is a persistent sidebar hidden on mobile', () => {
  for (const f of ['src/views/user/UserLayout.vue', 'src/views/admin/AdminLayout.vue']) {
    const vue = readProjectFile(f)
    assert.match(vue, /hidden h-dvh w-6[48]|hidden h-dvh w-72/, f)
    assert.match(vue, /md:flex/, f)
  }
})

test('admin user table switches to card list under md breakpoint', () => {
  const vue = readProjectFile('src/views/admin/AdminUserView.vue')

  assert.match(vue, /hidden overflow-hidden rounded-xl border border-border md:block/)
  assert.match(vue, /grid gap-2\.5 md:hidden/)
})

test('member expire dialog keeps datetime-local precision and backend error surfacing', () => {
  const vue = readProjectFile('src/views/admin/AdminUserView.vue')

  assert.match(vue, /type="datetime-local"/)
  assert.match(vue, /step="1"/)
  assert.match(vue, /catch \(error\)\s*\{[\s\S]*?errorText\.value = error\.response\?\.data\?\.message/)
  assert.match(vue, /errorText[\s\S]*?variant="destructive"/)
})
