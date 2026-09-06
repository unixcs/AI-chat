# Thallo Brand Guidelines（品牌视觉规范）

> 2026-09-06 随 Part D 全站重构落地（PREF-UI-REDESIGN-PLAN.md §3.2）。实现载体：
> `frontend/src/style.css`（CSS-first token）+ `frontend/src/components/ui/`（shadcn-vue 组件）。
> 修订本文档时必须同步 style.css 的 token 与 ui/ 组件，三者是同一事实的三个面。

## 1. 色彩

沿用品牌既有「雾感 + 雾林绿」识别，token 化到 shadcn 语义变量。

### 语义 token（亮色 `:root` / 暗色 `.dark`）

| Token | 亮色 | 暗色 | 用途 |
|---|---|---|---|
| `--primary` | `#3F7D4E` 雾林绿 | `#6FAF7F` | 主行动、品牌强调、有效状态 |
| `--primary-hover` | `#356A42` | `#7FBF8E` | primary hover |
| `--background` | `#F5F4EF` 暖雾白 | `#131614` 深雾 | 页面底层 |
| `--card` / `--popover` | `#FFFFFF` | `#1A1E1B` / `#1C211D` | 卡面/浮层 |
| `--foreground` | `#1E241F` | `#E8ECE8` | 正文 |
| `--muted-foreground` | `#5C665E` | `#9AA79D` | 次要文本（对 `--card` 对比度 ≥ AA） |
| `--text-faint` | `#8A948B` | `#6E7A70` | 辅助微字 |
| `--destructive` | `#B4533C` | `#D97B64` | 危险/错误 |
| `--warning` | `#B98A2E` | `#D9A94C` | 警示 |
| `--info` | `#3E6E8E` | `#7AA8C7` | 信息蓝 |
| `--border` / `--input` | 雾绿灰 12%/16% 透明 | 白 10%/14% 透明 | 边框 |
| `--ring` | primary 40% 透明 | primary 45% 透明 | 焦点环 |

- success 语义 = primary 系（有效会员、成功 toast 用 primary/accent，不另设绿色）。
- 旧雾感变量（`--text-main` 等）在 style.css 的 LEGACY COMPAT 区保留为别名，过渡期供未迁移 scoped 样式解析；**新代码禁止直接引用旧变量**，全部走语义 token / tailwind 语义类（`bg-card`、`text-muted-foreground`、`bg-accent`…）。

## 2. 字体

- 中文优先系统栈：`-apple-system, 'PingFang SC', 'HarmonyOS Sans SC', 'Microsoft YaHei', 'Noto Sans SC', 'Segoe UI', sans-serif`（token：`--font-sans`）。
- 代码：`ui-monospace, 'SF Mono', 'Cascadia Mono', 'Courier New', monospace`（token：`--font-mono`）。
- 字号阶：12 / 13 / 14 / 15 / 17 / 20 / 24；正文 15px、行高 1.75（`@layer base` body 设定）。
- 标题不加字重 900；最高 bold(700)。

## 3. 几何

- 圆角阶（token）：`--radius-sm` 8 / `--radius-md` 12 / `--radius-lg` 16 / `--radius-xl` 22。气泡与卡面用 xl/lg，输入控件用 md。
- 间距：4px 基准；组件内 8/12/16，区块 24/32（tailwind p-4/p-6、gap-2/3/4）。
- 阴影两层语义 token：`--shadow-elevated`（卡面）、`--shadow-overlay`（浮层），柔和低透明度。

## 4. 层级与交互

- 层级：背景层（body 雾感渐变）→ 卡面层（Card）→ 浮层（Dialog/Sheet/Popover/Select，z-50，overlay 遮罩 + backdrop-blur-[2px]）。
- 动效统一 150–200ms ease（`duration-150/200` + `tw-animate-css` 的 data-[state] 动效类）；**禁止 >300ms**。
- 焦点环：2px `--ring`（`focus-visible:ring-2 focus-visible:ring-ring`），键盘可达。
- 文字选中（`::selection`）：彩色底内容卡（如聊天气泡）必须显式声明 `::selection` 高亮，禁止依赖全局默认——高亮与底色须明度反转（R14 教训：绿底白字气泡上全局绿高亮=隐形）；取值集中在 `style.css` 的 `chat-*-selection-*` 变量并保证 `.dark` 覆盖。
- 浮层关闭钮右上角，`sr-only` 中文标签「关闭」。

## 5. 移动优先

- 断点：`sm` 640 / `md` 960（`--breakpoint-md` 已覆盖为 960；960 是本应用桌面/移动的判定线，与 ChatView `window.innerWidth > 960` 一致）。
- 触控目标 ≥44px：主按钮 `h-11`（Button size default / icon）；密集表格内允许 `size="sm"`（h-9）但仅限桌面。
- 底部 composer 贴 safe-area：`pb-[max(12px,env(safe-area-inset-bottom))]`。
- 管理端表格 <960px 折叠为卡片列表（`hidden md:block` 表格 + `md:hidden` 卡片）。
- 移动导航：用户端与管理端均用 Sheet（side left），桌面常驻侧栏。

## 6. 语气

- 文案克制，不用感叹号堆叠；按钮动词直给（保存/删除/恢复预置）。
- 空状态 = 一句引导 + 一个主行动（如聊天空态「告诉我你有什么想法 / 开始新的对话吧」）。
- 破坏性操作一律 AlertDialog 确认，正文说明不可逆后果（如「删除后对话与其中全部消息不可恢复」）。

## 7. 组件消费约定（shadcn-vue 哲学）

- 组件代码复制进仓库（`src/components/ui/`，22 组），代码归项目所有；不引 CLI、不引整库皮肤。
- 业务视图只允许：ui 组件 + tailwind 语义类 + lucide-vue-next 图标；禁止新写与 ui 组件重复的手造控件样式。
- `cn()`（clsx + tailwind-merge）是唯一类名合并入口（`@/lib/utils`）。
- 暗色双机制：`utils/theme.js` 同时写 `data-theme` 属性与 `html.dark` class；CSS 里语义 token 定义在 `:root` 与 `.dark`，旧 scoped 样式继续被 `[data-theme='dark']` 别名覆盖。`tests/verify-dark-mode.py` 断言两机制并存。
