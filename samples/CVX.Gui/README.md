# CVX.Gui

WPF 示例客户端，演示 `CVX.Client` 生命周期与 `CVX.Sdk.DotNet` 业务接口。不是可发布的 NuGet 产物。

```sh
dotnet run --project samples/CVX.Gui -c Release
```

首次安装组件时需要更新服务地址和组件 Token。可设置环境变量 `CVX_UPDATE_SERVER`、`CVX_COMPONENT_TOKEN`，或写入 `%LOCALAPPDATA%\CVX\gui-credentials.json`。已安装且校验通过的组件不要求 Token。

自动版本修复会关闭其他目录中的微信会话。关闭窗口只释放本客户端，不会停止微信。
