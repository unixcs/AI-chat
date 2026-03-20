# Demo AI Chat Platform

一个前后端分离的 AI 对话与会员管理系统，包含用户端与管理后台。用户通过手机号注册登录，使用兑换码开通会员后可调用 DeepSeek 进行流式对话；管理员可管理用户、兑换码、兑换记录与完整会话内容。

## 1. 功能概览

### 用户端
- 首页/登录/注册（手机号唯一，不需要短信验证码）
- AI 对话（SSE 流式输出）
- 历史会话与消息查看
- 个人中心（昵称/头像）
- 修改密码
- 兑换码激活会员

### 会员规则
- 兑换后：`memberExpireAt = 激活时间 + 兑换码时长(月)`
- 默认兑换码时长：1 个月
- 非会员或会员过期：禁止发送对话消息

### 管理端
- 管理员登录
- 控制台统计
- 用户管理（查询、启停用、重置密码、编辑）
- 角色管理、菜单管理
- 会员管理
- 兑换码管理（生成、作废、筛选、分页、导出）
- 兑换记录管理
- 会话管理（可查看完整消息内容）

## 2. 技术栈与部署技术

### 前端
- Vue 3
- Vue Router 4
- Pinia
- Vite 8
- Axios
- Day.js
- Sass

### 后端
- Node.js + Express 5
- JWT（`jsonwebtoken`）
- 密码加密（`bcryptjs`）
- CORS
- dotenv
- Day.js
- SQLite（使用 Node 内置 `node:sqlite` 的 `DatabaseSync`）

### AI 模型接入
- DeepSeek Chat API
- SSE 流式响应
- 上下文拼接（最近多轮）

### 部署与运维（推荐）
- Linux 云服务器（Ubuntu 22.04/24.04）
- Nginx（前端静态资源 + API 反向代理）
- systemd（守护后端进程）
- Docker / Docker Compose（容器化部署方案）

## 3. 项目结构

```text
demo/
├─ frontend/                 # Vue 前端
│  ├─ src/
│  ├─ package.json
│  └─ vite.config.js
├─ backend/                  # Express 后端
│  ├─ server.js
│  ├─ db.js
│  ├─ auth.js
│  ├─ data.sqlite            # SQLite 数据文件（运行后生成/使用）
│  ├─ data.json              # 初始种子数据来源（首次初始化时可导入）
│  └─ package.json
└─ README.md
```

## 4. 运行环境要求

- Node.js：`>= 22`（重要：项目使用 `node:sqlite`）
- npm：`>= 10`（或 pnpm `>= 9`）
- Git
- 可选：Docker `>= 24`，Docker Compose `>= 2.20`

## 5. 本地开发启动

### 5.1 克隆项目

```bash
git clone https://github.com/unixcs/AI-chat
cd AI-chat
```

### 5.2 配置后端环境变量

在 `backend/.env` 写入：

```env
DEEPSEEK_API_KEY=你的_deepseek_api_key
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1

# 可选：并发与队列
MODEL_CONCURRENCY=6
MODEL_QUEUE_MAX=50
```

### 5.3 启动后端

```bash
cd backend
npm install
npm run dev
```

后端默认地址：`http://localhost:3001`

### 5.4 启动前端

```bash
cd ../frontend
npm install
npm run dev
```

前端默认地址：`http://localhost:5173`

说明：前端开发代理已配置 `/api -> http://localhost:3001`（见 `frontend/vite.config.js`）。

## 6. 默认账号与隐私说明

### 用户端
- 默认测试账号信息已隐藏（隐私保护）
- 如需测试，请在部署后由管理员在后台创建测试账号，或联系项目维护者获取临时账号

### 管理端
- 地址：`/admin/login`
- 管理员初始账号信息已隐藏（生产环境请首次登录后立即修改）

## 7. GitHub 拉取与更新流程

### 7.1 首次拉取

```bash
git clone https://github.com/unixcs/AI-chat
cd AI-chat
```

### 7.2 后续更新

```bash
git pull origin main
```

如果你的默认分支不是 `main`，请改成对应分支名（如 `master`）。

## 8. 云服务器部署（非 Docker，推荐生产）

以下示例基于 Ubuntu 22.04。

### 8.1 安装基础环境

```bash
sudo apt update
sudo apt install -y git curl nginx
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs
node -v
npm -v
```

### 8.2 拉取项目

```bash
cd /var/www
sudo git clone https://github.com/unixcs/AI-chat demo
sudo chown -R $USER:$USER /var/www/demo
cd /var/www/demo
```

