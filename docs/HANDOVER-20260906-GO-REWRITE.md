# AI-chat 变动接收文档（2026-09-06 · Go 重构版 → 测试服部署完成）

> 本文档面向 **Tencent 主控服务器** 的后续操作者：本文记录 2026-09-06 下午至晚间的全部代码变动、
> 三套环境的当前状态、以及后续把新版升级到 yun 生产 8181 的操作要点。
> 代码事实以 git 为准：`go-rewrite` 分支 HEAD = `8b419cf`（GitHub unixcs/AI-chat 已同步）。

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
| `8b419cf` | **R14 三项用户反馈**：移除消息复制按钮（真机多轮修复仍不可用，clipboard 工具一并下线）；气泡选中对比度（补 `::selection` 规则 + `.dark` 聊天变量覆盖，修 AI 气泡深色过亮）；个人资料/卡密充值提示拆独立状态杜绝串位。**纯前端改动，后端零变更** |

三块功能：①消息复制修复（桌面+移动，iOS 专项）——**R14 起复制按钮按用户决定整体移除**；②三维度回答偏好（长度/风格/输出格式，后台 10 张 Prompt 卡片热改）；③全站 UI 重建（shadcn-vue + brand-guidelines，前台后台统一视觉）。

## 2. 三套环境现状（2026-09-06 晚）

| 环境 | 位置 | 版本 | 状态 |
|---|---|---|---|
| **yun 生产** | yun 服务器 `8181`(前端)/`3001`(后端)，`/opt/AI-chat`，数据 `/srv/ai-chat/data` | **Node.js 老版（未动！）** | Up 27h+，红线保护 |
| **yun 测试服** | yun 服务器 `8189`(前端)/`3002`(后端)，`/home/admin/ai-chat-go-test`，数据=**老 8189 测试库已迁入**（见 §3 演练记录） | **Go 新版，前端 `8b419cf`** | 已验收，对外可访问 http://121.41.192.80:8189 |
| **yun1 测试位** | yun1 `8189`/`3002`，`/opt/ai-chat-go-test` | **Go 新版，前端 `8b419cf`** | 已验收 |

说明：yun 生产从未被本次任何操作触碰（容器未重建、数据/env 未改）；yun 测试服的 `DEEPSEEK_API_KEY` 等从生产 `.env` **只读提取**写入新目录（生产文件本身零改动，chmod 600）。

## 3. 无感升级 SOP（2026-09-06 已在 yun 8189 实盘演练验证）

**原则（用户要求，硬性）**：生产升级的前提是保证全部业务数据，做到用户无感——数据零丢失、行数对账一致、用户登录态不掉（JWT_SECRET 相同则浏览器 token 直接有效，连重新登录都不需要）。

**迁移机制（已验证）**：Go 版 schema 与 Node 版同表名兼容，迁移为纯增量（`CREATE TABLE IF NOT EXISTS` 新表 + `ADD COLUMN` 新列，幂等）；旧库 7 张表原样保留，不删列不改行。

**2026-09-06 实盘演练记录（yun 8189：老 Node 测试库 → Go 后端）**：

| 表 | 迁移前 | 迁移后 | 结果 |
|---|---|---|---|
| users | 32 | 32 | 一致 |
| conversations | 24308 | 24308 | 一致 |
| messages | 152216 | 152216 | 一致 |
| redeemCodes / redeemRecords | 62 / 40 | 62 / 40 | 一致 |
| roles / menus | 2 / 8 | 2 / 8 | 一致 |
| settings / announcements / promptRevisions / announcementReads | 无表 | 10 / 0 / 0 / 0 | 新增（预置卡片等） |
| users.answerFormat 列 | 无 | 已加，旧用户 NULL | 前端 standard 兜底 |

演练同时确认：老用户账号密码可直接登录（bcrypt 双向兼容），历史会话/消息在新 UI 正常可见。

