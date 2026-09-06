import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'
import { execSync } from 'node:child_process'
import { resolve } from 'node:path'

const root = resolve(import.meta.dirname, '..')

const readProjectFile = (relativePath) => {
  return readFileSync(resolve(root, relativePath), 'utf8')
}

// ---------- Part D 新体系：token / 组件库存 / 暗色双机制 ----------

test('style.css is the tailwind v4 entry with brand tokens and dark variant', () => {
  const css = readProjectFile('src/style.css')

  assert.match(css, /@import "tailwindcss";/)
  assert.match(css, /@import "tw-animate-css";/)
  assert.match(css, /@custom-variant dark/)
  // 品牌语义 token（雾林绿 + 圆角阶 + 960 断点）
  assert.match(css, /--primary:\s*#3f7d4e/i)
  assert.match(css, /--radius-lg:\s*16px/)
  assert.match(css, /--breakpoint-md:\s*960px/)
  // shadcn 语义变量映射进 tailwind 工具
  assert.match(css, /--color-card:\s*var\(--card\)/)
  assert.match(css, /--color-muted-foreground:\s*var\(--muted-foreground\)/)
  // 暗色 token 挂在 .dark（shadcn 标准）而非仅 data-theme
  assert.match(css, /^\.dark \{[\s\S]*?--background:/m)
  // 旧雾感变量保留为兼容别名（过渡期未迁移 scoped 样式仍可解析）
  assert.match(css, /--text-main:/)
  assert.match(css, /LEGACY COMPAT/)
})

test('shadcn-vue ui component inventory: 22 component groups copied into the repo', () => {
  const groups = [
    'button', 'card', 'input', 'textarea', 'label',
    'dialog', 'alert-dialog', 'sheet', 'popover', 'dropdown-menu',
    'select', 'switch', 'badge', 'separator', 'scroll-area',
    'table', 'tabs', 'avatar', 'tooltip', 'sonner',
    'skeleton', 'alert',
  ]
  for (const dir of groups) {
    assert.equal(existsSync(resolve(root, 'src/components/ui', dir)), true, `ui/${dir} missing`)
  }
  // 关键文件抽查
  for (const f of [
    'src/components/ui/button/Button.vue',
    'src/components/ui/dialog/DialogContent.vue',
    'src/components/ui/alert-dialog/AlertDialogAction.vue',
    'src/components/ui/sheet/SheetContent.vue',
    'src/components/ui/popover/PopoverContent.vue',
    'src/components/ui/dropdown-menu/DropdownMenuItem.vue',
    'src/components/ui/select/SelectItem.vue',
    'src/components/ui/table/TableRow.vue',
    'src/components/ui/sonner/Sonner.vue',
    'src/lib/utils.js',
  ]) {
    assert.equal(existsSync(resolve(root, f)), true, `${f} missing`)
  }
  assert.match(readProjectFile('src/lib/utils.js'), /tail-merge|twMerge/)
})

test('theme.js drives both dark mechanisms in lockstep (data-theme + html.dark)', () => {
  const theme = readProjectFile('src/utils/theme.js')

  assert.match(theme, /dataset\.theme = nextTheme/)
  assert.match(theme, /classList\.toggle\('dark', nextTheme === 'dark'\)/)
  assert.match(theme, /localStorage\.setItem\('theme', nextTheme\)/)
})

test('App.vue mounts the sonner Toaster', () => {
  const app = readProjectFile('src/App.vue')

  assert.match(app, /import Toaster from '@\/components\/ui\/sonner\/Sonner\.vue'/)
  assert.match(app, /<Toaster/)
})

// ---------- 完整性门禁（Plan §3.4 评审 F1）：18 视图 + 2 布局全部换装，无死区 ----------

test('completeness gate: 20 view/layout files, none still using legacy global classes', () => {
  const files = execSync('find src/views -name "*.vue"', { cwd: root, encoding: 'utf8' })
    .trim().split('\n').sort()
  assert.equal(files.length, 20, `expected 20 view/layout files, got ${files.length}: ${files.join(', ')}`)

  // 旧全局类在新模板中必须绝迹（style.css 的 LEGACY COMPAT 区单独由 CSS 测试断言）
  const banned = [
    'panelShell', 'sectionLabel', 'primaryBtn', 'ghostBtn', 'softInput',
    'formItem', 'toolbarRow', 'tableWrap', 'pageSection', 'contentContainer',
    'tagSuccess', 'tagWarn', 'tagDanger', 'mutedText', 'dangerText',
  ]
  const offenders = []
  for (const rel of files) {
    const src = readProjectFile(rel)
    for (const cls of banned) {
      const staticHit = new RegExp(`class="[^"]*\\b${cls}\\b`).test(src)
      const boundHit = new RegExp(`:class="[^"]*\\b${cls}\\b`).test(src)
      if (staticHit || boundHit) {
        offenders.push(`${rel}: ${cls}`)
      }
    }
  }
  assert.deepEqual(offenders, [], `legacy classes remain in: ${offenders.join('; ')}`)
})

// ---------- 核心页面结构 ----------

test('landing page renders hero Card with the login CTA', () => {
  const vue = readProjectFile('src/views/public/LandingView.vue')

  assert.match(vue, /import Card from '@\/components\/ui\/card\/Card\.vue'/)
  assert.match(vue, /import Button from '@\/components\/ui\/button\/Button\.vue'/)
  assert.match(vue, /to="\/login"/)
  assert.match(vue, /开启灵感/)
})

test('auth pages share AuthShell with ui inputs and submit buttons', () => {
  for (const f of [
    'src/views/public/LoginView.vue',
    'src/views/public/RegisterView.vue',
    'src/views/admin/AdminLoginView.vue',
  ]) {
    const vue = readProjectFile(f)
    assert.match(vue, /import AuthShell from '@\/components\/layout\/AuthShell\.vue'/, f)
    assert.match(vue, /import Input from '@\/components\/ui\/input\/Input\.vue'/, f)
    assert.match(vue, /@submit\.prevent/, f)
  }
})

test('user layout keeps desktop sidebar + mobile Sheet navigation with lucide icons', () => {
  const vue = readProjectFile('src/views/user/UserLayout.vue')

  // 业务逻辑原样保留的锚点
  assert.match(vue, /const memberTag = computed/)
  assert.match(vue, /const logout = \(\) => \{/)
  // 双形态导航
  assert.match(vue, /import Sheet from '@\/components\/ui\/sheet\/Sheet\.vue'/)
  assert.match(vue, /side="left"/)
  assert.match(vue, /component :is="item\.icon"/)
  assert.match(vue, /menuItems = \[/)
  assert.match(vue, /path: '\/app\/chat'/)
  assert.match(vue, /path: '\/app\/profile'/)
})

test('chat view keeps behavioral anchors and adopts the new overlay system', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  // 三维度偏好（Part B/C 锚点，行为不变）
  assert.match(vue, /answerLengthOptions/)
  assert.match(vue, /answerStyleOptions/)
  assert.match(vue, /answerFormatOptions/)
  assert.match(vue, /persistPreferences/)
  // 复制（Part A 锚点）
  assert.match(vue, /import \{ copyText \} from '\.\.\/\.\.\/utils\/clipboard'/)
  assert.match(vue, /copiedMessageId/)
  // 智能滚动（逻辑不变）
  assert.match(vue, /messageIsNearBottom/)
  assert.match(vue, /isNearBottom/)
  // 新体系：历史 Sheet / 删除 AlertDialog / 公告 Dialog / 偏好 Popover / 发送停止互换
  assert.match(vue, /openHistoryDrawer/)
  assert.match(vue, /import Sheet from '@\/components\/ui\/sheet\/Sheet\.vue'/)
  assert.match(vue, /import AlertDialog from '@\/components\/ui\/alert-dialog\/AlertDialog\.vue'/)
  assert.match(vue, /pendingDeleteId/)
  assert.match(vue, /import Dialog from '@\/components\/ui\/dialog\/Dialog\.vue'/)
  assert.match(vue, /acknowledgeAnnouncement/)
  assert.match(vue, /import Popover from '@\/components\/ui\/popover\/Popover\.vue'/)
  assert.match(vue, /stopGenerating/)
  // 桌面端只保留主对话区，历史一律经按钮加载
  assert.doesNotMatch(vue, /historySidebar/)
})

test('admin layout keeps menu wiring incl. prompt page and adds logout confirm', () => {
  const vue = readProjectFile('src/views/admin/AdminLayout.vue')

  assert.match(vue, /label: '提示词管理', path: '\/admin\/prompt'/)
  assert.match(vue, /label: 'AI 状态', path: '\/admin\/ai'/)
  assert.match(vue, /showLogoutConfirm/)
  assert.match(vue, /authStore\.logoutAdmin\(\)/)
  assert.match(vue, /component :is="item\.icon"/)
})

test('admin prompt page keeps pref cards + revision restore with AlertDialog confirms', () => {
  const vue = readProjectFile('src/views/admin/AdminPromptView.vue')

  // 偏好卡片（阶段一功能锚点）
  assert.match(vue, /dimensionTitles/)
  assert.match(vue, /answerFormat: '输出格式'/)
  assert.match(vue, /card\.customized/)
  assert.match(vue, /预置：\{\{ card\.preset \}\}/)
  assert.match(vue, /loadPrefCards/)
  assert.match(vue, /resetAdminPrefPrompt/)
  // 版本历史
  assert.match(vue, /restoreAdminPrompt/)
  assert.match(vue, /pendingRestoreRevision/)
  // window.confirm 全面替换为 AlertDialog（注释提及不受限，锁定的是调用）
  assert.doesNotMatch(vue, /window\.confirm\(/)
})

test('admin user page keeps member-expire logic and renders dual table/cards', () => {
  const vue = readProjectFile('src/views/admin/AdminUserView.vue')

  assert.match(vue, /formatMemberExpireAtForInput/)
  assert.match(vue, /getMemberExpireDisplayMeta/)
  assert.match(vue, /memberExpireState\.submitting/)
  // 桌面表格 + 移动卡片（双渲染形态）
  assert.match(vue, /md:block/)
  assert.match(vue, /md:hidden/)
  // 空值哨兵映射（reka-ui SelectItem 禁止空 value）
  assert.match(vue, /statusFilter/)
  assert.doesNotMatch(vue, /<SelectItem value=""/)
})

test('markdown body styles survive in scoped chat styles for prose rendering', () => {
  const vue = readProjectFile('src/views/user/ChatView.vue')

  assert.match(vue, /\.markdownBody/)
  assert.match(vue, /renderAssistantContent/)
  assert.match(vue, /renderMarkdownToSafeHtml/)
})
