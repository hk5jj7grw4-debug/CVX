# CVX.Manager 对接文档

本文面向需要复用微信机器人能力的 Windows 宿主 App。RobotManager 只管理微信、注入组件和本地 gVx API 的生命周期；消息、群聊、联系人等业务接口由宿主 App 直接调用 CVX SDK。

## 1. 架构与职责

```text
宿主 App
├─ 管理器安装与更新
├─ HTTP 生命周期调用 ───────> RobotManager 127.0.0.1:19087
├─ HTTP 微信业务调用 ───────> gVx API      127.0.0.1:19088
└─ HTTP 消息回调接收 <────── gVx API      callbackUrl

RobotManager
├─ 下载、校验和安装机器人组件
├─ 启动或停止便携微信
├─ 单次运行 inject.exe 完成注入
├─ 检测 gVx API 是否就绪
└─ 微信异常退出时按期望状态自动恢复
```

边界约定：

- RobotManager 不代理 gVx 业务 API，也不缓存、排队或重试微信消息。
- 宿主 App 退出或更新时，不应停止 RobotManager 和微信。
- RobotManager 退出时不会停止微信；只有 `POST /v1/stop` 会明确停止微信。
- RobotManager 不更新自身。宿主 App 负责下载和切换 RobotManager 版本。
- RobotManager 下载机器人组件时不使用系统代理。

## 2. 默认地址与进程

| 项目 | 默认值 |
| --- | --- |
| Manager API | `http://127.0.0.1:19087` |
| gVx API | `http://127.0.0.1:19088` |
| 消息回调 | `http://127.0.0.1:5000/api/recvMsg` |
| Manager 程序 | `CVX.Manager.exe` |
| 组件数据目录 | `%LOCALAPPDATA%\Sao\WechatRobot` |

Manager 使用 `WinExe` 构建。宿主 App 应隐藏启动它，不会出现控制台黑窗口。

## 3. 推荐初始化流程

宿主 App 每次启动执行以下幂等流程：

1. 查找本地已安装的最新 RobotManager 版本。
2. 启动 `CVX.Manager.exe`，不要等待该进程退出。
3. 轮询 `GET /v1/health`，建议最多等待 20 秒。
4. 启动宿主 App 的本地回调服务。
5. 调用 `PUT /v1/client` 注册本次客户端实例及回调地址。
6. 调用 `PUT /v1/config` 写入当前宿主配置。
7. 调用 `POST /v1/reconcile`。
8. 轮询 `/v1/status`，直到 `gvxApiReady` 和 `httpCallbackReady` 都为 `true`。
9. 业务请求直接发送到 gVx API，消息由 gVx API 主动回调宿主 App。

`/v1/reconcile` 会将期望状态设为 `running`，并依次完成缺失组件安装、文件修复、组件更新检查、微信启动和注入。重复调用是安全的。

如果用户此前明确执行过 `/v1/stop`，而宿主 App 希望保留停止状态，则启动时只读取 `/v1/status`，不要自动调用 `/v1/reconcile`。

## 4. 鉴权

Manager 仅允许监听本机回环地址，所有生命周期接口均无需 Token。

仅组件版本查询和下载使用 `componentToken`，由 Manager 保存，并在访问组件服务时携带 Bearer 凭据。Client 下载 Manager 本体采用匿名访问，不需要另外的下载 Token。

旧配置中的 `apiToken` 已废弃，不会阻止本机请求。

## 5. 配置接口

### `PUT /v1/config`

请求字段均可选，仅更新传入的字段：

```json
{
  "updateServerUrl": "http://update-server.example:38081",
  "componentToken": "component-download-token",
  "autoStart": true,
  "gvxApiPort": 19088,
  "httpCallbackUrl": "http://127.0.0.1:5000/api/recvMsg",
  "startTimeoutSeconds": 30
}
```

约束：

- `gvxApiPort`：`1..65535`。
- `httpCallbackUrl`：必须是绝对 HTTP 或 HTTPS 地址。
- `startTimeoutSeconds`：自动限制在 `5..300` 秒。
- 修改 `gvxApiPort` 或 `httpCallbackUrl` 后调用 `/v1/reconcile`；Manager 会在需要时重新注入。
- Token 只在本地配置文件中保存，`GET /v1/config` 仅返回是否已配置，不返回明文。

