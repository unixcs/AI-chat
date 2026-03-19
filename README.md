# demo

这是一个按你需求搭建的前后端演示项目，包含：

- 用户端：首页、登录、注册、对话、侧边栏、个人中心、修改密码
- 会员逻辑：未开通/过期时禁止对话发送
- 兑换码逻辑：兑换后到期时间 = 激活时间 + 兑换码有效月数（默认 1 个月）
- 管理端：后台登录、首页统计、用户管理、角色管理、菜单管理、会员管理、兑换码管理、兑换记录、会话管理（可查看完整对话）

## 项目结构

- `frontend`：Vue 3 + Vite + Pinia + Vue Router
- `backend`：Express + SQLite

## 启动方式

### 1) 启动后端

```bash
cd backend
npm install
npm run dev
```

后端地址：`http://localhost:3001`

后端会读取 `backend/.env`，请填入：

```env
DEEPSEEK_API_KEY=你的_deepseek_api_key
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
```

### 2) 启动前端

```bash
cd frontend
pnpm install
pnpm dev
```

前端地址：`http://localhost:5173`

前端已配置代理 `/api -> http://localhost:3001`。

## 默认账号

### 用户端

- 手机号：`15555166986`
- 密码：`simple1270.`

### 管理端

- 地址：`/admin/login`
- 账号：`admin`
- 密码：`admin123`

## 说明

- 后端数据文件：`backend/data.sqlite`
- 流式对话接口：`GET /api/chat/conversations/:id/stream`（SSE）
- 管理后台已支持：编辑用户、禁用/启用、重置密码、兑换码作废、筛选、分页、导出
- 这是可运行的原型版，便于你快速确认流程和界面。
- 对话已支持 DeepSeek 实时流式返回。
