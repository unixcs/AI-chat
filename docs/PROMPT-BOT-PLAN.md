# 内置提示词管理 + Telegram Bot 高频管理 — 实施 Plan（v2，评审 PASS 8.5/10 修订版）

> 2026-09-06。依据用户需求整理为最终 Plan；从第一性原理取默认推荐最优解。
> v2：吸收评审 F1-F8（AdminLayout 菜单硬编码接线、版本分配改 BEGIN IMMEDIATE、TG env 部署口径、"30天/1日历月"决策、预设解释声明、冗余索引清理、现状描述校准、测试基建依赖补充）。
> 范围：backend-go / frontend / bot。约束：Node 契约不破坏（本改动全部为增量 API）；yun 生产零接触；secret 只走环境变量。

---

## §0 背景与目标

Go 重构版已部署 yun1 测试环境。本轮新增两块：

- **功能 A：管理员后台「内置提示词管理」** — 把写死在启动配置里的 System Prompt 变成 Web 后台可热改的 SQLite 配置，带版本历史与恢复。
- **功能 B：Telegram Bot 补齐「高频管理员操作」** — Bot 骨架已存在（轮询 + ID 白名单 + audit），本轮补齐需求清单中的缺口命令，全部走统一 Service 层。

## §1 现状盘点（已核实）

| 事实 | 位置 | 对本设计的影响 |
|---|---|---|
| SystemPrompt 启动时从 `DEEPSEEK_SYSTEM_PROMPT_FILE` 或默认值载入 `cfg.SystemPrompt` | `internal/config/config.go:116` | 配置降级为「初始默认值」 |
| `BuildSystemPrompt` 每请求在**流式请求准备阶段调用一次**，消息数组随即冻结 | `internal/service/chat.go:122` | “进行中请求不换 Prompt”**天然满足**，无需额外机制 |
| 用户级「回答长度/风格」以 `answerModeSuffix` 追加在系统提示词尾部 | `chat.go:170` | 管理员 Prompt 与用户偏好分层注入，互不冲突——需求 6 满足 |
| Bot 已有：白名单、[tg-audit] 日志、`/status /ai /code /codes /user /ban /unban /announce /reg /redeem`，邀请码固定 30 天，operator=`tg:<id>` | `internal/bot/telegram.go` | 功能 B 只补缺口，不重写 |
| `settings` 表（key-value upsert）、`announcementReads` 等增量表模式已验证 | `internal/store/announcements.go` | 新表沿用同一 `CREATE TABLE IF NOT EXISTS` 迁移模式，Node 时代库可平滑接管 |
| `ensureMenus()` 为 Node 时代库补种新菜单 | `store.go:164` | 新增「提示词管理」菜单照此办理 |
| `AdminRedeemRecords(page,pageSize,phone,start,end)` 已存在 | `service/admin.go:345` | Bot 兑换记录查询直接复用 |

## §2 功能 A：内置提示词管理

### D-A1 存储：`promptRevisions` 追加表（append-only）
```sql
CREATE TABLE IF NOT EXISTS promptRevisions (
  id TEXT PRIMARY KEY, version INTEGER UNIQUE, content TEXT NOT NULL,
  operator TEXT NOT NULL, createdAt TEXT NOT NULL);
```
- **当前生效版本 = 最大 version 行**。不设 “current 指针” 字段——指针方案需要额外的回滚语义（回滚=指针回拨，历史分叉），追加表把「恢复」实现为「旧内容生成新版本」，历史永远线性、可审计、无分叉。拒绝项：单行 settings 存 prompt（无历史，不满足需求 8）；指针+历史双写（回滚语义复杂，收益为零）。
- version 分配：**`BEGIN IMMEDIATE` 写事务**内 `SELECT MAX(version)` + 1 + INSERT。评审 F2 指出：连接池多连接 + DEFERRED 事务读-升级-写会产生 SQLITE_BUSY_SNAPSHOT（busy_timeout 不重试），并发保存会 500——故 store 新增 `inTxImmediate` 辅助函数（`s.DB.BeginTx` + `_tx_immediate` pragma 或 modernc 支持的 BEGIN IMMEDIATE），版本分配必须走它；10 并发保存测试（§7-T6）锁定该行为。
- （v1 的冗余 `CREATE UNIQUE INDEX` 已删：`version INTEGER UNIQUE` 列约束自动建索引。）

### D-A2 生效语义：每聊天请求读库一次
- `Store.CurrentPromptRevision()` → 最高版本行；无行时回退 `cfg.SystemPrompt`（env/文件初始值），`source` 标记 `db|env`。
- `BuildSystemPrompt` 改为：`base := s.EffectiveSystemPrompt()`（=读库+回退），再拼 `answerModeSuffix`。
- 性能：单行索引读，进程内 SQLite，微秒级；本产品并发量级（MODEL_CONCURRENCY 默认 6）完全无压力。拒绝项：带 TTL 的内存缓存（多一个失效面，读库本身已够快）。

