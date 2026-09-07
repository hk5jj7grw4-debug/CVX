# CVX

> 此为私人项目依赖包，无实际内核程序，请勿下载使用。

## 项目

| 项目 | 职责 | 产物 |
|---|---|---|
| `CVX.Client` | 本地组件安装、启动注入、自动恢复与消息回调 | DLL / NuGet |
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
runtime.StatusChanged += (_, state) =>
    Console.WriteLine($"{state.Phase}: 微信 {state.DetectedWechatVersion ?? "未知"} / 要求 {state.RequiredWechatVersion}, {state.Error}");

var status = await runtime.ConnectAsync();
Console.WriteLine(status.IsLoggedIn ? "已登录" : "内核就绪，等待登录");

// DLL 已自动监控；UI 的“重新接入”按钮也可以调用，无需重建 Runtime 或订阅。
await runtime.ConnectAsync();

// UI 的“重启微信”按钮可以调用：
await runtime.RestartAsync();
```

`ConnectAsync` 先启动回调监听，再通过 `POST /api/check_login` 验证内核响应。正常且进程身份、端口、回调记录一致时直接复用；内核不可用时准备本地组件、启动并注入。账号尚未登录也属于内核就绪，不会因此反复重启。

`RestartAsync` 明确停止组件目录内的受管微信，再重新注入。受管内核的异常 API 响应也会由自动检查恢复；无法识别为受管进程的在线服务不会被停止。此操作会中断当前微信会话。

首次 `ConnectAsync` 后，DLL 每 3 秒异步探测一次，单次接口请求超时 2 秒。通信失败、超时、HTTP 错误或无效协议响应都立即检查本地版本；正常通信时不扫描磁盘版本，即使磁盘已经更新也不修复。检查和恢复尚未完成时跳过定时检查，不重复排队。首次连接失败后监控仍可重试，停止监控请释放 Runtime。

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

当前目标微信为 `4.1.8.27`。兼容性按清单 `supportedWechatVersions` 的四段版本精确白名单判断，空列表不限制。连接异常时优先检查启动器旁四段目录中的 `Weixin.dll` 固定数字版本，优先选目录与 DLL 版本一致的最高完整候选；没有完整候选时使用可读 DLL 的版本，再依次退回最高四段目录和启动器固定数字版本；确认版本不匹配则关闭所有目录中的微信主进程、扩展及更新进程（包括其他微信会话），确认退出后用校验通过的本地 ZIP 恢复；关闭失败则停止修复，不覆盖文件。版本兼容的普通故障仅处理受管微信。包缺失或损坏才请求 `updates/latest` 下载。已安装且校验通过的组件直接使用，不要求 Token，不主动同步远程版本。组件缺失或损坏时使用 Bearer Token 下载，校验大小、SHA-256、清单文件及便携微信兼容版本后安装。Token 不发送到本机内核，也不写入运行状态文件。

`Status` 返回最近状态，`StatusChanged` 通知状态变化；`ConnectAsync` / `RestartAsync` 返回同一类型 `WechatRobotStatus`。包含 `ApiReady`、`IsLoggedIn`、`ComponentVersion`、`CallbackUrl`、`RequiredWechatVersion`、`DetectedWechatVersion`、`VersionMatch`、`Phase`、`CheckedAt`、`Error` 和 `DownloadProgress`（0–100，未知时为空）。`IsLoggedIn` 只有在 `ApiReady` 时有意义。

检测版本来自磁盘运行时 DLL（兜底为版本目录或启动器），不代表已加载模块的运行时证明。健康复用时沿用持久化的版本信息，旧记录没有版本则返回未知；`CheckedAt` 是最近状态检查时间，不意味着重新扫描了版本。正常探测不计算 ZIP 哈希或解压。

状态事件可能来自后台线程，GUI 应通过 Dispatcher / SynchronizationContext 切回 UI 线程，事件处理器不应同步等待连接或释放。主动调用失败通过异常返回，后台失败通过 `Status.Error` 展示。

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

使用 `DisposeAsync` / `await using` 关闭连接。退出会停止并等待监控、取消正在执行的连接并关闭回调监听，不结束微信或内核；所有业务客户端退出后不提供后台自动恢复，下次连接时再检测和恢复。

运行记录 `client-runtime.json` 保存受管进程 PID、启动时间、可执行路径、内核端口、已检查的微信版本和已应用回调地址。复用时核对端口实际监听 PID 及进程身份，辅助进程不会被当成多个实例；回调地址变化时通过重新注入应用。HTTP 探活和已应用记录不等于端到端消息投递确认。

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
