# Changelog

## 0.3.2

- Client 内部每 3 秒异步监控，接口超时 2 秒；正常通信不扫描版本，通信异常立即检查并恢复。
- 新增 Status / StatusChanged，提供版本、连接阶段、下载进度和错误信息。
- 以实际监听 PID 识别内核，兼容微信多进程和客户端重新接入。
- 微信版本优先读取运行时 Weixin.dll 的四段数字版本，区分完整安装与未完成更新。
- 版本不兼容时关闭所有目录中的微信主进程、扩展及更新进程，退出完成后用本地包恢复；包缺失或损坏才从 updates/latest 下载。
- 关闭微信失败时停止修复；DisposeAsync 停止监控，保留微信运行。

注意：自动版本修复会关闭其他目录中的微信会话。GUI 状态事件需切回 UI 线程处理。

## 0.3.1

- 默认组件根目录改为 `%LOCALAPPDATA%\CVX`，不再默认读取或写入 Sao 目录。
- 便携包直接解压到组件版本目录，严格使用 manifest 指定的启动器，保留官方版本子目录。
- 调整受管进程路径识别及便携包修复，保留版本目录内的注入组件与归档。

迁移提醒：默认目录已变化；请退出旧目录运行的微信，再将组件复制到 CVX 或重新下载。新版本不会自动接管 Sao 目录。

## 0.3.0

### Client

- 将组件下载、校验、启动注入和恢复合并进 `CVX.Client`，移除独立 Manager 项目及其 HTTP 协议。
- `ConnectAsync()` 检测并复用正常内核，可在同一 Runtime 上恢复崩溃的微信。
- `RestartAsync()` 手动重启受管微信并重新注入，保留消息订阅。
- 使用实际 HTTP 响应验证内核，区分未登录和内核不可用；停止前校验受管进程身份。
- 仅组件服务请求使用 `ComponentToken`，本机内核调用不发送组件凭据。
- 支持连接取消、失败重试、跨进程独占和已应用回调记录，退出 Client 不关闭微信。

### Breaking changes

- 移除 `RobotManagerClient` 及 Manager 协议类型。
- `WechatRobotOptions` 使用 `ComponentDirectory`、`GvxApiPort` 和组件下载配置；移除 Manager 地址和本体安装配置。
- `ConnectAsync()` 返回 `WechatRobotStatus`，提供 API 就绪、登录状态、组件版本及回调地址。
- 迁移前退出旧 Manager，禁用其外部自启动；已有组件目录继续复用，首次连接旧运行实例会重新注入。

### Build and validation

- Client 与 .NET SDK 统一使用发布标签版本；普通 CI 使用独立预发布版本。
- 发布 DLL / NuGet，不再构建独立 Manager EXE。
- 26 项本地检查覆盖恢复、重启、回调、取消和组件下载校验；真实 Windows 注入仍需在目标环境验证。