### D-A3 快照语义
- 消息数组在流式请求准备阶段构建并冻结（现状已如此），流中途改 Prompt 不影响进行中的生成。验收断言见 §7-T3。

### D-A4 API（全部 adminAuth）
| 方法/路径 | 语义 |
|---|---|
| `GET /api/admin/prompt` | `{content, version, updatedAt, operator, source}`；source=env 时 version=null |
| `PUT /api/admin/prompt` | body `{content}`；非空、≤20000 字节、UTF-8 有效；写入新版本，返回新版本元数据；operator 记录 `web-admin`（复用 auditLog） |
| `GET /api/admin/prompt/revisions` | 最近 50 个版本：`{version, content, operator, createdAt}`（倒序） |
| `POST /api/admin/prompt/restore` | body `{version}`；将该版本内容**复制为新版本**；目标版本不存在→404 |

- 与 Node 的路由不冲突（全新路径）。响应沿用 `{code:0,data}` 信封。

### D-A5 前端：`AdminPromptView.vue`
- 上半：当前生效 Prompt 编辑器（textarea，等宽字体）+ 当前版本/最后修改时间/操作人徽标 + 保存按钮（未修改时禁用）。
- 下半：版本历史列表（版本号、时间、操作人、内容预览、**恢复**按钮）。恢复前 `window.confirm`。
- 路由 `/admin/prompt`，菜单名「提示词管理」，分组「系统管理」。
- **接线两处（评审 F1）**：①`defaultMenus()`/`ensureMenus()` 补种 DB 菜单行（影响「菜单管理」页数据）；②**`AdminLayout.vue` 的 menuGroups 硬编码数组**——侧边栏实际渲染自该常量而非 /api/admin/menus，必须在「系统管理」组显式加「提示词管理」条目，否则页面可达但导航不可见（公告/AI 状态页即两处同时加的先例）。
- 保存失败在页面内可见（errorText），不再复刻“静默失败”教训。

### D-A6 Bot 侧 Prompt 命令（只读 + 切换）
- `/prompt` → 当前版本、来源、最后修改时间、操作人。
- `/prompt set <version>` → 切换（=复制为目标新版本，走同一 Service），等价于 Web 端恢复。
- **「预设切换」的解释声明（评审 F5）**：第一版把"预设"解释为**历史版本之间的切换**（每次切换留下审计版本行，可再切回），不新建"命名预设实体"。若用户实际想要的是少量固定命名预设（如"简洁助手/深度推演"），属二阶段需求，记入偏差表待验收确认。完整编辑只在 Web（Telegram 不适合长文本编辑）。

### D-A7 权限与审计
- 全部走 `adminAuth`（Web）/ ID 白名单（Bot）；每次写操作 `auditLog`（action=`update-prompt` / `restore-prompt`，detail 带 version 与字节数）。

### D-A8 明确不做（第一版）
角色系统、角色市场、用户自定义角色、Prompt 工作流/编排、复杂发布系统（灰度/定时）、diff 高亮、多语言 Prompt。

## §3 功能 B：Telegram Bot 高频管理（补缺口）

### 需求对照（已有 → 保留；缺口 → 新增）
| 需求 | 现状 | 本轮动作 |
|---|---|---|
| 1 邀请码生成（固定 30 天） | `/code`、`/codes <n>` 已有；实际语义 **durationMonths=1（1 个日历月，28-31 天）**，Bot 文案却写"30 天" | **决策（评审 F4）：保留 1 个日历月语义**（Node 契约 AddMonths，且已部署验收过），仅修正 Bot 文案为「1 个月（约 30 天）」。改精确 30 天会偏离 Node 兼容的到期计算，收益不抵契约风险，记入偏差表 |
| 2 用户查询/封禁/解封 | `/user /ban /unban` 已有；`/user` 的会员输出无法区分"未开通/已过期"（MembershipValid 对两者均 false） | `/user` 输出改三态：未开通 / 有效期至 <日期> / 已过期 <日期> |
| 3 兑换记录查询 | **缺口** | 新增 `/redeemlog [n]`、`/codeinfo <码>` |
| 4 注册/兑换开关 | `/reg /redeem` 已有 | 修复其缺陷：setter 失败必须回显错误（现 `_ =` 吞错）；未知参数回显当前态 |
| 5 公告发布 | `/announce` 已有 | 保留 |
| 6 服务状态 | `/status` 已有 | 保留 |
| 7 AI Gateway 状态 | `/ai` 已有；**已输出当前模式与"❄️冷却中"标记**（评审 F7 校准） | 确认现状已满足，仅补测试锁定 |
| 8 Prompt 状态/预设切换 | **缺口** | `/prompt`、`/prompt set <v>`（见 D-A6） |