配置保存在：

```text
%LOCALAPPDATA%\Sao\WechatRobot\manager-settings.json
```

## 6. 生命周期 API

| Method | Path | 使用场景 |
| --- | --- | --- |
| `GET` | `/v1/health` | 判断 Manager 是否启动并读取 Manager 版本 |
| `GET` | `/v1/status` | 读取微信版本、运行状态和最近错误 |
| `GET` | `/v1/operation` | 读取当前或最近一次生命周期操作进度 |
| `PUT` | `/v1/client` | 注册或原子替换当前客户端实例 |
| `DELETE` | `/v1/client` | 仅在实例 ID 匹配时注销当前客户端 |
| `POST` | `/v1/initialize` | 等同于 `/v1/reconcile` |
| `POST` | `/v1/reconcile` | 启动时进行完整校准并保证机器人在线 |
| `POST` | `/v1/start` | 使用已安装组件启动微信机器人 |
| `POST` | `/v1/stop` | 停止微信并保存期望状态 `stopped` |
| `POST` | `/v1/restart` | 停止微信后重新启动和注入 |
| `POST` | `/v1/repair` | 校验并修复组件，按原期望状态决定是否启动 |
| `GET` | `/v1/component/update` | 仅检查机器人组件更新 |
| `POST` | `/v1/component/update` | 安装机器人组件更新，必要时重启微信 |
| `GET` | `/v1/config` | 查询脱敏配置 |
| `PUT` | `/v1/config` | 局部更新并持久化配置 |
| `POST` | `/v1/host/shutdown` | 仅退出 Manager，保持微信和 gVx API 运行 |

所有会改变生命周期的操作在 Manager 内部串行执行。宿主 App 应禁用重复操作按钮，并为启动、修复和更新请求设置不少于 10 分钟的超时。

### 操作进度

宿主 App 调用生命周期 API 后，可以每 300 到 500 毫秒轮询 `GET /v1/operation`：

```json
{
  "running": true,
  "operation": "repair",
  "phase": "downloading",
  "phaseText": "正在下载机器人组件 42%",
  "progress": 33,
  "startedAt": "2026-08-31T12:10:20Z",
  "error": null
}
```

`progress` 是整个生命周期操作的总体进度；无法计算时为 `null`，宿主应显示不确定进度动画。`phase` 可能为 `starting`、`checking`、`downloading`、`verifying`、`extracting`、`stopping_wechat`、`starting_wechat`、`waiting_gvx_api`、`completed` 或 `failed`。

## 7. 响应模型

### Manager 健康状态

`GET /v1/health`：

```json
{
  "ok": true,
  "service": "CVX.Manager",
  "version": "0.2.0",
  "protocolVersion": 2,
  "processId": 8240
}
```

### 机器人运行状态

`GET /v1/status` 及大多数生命周期操作返回：

```json
{
  "ok": true,
  "desiredState": "running",
  "runtimeState": "online",
  "wechatRunning": true,
  "gvxApiReady": true,
  "gvxApiPort": 19088,
  "wechatProcessIds": [10936],
  "wechatVersion": "4.1.12.55",
  "componentVersion": "4.1.12.55",
  "versionStatus": "compatible",
  "supportedWechatVersions": ["4.1.12.55"],
  "httpCallbackReady": true,
  "lastError": null
}
```

字段说明：

| 字段 | 含义 |
| --- | --- |
| `desiredState` | 持久化期望状态：`running` 或 `stopped` |
| `runtimeState` | 当前状态：`online`、`starting` 或 `stopped` |
| `wechatRunning` | 是否存在微信进程 |
| `gvxApiReady` | gVx API 端口是否可连接；业务可用性的主要判断字段 |
| `httpCallbackReady` | 当前客户端实例、已注入回调地址和回调探活是否全部一致 |
| `wechatVersion` | 当前组件内便携微信版本 |
| `componentVersion` | 已安装机器人组件版本 |
| `versionStatus` | `compatible`、`incompatible` 或 `unknown` |
| `lastError` | Manager 最近一次生命周期操作的错误信息 |

注意：`ok` 当前与 `gvxApiReady` 一致。判断 Manager 本身是否存活应使用 `/v1/health`，不要使用状态响应中的 `ok`。

