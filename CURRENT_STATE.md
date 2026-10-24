# AI-Chat 当前状态总览

> 更新时间：2026-10-24（Android 1.1.0 发布；§0/§0.1/§0.2 为最新权威，§1 及以后为 2026-09-05 历史记录）

## 0. 服务器职责划分（2026-10-24 用户决定，最高优先级，实地核查确认）

Terraform/Tailscale 网络内两台阿里云服务器，**生产与测试物理分离，是两台不同机器**：

| 机器 | Tailscale IP | 公网 IP | 主机名 | 职责 |
|---|---|---|---|---|
| **yun** | 100.111.130.0 | 121.41.192.80 | iZbp15kdyy928z0o7i8p1nZ | **生产专用**：`8181/3001`（Go 版，`/opt/AI-chat`）。🚫 **禁止任何测试改动** |
| **yun1** | 100.96.106.50 | 121.41.206.32 | cl | **测试专用**：`8189/3002`（Go 测试服，`/opt/ai-chat-go-test`）。今后**所有测试一律用 yun1 8189** |

yun1 上的环境（2026-10-24 实测）：
- `ai-chat-go-*`（8189/3002）：现行唯一测试服，Up。
- `ai-chat-*`（8181/3001）：**Node 老版本，已决定退役**——与 yun 生产 8181 端口号相同但属不同机器，极易混淆。退役属停机操作，待用户单独确认后执行（`docker compose down`，保留数据，不加 `-v`）。

历史遗留说明：
- yun 上的老测试服（`/home/admin/ai-chat-test`、`/home/admin/ai-chat-go-test`，8189/3002）已于 2026-09 关闭，容器 Exited、数据保留不删。**yun 从此不再承载任何测试流量**。

## 0.1 Android 1.0.0 与生产后端补丁（2026-10-24）

### Android 发布

| 项 | 值 |
|---|---|
| Release | <https://github.com/unixcs/AI-chat/releases/tag/android-v1.0.0>（tag 指向 `ad841ac`） |
| APK | `aichat-1.0.0.apk`，8,975,733 字节 |
| SHA-256 | `a7554786891d5bacc24ce96925d898fd9bcc831bb6333dea98552cede220a61d`（与 GitHub asset digest 一致） |
| 包信息 | `com.unixcs.aichat`，versionName `1.0.0`，versionCode `7`；minSdk 26 / targetSdk 35 |
| 签名 | APK Signing Block v2 已验证（minSdk 26 无需 v1） |
| release 端点 | `http://121.41.192.80:8181`（生产）；APK 内未出现测试地址 |
| debug 端点 | `http://121.41.206.32:8189`（yun1，仅开发调试） |

CI：`Android_Debug`（run 40032342820）与 `Android_Release`（run 40032822448）均 success；release 由 tag 触发，keystore 从 GitHub Secrets 临时解码并在构建后删除。

功能范围：登录、SSE 流式对话、Markdown、Room 历史、分组/搜索/重命名/删除、三维度偏好、资料/会员状态、AES-256-GCM 加密备份导入导出、深色模式。构建与发布细节见 `android/README.md`。

**未完成项**：本环境无法连接真机，尚未做真机冒烟。建议至少验证：登录、流式对话、断网重连、历史重命名/删除、导出后清数据再导入、深色模式。

### 生产后端补丁

为支持 Android 会话重命名，生产后端升级为加法端点 `PUT /api/chat/conversations/{id}/title`（`7be22ed`，不改变既有字段/路径/错误格式）。

| 项 | 值 |
|---|---|
| 生产镜像 | `ai-chat-go-backend:prod-b243734`（二进制 sha256 `861784e2…87bd3`） |
| 切换时间 | 2026-10-23T22:13:27Z，`up -d --no-build`，后端停机约 4–5 秒，前端全程在线 |
| 数据备份 | `/home/admin/backup-pre-android-20261024-061216/data.sqlite`（sha256 `74a1e9d9…b194`） |
| compose 备份 | `/opt/AI-chat/docker-compose.override.yml.bak-pre-android-20261024-061216` |
| 回滚镜像 | `ai-chat-go-backend:prod-29bb956` |
| 验证 | 切换前后表行数一致；仅备份副本预演迁移；`tools/health-yun.sh all` ALL GREEN；rename 未授权返回 401 |

