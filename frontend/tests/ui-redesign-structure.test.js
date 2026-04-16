import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const readProjectFile = (relativePath) => {
  return readFileSync(resolve(root, relativePath), 'utf8')
}

test('global style sheet exposes the new mist design tokens and shared shells', () => {
  const css = readProjectFile('src/style.css')

  assert.match(css, /--surface-elevated:/)
  assert.match(css, /--glass-border:/)
  assert.match(css, /--radius-xl:/)
  assert.match(css, /\.panelShell\s*\{/) 
  assert.match(css, /\.softInput\s*\{/) 
})

test('user layout defines the new workspace shell and sidebar sections', () => {
  const vue = readProjectFile('src/views/user/UserLayout.vue')

  assert.match(vue, /class="userWorkspaceShell pageWrap"/)
  assert.match(vue, /class="userSidebar card"/)
  assert.match(vue, /class="sidebarFooter"/)
  assert.match(vue, /class="mobileWorkspaceBar card"/)
})

test('chat view exposes the redesigned stage, composer shell, and history panel', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /class="chatStage"/)
  assert.match(vue, /class="messageViewport"/)
  assert.match(vue, /class="composerShell"/)
  assert.match(vue, /class="historyDrawerPanel card"/)
})

test('public auth pages adopt the new hero and surface structure', () => {
  const loginVue = readProjectFile('src/views/public/LoginView.vue')
  const registerVue = readProjectFile('src/views/public/RegisterView.vue')

  assert.match(loginVue, /class="authShell pageWrap"/)
  assert.match(loginVue, /class="authSurface card"/)
  assert.doesNotMatch(loginVue, /class="authHeroPanel"/)
  assert.match(registerVue, /class="authShell pageWrap"/)
  assert.match(registerVue, /class="authSurface card"/)
  assert.doesNotMatch(registerVue, /class="authHeroPanel"/)
})

test('admin layout defines the unified workspace shell and navigation panel', () => {
  const vue = readProjectFile('src/views/admin/AdminLayout.vue')

  assert.match(vue, /class="adminWorkspaceShell pageWrap"/)
  assert.match(vue, /class="adminSidebar card"/)
  assert.match(vue, /class="adminTopbar card"/)
  assert.match(vue, /class="adminNavFooter"/)
})

test('admin dashboard and role views expose refined spotlight sections', () => {
  const dashboardVue = readProjectFile('src/views/admin/AdminDashboardView.vue')
  const roleVue = readProjectFile('src/views/admin/AdminRoleView.vue')

  assert.match(dashboardVue, /class="dashboardHero card panelShell"/)
  assert.match(dashboardVue, /class="metricCard card panelShell"/)
  assert.match(roleVue, /class="rolePanelWrap"/)
  assert.match(roleVue, /class="roleTablePanel card panelShell"/)
  assert.match(roleVue, /class="passwordPanel card panelShell"/)
})

test('admin conversation view exposes the upgraded query rail and transcript surface', () => {
  const vue = readProjectFile('src/views/admin/AdminConversationView.vue')

  assert.match(vue, /class="conversationPanel"/)
  assert.match(vue, /class="queryCluster"/)
  assert.match(vue, /class="conversationHistoryCard"/)
  assert.match(vue, /class="transcriptPanel card panelShell"/)
  assert.match(vue, /class="transcriptBubble"/)
})

test('remaining admin list pages share upgraded section shells and dense panels', () => {
  const userVue = readProjectFile('src/views/admin/AdminUserView.vue')
  const memberVue = readProjectFile('src/views/admin/AdminMemberView.vue')
  const menuVue = readProjectFile('src/views/admin/AdminMenuView.vue')
  const redeemCodeVue = readProjectFile('src/views/admin/AdminRedeemCodeView.vue')
  const redeemRecordVue = readProjectFile('src/views/admin/AdminRedeemRecordView.vue')

  assert.match(userVue, /class="card panelShell panel userPanel"/)
  assert.match(memberVue, /class="card panelShell panel memberPanel"/)
  assert.match(menuVue, /class="card panelShell panel menuPanel"/)
  assert.match(redeemCodeVue, /class="card panelShell panel generatorPanel"/)
  assert.match(redeemCodeVue, /class="card panelShell panel listPanel"/)
  assert.match(redeemRecordVue, /class="card panelShell panel recordPanel"/)
})

