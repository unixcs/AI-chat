# 三合一需求 Plan：复制修复 + 三维度回答偏好 + shadcn/ui·brand-guidelines 全站重构

> 2026-09-06。依据用户三条需求整理；流程：本 Plan → 子代理评审（通过方可实施）→ **阶段一：A+B+C 实施 → 碰撞审查 → 版本锚点（tag+push GitHub）** → **阶段二：全站 UI 重建 → 碰撞审查 → 部署 yun1 → 推 GitHub → 同步文档**。
> 基线：go-rewrite @ 293cb49。红线：yun 生产零触碰；Node API 契约不破坏（全部增量）；secret 不进 git。

---

## §0 三个交付块总览

| 块 | 内容 | 性质 |
|---|---|---|
| A | 修复消息复制 Bug（用户消息+AI回复，桌面+移动端，Clipboard API 优先+兼容方案，"已复制"提示，复制原始文本） | Bug 修复 |
| B | 三维度回答偏好：回答长度（精简/适中/详细，默认适中）、回答风格（大白话/标准/专业/严谨/鼓励，默认标准）、**输出格式（标准/纯文字，默认标准）**；最终指令 = 基础 Prompt + 长度 + 风格 + 格式 | 功能（B+C 合一实现） |
| C | 后台 10 张偏好提示词卡片（3长度+5风格+2格式），预置默认文案，热改即效、进行中请求不中断 | 功能 |
| D | 前台聊天 + 后台管理**全站 UI 重构**：shadcn/ui（shadcn-vue）组件体系 + brand-guidelines 视觉规范，移动端优先，功能逻辑不变 | 重构 |

**明确不做**：不新增角色系统/复杂公告中心/分享/重新生成等（Plan 原始 §22 不变）；UI 重构不夹带新功能；不动后端 Router/AI 层（B 只动 Prompt 组装与偏好存储）。

## §1 Part A：复制修复

### 1.1 取证结论（headless Chromium 实测 yun1 实站）
- 实站 `http://…:8189` 为 **insecure context**：`navigator.clipboard === undefined`，永远走 `execCommand` 兜底路径；
- 桌面 Chromium：点击复制 → `execCommand('copy')` 返回 **true**，按钮翻转"已复制"，零报错 → **桌面链路本身是通的**；
- 根因判定（按证据排序）：
  1. **iOS Safari/WebView**：`readonly` textarea 的 `select()/setSelectionRange()` 在 iOS WebKit 上不产生可用选区（已知 WebKit 行为）→ `execCommand('copy')` 静默失败或复制空 → 按钮却显示"已复制"；
  2. **secure-context 分支的 `await` 丢失 user activation**：Safari 下 `await navigator.clipboard.writeText` 被拒后落入 execCommand 已跨微任务，activation 失效 → 静默失败；
  3. 部分安卓 WebView 需要元素 focus + Range 选区双保险。
- 现实现的 UX 缺陷：失败只有行内 errorText（弹层外不可见）；成功提示 1.5s 状态文字（无全局轻提示）。

### 1.2 修复设计（重写 `utils/clipboard.js`）
1. **优先 Clipboard API**：只要 `navigator.clipboard?.writeText` 存在就先试（不再与 isSecureContext 与门），rejection 落入兼容路径；
2. **同步可达约束（评审 F9）**：execCommand 兜底必须保持在首次 `await` 之前**同步可达**（insecure 路径零 await 直达），代码注释+测试锁定该性质——这是 user activation 存活的关键；
3. **iOS 兼容的 execCommand 路径**：textarea 不加 readonly、叠加 `contentEditable=true`；`focus()` + `select()` + `setSelectionRange` **并同时用 Range API**（`getSelection().removeAllRanges(); addRange(range)`）双保险；textarea 定位到触发元素附近（防 WebKit 滚动跳动）；copy 后 `blur()`、移除节点、恢复原焦点/选区；
4. **返回真值语义**：`execCommand('copy')` 返回 false 或抛异常 → 返回 false；
5. **轻提示**：成功 → 全局 toast「已复制」（Part D 上线 vue-sonner 后统一切 toast；过渡期维持按钮态）；失败 → toast「复制失败，请长按文本手动复制」；
6. **复制原始文本**：继续传 `msg.content`（Markdown 源码），绝不取 DOM innerText/v-html 渲染结果（新增测试锁定）。