回滚方式：把 `docker-compose.override.yml` 中 backend 的 `image` 改回 `ai-chat-go-backend:prod-29bb956`，执行 `docker compose -p ai-chat up -d --no-build backend`。

## 0.2 Android 1.1.0（R15 迭代，2026-10-24）

### 发布

| 项 | 值 |
|---|---|
| Release | <https://github.com/unixcs/AI-chat/releases/tag/android-v1.1.0>（tag 指向 `0f457a5`） |
| APK | `aichat-1.1.0.apk`，8,994,657 字节 |
| SHA-256 | `9d51933f8f2be7729ef0b58a19e73c7d2fbeb676e563b41f6fd0dfe2e3bc1126`（与 GitHub asset digest 一致） |
| 包信息 | `com.unixcs.aichat`，versionName `1.1.0`，versionCode `8`；应用名「塔罗 Tarot」 |
| 签名 | APK Signing Block v2 已验证；keystore 复用 v1.0.0 同一密钥（GitHub Secrets） |
| release 端点 | `http://121.41.192.80:8181`（生产）；已确认 APK dex 内不含测试地址 8189 |
| CI | `Android_Debug` run 40049302809、`Android_Release` run 40049650672，均 success |
| 后端改动 | 无——兑换码/昵称直接复用既有 `POST /api/user/redeem`、`PUT /api/user/profile` |

### 功能范围（R15）

1. **抽屉空白假死修复**：侧边栏概率性进入空白且返回手势失效，根因是 Compose BOM 2024.10.01 所带 material3 1.3.0 的 ModalNavigationDrawer 缺陷；升级 BOM 至 2024.12.01（material3 1.3.1）解决，抽屉代码本身未动。
2. **兑换码充值 + 昵称修改**：设置页账户区新增两项；兑换走 `POST /api/user/redeem` 成功后用返回 profile 即时刷新会员信息；昵称走 `PUT /api/user/profile`，客户端回传现有 avatarUrl 以规避后端全量覆盖语义误清头像；401 统一回登录页。
3. **深浅色 + 主题色**：聊天页顶栏一键深浅切换（DataStore 持久化）；设置页可选跟随系统/浅色/深色与 6 套柔和色板（雾紫/星夜蓝/月白青/秘仪金/暮玫/松间绿）；状态栏前景色随深浅自动反转。
4. **消息快捷复制**：气泡下方按钮（气泡外，避免与长按选择冲突——沿用 R14 网页版教训），AI 消息左下、用户消息右下镜像布局，点击后打勾反馈 1.5 秒。
5. **品牌与精修**：应用名改「塔罗 Tarot」，星月启动图标；历史抽屉条目 18dp 圆角、搜索框+新建按钮重排；气泡 20dp 圆角。

### 未完成项

- 真机冒烟（本环境无真机）：抽屉反复开合复现原 bug 场景、兑换码（yun1 有测试卡密则先测测试服）、昵称修改、主题切换与深浅切换、复制按钮、v1.0.0 覆盖安装。
- Telegram 下载地址推送被阻塞：生产/yun1 容器与环境文件中不存在任何 TG bot token / chat id（符合「TG_BOT_TOKEN 仅环境变量，不进 Git」约定），需用户提供后才能推送。

## 1. 仓库和代码在哪儿

| 对象 | 位置 | 状态 |
|---|---|---|
| GitHub 远程仓库 | `git@github.com:unixcs/AI-chat.git` | 本地 `master` 已与 `origin/main` 同步（commit `4c667ec`） |
| 本地代码仓库（我们正在开发的地方） | `/mnt/vps/yun/AI-chat/` | `master` 分支，commit `4c667ec`（含 89b3a94 及后续 stream 修复、文档、工作流总结） |
| 服务器生产代码 | `/opt/AI-chat/` | 已更新为新代码（2026-09-05） |
| 服务器生产容器 | `ai-chat-frontend` / `ai-chat-backend` | 运行中，端口 `8181` / `3001` |
| 服务器测试容器 | `/home/admin/ai-chat-test`（`ai-chat-test-*`） | 已关闭，数据保留在 `data/`，端口 `8189` / `3002` 已释放 |

## 2. 当前最重要的一句话

**生产已上线新代码，测试服已关闭，GitHub 已同步。**

