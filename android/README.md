# AI Chat Android 客户端

原生 Android 客户端，与网页版共用同一套 Go 后端（`/api` 契约完全一致）。当前版本 **1.0.0（versionCode 7）**，包名 `com.unixcs.aichat`。

- 下载：<https://github.com/unixcs/AI-chat/releases/tag/android-v1.0.0>
- APK：`aichat-1.0.0.apk`，8,975,733 字节
- SHA-256：`a7554786891d5bacc24ce96925d898fd9bcc831bb6333dea98552cede220a61d`

## 功能

- 手机号 + 密码登录；JWT 保存在 DataStore，启动时校验会话，失效自动回登录页
- SSE 流式对话，逐字渲染；会话与消息写入 Room，弱网/重启后可继续查看
- AI 回复 Markdown 渲染：标题、列表、引用、代码块、行内样式、链接（流式过程中先按纯文本显示，结束后再渲染）
- 抽屉式历史：按时间分组、搜索、重命名、删除
- 回答偏好：长度 / 风格 / 输出格式三维度组合，发送时即时生效；可隐藏聊天页入口
- 个人资料、会员到期、卡密充值入口、修改密码（跳转网页版对应页面）
- 本地加密备份导出/导入（SAF 选择文件位置，AES-256-GCM）
- 浅色 / 深色主题跟随系统

## 技术栈

| 层 | 选型 |
|---|---|
| 语言 / UI | Kotlin 2.0.21、Jetpack Compose（BOM 2024.10.01）、Material 3 |
| 架构 | MVVM + StateFlow + Navigation Compose 2.8.4 |
| 依赖注入 | Koin 4.0.0 |
| 本地存储 | Room 2.6.1（KSP）+ DataStore Preferences |
| 网络 | OkHttp 4.12、kotlinx-serialization 1.7.3 |
| 构建 | AGP 8.7.3、Gradle 8.9、JDK 17 |
| SDK | minSdk 26、compileSdk/targetSdk 35 |

## 构建

前置：JDK 17、Android SDK Platform 35。

```bash
cd android
./gradlew assembleDebug          # 产物 app/build/outputs/apk/debug/app-debug.apk
./gradlew testDebugUnitTest      # 单元测试
```

Debug 构建固定指向 **yun1 测试服** `http://121.41.206.32:8189`，不会访问生产环境。

### Release 签名构建

签名材料只从环境变量读取，不落仓库：

```bash
export ANDROID_KEYSTORE_PATH=/path/to/aichat-release.jks
export ANDROID_KEYSTORE_PASSWORD=...
export ANDROID_KEY_ALIAS=...
export ANDROID_KEY_PASSWORD=...
./gradlew assembleRelease
```

未提供 `ANDROID_KEYSTORE_PATH` 时 release 变体也可以编译，但产物未签名，不能直接安装分发。CI 的 `Android Release` 工作流会从 GitHub Secrets 解码临时 keystore，构建后立即删除。

## 服务器地址

| 构建类型 | 地址 | 用途 |
|---|---|---|
| release | `http://121.41.192.80:8181` | yun 生产（前端 Nginx 反代 `/api`） |
| debug | `http://121.41.206.32:8189` | yun1 测试服 |

`res/xml/network_security_config.xml` 将明文 HTTP 限定为上述两个主机，其他域名默认禁止明文流量。

## 发布流程

```bash
# 1. 更新 android/app/build.gradle.kts 的 versionCode / versionName
# 2. 本地跑单测与构建（可选）
cd android && ./gradlew testDebugUnitTest assembleDebug
# 3. 提交并打 tag（tag 必须带 android-v 前缀）
git tag android-v1.0.0
git push origin main
git push origin android-v1.0.0
```

tag 推送触发 `.github/workflows/android-release.yml`：解码 keystore → 单测 → `assembleRelease` → 把 APK 传到对应 tag 的 GitHub Release。`Android_Debug` 工作流在每次 `android/**` 变更推送到 `main` 时运行单测并上传 debug APK artifact。

## 加密备份格式

```
"AICHAT1" (7B) | version (1B) | salt (16B) | iv (12B) | AES-256-GCM ciphertext
```

- 密钥：PBKDF2-HMAC-SHA256，120,000 次迭代，256 位；盐与 IV 每次导出都重新随机生成
- 明文：ZIP（`manifest.json` + `data.json`），导入时拒绝未知条目、限制条目/总量/会话/消息数量
- 密码错误与文件损坏统一提示「密码错误或文件已损坏」，不区分两者
- 备份密码与账号密码无关，只用于本地加解密，不上传服务器
- 导入采用合并策略（不删除本机较新的会话）；服务器仍是权威数据源，刷新后会清理服务端已不存在的会话

## 代码结构

```text
android/app/src/main/java/com/unixcs/aichat/
├─ data/api/        # OkHttp + 统一 envelope 解析
├─ data/chat/       # 会话仓库、SSE 客户端与解析器
├─ data/local/      # Room 实体/DAO/数据库
├─ data/backup/     # 备份格式、加解密、SAF 导入导出
├─ data/session/    # JWT 会话存储
├─ data/prefs/      # 偏好设置
├─ di/              # Koin 模块
└─ ui/              # 登录、聊天、历史抽屉、设置、备份、Markdown 渲染
```

## 已知边界

- 生产后端目前是明文 HTTP（受网络白名单限制），后续上 HTTPS 后应同步收紧 `network_security_config.xml`
- 备份只覆盖本机会话与消息，不含图片/附件（当前后端也没有附件能力）
- 未接入推送；会员到期、公告更新需要打开 App 后刷新
- 真机冒烟需人工完成：登录、流式对话、断网重连、历史重命名/删除、导出后清数据再导入、深色模式
