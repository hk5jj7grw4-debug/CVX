# CVX.Gui

WPF 示例客户端，演示 `CVX.Client` 生命周期与 `CVX.Sdk.DotNet` 业务接口。不是可发布的 NuGet 产物。

```sh
dotnet run --project samples/CVX.Gui -c Release
```

首次运行在窗口「配置」里填写组件服务 URL 和 Token，点「保存配置」后再接入。也可使用环境变量 `CVX_UPDATE_SERVER`、`CVX_COMPONENT_TOKEN`。已安装且校验通过的组件不要求 Token。

自动版本修复会关闭其他目录中的微信会话。关闭窗口只释放本客户端，不会停止微信。