## 8. 错误处理

错误统一返回：

```json
{
  "ok": false,
  "error": "错误原因"
}
```

| HTTP 状态 | 含义 | 宿主处理建议 |
| --- | --- | --- |
| `401` | 组件下载 Token 无效 | 更新 Manager 的组件 Token |
| `409` | 配置、组件状态或操作前提不满足 | 展示服务端 `error`，不要盲目重试 |
| `500` | 未分类内部错误 | 展示并记录 `error`，允许用户重新修复 |
| `504` | 微信启动或 gVx API 就绪超时 | 展示错误，保留“检查修复”和“重启”操作 |

网络连接失败表示 Manager 未运行、正在更新或异常退出。宿主 App 可重新启动 Manager 并再次等待 `/v1/health`，但不要因此直接结束微信进程。

## 9. Manager 本体更新

Manager 不能可靠覆盖正在运行的自身程序，因此由宿主 App 完成版本切换。建议使用不可变版本目录：

```text
%LOCALAPPDATA%\YourApp\WechatRobotManagerHost
├─ downloads\
└─ versions\
   ├─ 0.1.1\CVX.Manager.exe
   └─ 0.1.2\CVX.Manager.exe
```

更新流程：

1. 宿主 App 向自己的更新服务检查 `wechat-robot-manager` 最新版本。
2. 下载 ZIP，并校验产物大小和 SHA-256。
3. 解压到新的版本目录，禁止覆盖当前目录。
4. 调用 `POST /v1/host/shutdown`。
5. 等待 `/v1/health` 无法连接，最多等待 10 秒。
6. 启动新目录中的 Manager。
7. 等待新 Manager `/v1/health` 就绪并验证 `version`。
8. 重新调用 `/v1/config` 和 `/v1/reconcile`。
9. 新版启动失败时，重新启动上一版本并报告更新失败。

`/v1/host/shutdown` 的成功响应：

```json
{
  "ok": true,
  "stopping": true,
  "wechatUntouched": true
}
```

这段切换只中断生命周期 API。微信、已注入组件和 `19088` 业务 API 保持运行，除非新 Manager 在 `/v1/reconcile` 中发现组件确实需要更新或修复。

## 10. 最小 C# 调用示例

```csharp
using System.Net.Http.Json;

var manager = new HttpClient
{
    BaseAddress = new Uri("http://127.0.0.1:19087"),
    Timeout = TimeSpan.FromMinutes(12),
};

var health = await manager.GetFromJsonAsync<ManagerHealth>("/v1/health");

using var configResponse = await manager.PutAsJsonAsync("/v1/config", new
{
    updateServerUrl,
    componentToken,
    autoStart = true,
    gvxApiPort = 19088,
    httpCallbackUrl = "http://127.0.0.1:5000/api/recvMsg",
    startTimeoutSeconds = 30,
});
configResponse.EnsureSuccessStatusCode();

using var reconcileResponse = await manager.PostAsync("/v1/reconcile", null);
reconcileResponse.EnsureSuccessStatusCode();
var status = await reconcileResponse.Content.ReadFromJsonAsync<RobotStatus>();

if (status?.GvxApiReady == true)
{
    // 后续消息、群聊等请求直接调用 http://127.0.0.1:19088。
}

public sealed record ManagerHealth(
    bool Ok,
    string Service,
    string Version,
    int ProtocolVersion,
    int ProcessId);

public sealed record RobotStatus(
    bool Ok,
    string DesiredState,
    string RuntimeState,
    bool WechatRunning,
    bool GvxApiReady,
    int GvxApiPort,
    int[] WechatProcessIds,
    string? WechatVersion,
    string? ComponentVersion,
    string VersionStatus,
    string[] SupportedWechatVersions,
    bool HttpCallbackReady,
    string? LastError);
```

生产接入时还应处理非成功 HTTP 响应中的 `error` 字段，并避免在日志中输出任何 Token。

## 11. UI 建议

普通用户界面只需要常驻显示：

```text
微信 4.1.12.55  ·  管理器 v0.1.1
```

管理操作建议提供启动、检查修复、重启和停止四项。停止操作应二次确认；错误详情应展示 Manager 返回的原始 `error` 或 `lastError`，并提供复制功能。
