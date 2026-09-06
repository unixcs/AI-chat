# AI-chat Go 重构任务全记录

> 2026-09-06 ｜ 执行分支 `go-rewrite` ｜ 最终 commit `293cb49` ｜ 部署：yun1 测试位（8189/3002）
> 本文档是"从原始 Plan 到上线验收"全流程的完整档案：方案评审、十轮对抗审查、验收 bug 修复、两项新功能、测试矩阵、部署与运维。

---

## 0. 任务与约束

**原始需求**：基于用户提供的完整 Plan（`GO-CHAT-PLAN-ORIGINAL.md`，24 节）对 Node.js 版 AI-chat 做 Go 重写：AI Provider 解耦 + 2s 首字 failover 路由 + Gateway/Official 双模式 + 模型健康度 + 成本可观测 + 前端聊天体验打磨 + TG 管理 Bot。

**过程要求**（用户指定）：①方案从第一性原理取默认最优解，先交子代理评审，通过方可实施；②完成后代码复核；③用子代理做对抗式审查测试，**最少 10 轮**，反复打磨；④全部完成后输出本文档。

**红线**：yun 生产零触碰；yun1 现有二节点（8181/3001）零触碰；镜像不在 yun1 构建（内存红线）；secret 不进 git。

## 1. 时间线与 commit 索引

| 阶段 | Commit / 产物 | 内容 |
|---|---|---|
| 方案 | `docs/GO-REFACTOR-PLAN.md` | 第一性原理方案（D1-D11 决策、十断言表、偏差声明） |
| 方案评审 R1 | FAIL 9/12 | 9 项缺陷 → 修正重审 |
| 方案评审 R2 | **PASS 10.5/12** | `3a08afb` + `a148cc1` 落地评审修复 |
| 实施 | `fc6410c` | backend-go 全量 + 前端 UX 迭代 |
| 对抗 W1-W4 | `b7d0d5e` `5bc66bc` `5884b60` `417dd07` | 对抗轮 R1-R8 的修复落地（见 §4） |
| 对抗 R10 | ALL GREEN 15/15 | 回归狩猎；BOM 首行 LOW 修复入 `4258d82` |
| 验收 bug | `acdbad0` | 用户白天验收发现的 6 项（见 §5） |
| 新功能 | `docs/PROMPT-BOT-PLAN.md`（v2 评审 PASS 8.5/10）+ `97ab9ee` | 提示词管理 + TG Bot 补全（见 §6） |
| 对抗 R11 | `293cb49` | 最终轮 1×P1 + 7×P3 全修（见 §7） |
| 部署 | `deploy-yun1-test.sh` + 冒烟 | yun1 重部署，双机健康门禁 ALL GREEN（见 §9） |

## 2. 方案要点（评审后的目标架构）

- **P1 写放大**：Node `writeDb()` 全表 DELETE+INSERT → Go 精准 SQL + 事务；WAL + busy_timeout 5000。
- **P2/P3 模型单点与重试错位**：`ProviderEntry` 有序池 + Router 状态机——2s 首字看门狗（官方兜底豁免）、6s 免费池累计预算（从获槽起算，排队不烧）、首字后绝不重放、客户端取消零健康伤害、402 立即终止、指数冷却按 10s 事故窗分组。
- **契约**：路径/状态码/`{code:0,data}`/错误文案逐字兼容 Node；SSE `data:{...}`/`[DONE]` 协议一致；bcrypt/JWT 双向互通（Node 库直接接管）。
- **并发**：信号量 + 队列上限，溢出 `MODEL_QUEUE_OVERFLOW`（同 Node 文案）。
- **明确不做**（Plan §22 对齐）：微服务/Redis/ORM/计费预算/复杂角色系统。

## 3. 对抗式审查十轮全记录

> 形式：独立子代理按攻击面猎杀，每轮输出 `VERDICT: SOLID/BREACH` + [P0-P3] 带证据发现项 → 修复 → 全量 `-race` 回归 → 下一轮。工作目录 `/tmp/adv-r1..r11`（仓库零污染）。

