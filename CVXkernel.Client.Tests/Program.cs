using System.Net;
using System.Text;
using CVXkernel.Client;

await TestMessageParserAsync();
await TestCallbackReceiverAsync();
await TestManagerClientAsync();
Console.WriteLine("全部测试通过");

static Task TestMessageParserAsync()
{
    const string json = """
        {
          "msgType": "1",
          "fromUserName": { "string": "48778929610@chatroom" },
          "senderWxid": "wxid_sender",
          "senderNick": "测试用户",
          "content": { "string": "wxid_sender:\nHELLO" },
          "newMsgId": "9001"
        }
        """;
    var message = WechatCallbackMessageParser.Parse(Encoding.UTF8.GetBytes(json));
    Equal("48778929610@chatroom", message.GroupId, "群 ID 解析");
    Equal("wxid_sender", message.SenderId, "发送人解析");
    Equal("HELLO", message.Text, "群消息正文解析");
    Equal("9001", message.MessageId, "消息 ID 解析");
    return Task.CompletedTask;
}

static async Task TestCallbackReceiverAsync()
{
    await using var receiver = new WechatCallbackReceiver();
    var completion = new TaskCompletionSource<WechatCallbackMessage>(TaskCreationOptions.RunContinuationsAsynchronously);
    receiver.MessageReceived += (_, message) => completion.TrySetResult(message);
    var callback = receiver.Start(new WechatCallbackReceiverOptions
    {
        CallbackUrl = "http://127.0.0.1:0/api/recvMsg",
    });

    using var http = new HttpClient();
    using var body = new StringContent("{\"type\":\"text\",\"text\":\"ping\",\"messageId\":\"1\"}", Encoding.UTF8, "application/json");
    using var response = await http.PostAsync(callback, body);
    Equal(HttpStatusCode.OK, response.StatusCode, "回调 HTTP 状态");
    var message = await completion.Task.WaitAsync(TimeSpan.FromSeconds(3));
    Equal("ping", message.Text, "回调事件正文");
    Equal(1L, receiver.ReceivedCount, "回调计数");
}

static async Task TestManagerClientAsync()
{
    var handler = new StubHandler(request =>
    {
        Equal("/v1/status", request.RequestUri?.AbsolutePath, "Manager 请求路径");
        const string json = """
            {
              "ok": true,
              "desiredState": "running",
              "runtimeState": "ready",
              "wechatRunning": true,
              "gvxApiReady": true,
              "gvxApiPort": 19088,
              "wechatProcessIds": [123],
              "wechatVersion": "4.0",
              "componentVersion": "1.0",
              "versionStatus": "supported",
              "supportedWechatVersions": ["4.0"],
              "lastError": null
            }
            """;
        return new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent(json, Encoding.UTF8, "application/json"),
        };
    });
    using var http = new HttpClient(handler);
    using var client = new RobotManagerClient(http, apiToken: "local-test-token");
    var status = await client.GetStatusAsync();
    Equal("local-test-token", handler.LastRequest?.Headers.GetValues("X-Robot-Token").Single(), "Manager 本地鉴权");
    Equal(true, status.GvxApiReady, "Manager 状态反序列化");
    Equal(19088, status.GvxApiPort, "GVx 端口反序列化");
}

static void Equal<T>(T expected, T actual, string name)
{
    if (!EqualityComparer<T>.Default.Equals(expected, actual))
        throw new InvalidOperationException($"{name}失败，期望 {expected}，实际 {actual}");
}

sealed class StubHandler(Func<HttpRequestMessage, HttpResponseMessage> responseFactory) : HttpMessageHandler
{
    public HttpRequestMessage? LastRequest { get; private set; }

    protected override Task<HttpResponseMessage> SendAsync(
        HttpRequestMessage request,
        CancellationToken cancellationToken)
    {
        LastRequest = request;
        return Task.FromResult(responseFactory(request));
    }
}