**生产 8181 升级标准流程（照此执行，不得跳步）**：
1. **备份**：`sqlite3 /srv/ai-chat/data/data.sqlite ".backup '/home/admin/backup-pre-upgrade.sqlite'"`（在线一致备份），并 `docker tag` 记录当前 Node 镜像 ID；
2. **副本预演**：把备份副本挂到测试容器（或 yun 8189 同款流程）跑一次迁移 + 行数对账 + 抽样登录，**对账不过不碰生产**；
3. **切换**（低峰）：停 `ai-chat-backend` → 换 `ai-chat-go-backend:latest` 镜像 → 起容器（Go 自动增量迁移，秒级）→ 前端镜像同理（前端无状态可随时换）；
4. **升级后对账**：§3 同款行数比对必须与切换前一致；`ops/health-yun.sh yun` 全绿；
5. **回滚**：换回记录的 Node 镜像 tag + 恢复备份库（升级产生的增量列/表对 Node 版无害，Node 忽略未知表）。

**已规避的教训**：2026-09-06 部署 yun 8189 时误按"废弃环境重建"新建空库，未提前确认老环境有数据（数据未丢，原目录 `/home/admin/ai-chat-test/` 完好保留含每日备份）。任何"新建/重建"动作前必须先查证旧数据存在性并向用户确认处置方式。

## 4. 后续把新版升级到 yun 生产 8181 的操作要点（修订版）

前置阅读：`DEPLOYMENT_WORKFLOW.md`（本仓库）+ Tencent 机 `/mnt/vps/yun/.claude/skills/ai-chat-safe-deploy/`（生产安全部署 skill，含 01-05 脚本）。

1. **测试服观察期**：yun 8189 / yun1 8189 稳定运行数天（登录/聊天/偏好/后台管理）再升级。
2. **数据决策**（关键！）：Go 版 schema 与 Node 版同表名兼容，但字段为**增量**（answerFormat 列、settings/promptRevisions 等新表，`CREATE TABLE IF NOT EXISTS` 幂等）——生产 SQLite 挂载进 Go 后端可直接接管（yun1 已演练老库迁移 E2E）。升级前必须跑 skill 的 `02-backup-sqlite.sh`。
3. **切换形态**：生产前端容器 8181:80 不变，把 `ai-chat-backend` 容器镜像换为 `ai-chat-go-backend:latest`、`ai-chat-frontend` 换 `ai-chat-go-frontend:latest`（镜像由 WSL 构建 save/load，**禁止在 yun 上构建**——1.6GB 内存红线）；`.env` 沿用生产现值（不覆盖）。R14 仅为前端镜像更新，生产切换时用最新 `ai-chat-go-frontend:latest` 即可。
4. **升级窗口**：低峰执行；后端停机时间 ≈ 容器重建 10 秒（SSE 连接会断，用户刷新即恢复）。
5. **验收**：升级后必跑 `ops/health-yun.sh yun` 全绿 + 线上冒烟（登录/流式聊天/公告/后台登录）。
6. **回滚**：原 Node 镜像 tag 保留（升级前 `docker tag` 记录当前镜像 ID），出问题一键换回；master 分支 = Node 版完整历史。

## 5. 已知限制 / 注意

- 复制按钮已按用户决定整体移除（R12a/R13 两轮修复后真机仍不可用）；用户复制文本改由长按选择，选中对比度已在 R14 强化（气泡内高亮可见）。
- 「纯文字」输出格式是指令层约束（拼入 system），不保证模型 100% 遵从（记偏差，见 GO-REWRITE-REPORT §10）。
- yun 测试服数据 = 老 8189 测试库原样迁入（32 用户/15.2 万消息），与生产数据无关；不要把生产数据目录挂给测试容器。

## 6. 联络锚点

- 完整验收/审查记录：本仓库 `docs/GO-REWRITE-REPORT.md`（§12/§13）、`docs/PREF-UI-REDESIGN-PLAN.md`（§6/§7/R13/§8-R14）、`docs/BRAND-GUIDELINES.md`
- 部署脚本：`deploy/deploy-yun1-test.sh`（`SSH_TARGET=yun REMOTE_DIR=/home/admin/ai-chat-go-test bash deploy/deploy-yun1-test.sh` 可重复执行更新测试服）
