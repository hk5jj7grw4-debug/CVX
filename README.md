# CVX

> 此为私人项目依赖包，无实际内核程序，请勿下载使用。

## 子项目

| 项目 | 职责 |
|---|---|
| `CVX.Manager` | 常驻管理器，负责下载、启动、注入与运行状态管理 |
| `CVX.Client` | 提供给 UI 客户端的 Manager 连接、启动和恢复能力 |
| `CVX.Sdk.DotNet` | .NET 机器人业务 SDK |
| `CVX.Sdk.Go` | Go 机器人业务 SDK |

| SDK 方法 | Robot Manager 接口 |
|---|---|
| `GetHealthAsync` | `GET /v1/health` |
| `GetStatusAsync` | `GET /v1/status` |
| `GetOperationAsync` | `GET /v1/operation` |
| `RegisterClientAsync` | `PUT /v1/client` |
| `UnregisterClientAsync` | `DELETE /v1/client` |
| `ConfigureAsync` | `PUT /v1/config` |
| `ReconcileAsync` | `POST /v1/reconcile` |
| `StartAsync` | `POST /v1/start` |
| `StopAsync` | `POST /v1/stop` |
| `RestartAsync` | `POST /v1/restart` |
| `RepairAsync` | `POST /v1/repair` |
| `ShutdownHostAsync` | `POST /v1/host/shutdown` |

## 标准接入流程

业务客户端通过一个 Runtime 连接并接收消息：

```csharp
using CVX.Client;

await using var runtime = new WechatRobotRuntime(new WechatRobotOptions
{
    // Manager 已安装并配置好时，可直接使用默认选项。
    // 首次安装时提供以下参数：
    UpdateServerUrl = "https://your-update-server.example",
});
runtime.MessageReceived += (_, message) => Console.WriteLine(message.Text);
runtime.ReceiveError += (_, error) => Console.Error.WriteLine(error.Message);
await runtime.ConnectAsync();
```

`ConnectAsync` 负责启动回调监听、确保 Manager 可连接、校验服务身份与协议、注册当前实例、应用回调地址并等待机器人就绪。已有兼容 Manager 会直接复用，不检查或执行升级。重复连接会查询状态，就绪时直接返回；不会启动后台自动重连，Manager 重启后可再次调用 `ConnectAsync` 注册。

`WechatRobotOptions` 是普通接入的唯一配置入口：`ManagerAddress` 用于连接，`CallbackUrl` 用于监听，`ClientId` 标识应用，`InstallDirectory` 指定 Manager 安装目录。`ConnectTimeout` 限制整个连接过程，默认 12 分钟（含首次下载）。未提供的更新地址保留 Manager 原配置；连接不会修改组件 Token、业务端口或自动启动策略。

使用 `await using` 异步释放。释放会取消正在进行的连接，限时注销本次实例并关闭监听，不会停止 Manager 或微信。消息事件在后台线程执行，UI 更新需要自行切回 UI 线程；HTTP 确认仅代表接收，不代表业务处理完成。当前消息分发不保证顺序，已经派发的处理可能在释放后结束。

管理界面按需单独使用 `RobotManagerClient` 调用启动、停止、重启、修复等接口；Runtime 不重复暴露这些方法。

### Token 配置

Client 连接本机 Manager、查询 Manager 版本及下载 Manager 安装包均不需要 Token。Manager 的版本查询和安装包地址必须允许匿名访问。

仅 Manager 检查、下载机器人组件时使用 `ComponentToken`，在 Manager 的 `appsettings.json` 的 `RobotManager` 节中配置，或通过 `PUT /v1/config` 设置 `componentToken`。普通 Runtime 不接收也不覆盖它。未配置时不能从组件服务检查或下载组件，但已安装可用组件仍可启动。

Manager 仅允许监听本机回环地址；旧配置中的 `apiToken` 不再生效。此调整针对 Client / Manager，不修改独立业务 SDK 的内核协议。

### 接口迁移

这是一次不兼容的公开接口收紧：

- `RobotManagerBootstrapperOptions` 改为 `WechatRobotOptions`：`ManagerApiBaseAddress` → `ManagerAddress`，`RootDirectory` → `InstallDirectory`。
- `ConnectAsync(config, callbackOptions)` 改为 `ConnectAsync(cancellationToken)`；更新地址和回调地址移到构造选项。
- `runtime.Messages.MessageReceived` / `ReceiveError` 改为 `runtime.MessageReceived` / `ReceiveError`。
- 删除 Runtime 的 `EnsureRunningAsync`、`Manager`、同步 `Dispose`；使用 `ConnectAsync`、独立 `RobotManagerClient` 和 `DisposeAsync`。
- Bootstrapper、回调监听器及解析器转为内部实现。消息发送仍使用 `CVX.Sdk`。
- 移除未使用的 `IRobotManagerClient`，管理调用直接使用 `RobotManagerClient`。

验证客户端连接流程：

```sh
dotnet run --project tests/CVX.Client.Tests -c Release
```

当前 SDK 协议版本为 `2`。配套 Manager 必须：

- 在 `GET /v1/health` 返回 `service: "CVX.Manager"` 和 `protocolVersion: 2`；
- 实现 `PUT /v1/client`，按 `clientId + instanceId` 注册或替换回调；
- 实现 `DELETE /v1/client`，仅在实例 ID 匹配时注销回调；
- 在状态响应中返回 `httpCallbackReady`；
- 使用 `GET` 请求回调地址完成可达性检查。

本协议不保存或补发客户端离线期间的消息。