### 1.3 测试（jsdom 单测 `tests/clipboard.test.js`）
- **stub 策略（评审 F9）**：jsdom 的 `document.execCommand` 未实现（调用即抛 Not implemented）——测试须 stub `document.execCommand`/`getSelection`/`navigator.clipboard`，验证调用序列与返回值传播；
- 场景矩阵：用户消息/AI 回复/长文本(>10KB)/Markdown 源文/emoji/多行/数字 content/空串拒绝；
- 路径锁定：clipboard API 存在→优先且只调一次 writeText；API 拒绝→回退 execCommand；API 不存在→直接 execCommand；execCommand false→返回 false；
- iOS 路径：验证 Range 选区被建立、焦点恢复、textarea 被移除、兜底在首次 await 前同步可达。

## §2 Part B+C：三维度偏好 + 后台提示词卡片

### 2.1 现状盘点（已核实）
- `users.answerLength`（concise/standard/detailed）、`users.answerStyle`（plain/standard/professional/rigorous/encouraging）列已存在，前端两维度选择器已有；
- `answerModeSuffix()` 把 4 种硬编码文案拼在系统提示词尾部（`chat.go:183`），**standard=适中/标准 无后缀**；
- `BuildSystemPrompt` 在 `PrepareStream` 每请求调用一次后冻结（进行中不换 Prompt 语义已天然成立）；
- `settings` 表（key-value upsert）+ `ensureSchema` 幂等迁移模式可直接复用。

### 2.2 数据模型
- **新增列** `users.answerFormat TEXT`（''/standard/plain；空=standard），走 `addUsersColumns` 非破坏迁移；
- **10 张卡片存 settings 表**，键 `prefPrompt.<dimension>.<value>`：
  - `prefPrompt.answerLength.concise|standard|detailed`
  - `prefPrompt.answerStyle.plain|standard|professional|rigorous|encouraging`
  - `prefPrompt.answerFormat.standard|plain`
- **预置默认**：启动时 `INSERT OR IGNORE` 种入 10 条精心编写的默认文案（管理员无需从零写；`INSERT OR IGNORE` 保证管理员编辑跨重启保留）；
- value 白名单硬编码在代码（dimension+value 枚举），非法键 404。

### 2.3 组装语义（chat.go）
```
最终 system = 基础内置Prompt（promptRevisions 头版本，无则 env）
           + "\n\n回答长度要求：" + 卡片(用户 answerLength)
           + "\n\n回答风格要求：" + 卡片(用户 answerStyle)
           + "\n\n输出格式要求：" + 卡片(用户 answerFormat)
```
- 三段独立解析，卡片内容为空串 → 跳过该段；读取失败 → 跳过该段并 log（不 500）；
- **默认语义变化声明（评审 F5）**：现状 standard（适中/标准）无后缀；改造后 10 张预置卡非空，**默认用户的 system prompt 将多出三段**（含 standard 段）——这正是需求"基础+长度+风格+格式"的本意，实现者不得沿用"standard 跳过"旧语义；P3 断言按此写；删除硬编码 `answerModeSuffix` 时，`service_test.go` 对其的直接单测须同步改写为卡片组装断言；
- 组装在 `BuildSystemPrompt`（PrepareStream 内）一次完成 → **进行中请求沿用开始时快照**（既有语义，测试锁定）。

