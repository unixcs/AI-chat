# AI Chat Android 客户端

原生 Android 客户端，与网页版共用同一套 Go 后端（`/api` 契约完全一致）。当前版本 **1.4.0（versionCode 11）**，包名 `com.unixcs.aichat`，应用名「塔罗 Tarot」。

- 下载：<https://github.com/unixcs/AI-chat/releases/tag/android-v1.4.0>
- APK：`aichat-1.4.0.apk`，9,010,885 字节
- SHA-256：`397e3165d989344c0fbe35665c85cf5356db3ca2a2a331b56c82c4174f5db779`（校验链：下载文件重算 SHA-256 + apksigner v2 验签通过 + 证书 SHA-256 与 v1.0.0–v1.3.0 逐版一致）

## 功能

- 手机号 + 密码登录；JWT 保存在 DataStore，启动时校验会话，失效自动回登录页
- SSE 流式对话，逐字渲染；会话与消息写入 Room，弱网/重启后可继续查看
- AI 回复 Markdown 渲染：标题、列表、引用、代码块、行内样式、链接（流式过程中先按纯文本显示，结束后再渲染）
- 抽屉式历史（DeepSeek 风格）：无边框圆角行、无时间戳、按时间分组、搜索、长按或 ⋯ 呼出重命名/删除、多选批量删除（进度反馈）
- 回答偏好：长度 / 风格 / 输出格式三维度组合，发送时即时生效；可隐藏聊天页入口
- 账户中心：昵称修改（APP 内直改）、兑换码充值（`POST /api/user/redeem`，成功后即时刷新会员信息）、会员到期展示
- 消息快捷复制：气泡下方按钮，AI 消息在左下、用户消息在右下（DeepSeek 式镜像布局），点击后打勾反馈 1.5 秒
- 深浅色跟随系统（Material 3 动态色）；主题为 v1.0.0 静态基线，无应用内切换（主题切换是 v1.1–v1.3 反复假死的触发器，已整体移除，见版本记录 1.4.0）
- 品牌：应用名「塔罗 Tarot」，星月启动图标
- 本地加密备份导出/导入（SAF 选择文件位置，AES-256-GCM）

## 技术栈

| 层 | 选型 |
|---|---|
| 语言 / UI | Kotlin 2.0.21、Jetpack Compose（BOM 2024.10.01）、Material 3 官方组件（ModalNavigationDrawer / ModalBottomSheet / AlertDialog，零自研手势代码） |
| 架构 | MVVM + StateFlow + Navigation Compose 2.8.5 |
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

