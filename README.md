# AI Chat

前后端分离的 AI 对话与会员管理平台：用户手机号注册登录，兑换码开通会员后进行流式 AI 对话；管理后台覆盖用户、会员、兑换码、公告、提示词等完整运营面。

2026-09 完成两轮大改并已上生产：

- **后端 Node.js → Go 重构**：API 契约 100% 兼容（前端零改动可跑），JWT 互相签发/校验兼容，SQLite 数据库同表名增量迁移（老库直接接管，用户无感升级）。
- **全站 UI 重建**：Tailwind CSS v4 + shadcn-vue 风格组件（reka-ui，组件源码进仓库），浅色/深色双主题，移动端自适应。

## 1. 功能概览

### 用户端
- 手机号注册/登录（JWT；同账号重复登录会使旧会话失效）
- AI 流式对话（SSE）+ Markdown 渲染，历史会话管理
- **三维度回答偏好**：发送按钮旁「调节」面板，长度 / 风格 / 输出格式 三个维度自由组合（默认「适中 + 标准」），对新对话即时生效；偏好对应的提示词卡片由管理员在后台维护、改完热生效
- 站内公告：进站一次性弹窗，管理员修改公告内容后自动对全员重推
- 个人资料、修改密码、卡密充值（兑换码激活会员）
- 对话文字支持长按选择复制（选中高亮对比度已按浅/深色分别调校）

### 会员规则
- 兑换码激活：`memberExpireAt = 激活时间 + 兑换码时长`
- 非会员或会员过期禁止发言，返回固定文案「会员过期，请续费后使用」

### 管理端（`/admin`，默认账号见 §4）
- 控制台统计
- 用户管理（查询、启停用、重置密码、会员时长调整）、角色管理、菜单管理
- 兑换码管理（批量生成、作废、筛选、导出）、兑换记录
- 会话审计（可查看完整消息内容）
- 公告管理、提示词版本管理（恢复 = 旧内容生成新版本，历史可追溯）
- AI 状态页

### AI 接入（Go Router，双模式）
- `AI_MODE=official`：仅 DeepSeek 官方 API
- `AI_MODE=gateway`：OpenAI 兼容免费模型池（`AI_PROVIDERS_JSON` 有序列表）→ 自动故障切换 → 官方兜底
- 切换策略：免费池单模型首字 2s 超时即换 → 池内累计预算 6s 后直接官方兜底；故障模型健康冷却 30s 起指数封顶 10min，到期自动复探测
- 可选 Telegram 管理机器人（`TG_BOT_TOKEN`，不配则不启动）

## 2. 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.27（标准库 `net/http`）、`modernc.org/sqlite`（纯 Go SQLite，无 CGO）、`golang-jwt/v5`、bcrypt |
| 前端 | Vue 3 + Vite、Tailwind CSS v4（CSS-first token）、reka-ui、Pinia、markdown-it + DOMPurify、lucide-vue-next |
| 存储 | SQLite 单文件（WAL）；schema 增量迁移幂等，老版本数据库可直接接管 |
| 部署 | Docker Compose；前端容器 Nginx 承载静态资源并反代 `/api` |

## 3. 项目结构

```text
AI-chat/
├─ backend-go/                  # Go 后端（现行版本）
│  ├─ cmd/server/               # 入口（默认端口 3001）
│  ├─ internal/
│  │  ├─ api/                   # HTTP 路由与 handler（含 SSE）
│  │  ├─ ai/                    # Provider/Router（免费池 + 官方兜底）
│  │  ├─ auth/                  # JWT
│  │  ├─ bot/                   # Telegram 管理 Bot（可选）
│  │  ├─ config/                # 环境变量配置
│  │  ├─ model/ service/ store/ # 领域模型 / 业务 / SQLite 存储
│  ├─ prompts/Prompt.md         # 内置基础提示词
│  └─ Dockerfile                # golang:1.27-alpine 多阶段 → alpine:3.20
├─ frontend/                    # Vue3 前端
│  ├─ src/components/ui/        # shadcn-vue 风格组件（源码入仓库）
│  ├─ tests/                    # node --test 回归 + 暗色双机制校验脚本
│  └─ Dockerfile                # node:22-alpine 构建 → nginx:alpine
├─ deploy/
│  ├─ deploy-yun1-test.sh       # 测试服一键部署/更新（本地构建镜像 → ssh load → --no-build 拉起）
│  └─ docker-compose.yun1-test.yml
├─ docs/                        # 现行文档（见 §7 文档索引）
├─ docker-compose.yml           # ⚠️ Node 老版遗留示例（build ./backend），勿用于 Go 版
└─ backend/                     # ⚠️ Node 老版后端，2026-09-06 已退役，仅存档（完整历史见 commit 4b7bfa9）
```

