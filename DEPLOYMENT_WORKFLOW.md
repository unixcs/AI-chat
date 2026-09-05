# AI-chat 开发 → 测试 → 生产 → 收尾 工作流总结

> 基于 2026-09-05 实际执行记录整理。
> 适用服务器：云服务器（当前生产 121.41.192.80，入口 admin）。
> 核心原则：**测试隔离、数据不丢、生产只做最小变更、先备份再重建、验收通过再发布。**

## 1. 一图流

```text
本地开发 / 修改
      │ git commit（本地分支 master）
      ▼
部署到测试服（8189/3002）：
  /home/admin/ai-chat-test  → docker compose up -d --force-recreate --build -> 人工验收
      │ 验收通过
      ▼
部署到生产服（8181/3001）：
  备份 → 同步最小文件 → 合并 compose → 更新 .env → 重建容器 → 自动验证 → 人工验收
      │ 验收通过
      ▼
Push 到 GitHub（main）
      ▼
关闭测试服：
  /home/admin/ai-chat-test → docker compose down（保留数据，不删 volume）
```

## 2. 云服务器目录/端口约定

| 环境 | 路径 | 端口 |
|---|---|---|
| 生产 | `/opt/AI-chat` | frontend `8181 -> 80`，backend `3001 -> 3001` |
| 测试 | `/home/admin/ai-chat-test` | frontend `8189 -> 80`，backend `3002 -> 3001` |
| 生产数据 | `/srv/ai-chat/data/data.sqlite` | 必须只备份、不动 |
| 测试数据 | `/home/admin/ai-chat-test/data` | 关闭测试服时保留 |

## 3. 标准发布步骤（可直接复用）

### 3.1 本地完成并提交

```bash
cd /mnt/vps/yun/AI-chat
git checkout master
git add .
git commit -m "fix: ..."
git status
```

### 3.2 部署到测试环境（8189）

```bash
# 服务器上确保测试目录已存在
ssh admin@121.41.192.80

# 代码同步：把本地后端/前端源码上传到 /home/admin/ai-chat-test
# 可以直接 scp 或工具同步（注意避开 node_modules、.git、data、.env）

# 如果新增了环境变量（如 DEEPSEEK_EXTRA_BODY），只追加到测试 .env
cd /home/admin/ai-chat-test
grep -q '^DEEPSEEK_EXTRA_BODY=' backend/.env || \
  echo 'DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":1000}' >> backend/.env

# 关键：必须 force-recreate，因为 docker restart 不重新加载 .env
docker compose up -d --force-recreate --build

# 验证
curl -s http://127.0.0.1:8189/api/health
docker exec ai-chat-test-frontend grep -E 'proxy_(read|send)_timeout' /etc/nginx/conf.d/default.conf
docker exec ai-chat-test-backend printenv DEEPSEEK_EXTRA_BODY
```

### 3.3 人工验收测试环境

- 登录、发消息、看流式回复首字延迟。
- 确认不再出现 60s 超时/「联系管理员」。
- 确认对话列表打开不卡。
- 注意核对 `8189` 与生产 `8181` 数据是独立库。

### 3.4 部署到生产环境（8181）——重点步骤