发布地址交付约定（2026-10-24 用户决定）：每次编译发布完成后，直接在会话中把 GitHub Release 下载地址推送给用户（Release 页 + APK 直链 + SHA-256）；Telegram 推送方案搁置。

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
├─ data/settings/   # 外观设置（深浅模式、主题色，DataStore）
├─ di/              # Koin 模块
└─ ui/              # 登录、聊天、历史抽屉、设置、备份、主题、Markdown 渲染
```

## 版本记录

- **1.4.0（android-v1.4.0，2026-10-25）**：R18 回滚根除。四版反复假死的第一性结论：假死自 1.1.0 引入主题切换起出现，跨越官方抽屉（1.1）与自研壳（1.2/1.3）两种实现始终复现——唯一共同变量是主题切换链（DataStore 写入 → MaterialTheme 全树重组 + `enableEdgeToEdge` 重设窗口标志，OEM 掉帧风暴下与手势/动画竞态）。1.0.0 零 bug 的真正原因：无主题切换 + 全部使用 M3 官方组件状态机。本次：①主题切换机整体移除（ThemeMode/AccentPalettes/外观点/顶栏深浅色按钮），Theme/Color/MainActivity 恢复 1.0.0 原文，深浅色跟随系统；②自研 DrawerShell/BottomSheetShell 整体移除，抽屉回到 ModalNavigationDrawer、偏好回到 ModalBottomSheet（BOM 回 2024.10.01）；③公告弹窗移除（可有可无功能，后端端点保留不受影响）；④顺手修偏好入口按钮首帧闪现（stateIn 初始 true→false）。保留：兑换码/昵称、批量删除、塔罗品牌、加密备份。57 项单元测试全绿（10 套件）；子代理对抗审查无 P0/P1。
- **1.3.0（android-v1.3.0，2026-10-24）**：R17 真机四项反馈连根修复——①设置切主题色后开抽屉仍假死：根因是帧时钟停摆时 `animateTo` 永不完成（OEM 掉帧风暴 × targetSdk 35 edge-to-edge × MaterialTheme 全树重组），`settleTo` 看门狗 1.2 秒超时后 `snapTo`（纯状态写入、不依赖帧时钟）兜底，数学上不存在可卡死的中间态；②侧边栏左滑回弹：settle 改速度优先（≥150dp/s 的轻扫直接顺扫向落位，不再只看位置阈值）；③偏好按钮失灵：`BottomSheetShell` 手势关闭落底后回调 `onClosed`（一次性、重开时复位），ChatScreen 据此复位 `showPreferences`，此前布尔永远为 true 导致按钮死亡；④弹层掉帧：每帧值只在 `graphicsLayer{}`（绘制期）读取，组合期经 `derivedStateOf` 归并为 3 个布尔，抽屉壳不再逐帧重组整棵聊天树。子代理对抗审查 3 项 P1 加固（看门狗不覆盖新动画、onClosed 防重、边缘热区锁存复位）。67 项单元测试全绿（新增 settleTarget 边界 5 测）。**（该版修复无效，1.4.0 已回滚替代。）**
- **1.2.0（android-v1.2.0，2026-10-24）**：根除抽屉/弹层假死——弃用 material3 ModalNavigationDrawer/ModalBottomSheet（AnchoredDraggable NaN 锚点 + 预测式返回手势与主题切换 enableEdgeToEdge 竞态所致），改为自研 Animatable 进度驱动的 DrawerShell/BottomSheetShell（单边缘热区拖拽、恒定 settle、无状态机可卡死）；历史抽屉对标 DeepSeek 全面重设计（无边框药丸行、无时间戳、填充式圆角搜索框、行内锚定菜单、长按=⋯菜单、首组标题批量管理入口、多选批量删除含进度条与 404 幂等）；新增商家公告弹窗（后台更新公告 → APP 自动弹出卡片 →「我知道了」确认，data:null 信封与 asOf 逐字回传均有单测锁定）；流式回复中删除当前会话先取消 SSE 再删服务端，杜绝缓存孤儿行。62 项单元测试全绿。
- **1.1.0（android-v1.1.0，2026-10-24）**：修复侧边栏抽屉概率性空白假死（升级 Compose BOM 至 2024.12.01 / material3 1.3.1）；新增兑换码充值与昵称修改（APP 内直连 `/api/user/redeem`、`PUT /api/user/profile`）；聊天页顶部深浅色快捷切换 + 设置内 6 套主题色；消息气泡下方快捷复制按钮（AI 左下/用户右下）；改名「塔罗 Tarot」+ 星月图标；历史抽屉与气泡圆角精修。
- **1.0.0（android-v1.0.0，2026-10-24）**：首个正式版——登录、SSE 流式对话、历史抽屉、回答偏好、加密备份、会员展示。

## 已知边界

- 生产后端目前是明文 HTTP（受网络白名单限制），后续上 HTTPS 后应同步收紧 `network_security_config.xml`
- 备份只覆盖本机会话与消息，不含图片/附件（当前后端也没有附件能力）
- 未接入推送；会员到期、公告更新需要打开 App 后刷新（公告弹窗在打开聊天页时拉取）
- 真机冒烟需人工完成：登录、流式对话、断网重连、历史重命名/删除、导出后清数据再导入、深色模式
- R16 新功能（自研抽屉手势、批量删除、公告弹窗）尚待真机回归