- 生产 `/opt/AI-chat` 已更新：nginx 3600s 超时、后端分页、`DEEPSEEK_EXTRA_BODY`、前端流式渲染优化均已生效。
- 测试服 `8189/3002` 已通过 `docker compose down` 关闭，测试数据保留未删除。
- 本地分支 `master` 已包含全部生产代码与文档，并已推送至 GitHub `main`（commit `4c667ec`）。

## 3. 已做工作

### 3.1 排查结论

- 卡顿根因：DeepSeek 上游首次响应超过 60s 时，nginx 默认 60s 超时会把 SSE 连接切断，前端表现为「没回复/刷新慢/联系管理员」。
- 费用上涨：DeepSeek 8/17 涨价 + 单次输出 token 变多（疑似出现无推理参数时还输出 `reasoning_content`）共同导致，排除了 key 泄露为主因的证据。

### 3.2 本地已完成的代码修改

commit `89b3a94`（本地）共 7 个文件：

| 文件 | 改动 | 解决的问题 |
|---|---|---|
| `frontend/nginx.default.conf` | 增加 `proxy_read_timeout 3600s; proxy_send_timeout 3600s; send_timeout 3600s;` | 解决上游 60s 超时导致的卡顿/无回复 |
| `backend/server.js` | 新增 `DEEPSEEK_EXTRA_BODY` 环境变量；SSE 跳过 `reasoning_content` | 控制是否禁用推理，降低输出 token |
| `backend/db.js` | 新增 `getConversationsByUser(userId, {limit, offset})` | 对话列表分页 |
| `backend/server.js` | `/api/chat/conversations` 改为分页返回 | 避免 2.4MB 全量列表拖慢前端 |
| `docker-compose.yml` | backend 增加 `dns: [223.5.5.5, 119.29.29.29]` | 防止 DNS 问题再次导致 DeepSeek 不可达 |
| `frontend/src/App.vue` | 健康检查 8s -> 30s | 降低无效轮询 |
| `frontend/src/utils/session-watch.js` | session 检查 30s -> 60s | 降低无效轮询 |

### 3.3 已完成验证

- `node -c backend/server.js` / `node -c backend/db.js`：通过。
- 后端单元测试 7/7 通过；1 个因本地 Node20 缺少 `node:sqlite` 跳过（服务器 Docker 内 Node22 可用）。
- 两轮对抗性 review：静态全通过，运行时边界 49 通过 + 1 个误报（admin 分页不受影响）。
- 本地没启动 Docker，因为当前 agent 机和 docker.sock 无权限。

## 4. 你可能的错误理解（澄清）

上一版我写的 `8189` 方案 **确实是在服务器上多起一套测试容器**，不是“本地测试”。它只是不动生产 8181。由于 `/opt` 没有写权限，实际创建在 `/home/admin/ai-chat-test`。

所以实际上有两条路线：

- **A. 服务器测试实例（推荐）**：已在 yun 服务器上创建 `/home/admin/ai-chat-test`，端口 8189/3002，不动生产。
- **B. 纯本地测试（不碰服务器）**：只在 `/mnt/vps/yun/AI-chat/` 本地起 Node22 + 前端，完全不动服务器；但需要本地 Node 22 和 DeepSeek key 等配置。

## 5. 后续行动顺序（等确认）

1. 你选择 A 或 B。
2. 根据选择：
   - A：已用 `/home/admin/ai-chat-test` 创建测试容器 `ai-chat-test-backend`/`ai-chat-test-frontend`，端口 8189/3002；生产 8181/3001 未动。
   - B：我在 `/mnt/vps/yun/AI-chat/` 本地搭 Node22/前端，你访问 `http://localhost:xxxx` 验收。
3. 你验收通过后，再执行：
   - 把本地 `89b3a94` 同步到服务器 `/opt/AI-chat` 并 `docker compose up -d --build`
   - push 到 GitHub
   - 健康检查 / 回滚预案

## 6. 验收清单

- 打开对应地址能登录
- 发消息能收到流式回复
- 连续发消息不再卡 60s 无回复
- `curl -s http://127.0.0.1:8189/api/health`（A 路线）
- `docker exec <frontend容器> grep proxy_read_timeout /etc/nginx/conf.d/default.conf` 能命中
- 后端日志不再出现 `upstream timed out`
- 对话多的账号打开会话列表明显更快

## 7. 关键文件