| 轮 | 攻击面 | 结论 | 关键发现与修复 |
|---|---|---|---|
| R1 | HTTP 契约 | BREACH | 11 项：CORS 缺失、兑换记录 code 空、幽灵消息、分页语义、null 契约、LIKE 通配符未转义等 → `b7d0d5e` |
| R2 | 认证/会话 | SOLID | 2 项加固：审计 operator 透传、非 user 令牌打 SSE 403 |
| R3 | 注入 | SOLID | 4 项：escapeLike、宽容 JSON 解析收紧等 |
| R4 | Router 状态机 | BREACH | P1：排队时间烧网关预算（预算时钟改从获槽起算）；取消错误码归位；冷却按事故分组防爆炸 → `5bc66bc` |
| R5 | SSE 通道 | 发现 | P1 写竞争（单写者 goroutine + done 屏障）；空响应假错误；数字 content 强转；BOM 容忍 → `5884b60` |
| R6 | 数据完整性 | 发现 | 3×P0：兑换并发双花（事务内 no-op 写锁串行化）、注册并发 twin 账号（UNIQUE 索引 insert-first）、孤儿消息（启动清扫） |
| R7 | 资源/限流 | SOLID | 2 项：慢速连接钉死（上游 headers-only 挂起）、总超时用例 |
| R8 | 前端 | BREACH | 超长密码 400、ack 404 契约、偏好回填竞态；P1 代理对劈裂记为已知限制 → `417dd07` |
| R9 | 配置/启动 | ⚠️ 未派生 | 平台并发限制两次启动失败；检查面（幂等迁移、并发初始化、优雅停机）并入 R10 门禁与部署冒烟覆盖 |
| R10 | 全量回归狩猎 | **ALL GREEN 15/15** | 前 8 轮修复全部实测复核（假上游 + 4 后端实例黑盒）；LOW：流首行 BOM 在 `data:` 前之前的整行被跳过 → `4258d82` 修复；另证冷却事故分组正确（5 并发打死入口 consec=1 而非 480s 爆炸） |
| R11 | 新功能+bug修复+回归 | 发现 | 1×P1 + 7×P3，全修，见 §7 → `293cb49` |

**R10 详细核对清单**（15 项全过）：断连零重放、排队不烧预算、官方 3s 首字豁免看门狗、取消零健康伤害、兑换并发恰一成功+叠码精确+记录 code 非空、20 并发注册恰一成功、上游 500 中断落库、流中删会话无孤儿、CORS 头、总超时 804ms 触发、health 裸对象、前端 build+58/58、Go 五包 -race、审计日志落盘、数字 content+BOM 渲染。

## 4. 验收 bug 修复（`acdbad0`，用户白天验收发现 → 标准流程：定位→建议→修复→对抗审查→测试）

| # | 现象 | 根因 | 修复 | 验证 |
|---|---|---|---|---|
| 1 | 设置会员到期时间，确定无反应 | `NormalizeMemberExpireAt` 无 `datetime-local`（T 分隔无时区）布局，解析失败被前端静默吞掉 | 新增 naive 布局按 +8 墙钟解析（所见即所得往返）；弹窗报错改为可见 | `member_expire_test.go` 全矩阵（无秒/带秒/空格/带 Z/偏移/非法/空串） |
| 2 | 公告重新设置后用户重登收不到 | 管理员编辑沿用原公告 id，用户已读记录仍在，未读查询被排除 | 标题/内容**实际变更**时同事务原子清 `announcementReads`（重推全员）；仅切 active 不清（不骚扰）；补 audit | 四态 store 测试 + R11 黑盒复测（§9 冒烟 B 段） |
| 3 | 首页中间多余历史列 | ChatView 桌面布局冗余 | 删除中间列；历史统一经主区域「历史」按钮抽屉加载（移动端入口保留） | 前端结构回归测试 |
| 4-6 | 三处管理端手机端 UI（用户管理字体拥挤按钮拉伸/会话管理竖排/公告管理竖排） | 表格在窄屏挤压 | 用户/公告表格 ≤640px 卡片化（data-label 标签行）；会话管理改左列表(5fr)右详情(7fr)分栏各自内滚 | 前端回归 55→57 |

## 5. 新功能（`97ab9ee`，Plan v2 评审 PASS 8.5/10，F1-F8 全部吸收）

