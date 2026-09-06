# AI-chat Go 重构方案（第一性原理版）

> 日期：2026-09-06 ｜ 状态：送审 ｜ 评审基准：`docs/GO-CHAT-PLAN-ORIGINAL.md`
> 本方案描述目标架构、每条决策的第一性原理推导、精确语义与偏差声明。评审通过后作为实施与验收的唯一基准。

---

## 0. 问题定义（现状的真实痛点）

在动手之前，先从现象剥离出根因。上一轮优化（commit 89b3a94/26f8749/4c667ec）解决的是**表层**（nginx 超时、轮询频率、流式渲染卡顿），以下四个是**结构层**的病：

| # | 病根 | 现有证据 | 后果 |
|---|---|---|---|
| P1 | **DB 写放大**：`writeDb()` 每次写 = 全表 DELETE+INSERT（messages 除外） | `backend/db.js` writeDb | 10万+消息库上每条消息/每次登录全库重写，锁竞争，写延迟随数据量线性涨 |
| P2 | **模型单点**：业务代码直接耦合 DeepSeek 官方 API + 写死模型名 | `server.js` callDeepseekWithRetry | 上游抖动=全站不可用；无成本逃生通道 |
| P3 | **重试策略错位**：共享一个 90s AbortController 重试 3 次 | 同上 | 重试期间用户持续空转；首字延迟无预算概念 |
| P4 | **运维面缺失**：模型健康度、切换次数、兜底成本全部不可观测 | 无 | 故障只能靠用户报障发现 |

**用户价值的本质**（第一性）：打开页面 → 发消息 → **尽快看到第一个字** → 流式不中断 → 不用重发 → 长期便宜 → 出问题管理员先知道。重构一切决策都服务这条链。

## 1. 不变量与变量

- **不变量**（动不得）：SQLite 单文件数据、现有 Vue 前端的 API 契约、现有用户数据兼容、/srv 式 volume 挂载部署形态、会员/邀请码规则。
- **变量**（这次动）：后端运行时（Node→Go）、DB 访问模式、AI 上游拓扑、可观测性、前端交互细节。
- **约束**：1.6GB 内存测试机（yun1）禁本机构建镜像；不引入 Redis/MQ/微服务；secret 不进 git。

## 2. 设计决策（每条：原理 → 决策 → 被否决的备选）

### D1 后端运行时 = Go 纯静态单二进制
- 原理：P1 的解法需要精确的连接/事务控制；长期维护成本 = 内存占用 + 依赖表面积。Go 编译期约束 + 单文件部署在此规模（单机单实例 SQLite）是复杂度最低点。
- 决策：Go 1.27，`net/http` 标准库路由（Go 1.22+ pattern routing），CGO_ENABLED=0，`modernc.org/sqlite`（纯 Go SQLite 驱动）。外部依赖仅 3 个：modernc.org/sqlite、golang-jwt/v5、x/crypto/bcrypt。
- 否决：gin/chi（路由能力标准库已够，省一个依赖树）；mattn/go-sqlite3（要 CGO，交叉构建与 alpine 静态化麻烦）；Rust/重写前端 BFF（超范围）。

### D2 DB 访问 = 精准语句 + 事务，同库同 schema
- 原理：数据是用户的，schema 演进必须非破坏；写放大是 P1 病根，杀它就够，不需要换数据库。
- 决策：
  - 表结构与 Node 版完全一致（users/roles/menus/redeemCodes/redeemRecords/conversations/messages），`CREATE TABLE IF NOT EXISTS`；
  - 新能力只用**增量列/新表**：`users.answerLength/answerStyle` 列 + `announcements`/`announcementReads`/`settings` 表；
  - WAL + busy_timeout 5000 + 连接池上限 8；
  - 语义级兼容：ID 格式（`毫秒时间戳_随机6位`）、时间格式（`2006-01-02T15:04:05.000Z`，UTC）、`AddMonths` 的月末溢出语义与 Node `setMonth` 一致（1-31 加一月 = 3-03，已有测试锁定）；
  - bcrypt 哈希双向兼容（bcryptjs `$2a/$2b` ↔ x/crypto），JWT HS256 同 secret 互验——**Node 签发的 token、存的密码，Go 进程直接接管**。
- 否决：换 Postgres（违反不变量）；GORM 等复杂 ORM（原 Plan 明确不做）；自建 JSON 文件层（回到 P1）。