### 8.3 部署后端

```bash
cd /var/www/demo/backend
npm install --production
```

创建环境变量文件：

```bash
cat > /var/www/demo/backend/.env << 'EOF'
DEEPSEEK_API_KEY=你的_deepseek_api_key
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
MODEL_CONCURRENCY=6
MODEL_QUEUE_MAX=50
EOF
```

创建 systemd 服务 `/etc/systemd/system/demo-backend.service`：

```ini
[Unit]
Description=Demo Backend Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/var/www/demo/backend
ExecStart=/usr/bin/node /var/www/demo/backend/server.js
Restart=always
RestartSec=5
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
```

启动并设置开机自启：

```bash
sudo systemctl daemon-reload
sudo systemctl enable demo-backend
sudo systemctl start demo-backend
sudo systemctl status demo-backend
```

### 8.4 部署前端静态资源

```bash
cd /var/www/demo/frontend
npm install
npm run build
```

构建产物在：`/var/www/demo/frontend/dist`

### 8.5 配置 Nginx

创建 `/etc/nginx/sites-available/demo.conf`：

```nginx
server {
  listen 80;
  server_name _; # 改成你的域名，例如 ai.example.com

  root /var/www/demo/frontend/dist;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api/ {
    proxy_pass http://127.0.0.1:3001/api/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_buffering off;
    proxy_cache off;
  }
}
```

启用站点并重载：

```bash
sudo ln -s /etc/nginx/sites-available/demo.conf /etc/nginx/sites-enabled/demo.conf
sudo nginx -t
sudo systemctl reload nginx
```

### 8.6 放行防火墙（如启用了 UFW）

```bash
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw enable
sudo ufw status
```

## 9. Docker 部署（完整命令与文件）

下面给出完整容器化方案（前端 + 后端 + Nginx 反代）。

### 9.1 服务器安装 Docker

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo $VERSION_CODENAME) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo systemctl enable docker
sudo systemctl start docker
docker --version
docker compose version
```

### 9.2 克隆项目并进入目录

```bash
git clone https://github.com/unixcs/AI-chat
cd AI-chat
```

### 9.3 创建后端环境变量

```bash
cat > backend/.env << 'EOF'
DEEPSEEK_API_KEY=你的_deepseek_api_key
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
MODEL_CONCURRENCY=6
MODEL_QUEUE_MAX=50
EOF
```

### 9.4 新建 Docker 文件

#### `backend/Dockerfile`

```dockerfile
FROM node:22-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --production
COPY . .
EXPOSE 3001
CMD ["node", "server.js"]
```

#### `frontend/Dockerfile`

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### `frontend/nginx.default.conf`

```nginx
server {
  listen 80;
  server_name _;

  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api/ {
    proxy_pass http://backend:3001/api/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_buffering off;
    proxy_cache off;
  }
}
```

#### 根目录 `docker-compose.yml`

```yaml
version: "3.9"
services:
  backend:
    build:
      context: ./backend
    container_name: demo-backend
    restart: always
    env_file:
      - ./backend/.env
    ports:
      - "3001:3001"
    volumes:
      - ./backend/data.sqlite:/app/data.sqlite

  frontend:
    build:
      context: ./frontend
    container_name: demo-frontend
    restart: always
    depends_on:
      - backend
    ports:
      - "80:80"
```

### 9.5 一键构建启动

```bash
docker compose up -d --build
docker compose ps
```

### 9.6 查看日志

```bash
docker compose logs -f backend
docker compose logs -f frontend
```

### 9.7 停止与重启

```bash
docker compose down
docker compose up -d
```

## 10. 常见问题排查

- 后端启动失败：确认 Node 版本是否 `>= 22`（`node -v`）
- AI 无响应：检查 `DEEPSEEK_API_KEY` 是否有效、余额是否充足
- 前端显示后端未启动：确认 `3001` 端口可访问，或检查 Nginx `/api` 反代
- Docker 下 SSE 不流式：确认反代关闭缓冲（`proxy_buffering off`）
- 数据保护：请定期备份 `backend/data.sqlite`

## 11. 安全建议（生产必做）

- 修改默认管理员账号密码
- 使用 HTTPS（Nginx + Certbot）
- 对 `.env` 做最小权限控制，不入库
- 使用云厂商安全组仅开放 `80/443`
- 增加日志轮转与监控告警

## 12. License

仅供学习与原型演示使用。