## 4. 本地开发

### 后端（Go）

```bash
cd backend-go
cp .env.example .env      # 至少填 DEEPSEEK_API_KEY、JWT_SECRET
go run ./cmd/server       # 默认 http://localhost:3001
```

- 健康检查：`GET /api/health`
- 空库首次启动自动创建最小系统数据（管理员、角色、菜单），不创建演示用户/兑换码/消息
- 默认管理员：`admin / admin123` —— **生产环境首次登录后立即修改**

### 前端（Vue3）

```bash
cd frontend
npm install
npm run dev               # http://localhost:5173，/api 代理到 localhost:3001
```

## 5. 测试

```bash
cd backend-go && go test ./...
cd frontend && node --test tests/*.test.js
cd frontend && python3 tests/verify-dark-mode.py   # 暗色双机制（data-theme + html.dark）校验
```

## 6. Docker 部署

生产做法（小内存服务器也适用）：**镜像在本地/CI 构建，传输到服务器 `docker load`，服务器上只 `--no-build` 拉起**——不要在小内存机器上原地 build。

```bash
# 本地构建
docker build -t ai-chat-go-backend:latest  ./backend-go
docker build -t ai-chat-go-frontend:latest ./frontend

# 传输（示例）
docker save ai-chat-go-backend:latest | gzip -1 | ssh HOST 'gunzip | docker load'
docker save ai-chat-go-frontend:latest | gzip -1 | ssh HOST 'gunzip | docker load'
```

服务器上准备目录与 compose（可直接参考 `deploy/docker-compose.yun1-test.yml`，改端口/项目名即可）：

- backend：`env_file` 挂 `.env`，数据卷 `<数据目录>:/app/data`，环境变量 `SQLITE_PATH=/app/data/data.sqlite`
- frontend：80 端口；容器内 Nginx 已配好 `/api/ → http://backend:3001/api/`，**SSE 必需项已带**（`proxy_buffering off` + 长超时），服务名必须叫 `backend`
- 拉起：`docker compose -p ai-chat up -d --no-build`

后端环境变量全量说明见 `backend-go/.env.example`（每键带注释）：端口/库路径、`AI_MODE`、DeepSeek 四件套、Router 调参（首字超时/预算/并发队列）、`AI_PROVIDERS_JSON` 免费池、`JWT_SECRET`、TG Bot 三键。**密钥只进服务器上的 `.env`（chmod 600），绝不进 git。**

### 数据备份与升级

- 备份：`sqlite3 <库文件> ".backup '<目标文件>'"`（WAL 感知的一致性备份，比直接 cp 可靠）
- 升级 SOP（无缝、保数据）：`sqlite3 .backup` 一致性备份 → 备份副本挂一次性容器预演迁移并逐行对账 → compose override 文件只覆 `image` 字段 → `up -d --no-build` 切换 → 终对账。完整步骤见 `docs/HANDOVER-20260906-GO-REWRITE.md`（2026-09-06 已在生产 8181 实盘执行：停机约 45 秒、旧 token 免重登、数据零丢失）

## 7. 文档索引

| 文档 | 内容 |
|---|---|
| `docs/HANDOVER-20260906-GO-REWRITE.md` | 交接总入口：环境/版本锚点、部署与生产升级 SOP、升级实录 |
| `docs/GO-REWRITE-REPORT.md` | Go 重构全记录：架构、十轮对抗审查、验收数据 |
| `docs/PREF-UI-REDESIGN-PLAN.md` | 三维度回答偏好设计 + 全站 UI 重建方案与验收 |
| `docs/BRAND-GUIDELINES.md` | 前端设计规范：token、组件消费约定、语气 |
| `docs/GO-REFACTOR-PLAN.md` / `docs/GO-CHAT-PLAN-ORIGINAL.md` | 重构计划（修订版/原始版） |
| `docs/PROMPT-BOT-PLAN.md` | Telegram 提示词 Bot 设计 |

## 8. 安全建议（生产必做）

- 修改默认管理员密码；`JWT_SECRET` 必须换成强随机值
- `.env` 权限 600，不入库、不进镜像构建上下文
- 反代层上 HTTPS；安全组/防火墙仅放行 80/443
- 定期执行 §6 的 SQLite 备份并异地存放
- 日志轮转与监控告警

## 9. License

仅供学习与原型演示使用。