| 文件 | 含义 |
|---|---|
| `/mnt/vps/yun/AI-chat/` | 本地 git 工作区（修改在 89b3a94） |
| `/mnt/vps/yun/AI-chat-TEST-ACCEPTANCE.md` | 独立测试验收方案说明 |
| `/opt/AI-chat/`（服务器） | 生产代码，未改动 |
| `/home/admin/ai-chat-test/`（服务器） | 测试实例，已运行 |

## 8. 卡顿根因与已完成的验证（2026-09-05 新增）

### 收到问题的第一性分析

用户的“发送后一直转圈”根因不是服务器 nginx 超时这一层，而是：

```
DeepSeek 当前模型会先输出 reasoning_content（思维链/推理内容）
  -> 这些 token 会被计费
  -> 但前端只展示最终 content
  -> 用户在最终 content 出现前一直看到转圈
  -> 若推理很长，还可能触发 DEEPSEEK_TIMEOUT_MS=90000 超时
```

### 实测数据（测试后端，实际请求 DeepSeek）

| 场景 | 首个 content 时间 | 是否出现 reasoning_content | 备注 |
|---|---|---|---|
| 简单“hi” | 约 0.5s | 无 | 上游本身不慢 |
| 实际“继续”上下文 | 约 8.4s | 有，且首 token 前已有 248KB chunk | 明显被“思维过程”拖住 |
| 同样请求加 `reasoning_effort=none` | 约 0.35s | 无 | 可大幅降低首字延迟和费用 |

### 已在测试环境应用

已在 `/home/admin/ai-chat-test/backend/.env` 增加：

```env
DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none"}
```

已重启 `ai-chat-test-backend`。当前测试环境已使用该配置，你可以再次发送消息感受差异。

### 建议的生产部署动作（尚未执行）

1. 生产 `/opt/AI-chat/backend/.env` 增加相同的：
   ```env
   DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none"}
   ```
2. 同步本地 `89b3a94` 代码（已包含 `DEEPSEEK_EXTRA_BODY` 支持与 nginx 3600s 超时）。
3. 重新构建生产 `docker compose up -d --build`。
4. 手动测试通过后 push 到 GitHub。

### 关于两个容器共用数据库的确认

- 数据库是 **SQLite**（`data.sqlite`），不是 MySQL。
- 生产容器使用 `/srv/ai-chat/data`，本次全新 8189 测试容器使用 `/home/admin/ai-chat-test/data`。
- 两个容器 **不共用一份 DB**，互相没有写入冲突。
- 你登录后仍有旧登录态，是因为浏览器 `localStorage` 里的旧 token 没清，且测试库是从生产复制来的，已有该用户的账号记录；不是“MySQL 共用”。

## 8.1 最新测试（启用 reasoning_effort=none 后）

测试后端实际请求数据：

| 指标 | 之前（有推理） | 现在（关闭推理） |
|---|---|---|
| 首个 content 出现时间 | 约 8.4s | 约 0.6s |
| 完整长回答耗时 | 可能超过 15s | 约 6.8s |
| reasoning_content | 有（会被计费） | 无 |

结论：目前“卡”的主要剩余原因是**最终回复本身较长，完整生成需要几秒到十几秒**，而不是一上来就卡死。
前端 SSE 会逐步推文字；如果你在页面上一直只看到转圈、没有文字滚动，需要再看浏览器 Network 是否 EventSource 一直在推进，或 markdown 渲染长文本是否有卡顿。

## 8.2 13:44 卡 40 秒的最终原因（已修正）

13:44 那条消息确实在数据库里出现了 42s 延迟，原因是：

- 我此前只修改了 `/home/admin/ai-chat-test/backend/.env` 并 `docker restart`；
- 但 Docker Compose 的 `env_file` 只在容器**创建**时生效，`docker restart` 不会重新加载 `.env`；
- 所以测试后端实际运行时仍然没有 `reasoning_effort:none`，DeepSeek 继续输出推理，首字要 40 秒左右。

已修正：

```bash
cd /home/admin/ai-chat-test
docker compose up -d --force-recreate backend
```

现在容器内已确认：

```text
DEEPSEEK_EXTRA_BODY={"reasoning_effort":"none"}
```

同一段真实请求返测：

- 首个 content 时间：约 0.44s
- 完整耗时：约 6.8s
- `reasoning_content`：无

生产部署时也要注意：**只用 `docker restart` 不足以加载环境变量变更，必须 `docker compose up -d --force-recreate` / `--build`。**
