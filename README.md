# CVXkernel

> 此为私人项目依赖包，无实际内核程序，请勿下载使用。

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

业务客户端应通过 `WechatRobotRuntime.ConnectAsync` 接入，不自行管理 Manager 进程：

```csharp
await using var runtime = new WechatRobotRuntime(options);
var status = await runtime.ConnectAsync(config, callbackOptions);
```

`ConnectAsync` 会依次完成 Manager 探活/安装/启动、服务身份与协议校验、当前客户端实例注册、回调地址更新、状态协调及 Ready 等待。客户端退出只注销本次客户端实例并关闭回调监听，不会停止 Manager 或微信机器人。

当前 SDK 协议版本为 `2`。配套 Manager 必须：

- 在 `GET /v1/health` 返回 `service: "Sao.WechatRobotManager"` 和 `protocolVersion: 2`；
- 实现 `PUT /v1/client`，按 `clientId + instanceId` 注册或替换回调；
- 实现 `DELETE /v1/client`，仅在实例 ID 匹配时注销回调；
- 在状态响应中返回 `httpCallbackReady`；
- 使用 `GET` 请求回调地址完成可达性检查。

本协议不保存或补发客户端离线期间的消息。