### 2.4 API（全部增量，Node 契约不破坏）
| 端点 | 语义 |
|---|---|
| `PUT /api/user/preferences` | 增量接受 `answerFormat`（standard/plain 白名单校验，非法 400）；**响应保持 `{code:0,data:true}` 不变**（现状即如此，profile 经 GET /api/user/profile 回读）；**字段语义（评审 R1 写死）**：body 用 `*string` 感知是否提供——未提供或空串不写列（沿用 nilIfEmpty），提供合法值才更新；前端每次 PUT 三字段全量发送；`{code:0,...}` 信封不变 |
| `GET /api/admin/pref-prompts` | 10 张卡片：`[{dimension,value,label,content,preset,customized}]` 按 3 组返回；adminAuth |
| `PUT /api/admin/pref-prompts/{dimension}/{value}` | body `{content}`（≤4000 字节、UTF-8、可空串=该维度关闭）；白名单外 404；audit `update-pref-prompt` |
| `POST /api/admin/pref-prompts/{dimension}/{value}/reset` | 删除 settings 行回到预置文案；audit |

- `/api/user/profile` 与登录响应的 profile 增量携带 `answerFormat`（旧前端忽略无害）。

### 2.5 预置默认文案（首版，管理员可改）
- 长度·精简：「回答务必精简：控制在 100 字以内，直接给结论，不展开解释。」
- 长度·适中：「回答篇幅适中：一般 100–300 字，先给结论再简要说明，用户追问时可展开。」
- 长度·详细：「回答详细展开：分点说明，给出充分的解释、细节和示例，把问题讲透。」
- 风格·大白话：「全程用通俗易懂的大白话，避免专业术语；必须用时紧跟一句生活化比喻。」
- 风格·标准：「用自然、清晰、友好的语气回答，不刻意口语化也不刻意书面化。」
- 风格·专业：「使用专业、准确的表达，术语规范、逻辑清晰，避免堆砌行话。」
- 风格·严谨：「保持严谨：区分事实与推测，不确定就明确说明，结论附理由，不臆断。」
- 风格·鼓励：「语气温和、以鼓励为主：先肯定思路，再指出改进方向，给出具体建议。」
- 格式·标准：「排版自由：根据内容选用最合适的排版（分段、列表、表格、代码块等）。」
- 格式·纯文字：「只输出纯文字：不使用 Markdown、列表、表格、标题、代码块、加粗、emoji 等任何特殊排版符号，用自然段落书写。」

### 2.6 前端
- **聊天界面**：回答偏好弹层改 3 组选择（长度/风格/格式），默认值 适中/标准/标准；选择即 `PUT preferences` 持久化；Profile/Settings 同步展示；
- **后台**：`AdminPromptView` 扩展「偏好提示词卡片」区：3 组 × 卡片（textarea + 保存 + 恢复预置），显示「已自定义」徽标与预置对照；保存失败页内可见。

### 2.7 测试断言
| # | 断言 | 测试 |
|---|---|---|
| P1 | PUT preferences 带 answerFormat=plain → profile 回读 plain | api/service 测试 |
| P2 | 非法 answerFormat → 400 文案 | 同上 |
| P3 | 组装顺序 = 基础+长度+风格+格式（**standard/适中段同样拼入**，缺省段仅指空串卡片），缺省段跳过 | service 测试（假 upstream 捕获 system 消息） |
| P4 | 10 卡片启动预置幂等（二次启动不重复/不覆盖管理员编辑） | store 测试 |
| P5 | PUT 卡片后新请求用新文案；**进行中请求不变**——工装前提：fakeOpenAI 需支持"阻塞首字"钩子（先发起流挂起 → PUT 卡片 → 放行 → 断言 system 仍为旧文案），无该钩子不得降级为只测新请求 | api 测试 |
| P6 | 白名单外 dimension/value → 404；audit 落盘 | api 测试 |
| P7 | reset 回预置 | service 测试 |
| P8 | 旧客户端（无 answerFormat）profile 返回 **null**（SafeUser 空串→nil 语义，与 answerLength/Style 一致），前端按 standard 兜底；API 口径据此锁定 | service 测试 |
| P9 | PUT preferences 三字段全量往返持久化；仅发一维时其余两维不被破坏（锁定 nilIfEmpty 语义） | api 测试 |

## §3 Part D：shadcn/ui + brand-guidelines 全站重构

