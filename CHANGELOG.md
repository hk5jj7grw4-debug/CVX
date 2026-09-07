# Changelog

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