test('global stylesheet includes dark-mode table and panel refinements for admin pages', () => {
  const css = readProjectFile('src/style.css')

  assert.match(css, /\[data-theme='dark'\] \.tableWrap \{/) 
  assert.match(css, /\[data-theme='dark'\] th \{/) 
  assert.match(css, /\.tableWrap tbody tr:last-child td \{/) 
})

test('profile membership accent card defines a dark-mode override', () => {
  const vue = readProjectFile('src/views/user/ProfileView.vue')

  assert.match(vue, /\.accentPanel\s*\{[\s\S]*background: linear-gradient\(180deg, rgba\(255, 251, 247, 0\.92\)/)
  assert.match(vue, /\[data-theme='dark'\] \.accentPanel\s*\{[\s\S]*background: linear-gradient\(180deg, rgba\(39, 45, 54, 0\.92\)/)
})

test('profile view removes the old explanatory intro copy', () => {
  const vue = readProjectFile('src/views/user/ProfileView.vue')

  assert.doesNotMatch(vue, /管理昵称、头像地址与账户资料，功能保持不变，只优化界面层次与使用感受。/)
  assert.doesNotMatch(vue, /输入兑换码后直接更新会员时长与账户状态，原有业务逻辑与接口请求保持不变。/)
})

test('landing page keeps the inspiration CTA centered', () => {
  const vue = readProjectFile('src/views/public/LandingView.vue')

  assert.match(vue, /<router-link class="primaryBtn landingCta" to="\/login">开启灵感<\/router-link>/)
  assert.match(vue, /\.landingPanel\s*\{[\s\S]*margin:\s*0 auto;/)
  assert.match(vue, /\.landingCta\s*\{[\s\S]*display:\s*inline-flex;[\s\S]*align-items:\s*center;[\s\S]*justify-content:\s*center;/)
})

test('chat empty state uses the refreshed prompt copy', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /<h2>告诉我你有什么想法<\/h2>/)
  assert.doesNotMatch(vue, /<h2>开始新的对话吧<\/h2>/)
})

test('admin pages remove explanatory intro copy and simplify redeem generator form', () => {
  const dashboardVue = readProjectFile('src/views/admin/AdminDashboardView.vue')
  const userVue = readProjectFile('src/views/admin/AdminUserView.vue')
  const memberVue = readProjectFile('src/views/admin/AdminMemberView.vue')
  const menuVue = readProjectFile('src/views/admin/AdminMenuView.vue')
  const roleVue = readProjectFile('src/views/admin/AdminRoleView.vue')
  const redeemCodeVue = readProjectFile('src/views/admin/AdminRedeemCodeView.vue')
  const redeemRecordVue = readProjectFile('src/views/admin/AdminRedeemRecordView.vue')
  const conversationVue = readProjectFile('src/views/admin/AdminConversationView.vue')

  assert.doesNotMatch(dashboardVue, /sectionIntro/)
  assert.doesNotMatch(userVue, /sectionIntro/)
  assert.doesNotMatch(memberVue, /sectionIntro/)
  assert.doesNotMatch(menuVue, /sectionIntro/)
  assert.doesNotMatch(roleVue, /sectionIntro/)
  assert.doesNotMatch(redeemRecordVue, /sectionIntro/)
  assert.doesNotMatch(conversationVue, /sectionIntro/)

  assert.doesNotMatch(redeemCodeVue, /<label>有效月数<\/label>/)
  assert.doesNotMatch(redeemCodeVue, /v-model\.number="formState\.durationMonths"/)
  assert.match(redeemCodeVue, /durationMonths:\s*1/)
  assert.match(redeemCodeVue, /grid-template-columns:\s*1fr;/)
  assert.doesNotMatch(redeemCodeVue, /sectionIntro/)
})

test('admin sidebar and conversation detail surfaces define dark-mode overrides', () => {
  const layoutVue = readProjectFile('src/views/admin/AdminLayout.vue')
  const conversationVue = readProjectFile('src/views/admin/AdminConversationView.vue')

  assert.match(layoutVue, /\[data-theme='dark'\] \.adminSidebar\s*\{[\s\S]*background: linear-gradient\(180deg, rgba\(25, 30, 37, 0\.96\)/)
  assert.match(layoutVue, /\[data-theme='dark'\] \.menuGroup button\s*\{[\s\S]*background: rgba\(255, 255, 255, 0\.04\)/)
  assert.match(layoutVue, /\[data-theme='dark'\] \.themeRoundBtn,\s*\[data-theme='dark'\] \.logoutBtn\s*\{[\s\S]*background: rgba\(255, 255, 255, 0\.04\)/)
  assert.match(conversationVue, /\[data-theme='dark'\] \.queryCluster\s*\{[\s\S]*background: rgba\(255, 255, 255, 0\.03\)/)
  assert.match(conversationVue, /\[data-theme='dark'\] \.transcriptBubble\.assistant\s*\{[\s\S]*background: rgba\(255, 255, 255, 0\.04\)/)
  assert.match(conversationVue, /\[data-theme='dark'\] \.markdownBody :deep\(pre\)\s*\{[\s\S]*background: rgba\(6, 10, 16, 0\.52\)/)
})
