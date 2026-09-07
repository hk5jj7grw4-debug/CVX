# CVX

> 此为私人项目依赖包，无实际内核程序，请勿下载使用。

## 项目

| 项目 | 职责 | 产物 |
|---|---|---|
| `CVX.Client` | 本地组件安装、启动注入、手动恢复与消息回调 | DLL / NuGet |
| `CVX.Sdk.DotNet` | 消息、联系人、群等业务接口 | DLL / NuGet |
| `CVX.Sdk.Go` | Go 业务接口 | Go 模块 |

无需独立管理进程或管理 HTTP 端口。`inject.exe`、内核 DLL 和便携微信来自组件服务，不由本仓库构建。

## 接入与恢复

```csharp
using CVX.Client;

await using var runtime = new WechatRobotRuntime(new WechatRobotOptions
{
    UpdateServerUrl = "https://your-component-server.example",
    ComponentToken = "component-download-token",
});
runtime.MessageReceived += (_, message) => Console.WriteLine(message.Text);
runtime.ReceiveError += (_, error) => Console.Error.WriteLine(error.Message);

var status = await runtime.ConnectAsync();
Console.WriteLine(status.IsLoggedIn ? "已登录" : "内核就绪，等待登录");

// 微信崩溃后，在当前程序中再次调用即可恢复，无需重建 Runtime 或订阅。
await runtime.ConnectAsync();

// UI 的“重启微信”按钮可以调用：
await runtime.RestartAsync();
```

`ConnectAsync` 先启动回调监听，再通过 `POST /api/check_login` 验证内核响应。正常且进程身份、端口、回调记录一致时直接复用；内核不可用时准备本地组件、启动并注入。账号尚未登录也属于内核就绪，不会因此反复重启。

`RestartAsync` 明确停止组件目录内的受管微信，再重新注入。异常 API 响应可以通过此入口恢复；无法识别为受管进程的在线服务不会被停止。此操作会中断当前微信会话。

连接和重启按 Runtime 串行执行，并使用组件目录中的跨进程独占锁；同一组件目录只允许一个活跃回调客户端。锁随连接保留，以免其他客户端覆盖回调，释放 Runtime 后解除。

## 配置

| 选项 | 默认值 / 用途 |
|---|---|
| `UpdateServerUrl` | 组件服务地址，首次安装或修复损坏文件时需要 |
| `ComponentToken` | 仅用于组件版本查询和安装包下载 |
| `ComponentDirectory` | `%LOCALAPPDATA%\CVX` |
| `GvxApiPort` | `19088` |
| `CallbackUrl` | `http://127.0.0.1:5000/api/recvMsg`，仅本机 HTTP |
| `ConnectTimeout` | 整个连接或重启过程最多 12 分钟 |
| `StartTimeout` | 启动注入后最多等待 30 秒 |

已安装且校验通过的组件直接使用，不要求 Token，不自动检查或升级。组件缺失或损坏时使用 Bearer Token 下载，校验大小、SHA-256、清单文件及便携微信兼容版本后安装。Token 不发送到本机内核，也不写入运行状态文件。

Runtime 返回 `WechatRobotStatus`：`ApiReady`、`IsLoggedIn`、`ComponentVersion` 和实际 `CallbackUrl`。启动失败、超时或取消通过异常返回；可在同一 Runtime 上重试。

## 组件目录

```text
%LOCALAPPDATA%\CVX\
├─ client-runtime.json
├─ .runtime-owner.lock
├─ downloads\
└─ versions\<组件版本>\
   ├─ inject.exe
   ├─ libGLESv1.dll
   ├─ manifest.json
   ├─ Weixin.zip
   └─ Weixin\
      ├─ Weixin.exe
      └─ 4.1.8.27\
```

便携包直接解压到组件版本目录，保留 ZIP 自带的目录结构；不额外添加 `weixin` 包装层。`manifest.json` 的 `exePath: Weixin/Weixin.exe` 相对组件版本目录解析。官方版本子目录和文件保持原样，组件版本与微信内部版本独立。

## 退出与回调

使用 `DisposeAsync` / `await using` 关闭连接。退出只取消正在执行的连接并关闭回调监听，不结束微信或内核；所有业务客户端退出后不提供后台自动恢复，下次连接时再检测和恢复。

运行记录 `client-runtime.json` 保存受管进程 PID、启动时间、可执行路径、内核端口和已应用回调地址。复用时检查这些记录；回调地址变化时通过重新注入应用。HTTP 探活和已应用记录不等于端到端消息投递确认。

消息事件在后台线程执行，UI 更新需切回 UI 线程。HTTP 确认仅代表接收，不代表业务处理完成；不保证事件处理顺序，已派发的处理可能在释放后结束。客户端离线期间不保存或补发消息。

## 从独立 Manager 迁移

- 部署新版前退出旧 Manager，并停止外部配置的旧 Manager 自启动，避免它和 Client 同时管理微信。
- 删除 `ManagerAddress`、`ClientId` 及 Manager 本体下载配置。默认组件根目录为 `%LOCALAPPDATA%\CVX`，不自动回退到 `%LOCALAPPDATA%\Sao\WechatRobot`。
- `RobotManagerClient` 及其协议模型已删除；普通恢复用 `ConnectAsync`，明确重启用 `RestartAsync`。
- 未显式设置组件服务地址或 Token 时，可从组件目录的旧 `manager-settings.json` 读取这两个字段。旧 API Token、自动启动及期望状态均不使用。
- 默认不会读取、写入或管理旧 Sao 目录。迁移前关闭旧目录运行的微信，可将旧目录的 `versions` 复制到 CVX；也可配置组件服务后重新下载。复制的 inject、DLL、manifest 和 ZIP 可复用，便携包会在新版本目录直接解压。不要同时运行两个根目录的微信，也不要复制旧的 `client-runtime.json`。
- 原来配置过非默认内核端口或回调地址的调用方，需要在新选项中明确设置。

## 构建与验证

```sh
dotnet build CVX.sln -c Release
dotnet run --project tests/CVX.Client.Tests -c Release --no-build
go -C CVX.Sdk.Go test ./...
```

Client 检查使用本地 HTTP 服务和模拟进程宿主，覆盖恢复、重启、回调及组件下载。真实 `inject.exe` 和微信进程生命周期需要在 Windows 上验证。

发布版本统一在 `Directory.Build.props` 设置本地默认值。CI 的普通构建使用 `<默认版本>-ci.<运行编号>.<重试次数>`；`vX.Y.Z` 或 `vX.Y.Z-prerelease` 标签构建使用标签去掉 `v` 后的版本。恢复、编译及两个 NuGet 包共用该版本，打包不重复编译；预发布标签创建预发布 GitHub Release。只有标签构建执行远端发布。