### 3.1 技术选型（第一性原理取默认最优解）
- **shadcn/ui 的 Vue 生态正解 = shadcn-vue 组件体系**：Tailwind CSS v4（`@tailwindcss/vite`，CSS-first tokens）+ **reka-ui**（无头原语，shadcn-vue v2 底座）+ class-variance-authority + clsx + tailwind-merge + lucide-vue-next（图标）+ vue-sonner（toast）。
- 采用 shadcn 的核心哲学「**组件复制进仓库、代码归我们所有**」：`src/components/ui/` 下按需引入 **22 个组件**源码（button/card/input/textarea/label/dialog/alert-dialog/sheet/popover/dropdown-menu/select/switch/badge/separator/scroll-area/table/tabs/avatar/tooltip/sonner/skeleton/alert），逐个适配品牌 token；不引 CLI、不锁第三方皮肤包。
- 否决：继续手写 CSS 打补丁（用户已否决 D8 偏差）；引入完整组件库（element-plus 等，与 shadcn 体系冲突）；React 生态 shadcn 原版（项目是 Vue）。

### 3.2 brand-guidelines（`docs/BRAND-GUIDELINES.md`，随实现落成）
仓库中无现成 brand 文档（原始 Plan 仅一句话提到），故本 Plan 正式定义并落地为文档+CSS token，评审代理重点裁定：
1. **色彩**（沿用品牌既有雾感识别，token 化到 shadcn 语义变量）：
   - Primary 主色：雾林绿 `#3F7D4E`（hover `#356A42`，浅底 `rgba(63,125,78,.08)`）；
   - 中性面：暖雾白背景 `#F5F4EF` / 卡面 `#FFFFFF` / 玻璃边 `rgba(255,255,255,.55)`（暗色对应深雾 `#171B18` 系）；
   - 语义：success=primary 系、danger `#B4533C`、warning `#B98A2E`、info 蓝 `#3E6E8E`；
   - 文本：main `#1E241F` / soft `#5C665E` / faint `#8A948B`（暗色反转），正文对比度 ≥ AA。
2. **字体**：中文优先系统栈 `-apple-system,"PingFang SC","HarmonyOS Sans SC","Microsoft YaHei",sans-serif`；代码 `ui-monospace`；字号阶 12/13/14/15/17/20/24，正文 15px/1.75。
3. **几何**：圆角阶 sm 8 / md 12 / lg 16 / xl 22（气泡/卡面）；间距 4px 基准（组件内 8/12/16，区块 24/32）；阴影三层柔和低透明度。
4. **层级与交互**：页面=背景层→卡面层→浮层(dialog/sheet/popover)；动效统一 150–200ms ease，禁 >300ms；焦点环 2px primary/40%。
5. **移动优先**：断点 640/960；触控目标 ≥44px；底部 composer 贴 safe-area；管理端表格 <640px 折叠为卡片列表。
6. **语气**：文案克制不用感叹号堆叠；空状态给一句引导+一个主行动。

### 3.3 改造范围与策略（18 视图 + 2 布局，逐一核对不得有死区）
**原则：`<script setup>` 业务逻辑（API/状态/流式/滚动）原样保留，只重写 template/style；API 层零改动。**
| 区域 | 视图 | 改造要点 |
|---|---|---|
| 公开页 | **LandingView（`/` 入口页，必须纳入——其依赖的全局类在 tailwind 化后会消失）** | Hero + CTA 用 Card/Button 重建，移动优先 |
| 认证 | LoginView / RegisterView / AdminLoginView | 居中 Card + Input/Label/Button，错误 Alert |
| 用户壳 | UserLayout | shadcn 侧栏模式：桌面常驻侧栏 + 移动 Sheet + 底部移动栏；导航项 lucide 图标 |
| 聊天 | ChatView | 消息列表（Avatar+气泡+复制按钮常驻≥44px+时间）、Markdown 体（样式迁到 tailwind prose 类）、composer（自动增高 Textarea、发送↔停止互换、三维度偏好 Popover、历史 Sheet、删除 AlertDialog、公告 Dialog、智能滚动逻辑不变） |
| 用户页 | ProfileView / (SettingsView 重定向) | Card 表单 + Select/Switch（如现有） |
| 管理壳 | AdminLayout | shadcn 侧栏（分组导航+图标）+ 移动 Sheet；登出确认 |
| 管理页 | Dashboard/User/Member/Role/Menu/RedeemCode/RedeemRecord/Conversation/Announcement/Prompt(+偏好卡片)/AiStatus | 桌面 Table + 移动卡片列表；操作 DropdownMenu；表单 Dialog/Sheet；反馈 vue-sonner toast；分页 Button 组 |