### D3 AI 层解耦 = ProviderEntry 池 + Router 三策略
- 原理：原 Plan 说"Bifrost 负责成熟 Gateway 能力，自己只做选择/健康/切换"。CPA、Bifrost、DeepSeek 官方全部讲 **OpenAI chat completions 协议**——所以业务代码只需要一个协议实现，拓扑差异全部退化为"有序的端点配置"。
- 决策：
  - `ProviderEntry{Name, BaseURL, APIKey, Model, FirstTokenTimeoutMs, Fallback}`：一个有序池；`fallback:true` 条目恒排最后、官方 DeepSeek 自动追加（业务代码零模型名）；
  - `AI_MODE=official` → 池只含官方；`AI_MODE=gateway` → 全池。运行时（NewRouter）二次强制，即使配置错了也不会让业务层感知差异；
  - `DEEPSEEK_EXTRA_BODY`（如 `reasoning_effort:none`）只附加到官方条目——免费池模型对 DeepSeek 私有参数的容忍度不可假设。
- 否决：在 Go 里集成 Bifrost SDK（把强耦合换了个地方）；为 CPA 写专用客户端（协议一样，多余）。

### D4 Failover 语义（精确状态机，Plan 第四节的工程化）
- 原理："大约 2 秒没有首字"必须落成确定性行为，否则测试无法锁定、线上无法解释。
- 决策：
  1. 逐条目按序尝试；每条目一个**看门狗**：`FirstTokenTimeoutMs`（默认 2000ms）内没有第一个**可见 content token** → cancel 该尝试，记 MODEL_TIMEOUT，进入下一条；**fallback（官方）条目豁免看门狗**——官方是最终保险，绝不被 2s 规则砍掉，其预算只有总超时（90s）；
  2. **累计网关预算** `AI_GATEWAY_BUDGET_MS`（默认 6000ms）：非 fallback 条目开始尝试前检查，超预算直接跳官方兜底——对应 Plan"三个模型累计超过 6 秒立即切官方"；
  3. **首字后不静默重路由**：一旦 token 流向客户端，上游中断 = SSE 错误事件（绝不重放，避免文本重复/跳变）——对应"用户尽可能无感"的正确边界：无感窗口只在首字前；
  4. 全部失败才返回错误；官方余额不足（402）立即终止不换池（换池救不了账户问题）；
  5. 客户端断开（停止生成/关页）→ 取消上游 → **已生成部分照常落库**（与 Node 行为一致，停止≠丢弃）；客户端取消**不属于上游故障**：不记健康失败、不给任何条目上冷却、不虚增 switches，后续条目不再尝试；
  6. **首字后 failover 窗口关闭**：任何 token 已到达客户端后再遇到上游失败，立即返回 STREAM_ERROR，**绝不重放后续条目**（否则客户端收到 "AAABBB" 式拼接）；该失败仍计入当事条目的健康档案（总是断流的坏模型不得永久占据首字窗口）；
- 否决：首字后也切（会文本重复）；每条目独立重试×3（放大延迟，Plan 不要）。

### D5 健康度 = 内存计数 + 指数冷却 + 自然复探测
- 原理：Plan 第七节要"连续失败暂时摘除、过段时间自动恢复"。独立探测器是额外复杂度——**冷却到期后条目自然恢复参选资格，下一次真实请求就是探测**，省一个组件且探测结果真实。
- 决策：每条目记 Successes/Failures/Timeouts/ConsecFails/LastErr/LastFirstToken/EwmaFirstToken；失败 → 冷却 `30s × 2^(连败-1)` 封顶 10min；成功清零。暴露 `GET /api/admin/ai/status`（模式、每条目快照、总请求/免费池成功/切换次数/官方兜底次数）——Plan 第八节的成本可观测。
- 否决：持久化健康历史（v1 不需要）；主动心跳探测（假阳性、费钱）。

### D6 并发 = 信号量 + 队列上限（沿用 Node 语义）
- `MODEL_CONCURRENCY=6` 在途上限、`MODEL_QUEUE_MAX=50` 排队上限，超出即 `MODEL_QUEUE_OVERFLOW`（同一文案），防雪崩防内存失控。

### D7 API 契约 = 100% 兼容 + 纯增量
- 原理：现有前端一行不改必须能跑——这是回归风险的下限策略；新功能只加端点，绝不动旧语义。
- 决策：全部既有端点的路径/方法/状态码/`{code:0,data}` 包裹/错误文案逐字保持（含"账号已在其他设备登录"、SESSION_KICKED 机器码、SSE `data:{"delta":...}`/`[DONE]`/错误事件协议）；新增端点：会话删除、用户偏好、公告（用户+管理）、设置开关、AI 状态。已存偏差均在测试中锁定（见 §6）。
- 否决：借重构之机"优化"旧接口（破坏兼容 = 部署必须前后端同步，风险×10）。

