# CVX.Manager

独立的微信机器人生命周期管理程序。它只负责保证注入组件和便携微信正确运行，不代理 CVX SDK 的帐号、群和消息 API。

其他应用接入时请先阅读 [INTEGRATION.md](./INTEGRATION.md)，其中包含完整启动流程、接口模型、更新交接与错误处理约定。

## API 边界

```text
业务 App ──生命周期 HTTP──> RobotManager :19087
业务 App ──微信业务 HTTP──> gVx DLL     :19088
gVx DLL  ──消息回调 HTTP──> 业务 App 配置的 callbackUrl
```

RobotManager 仅允许监听本机回环地址，默认 `127.0.0.1:19087`，本地接口无需 Token。仅检查、下载机器人组件使用 `ComponentToken`；旧 `ApiToken` 配置不再生效。

## 生命周期 API

| Method | Path | 作用 |
| --- | --- | --- |
| GET | `/v1/health` | Manager 进程健康状态 |
| GET | `/v1/status` | 微信、gVx API 和组件状态 |
| GET | `/v1/operation` | 当前或最近一次生命周期操作进度 |
| PUT | `/v1/client` | 注册或替换当前客户端实例及回调地址 |
| DELETE | `/v1/client` | 按客户端和实例 ID 注销当前回调 |
| POST | `/v1/initialize` | 安装/修复缺失组件并启动 |
| POST | `/v1/reconcile` | 检查组件更新、修复并保证机器人在线 |
| POST | `/v1/start` | 启动；19088 已在线时直接复用 |
| POST | `/v1/stop` | 明确停止微信并关闭自动拉起 |
| POST | `/v1/restart` | 重启微信并重新注入 |
| POST | `/v1/repair` | 校验和修复组件，按期望状态启动 |
| GET | `/v1/component/update` | 检查组件更新 |
| POST | `/v1/component/update` | 下载、校验、安装并启动新版 |
| GET | `/v1/config` | 查询非敏感配置 |
| PUT | `/v1/config` | 更新生命周期配置 |
| POST | `/v1/host/shutdown` | 仅退出 Manager，不关闭微信 |

`PUT /v1/config` 接受局部字段：

```json
{
  "updateServerUrl": "http://203.195.196.180:38081",
  "componentToken": "token",
  "autoStart": true,
  "gvxApiPort": 19088,
  "httpCallbackUrl": "http://127.0.0.1:5000/api/recvMsg",
  "startTimeoutSeconds": 30
}
```

修改 `gvxApiPort` 或 `httpCallbackUrl` 后调用 `/v1/reconcile`。Manager 会在需要时重启微信并重新注入，使新配置生效。

## 运行规则

- `inject.exe` 只执行一次，注入完成后可以正常退出。
- Manager 以 `19088` 是否可连接作为机器人就绪判断。
- Manager 自己退出时不会关闭微信。
- 业务 App 更新或退出不会影响 Manager 和微信。
- 微信异常退出且期望状态为 `running` 时，Manager 会重新启动并注入。
- 调用 `/v1/stop` 后期望状态保存为 `stopped`，Manager 重启也不会自动拉起微信。
- 消息回调不经过 Manager，不保存、不排队、不重试。
- 客户端每次启动都使用新的 `instanceId` 注册；旧实例的注销请求不会清除新实例。

## 数据目录

默认复用现有目录：

```text
%LOCALAPPDATA%\Sao\WechatRobot
├─ manager-settings.json
├─ downloads\
└─ versions\<component-version>\
```

组件包格式继续兼容现有 `manifest.json + inject.exe + DLL + wechatBundle`。

## 发布

```powershell
dotnet publish .\CVX.Manager.csproj -c Release -r win-x64 --self-contained true \
  -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true
```

项目使用 `WinExe`，直接运行时不会弹出控制台黑窗口。