### 3.4 工程与回归策略
- `style.css` → tailwind v4 入口（`@import "tailwindcss"` + `@theme`/`:root` 品牌 token）+ `tw-animate-css`（shadcn-vue 浮层动效类依赖，官方 setup 必含）；旧雾感 token 保留为兼容别名，逐步删；
- **暗色模式（评审 F2 修正）**：现状是 `data-theme` 属性机制（utils/theme.js 写 `documentElement.dataset.theme`，verify-dark-mode.py 按此驱动），**不是 `.dark` class**。改造方案：theme.js 在保留 `data-theme` 的同时**同步切换 html 的 `dark` class**（shadcn 生态标准变体），CSS token 双变量同时定义；verify-dark-mode.py 同步更新为断言两机制并存；
- **完整性门禁（评审 F1 修正）**：以 `frontend/src/views/` 实际文件清单逐一核对（18 视图 + 2 布局），每项必须出现在 §3.3 改造表或显式豁免清单中——tailwind 化后未改造视图的全局类会消失，不允许存在死区；
- **测试**：行为测试（chat-state/chat-streaming/markdown/chat-entry/admin-member-expire）必须原样绿；结构测试（ui-redesign-structure/ui-layout-regressions/ui-copy-regression）重写为断言新体系（ui/ 组件存在、token 定义、copy util 行为、三维度选择器、卡片页结构）；
- 构建产物体积门禁：gzip 后 JS 增量 ≤ +120KB、CSS 增量 ≤ +30KB（reka-ui 按 tree-shake）。

## §4 实施顺序（**用户修订：前两块先做扎实，全站 UI 重建放最后 + 版本锚点回滚保险**）

**阶段一：Part A+B+C 做扎实**
1. 后端：answerFormat 列 + 卡片存储/预置/API + 组装改造（§2；同步改写 answerModeSuffix 直接单测）→ `go test ./... -race` 绿；
2. 前端 Part A+B（clipboard 重写 + 三维度选择器/卡片页；此步的偏好弹层为过渡实现，Part D 只换皮不改逻辑）→ build+测试绿；
3. 碰撞审查 R12a（攻 A/B/C + 回归狩猎）→ 修复发现项 → 全量回归绿；
4. **版本锚点（回滚保险）**：commit + `git tag pre-ui-redesign` + **push GitHub**——UI 重建前记录"最终可用版本"，UI 重构后若不满意可 `git reset --hard pre-ui-redesign`（或 revert 系列提交）一键回滚，另部署该锚点镜像到 yun1 做基线验证。

**阶段二：Part D 全站 UI 重建**
5. Part D 基建（tailwind v4 + tw-animate-css + ui/ 22 组件 + brand token + sonner + 暗色双机制）→ 布局/认证/Landing → ChatView → 管理壳 + 各管理页 → 每批 build+测试绿；
6. 碰撞审查 R12b（攻 Part D：视觉体系一致性/移动端/暗色/行为回归）→ 修复；
7. 全量回归 → 部署 yun1 → 冒烟（复制/三维度/卡片/重构页面双端渲染）→ health-yun 双机 ALL GREEN；
8. 推 GitHub（final）+ 文档同步（GO-REWRITE-REPORT 增章 / BRAND-GUIDELINES / CURRENT.md）。