### D8 前端 = 保持既有设计语言 + Plan 九~十三逐条落地
- 原理：Plan 明确"重点不是增加大量功能，而是把现有 Chat 做舒服"；shadcn/ui 是手段不是目的，现有雾感设计语言已被品牌接受，整体换装 = 全量回归风险无增量价值。
- 决策：复制按钮（用户+AI，含 `execCommand` 降级——生产是 http 非 secure context，`navigator.clipboard` 不可用）；生成中发送↔停止单键互换；智能滚动（近底 80px 内跟随，上滑不打扰）；桌面常驻历史侧栏 + 移动抽屉 + 会话删除；回答模式选择器（偏好持久化到用户行）；一次性公告弹窗；管理端公告管理 + AI 状态页。
- **偏差声明（对 Plan 二）**：不引入 shadcn/ui，理由如上。此为有意识取舍，请评审裁定。

### D9 安全与秘钥
- JWT secret 沿用 env（未设则警告 + fallback，与 Node 行为一致）；bcrypt 不明文存密码；SQL 全部参数化（无字符串拼接用户输入进 SQL）；body 1MB 上限；SSE 行 1MB 上限；Bot 不碰 SQL，走 service 层，白名单 TG_ADMIN_IDS，操作走同一套业务校验；敏感管理操作（封禁/解封/改会员/重置密码/生成邀请码/公告/开关）统一在 service 层打 `[audit]` 结构化日志（Bot 路径 operator=`tg:<id>` 透传进同一条 [audit] 行，未授权命令在 [tg-audit] 留痕）；secret 只存服务器 `.env`（600），git 内只有 `.env.example`。

### D10 部署与回滚（分层爆炸半径）
- 构建：WSL `docker build`（go.mod 层缓存 + goproxy.cn），`docker save | gzip | ssh docker load`（yun1 内存红线禁本机构建）；
- 测试位：yun1 `/opt/ai-chat-go-test`，compose 项目 `ai-chat-go-test`，容器 `ai-chat-go-*`，端口 8189(front)/3002(back)，数据全新初始化——**与现有 ai-chat 二节点（8181/3001）、与 yun 生产物理隔离**；
- 回滚 = `docker compose -p ai-chat-go-test down`，爆炸半径为零；
- 生产切换（未来）：沿用 `DEPLOYMENT_WORKFLOW.md` 备份→最小同步→重建→验证流程，不在本方案范围内。

### D11 明确不做（与 Plan 二十二对齐）
微服务/Redis/MQ/复杂 ORM/Token 计费/每日预算/配额/复杂角色后台/语音/n8n 依赖/搜索标签置顶收藏。Telegram Bot 未配置时零开销。

## 3. 架构

```
Vue3 前端(nginx) ──/api──> Go 单体（net/http）
                            ├─ api 层（HTTP/SSE，契约兼容层）
                            ├─ service 层（业务，Bot 与 HTTP 共用）
                            ├─ store 层（精准 SQL，WAL，事务）
                            └─ ai 层（ProviderEntry 池 → Router 状态机 → OpenAI 兼容上游们）
                                                    ├─ 免费池（CPA/Bifrost，可选）
                                                    └─ DeepSeek 官方（兜底）
SQLite（同 schema + 增量列/表）    Telegram Bot ──> service 层（不碰 SQL）
```

## 4. 关键语义精确定义（十断言，逐条映射自动化测试）

| # | 断言 | 锁定测试 |
|---|---|---|
| 1 | **首字**=第一个非空 `choices[0].delta.content`；`reasoning_content` 永不转发、不计首字 | ai/TestReasoningContentFiltered |
| 2 | **看门狗**：非 fallback 条目 2s 无首字 → cancel → 下一条；fallback 豁免看门狗（首字 3s 的官方仍成功）；总超时 90s 报 MODEL_TIMEOUT | ai/TestRouterFailoverFromSlowToFast、ai/TestFallbackExemptFromFirstTokenWatchdog |
| 3 | **预算检查**在每条目尝试开始前做；超预算跳转官方；被跳条目计入 switches | ai/TestGatewayBudgetExhausted、ai/TestGatewayBudgetAllowsHealthySecondEntry |
| 4 | **冷却中条目**直接跳过（0ms）；冷却到期自动恢复参选；**首字后失败绝不重放**（收集到的 delta 只含首条目部分内容 + STREAM_ERROR）；**客户端取消零健康伤害** | ai/TestHealthCooldownAndRecovery、ai/TestNoReplayAfterFirstToken、ai/TestClientCancelNoHealthDamage |
| 5 | **停止生成** = 客户端断开 → ctx cancel → 上游取消 → 部分内容落库 → 不发 [DONE] | api/TestStopGenerationPersistsPartial |
| 6 | **兑换**单事务 CAS：并发兑换同一码恰一成功；串行双花必败 | store/TestRedeemFlow、store/TestRedeemConcurrentDoubleSpend |
| 7 | **会话删除**级联删消息，同一事务，二次删除 = 404 | store/TestConversationDeleteCascades |
| 8 | **会话踢出**：登录换 sessionId，旧 token 401+SESSION_KICKED；SSE 端点同语义；**非 user 令牌打 SSE = 403 INVALID_USER_TOKEN（无 panic）** | api/TestSessionKickAndAuthGuards、api/TestStreamRejectsAdminToken |
| 9 | 队列满 → MODEL_QUEUE_OVERFLOW（当前请求较多，请稍后再试） | ai/TestQueueOverflow |
| 10 | 官方 402 余额不足 → 立即终止不换池；SSE 错误文案按 Node 映射逐字还原 | ai/TestOfficial402StopsImmediately、api/TestFullUserJourney（SSE 断言） |

