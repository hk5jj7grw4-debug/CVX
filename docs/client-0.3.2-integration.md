# CVX.Client 0.3.2：GUI 新增功能对接

适用于从 0.3.1 升级的 .NET 8 GUI。业务 SDK 接口、消息回调订阅方式不变。

## GUI 需要改什么

1. 更新 `CVX.Client` 至 `0.3.2`。
2. 全局保留一个 `WechatRobotRuntime`，在第一次连接前订阅 `StatusChanged`。
3. 删除 GUI 自己的连接轮询、版本检查、解压和自动注入逻辑，交给 DLL。
4. 从 `Status` / `StatusChanged` 展示状态和版本；事件中异步切回 UI 线程。
5. 程序退出时 `await DisposeAsync()`。

```sh
dotnet add package CVX.Client --version 0.3.2
```

## 新增接口

```csharp
public WechatRobotStatus Status { get; }
public event EventHandler<WechatRobotStatus>? StatusChanged;
```

`Status` 是最近一次状态快照，读取不会触发检测。事件包含同一类型的快照；即便阶段未变化，探测完成也可能再次通知，不要每次弹窗。

现有接口继续使用：

| GUI 操作 | 调用 |
|---|---|
| 启动时接入 | `await runtime.ConnectAsync()` |
| 用户点击“重新接入” | `await runtime.ConnectAsync()` |
| 用户点击“重启微信” | `await runtime.RestartAsync()` |
| 退出、停止自动监控 | `await runtime.DisposeAsync()` |

首次调用 `ConnectAsync()` 即启动后台监控，首次连接抛异常后仍会自动重试。取消单次连接不会停止后台监控；彻底停止需要释放 Runtime。不要在每次错误后新建 Runtime。

## 状态字段

| 字段 | 展示语义 |
|---|---|
| `Phase` | 当前阶段，见下表 |
| `ApiReady` | 内核接口是否就绪 |
| `IsLoggedIn` | 微信是否登录，仅 `ApiReady == true` 时有意义 |
| `ComponentVersion` | 注入组件包版本，不是微信版本 |
| `RequiredWechatVersion` | 当前目标微信版本 `4.1.8.27` |
| `DetectedWechatVersion` | 检测到的磁盘微信运行时版本；无结果时为 null |
| `VersionMatch` | `Unknown` / `Matching` / `Mismatched` |
| `CheckedAt` | 最近状态检查时间；不表示该时刻重新读取了磁盘版本 |
| `DownloadProgress` | 下载百分比 0–100；null 时展示不确定进度 |
| `Error` | 最近失败信息；由 `Failed` 阶段展示 |
| `CallbackUrl` | 实际使用的回调地址 |

健康连接会沿用运行记录中的版本信息，不重复扫描；旧记录没有版本时显示“未知”，不能自行用 `ComponentVersion` 填充。

| Phase | 建议文案 |
|---|---|
| `Idle` | 尚未连接 |
| `Checking` | 正在检查微信版本 |
| `Downloading` | 正在下载组件 |
| `Repairing` | 正在恢复微信 |
| `Connecting` | 正在启动并连接内核 |
| `Connected` | 根据登录状态显示“已连接”或“内核已就绪，请登录微信” |
| `Failed` | 展示 `Error`，可提供“重新接入”按钮 |
| `Disposed` | 已停止监控 |

`Connected` 且 `IsLoggedIn == false` 表示等待扫码登录，不应触发 GUI 重启。不要把 `VersionMatch == Unknown` 单独视为连接失败，以 `ApiReady` 和 `Phase` 展示连接结果。

## WPF 接入示例

以下成员放在现有窗口或控制器里。`RenderStatus` 负责把状态映射到现有控件或 ViewModel。

```csharp
using CVX.Client;

private readonly WechatRobotRuntime _runtime = new(new WechatRobotOptions
{
    UpdateServerUrl = "https://your-component-server.example",
    ComponentToken = "your-component-token",
});
private bool _closing;

private async Task InitializeClientAsync()
{
    _runtime.StatusChanged += OnRuntimeStatusChanged;
    RenderStatus(_runtime.Status);
    try
    {
        await _runtime.ConnectAsync();
    }
    catch (Exception)
    {
        // 错误已包含在 Status 中；后台监控仍会继续。
        RenderStatus(_runtime.Status);
    }
}

private void OnRuntimeStatusChanged(object? sender, WechatRobotStatus status)
{
    if (_closing) return;
    Dispatcher.BeginInvoke(new Action(() =>
    {
        if (!_closing) RenderStatus(status);
    }));
}

private async Task ReconnectAsync()
{
    try { await _runtime.ConnectAsync(); }
    catch (Exception) { RenderStatus(_runtime.Status); }
}

private async Task RestartWechatAsync()
{
    try { await _runtime.RestartAsync(); }
    catch (Exception) { RenderStatus(_runtime.Status); }
}

private async Task ShutdownClientAsync()
{
    _closing = true;
    _runtime.StatusChanged -= OnRuntimeStatusChanged;
    await _runtime.DisposeAsync();
}
```

初始化只执行一次；关闭窗口时在现有异步退出流程中等待 `ShutdownClientAsync()`。不要在事件中使用 `.Wait()` / `.Result`，也不要同步调用 Dispatcher，否则可能阻塞恢复或退出。WinForms 对应使用 `BeginInvoke`，其他框架使用各自 UI Dispatcher。

按钮操作期间应禁用重复点击。DLL 会串行处理手动请求；自动检查则在已有操作执行时跳过，GUI 无需再次调用连接。

## DLL 自动执行的流程

- 每 3 秒异步探测内核，接口超时 2 秒。
- 通信正常直接复用，不检查磁盘版本，不因磁盘更新而中断当前微信。
- 通信失败、超时、HTTP 错误或无效协议响应，立即检查磁盘版本。
- 版本兼容：只恢复受管微信，不关闭其他微信会话。
- 确认版本不兼容：关闭所有目录中的微信主进程、扩展及更新进程；退出失败则停止修复，不能覆盖文件。
- 退出完成后使用校验通过的本地微信 ZIP 恢复；本地包缺失或损坏才使用 Token 请求 `updates/latest` 下载。
- 完成后重新启动、注入、连接，自动更新状态。

版本修复会中断其他目录中的微信会话，GUI 应在功能说明中明确这个行为。端口属于非受管程序时会报错，不会关闭无关程序。

下载和完整恢复不受单次接口的 2 秒超时限制，仍使用已有 `ConnectTimeout`（默认 12 分钟）和 `StartTimeout`（默认 30 秒）。Token 只用于远程组件请求。

不需要 GUI 增加定时器、查询远程版本、读取进程模块或执行解压程序。释放 Runtime 停止监控和回调，不会主动关闭微信。