**回滚预案**：UI 不满意 → 前端回 `pre-ui-redesign` tag（后端 A/B/C 保留——卡片 API 与 answerFormat 为纯增量，旧 UI 不依赖也不受影响）；或仅 revert UI 相关 commit 系列，前端镜像单独重建。

## §5 风险与开放点
- reka-ui/tailwind 依赖下载失败 → npm 镜像兜底；锁版本保可复现；
- ChatView 重写回归风险最大 → 逻辑零改动原则 + 行为测试护航 + 上线前双端冒烟；
- 旧用户 profile 无 answerFormat → 返回 null + 前端 standard 兜底，无迁移需要；
- **`纯文字` 验收口径**：验收 = 指令确实按序拼入 system（P3/P5 锁定），**不把模型遵从性当验收门**（不做输出后处理，超出第一版范围，记偏差待验收）；
- Part D 前的版本锚点（tag pre-ui-redesign + GitHub push）是 UI 不满意时的快速回滚保险，阶段一完成即执行。

## §6 R12a 对抗审查结果与修复记录（2026-09-06）

**VERDICT: BREACH** → 全部发现项已修复，真机回归 ALL PASS。8 向量中 7 个 PASS（内容正确性/偏好并发契约/组装语义/后台卡片页/回归狩猎/新端点契约/种子迁移 E2E），复制兜底路径 FAIL。

| 级别 | 发现 | 修复 | 证据 |
|---|---|---|---|
| P0 | 复制兜底整体失效：`removeAllRanges()+selectNodeContents(textarea)+addRange` 中 textarea 无子节点 → range 恒空，且 removeAllRanges 销毁 select()/setSelectionRange() 刚建立的真选区；execCommand 返回 true 但 OS 剪贴板不变（假成功） | 删除整块 Range 选区逻辑；改为 clipboard.js 式实战验证链：readonly→focus→**移除 readonly**→select→setSelectionRange→execCommand；窗口选区仅在 copy 后清理并恢复用户原选区 | Chromium 149 CDP：种入 STALE → 兜底复制 → readText=新文本（5 场景：用户消息/AI markdown/2万字/emoji/多行）；insecure origin（copyhost.test 模拟生产 http://IP）选区在 copy 瞬间覆盖全文 |
| P1 | clipboard.test.js「iOS path」假绿灯：createRange stub 为 no-op，未验证 range 覆盖内容 | 重写为诚实的结构断言：生命周期顺序锁（focus→readonly移除→select→setSelectionRange→execCommand）、`selectNodeContents` 在源码中明令禁止（剥离注释后匹配）、原选区恢复断言；真机端到端证据由 CDP 工装提供 | 73/73 jsdom 绿 + 12/12 CDP PASS |
| P3 | 注释与代码矛盾（声称非 readonly 实挂 readonly） | readonly 仅在 focus 前存在（防移动端键盘弹出），选区前移除（iOS readonly 选区缺陷） | 已实现 |
| P3 | Plan §1.2(3) 两处未做：textarea 未贴滚动位置；copy 后未恢复原选区 | 均已落实：`top=pageYOffset`（iOS 只复制视口内选区）；`prevRange` 捕获并在 copy 后恢复 | 真机 PASS |
| P3 | Plan §2.6「Profile/Settings 同步展示」未实现 | ProfileView + SettingsView 各增只读「回答偏好」展示块（标签映射与 ChatView 一致，standard 兜底） | — |
| P3 | PUT preferences 空/缺省由"清列"变"不改列" | 有意加固（动态 SET），无现存前端路径发空串，记为语义偏差 | R12a 30 线程并发混合 PUT 无丢字段 |
| P3 | utf8.ValidString 在 JSON 入口不可达（json 已强转 U+FFFD） | 保留为直调防御 + 语义注释 | — |

**遗留声明**：jsdom 无法证明 OS 剪贴板端到端（R12a 的核心教训），真实浏览器证据由 CDP 工装（/tmp 工装思路已固化进测试注释与本文档）提供；UI 重建（Part D）若触碰 ChatView 复制按钮，须复跑同款 CDP 验证。