### 5.1 管理员后台「内置提示词管理」
- **存储**：`promptRevisions` 追加表（version UNIQUE，当前生效=最大版本行）；"恢复"= 旧内容生成新版本，历史永远线性无分叉；版本分配走进程内写互斥（`promptVersionMu`），10 并发保存版本严格 1..10。
- **生效**：`EffectiveSystemPrompt()` 每聊天请求读库一次（微秒级，无缓存失效面）；消息数组在流式准备阶段冻结 → 新聊天即效、进行中请求保持起始版本；读库失败降级 env 配置并记日志。用户「回答长度/风格」仍叠加在尾部，互不冲突。
- **API**（全 adminAuth + audit）：`GET/PUT /api/admin/prompt`、`GET /api/admin/prompt/revisions`（近 50）、`POST /api/admin/prompt/restore`；≤20000 字节校验。
- **前端**：`AdminPromptView.vue`（编辑器 + 版本/时间/操作人徽标 + 字节计数 + 版本卡片恢复，confirm 二次确认）；菜单接线两处（DB 种子 + AdminLayout 硬编码组）。
- **Bot 侧**：`/prompt` 查状态、`/prompt set <v>` 切换（=Web 恢复，同一 Service 层）。

### 5.2 Telegram Bot「高频管理员操作」补全
- 新增：`/redeemlog [n]`（最近兑换记录）、`/codeinfo <码>`（未使用/已作废/已使用三态：兑换人/时间/会员至）、`/prompt`、`/prompt set <v>`；`/user` 会员三态（未开通/有效期至/已过期）；`/reg`、`/redeem` setter 错误回显（不再 `_ =` 吞错）。
- 既有保留：`/code /codes`（固定 1 个日历月≈30 天，文案已对齐）、`/ban /unban /announce /status /ai`。
- 安全架构：Bot → 统一 Admin Service → Store（Bot 零 SQL）；`TG_ADMIN_IDS` int64 白名单（UNAUTHORIZED 落 [tg-audit]）；token 仅环境变量；所有关键操作 audit（operator=`tg:<id>`）；输出 4096 上限防御截断（字节预算 + rune 边界回退）；未配 `TG_BOT_TOKEN` 零开销不启动。

## 6. R11 最终轮发现与修复（`293cb49`）

| 级别 | 问题 | 修复 |
|---|---|---|
| P1 | 并发保存提示词时 PUT/restore 响应的 version 是"落库后重读头版本"，两个客户端可收到相同（错误）版本号，T6 门禁 40-50% 闪烁 | `CreatePromptRevision` 返回本次写入整行，响应由该行构造，绝不重读头版本；T6 连跑 20 次 -race 全绿 |
| P3 | TG 截断按字节切，多字节中文切成非法 UTF-8 | 截断点回退 rune 边界（字节预算 ≤4000 恒 ≤4096 UTF-16 单元）+ 新测试锁 astral 字符 |
| P3 | `announcementReads` 无唯一约束，重复 ack 虚增已读人数 | 迁移去重 + 唯一索引 `idx_read_ann_user_uniq` |
| P3 | ack 与公告改版竞态：用户读旧版、admin 改版清已读、ack 晚到 → 新内容被标已读 | ack 携带 `asOf`（用户所见 updatedAt）指纹，过期 ack 静默丢弃保持未读；前端随行发送；无指纹的旧客户端行为不变 |
| P3 | 读库失败静默降级 env 无日志 | 降级前 log 警告一行 |
| P3 | `BotPromptSet` 死代码 | 删除（Bot 直接走 `AdminRestorePrompt` 同一 Service） |
| P3 | 恢复确认弹窗预推算"vN+1"并发下会错 | confirm 文案去掉具体版本号 |
| P3 | 编辑器字数按字符与 20000 字节上限口径不一 | 改 `TextEncoder` 字节计数 `/ 20000 字节` |

## 7. 测试矩阵（最终态全绿）

| 套件 | 数量/结果 |
|---|---|
| `go test ./... -race`（backend-go 五包） | ai 14 / api 11 / bot 7 / service 7 / store 11 = **50 个测试函数全绿**（ai 19s / api 26.7s / bot 6.4s / service 3.5s / store 12.6s） |
| T6 并发版本门禁 | `-race -count=20` **20/20**（修复前 40-50% 闪烁） |
| 前端 | vite build 389ms + `node --test` **58/58** |
| R10 黑盒回归 | 15/15（4 后端实例 + 假上游工装，/tmp/adv-r10） |
| 真实 DeepSeek E2E | 本地首字 1.77s；yun1 冒烟 SSE 含 [DONE]、零错误事件、ai status totalRequests 增长 |