### D-B1 新命令规格
- `/redeemlog [n]`：最近 n 条（默认 5，上限 20）兑换记录：`码 → 手机号 @时间（会员至 …）`。复用 `AdminRedeemRecords`（pageSize=n），Bot 不碰 SQL。
- `/codeinfo <码>`：单码查询——未使用（有效/已作废）/已使用（兑换人、兑换时间、兑换后到期时间）。实现：service 新增 `BotRedeemCodeInfo(code)`（查 redeem_codes + redeem_records，null 安全），Bot 仍不直接查库。
- 全部命令输出遵守 Telegram 4096 字符上限：超长截断并提示（列表类天然受限，防御性实现）。

### D-B2 安全（沿用既有并固化）
- `TG_ADMIN_IDS` 白名单（已实现，非白名单请求记 `[tg-audit] UNAUTHORIZED`）。
- `TG_BOT_TOKEN` 仅环境变量，不进 Git；compose 注入 `${cred:}` 引用。
- 所有关键操作已记 audit（本轮新增命令同样强制）。
- Bot 仅经 Service 层（架构红线，见 telegram.go 包注释），不新增任何直连 SQL。
- 轮询失败退避 15s、不影响主服务（已实现，保留）。

### D-B3 架构不变式
```
Telegram Bot → Admin Service（统一业务规则）→ Store → SQLite
Web 后台    → 同一套 Admin Service
```
Bot 新增能力一律先在 service 层落函数（可被 Web 复用），Bot 只做参数解析与文案。

## §4 数据模型与迁移

- 新表 `promptRevisions`（§2 D-A1）加入 `store.go` schema 列表；`CREATE TABLE IF NOT EXISTS` + 唯一索引，对 Node 时代库零破坏（幂等）。
- 菜单补种：`defaultMenus()` 增 `{Name:"提示词管理", Path:"/admin/prompt", MenuGroup:"系统管理"}`；`ensureMenus()` 自动给老库补行。
- 不改任何既有表结构；无破坏性迁移；回滚 = 旧二进制直接替换（新表被忽略）。

## §5 测试与验收断言（先写死再实现）

| # | 断言 | 测试 |
|---|---|---|
| T1 | PUT 后新聊天立即用新 Prompt（integration：改 prompt → SSE → upstream 收到的 system 消息含新内容） | api/prompt_flow_test.go |
| T2 | 无版本行时回退 cfg.SystemPrompt，source=env | api/prompt_flow_test.go |
| T3 | 流式中途改 Prompt，进行中请求的 system 消息不变（请求开始时已冻结） | api/prompt_flow_test.go |
| T4 | restore 把旧版本内容生成为新版本，历史行数只增不减，Current 指向新版本 | service/prompt_test.go |
| T5 | 空内容 400、超长 400、非法 JSON 400；版本不存在 404 | api/prompt_flow_test.go |
| T6 | 并发保存产生连续不重复 version（10 goroutine，锁定 BEGIN IMMEDIATE 语义） | service/prompt_test.go |
| T7 | `/prompt set <v>` 后 `GET /api/admin/prompt` 与 Bot 视角一致（同一 Service） | bot/telegram_test.go |
| T8 | `/redeemlog`、`/codeinfo` 输出与库内记录一致；未使用/已作废/已使用三态正确 | bot/telegram_test.go |
| T9 | `/reg xyz` 等非法参数回显当前态且不改动；setter 失败有错误回显 | bot/telegram_test.go |
| T10 | 菜单补种幂等（老库二次启动不多行） | store 迁移测试（沿用 ensureMenus 现测风格） |
| T12 | Bot 输出 >4096 字符截断（防御性行为锁定） | bot/telegram_test.go |
| T13 | `/codeinfo 不存在的码`、`/prompt set 不存在的版本` 的错误文案 | bot/telegram_test.go |
| T14 | 流首行 BOM 剥离后 "data:" 前缀仍可识别（R10 LOW 修复回归锁） | ai/router_test.go |
| T11 | 全量 `go test ./... -race` 四包绿 + 前端 build + 回归用例绿 | 收口门禁 |

> 测试基建依赖（评审 F8）：T1/T3 需要扩展 `fakeOpenAI`——捕获最近请求体（互斥保护）+ 可控延迟/阻塞钩子，以断言"upstream 收到的 system 消息"与"流式中途改 Prompt"。

## §6 部署注意

- 新端点/新表全部增量，旧镜像可回滚。
- yun1 测试环境经 `deploy/deploy-yun1-test.sh` 重建（WSL 构建镜像 → save/load，不在 yun1 上构建）。
- **Bot 环境变量口径（评审 F3 校准）**：compose 模板与部署脚本**未**预置 TG_BOT_TOKEN/TG_ADMIN_IDS/TG_PROXY——要在 yun1 启用 Bot，需把三者加入远端 `backend/.env`（经 env_file 注入容器）。不配置 = bot 不启动（telegram.go 启动守卫），零风险。token 值不进 Git。

## §7 风险与开放点

- Prompt 内容 XSS：管理端编辑、用户端不渲染（仅注入给模型），前端展示用文本插值——无 XSS 面。
- 大 Prompt 性能：20000 字节上限，单行存储，无风险。
- Bot 输出隐私：`/codeinfo` 只在白名单会话内响应，泄露面等同现有 /user 命令。
