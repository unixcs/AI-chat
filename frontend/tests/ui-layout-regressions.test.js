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

test('copy button keeps a >=44px touch target on messages', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /min-h-11[^"]*'?\s*cursor-pointer/)
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