## 5. 测试与验收策略

1. 单元：Router 状态机（假上游：慢/死/快/预算/冷却/取消/reasoning 过滤/official-only）、store（seed/兑换/级联/公告/设置/AddMonths 语义）、service（上下文轮次/偏好后缀/日期归一/会员判定）。
2. 集成：`httptest` 全栈（真 SQLite 临时库 + 假 OpenAI 上游）：完整用户旅程、会话踢出、管理面、SSE 端到端。
3. `go test -race ./...` 全绿为合入线。
4. 前端：vite build + 既有回归套件（55 项）全绿。
5. 对抗式审查：独立子代理按攻击面轮次猎杀（契约/认证/Router/数据/SSE/资源/注入/前端/配置/部署），每轮发现 → 修复 → 全量回归 → 下一轮，≥10 轮。
6. 部署验收：yun1 全流程 + `health-yun.sh` 双机绿 + 生产/二节点无损确认。

## 6. 与原始 Plan 的偏差声明（请评审重点裁定）

| 偏差 | 理由 | 风险控制 |
|---|---|---|
| 不引入 shadcn/ui，保持既有设计语言 | D8；换装=全量回归风险，Plan 本意是"好用"不是"换框架" | Plan 九~十三逐条落地，回归测试锁定 |
| 兑换沿用 `durationMonths`（=1 月，约 30 天）而非字面"固定 30 天" | 兼容存量 redeemCodes 数据与管理端语义；durationMonths=1 与 30 天在用户感知上等价（月长 28-31 天） | AddMonths 语义与 Node 逐例一致；如需严格 30 天只改 `Redeem` 一处 |
| Bifrost/CPA 未在本次实际接入（gcp 侧未就绪），以配置位就绪 | D3；协议位一致，接入=填一段 JSON + 重启 | 假上游 E2E 已锁 failover 语义 |
| Telegram Bot 完整实现但默认关闭（大陆服务器直连 api.telegram.org 不可达） | 不影响主链路；提供 TG_PROXY 出口 | 全部命令走 service 层，白名单 |
| 桌面端侧栏为自绘（非 shadcn 组件） | 同 D8 | 响应式断点与移动抽屉行为有测试 |
| 兑换基数：Node 以激活时刻起算（覆盖剩余会员）；Go 以未过期到期时间起算叠加 | 用户预期"续费"而非"重置"；存量数据无迁移（beforeExpireAt 照记） | 行为差异记录于本表；回退只需改 service.Redeem 基数一行 |
| 会话标题时区：Node 用服务器本地（容器 UTC），Go 用 +8 生成 | 标题可读性（用户在 +8 时区） | 纯展示层，无数据语义 |
| 病态上游把 emoji 按 UTF-16 代理对劈进不同 delta 时，Go 渲染 U+FFFD（Node 因 JS code-unit 语义可拼接复原） | 无损修复需要跨 delta 的 code-unit 状态机，复杂度/风险不匹配；真实 DeepSeek/合规中转按完整字符分块 | 记录为已知限制；SSE、持久化、上下文回传三者内容一致，无安全影响 |

## 7. 验收清单（评审/测试代理逐项核对）

- [ ] Node→Go 契约逐端点比对通过（含错误文案逐字）
- [ ] `go test ./... -race` 全绿；前端 build + 回归 55/55
- [ ] §4 十项断言全部映射到具名自动化测试且全绿
- [ ] 本地真实 DeepSeek 流式 OK；假免费池 failover 实测 OK
- [ ] yun1 测试位全流程 OK；health-yun 双机绿；yun 生产与二节点零触碰
- [ ] secret 不入 git；.env 600；镜像不在 yun1 构建
