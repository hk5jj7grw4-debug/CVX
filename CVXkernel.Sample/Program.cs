using GVxSdk;
using CVXkernel.Client;

if (!OperatingSystem.IsWindows())
{
    Console.Error.WriteLine("此示例只能在安装了微信的 Windows 环境运行。");
    return;
}

var updateServerUrl = Environment.GetEnvironmentVariable("SAO_ROBOT_UPDATE_SERVER")
    ?? throw new InvalidOperationException("请设置 SAO_ROBOT_UPDATE_SERVER");
var downloadToken = Environment.GetEnvironmentVariable("SAO_ROBOT_DOWNLOAD_TOKEN") ?? "";
var componentToken = Environment.GetEnvironmentVariable("SAO_WECHAT_COMPONENT_TOKEN") ?? "";
var localApiToken = Environment.GetEnvironmentVariable("SAO_ROBOT_API_TOKEN") ?? "";

await using var runtime = new WechatRobotRuntime(new RobotManagerBootstrapperOptions
{
    DownloadToken = downloadToken,
    ApiToken = localApiToken,
});

runtime.Messages.MessageReceived += (_, message) =>
    Console.WriteLine($"[{message.ReceivedAt:HH:mm:ss}] {message.GroupId}/{message.SenderName}: {message.Text}");
runtime.Messages.ReceiveError += (_, exception) =>
    Console.Error.WriteLine("回调处理错误: " + exception.Message);

var progress = new Progress<int>(value => Console.WriteLine($"Manager 下载进度: {value}%"));
var status = await runtime.EnsureRunningAsync(
    new RobotManagerConfig
    {
        UpdateServerUrl = updateServerUrl,
        ComponentToken = componentToken,
        GvxApiPort = 19088,
    },
    new WechatCallbackReceiverOptions
    {
        // 使用 0 可由系统自动选择未占用端口，实际地址会自动下发给 Manager。
        CallbackUrl = "http://127.0.0.1:0/api/recvMsg",
    },
    progress);

Console.WriteLine($"Manager 已连接，微信状态: {status.RuntimeState}，GVx API: {status.GvxApiPort}");

// 生命周期 SDK 不包装 GVx。业务程序直接使用官方/现有 GVxSdk.Client。
using var gvx = new GVxClient($"http://127.0.0.1:{status.GvxApiPort}");
await gvx.Message.SendTextAsync("filehelper", "CVXkernel.Client 接入测试");

Console.WriteLine("正在接收消息，按 Enter 退出……");
Console.ReadLine();
