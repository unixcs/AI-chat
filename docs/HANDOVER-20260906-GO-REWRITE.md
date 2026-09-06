# AI-chat 变动接收文档（2026-09-06 · Go 重构版 → 测试服部署完成）

> 本文档面向 **Tencent 主控服务器** 的后续操作者：本文记录 2026-09-06 下午至晚间的全部代码变动、
> 三套环境的当前状态、以及后续把新版升级到 yun 生产 8181 的操作要点。
> 代码事实以 git 为准：`go-rewrite` 分支 HEAD = `7be284d`（GitHub unixcs/AI-chat 已同步）。

## 1. 本次变动范围（git 区间 pre-ui-redesign..HEAD）

| 提交 | 内容 |
|---|---|
| `pre-ui-redesign`（tag） | UI 重建前的完整可用版本（复制修复 + 三维度偏好已完成）——**回滚锚点** |
| `73fe394` | Part D 基建：Tailwind v4 + 22 组 shadcn-vue 组件 + 品牌 token + 暗色双机制；公开页/认证页/用户壳/ChatView 重写 |
| `cc4da59` | 管理端：AdminLayout（分组导航+登出确认）+ AdminPromptView（偏好卡片）+ AdminUserView（双端表格） |
| `e31b8d0` | Part D 完成：20 视图/布局全部换装；完整性/体积门禁入测试 |
| `0620714` | R12b 审查记录（SOLID） |
| `831c1dc` | GO-REWRITE-REPORT §12 验收章 |
| `7be284d` | **R13 复制二轮修复**：iOS/iPadOS 走 contentEditable span+Range 路径 |

三块功能：①消息复制修复（桌面+移动，iOS 专项）；②三维度回答偏好（长度/风格/输出格式，后台 10 张 Prompt 卡片热改）；③全站 UI 重建（shadcn-vue + brand-guidelines，前台后台统一视觉）。

## 2. 三套环境现状（2026-09-06 晚）

| 环境 | 位置 | 版本 | 状态 |
|---|---|---|---|
| **yun 生产** | yun 服务器 `8181`(前端)/`3001`(后端)，`/opt/AI-chat`，数据 `/srv/ai-chat/data` | **Node.js 老版（未动！）** | Up 25h+，红线保护 |
| **yun 测试服**（新） | yun 服务器 `8189`(前端)/`3002`(后端)，`/home/admin/ai-chat-go-test`，数据全新 | **Go 新版 `7be284d`** | 已验收，对外可访问 http://121.41.192.80:8189 |
| **yun1 测试位** | yun1 `8189`/`3002`，`/opt/ai-chat-go-test` | **Go 新版 `7be284d`** | 已验收 |

说明：yun 生产从未被本次任何操作触碰（容器未重建、数据/env 未改）；yun 测试服的 `DEEPSEEK_API_KEY` 等从生产 `.env` **只读提取**写入新目录（生产文件本身零改动，chmod 600）。

## 3. 后续把新版升级到 yun 生产 8181 的操作要点

前置阅读：`DEPLOYMENT_WORKFLOW.md`（本仓库）+ Tencent 机 `/mnt/vps/yun/.claude/skills/ai-chat-safe-deploy/`（生产安全部署 skill，含 01-05 脚本）。

1. **测试服观察期**：yun 8189 / yun1 8189 稳定运行数天（登录/聊天/复制/偏好/后台管理）再升级。
2. **数据决策**（关键！）：Go 版 schema 与 Node 版同表名兼容，但字段为**增量**（answerFormat 列、settings/promptRevisions 等新表，`CREATE TABLE IF NOT EXISTS` 幂等）——生产 SQLite 挂载进 Go 后端可直接接管（yun1 已演练老库迁移 E2E）。升级前必须跑 skill 的 `02-backup-sqlite.sh`。
3. **切换形态**：生产前端容器 8181:80 不变，把 `ai-chat-backend` 容器镜像换为 `ai-chat-go-backend:latest`、`ai-chat-frontend` 换 `ai-chat-go-frontend:latest`（镜像由 WSL 构建 save/load，**禁止在 yun 上构建**——1.6GB 内存红线）；`.env` 沿用生产现值（不覆盖）。
4. **升级窗口**：低峰执行；后端停机时间 ≈ 容器重建 10 秒（SSE 连接会断，用户刷新即恢复）。
5. **验收**：升级后必跑 `ops/health-yun.sh yun` 全绿 + 线上冒烟（登录/流式聊天/复制/公告/后台登录）。
6. **回滚**：原 Node 镜像 tag 保留（升级前 `docker tag` 记录当前镜像 ID），出问题一键换回；master 分支 = Node 版完整历史。

## 4. 已知限制 / 注意

- 复制按钮：insecure context（http://IP）无 Clipboard API，走 execCommand 兜底——iOS/iPadOS（含微信 WKWebView）已专项适配（span+Range 路径）；若真机仍有失败个案，收集「设备+浏览器+是否出现'复制失败'提示」反馈。
- 「纯文字」输出格式是指令层约束（拼入 system），不保证模型 100% 遵从（记偏差，见 GO-REWRITE-REPORT §10）。
- yun 测试服数据是全新空库，与生产数据无关；不要把生产数据目录挂给测试容器。

## 5. 联络锚点

- 完整验收/审查记录：本仓库 `docs/GO-REWRITE-REPORT.md`（§12）、`docs/PREF-UI-REDESIGN-PLAN.md`（§6/§7/R13）、`docs/BRAND-GUIDELINES.md`
- 部署脚本：`deploy/deploy-yun1-test.sh`（`SSH_TARGET=yun REMOTE_DIR=/home/admin/ai-chat-go-test bash deploy/deploy-yun1-test.sh` 可重复执行更新测试服）
