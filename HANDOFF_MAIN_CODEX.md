# 交接文档：AI-chat 修复与测试环境

> 写给主会话 Codex / 后续接手者。
> 更新时间：2026-09-05 15:00 左右。

## 1. 一句话现状

本地已修好 `nginx 超时 + 关闭 DeepSeek 推理 + 限长输出`，测试环境 `http://121.41.192.80:8189` 已可验收；生产 `8181/3001` **未动**。

## 2. 现在的仓库与 commit

- 项目：`https://github.com/unixcs/AI-chat`
- 本地工作区：`/mnt/vps/yun/AI-chat/`
- 本地 master：`d748ab6`
  - 其中包含：`89b3a94`（nginx 超时/DNS持久化/分页/轮询降频/DEEPSEEK_EXTRA_BODY 支持）
  - 和 `d748ab6`（文档 + `.env.example` 说明）
- **本地未 push**，GitHub 仍是最旧状态，**服务器生产 /opt/AI-chat 也还是旧代码**。

## 3. 已部署的测试实例（服务器）

| 项 | 值 |
|---|---|
| 路径 | `/home/admin/ai-chat-test/`（因为 `/opt` 无写权限，没放在 `/opt`） |
| 前端容器 | `ai-chat-test-frontend`，端口 `8189 → 80` |
| 后端容器 | `ai-chat-test-backend`，端口 `3002 → 3001` |
| 数据 | `/home/admin/ai-chat-test/data/data.sqlite`（生产库 `sqlite3 .backup` 快照） |
| 生产 `ai-chat-frontend/ai-chat-backend` | 未改动 |
| 生产端口 `8181/3001` | 未改动 |

## 4. 测试环境已生效的配置

`/home/admin/ai-chat-test/backend/.env` 已加：

```env
DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":500}
```

⚠️ 环境变量依赖 `.env` 在 **容器创建时** 生效：

```bash
cd /home/admin/ai-chat-test
docker compose up -d --force-recreate backend
```

只 `docker restart` 不生效（之前踩过的坑）。

## 5. 已确认的效果

同样请求在关闭推理并设置 `max_tokens:500` 后：

- 健康检查：`GET http://121.41.192.80:8189/api/health` → `{"ok":true}`
- 容器内 `DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":500}`
- 探针请求结果：
  - `firstTokenMs: ~0.4~0.8s`
  - `totalMs: ~5.5~6.8s`
  - `hasReasoning: false`
  - `reasoningChars: 0`

## 6. 本地已完成的代码改动列表

### 已提交 `89b3a94`

| 文件 | 改动 |
|---|---|
| `frontend/nginx.default.conf` | 加 `proxy_read_timeout 3600s` 等 |
| `backend/server.js` | 读 `DEEPSEEK_EXTRA_BODY`；SSE 跳过 `reasoning_content`；对话分页 |
| `backend/db.js` | 新增 `getConversationsByUser` |
| `docker-compose.yml` | backend 容器 DNS 指定 `223.5.5.5` |
| `frontend/src/App.vue` | 健康检查 8s → 30s |
| `frontend/src/utils/session-watch.js` | session 检查 30s → 60s |

### 已提交 `d748ab6`

- `backend/.env.example`：补充 `DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":500}` 说明
- `CURRENT_STATE.md`：记录排查过程

未进 Git 的环境变量（不在仓库里）：

```env
DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":500}
```

## 7. 后续部署动作（等用户测试通过）

```bash
# 0 备份
cd /opt/AI-chat
cp -a backend/.env backend/.env.bak.$(date +%F)
cp -a docker-compose.yml docker-compose.yml.bak.$(date +%F)

# 1 同步代码到生产（只改文件，不动数据/.env；用本地到服务器 rsync 或直接替换指定文件）

# 2 修改 /opt/AI-chat/backend/.env
#   DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":500}

# 3 重建生产容器（注意是 force-recreate）
cd /opt/AI-chat
docker compose up -d --force-recreate --build

# 4 验证
docker exec ai-chat-frontend grep proxy_read_timeout /etc/nginx/conf.d/default.conf
curl -s http://127.0.0.1:8181/api/health

# 5 push GitHub
cd /mnt/vps/yun/AI-chat
git push origin master
```

## 8. 交接提示词（给主 Codex）

> 你正在接手 AI-chat 项目的收尾部署。当前主分支在本地 `/mnt/vps/yun/AI-chat`（master，最近 commit `d748ab6`），尚未 push；服务器生产 `/opt/AI-chat` 还是旧代码，生产容器端口 8181/3001 未动；测试环境 `http://121.41.192.80:8189` 已上线，后端已配置 `DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none","max_tokens":500}`。验证详情见 `/mnt/vps/yun/AI-chat/HANDOFF_MAIN_CODEX.md` 和 `CURRENT_STATE.md`。等待用户验收后执行生产部署：同步代码到 `/opt/AI-chat`、备份并更新 `.env`、`docker compose up -d --force-recreate --build`、验证、push GitHub。