```bash
# 1. 进入生产目录
cd /opt/AI-chat

# 2. 备份当前状态（必须）
STAMP=$(date +%Y%m%d-%H%M%S)
mkdir -p backups/predeploy-$STAMP
sqlite3 "file:/srv/ai-chat/data/data.sqlite?mode=ro" ".backup 'backups/predeploy-$STAMP/data.sqlite'"
sqlite3 backups/predeploy-$STAMP/data.sqlite "PRAGMA integrity_check;"

cp -a backend/.env              backend/.env.bak.$STAMP
cp -a docker-compose.yml        docker-compose.yml.bak.$STAMP
cp -a backend/prompts/Prompt.md backend/prompts/Prompt.md.bak.$STAMP
docker tag ai-chat-backend:latest  ai-chat-backend:predeploy.$STAMP
docker tag ai-chat-frontend:latest ai-chat-frontend:predeploy.$STAMP

# 3. 同步生产应该同步的最小文件（不要 copy 整个目录覆盖）
#    例如：backend/db.js、backend/server.js、frontend/nginx.default.conf、
#    frontend/src/... 以及必要的 Prompt.md

# 4. 合并 docker-compose（保留 8181:80）
#    用 Python 脚本检查是否已有 dns，如果没有则插入 backend dns。
#    切勿直接用 GitHub 的 docker-compose.yml，因为它的 frontend 端口是 80:80。

# 5. 更新生产 .env（只追加，不覆盖）
grep -q '^DEEPSEEK_EXTRA_BODY=' backend/.env || \
  echo 'DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":1000}' >> backend/.env
# 注意：一定用 UTF-8 写，并检查不要带转义反斜杠。
python3 -c "import json,re; p='/opt/AI-chat/backend/.env'; s=open(p).read(); m=re.search(r'^DEEPSEEK_EXTRA_BODY=(.*)$', s, re.M); print(json.loads(m.group(1)))"

# 5. 重建
docker compose up -d --force-recreate --build

# 6. 验证
docker compose ps
curl -s http://127.0.0.1:8181/api/health
docker exec ai-chat-frontend grep -E 'proxy_(read|send)_timeout|send_timeout' /etc/nginx/conf.d/default.conf
docker exec ai-chat-backend printenv DEEPSEEK_EXTRA_BODY
docker compose logs --since=2m backend frontend | tail -50
# 数据库计数对比
sqlite3 "file:/srv/ai-chat/data/data.sqlite?mode=ro" \
  "SELECT (SELECT count(*) FROM conversations), (SELECT count(*) FROM messages), (SELECT count(*) FROM users);"
```

### 3.5 生产人工验收

- 在正式地址 `http://121.41.192.80:8181` 发消息。
- 确认首字更快、无 60s 卡顿、历史聊天/登录态正常。
- 确认后台账单后续 token 量下降（可等 1~2 天看 DeepSeek 官方后台）。

### 3.6 关闭测试服（只关闭服务，保留数据）

```bash
ssh admin@121.41.192.80
cd /home/admin/ai-chat-test
docker compose down          # 不要加 -v，否则删除数据

# 验证端口释放
ss -ltnp | grep -E ':(8189|3002)' || echo 'ports free'
docker ps -a --filter name=ai-chat-test
```

### 3.7 推送到 GitHub

```bash
cd /mnt/ai-chat
git push origin master:main
```

## 4. 踩过的坑与要点

1. **`.env` 变更必须重建容器**：`docker restart` 不会重新读取 `env_file`，必须 `docker compose up -d --force-recreate`。
2. **生产 compose 端口不要被 GitHub 覆盖**：仓库默认是 `80:80`，生产必须保留 `8181:80`。
3. **生产 `.env` 只追加**：不要整文件覆盖，否则会覆盖 API key、JWT_SECRET 等。
4. **追加环境变量如果用 shell echo 要注意 JSON 转义**：最好用文件写入工具/Python 写，避免在 .env 里出现 `\"`。
5. **数据库永远只备份、不迁移**：生产 DB 在 `/srv/ai-chat/data`，部署过程不触碰。
6. **备份时一定要做 `integrity_check`**，确保备份可用。
7. **生产部署顺序**：先备份→再改代码→再重建→再验证→最后 push。
8. **测试服关闭用 `docker compose down`，不要 `-v`**，否则测试数据丢了之后无法快速复现问题。
9. **对着 GitHub push 前先确认本机已配好凭据**（PAT 或 SSH key），否则 push 会因认证失败中断。
10. **DNS 根治**：如果上游 API（如 DeepSeek）出现过 DNS 故障，建议在 compose 里给 backend 指定国内 DNS（如 `223.5.5.5`, `119.29.29.29`）并持久化。
11. **DeepSeek 费用优化**：生产已开启 `DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":1000}`，关闭推理、限制输出长度，能明显降费提速度。

## 5. 后续建议

- 把以上步骤固化成 `deploy.sh` / `Makefile`，部署时只跑一个脚本（但每次变更都要重新审）。
- 后续若做 Go 重写，可先起 `future/go-backend` 分支，保持前端不变，只替换 SSE/API 后端。