## 8. 部署与线上验收（2026-09-06 11:06）

- 流程：WSL 构建 `ai-chat-go-backend/frontend:latest` → `docker save | gzip | ssh docker load` → `deploy-yun1-test.sh` 重建 `ai-chat-go-*` 容器（compose 项目 `ai-chat-go-test`，/opt/ai-chat-go-test，数据卷保留 → 昨日验收数据/公告无损迁移，新表/索引幂等生效）。
- **冒烟验收 25 项有效断言全过**：health；admin 登录；prompt PUT（响应版本=本次写入 v1）/revisions/restore（v2）/空内容 400/无 token 拒绝；注册→公告未读到达→**过期 asOf ack 被丢弃仍提示**→正确 ack 成功→**重复 ack readCount 恒 1**；邀请码兑换→**真实 DeepSeek 流式**（"1+1等于几？"→"2。"，[DONE]，无错误事件）→ai status；**公告改版重推**（改内容即重新通知、仅切 active 不重推）；清理冒烟数据。
  - 记录：首份冒烟脚本 2 个 ❌ 均为脚本自身断言缺陷（①未考虑库内尚有昨日测试公告未读——产品行为正确；②DELETE 少传参触发 set -u），复核后以冒烟 B 段复验通过。
- **健康门禁**：`ops/health-yun.sh all` → yun + yun1 **ALL GREEN**（DNS/容器/health 端点）；yun 生产与 yun1 二节点（8181/3001）全程零触碰（`ai-chat-frontend/backend` Up 不间断）。

## 9. 运维手册要点

```bash
# 更新部署（WSL 执行；镜像本地构建，yun1 零构建）
bash deploy/deploy-yun1-test.sh          # 数据保留；回滚=重载旧镜像 tag 或 compose down

# 双机健康门禁（任何 yun/yun1 运维操作后必跑）
bash /mnt/vps/tencent/Remote\ AI\ Coding/ops/health-yun.sh all

# 启用 TG Bot（yun1 backend/.env 追加，token 不进 git）
TG_BOT_TOKEN=...  TG_ADMIN_IDS=12345678  TG_PROXY=http://...   # 不配 = Bot 不启动

# Gateway 免费池接入（现 AI_MODE=official）
AI_MODE=gateway  AI_PROVIDERS_JSON='[{"name":"cpa","baseURL":"...","apiKey":"...","model":"..."}]'
```

- 验收入口：`http://121.41.206.32:8189`（tailnet `http://100.96.106.50:8189`）；管理 `admin/admin123`；演示邀请码 `VIP-20260905-Q5YV2Y`。
- 生产切换（未来）：合并 go-rewrite → main 后按 `DEPLOYMENT_WORKFLOW.md` 备份→最小同步→重建→验证，不在本次范围。

## 10. 已知限制与偏差声明（全部有意取舍，详见 GO-REFACTOR-PLAN §6）

- 不引入 shadcn/ui（保持既有设计语言）；兑换按 durationMonths=1 日历月（非字面 30 天）；会话标题 +8 时区（Node 为容器 UTC）；兑换基数从到期时间叠加（Node 为激活时刻重置）；Bifrost/CPA 配置位就绪未实接；**病态上游把 emoji 按 UTF-16 代理对劈进不同 delta 时渲染 U+FFFD**（Node 可复原；真实 DeepSeek 不触发）；TG"预设切换"= 历史版本切换，命名预设留二阶段。

## 11. 产物索引

| 文件 | 内容 |
|---|---|
| `docs/GO-CHAT-PLAN-ORIGINAL.md` | 用户原始 Plan 存档（评审基准） |
| `docs/GO-REFACTOR-PLAN.md` | 重构方案（评审通过版） |
| `docs/PROMPT-BOT-PLAN.md` | 新功能 Plan v2（评审 PASS 8.5/10） |
| `docs/GO-REWRITE-REPORT.md` | **本文档** |
| `backend-go/` | Go 后端（internal/{api,service,store,ai,bot,auth,config,model}） |
| `deploy/deploy-yun1-test.sh` + `docker-compose.yun1-test.yml` | 测试位部署 |
| `/tmp/adv-r1..r11`（服务器本地，不入库） | 十一轮对抗审查工装与证据 |
