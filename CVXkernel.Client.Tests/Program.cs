using System.Net;
using System.Text;
using CVXkernel.Client;

await TestMessageParserAsync();
await TestCallbackReceiverAsync();
await TestManagerClientAsync();
TestRuntimeOwnership();
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
    using var healthResponse = await http.GetAsync(callback);
    Equal(HttpStatusCode.OK, healthResponse.StatusCode, "回调健康检查状态");
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
        var path = request.RequestUri?.AbsolutePath;
        var json = path switch
        {
            "/v1/health" => """
                {
                  "ok": true,
                  "service": "Sao.WechatRobotManager",
                  "version": "1.0.0",
                  "protocolVersion": 2,
                  "processId": 456
                }
                """,
            "/v1/client" when request.Method == HttpMethod.Put => """
                {
                  "ok": true,
                  "clientId": "test-client",
                  "instanceId": "test-instance",
                  "callbackReady": true
                }
                """,
            "/v1/client" when request.Method == HttpMethod.Delete => "{}",
            "/v1/status" => """
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
                  "httpCallbackReady": true,
                  "lastError": null
                }
                """,
            _ => throw new InvalidOperationException("未预期的 Manager 请求: " + request),
        };
        return new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent(json, Encoding.UTF8, "application/json"),
        };
    });
    using var http = new HttpClient(handler);
    using var client = new RobotManagerClient(http, apiToken: "local-test-token");
    var health = await client.GetHealthAsync();
    Equal(2, health.ProtocolVersion, "Manager 协议版本反序列化");
    var session = await client.RegisterClientAsync(new RobotClientRegistration(
        "test-client",
        "test-instance",
        "http://127.0.0.1:5000/api/recvMsg"));
    Equal(true, session.CallbackReady, "客户端实例注册");
    var status = await client.GetStatusAsync();
    Equal("local-test-token", handler.LastRequest?.Headers.GetValues("X-Robot-Token").Single(), "Manager 本地鉴权");
    Equal(true, status.GvxApiReady, "Manager 状态反序列化");
    Equal(true, status.HttpCallbackReady, "回调状态反序列化");
    Equal(19088, status.GvxApiPort, "GVx 端口反序列化");
    await client.UnregisterClientAsync("test-client", "test-instance");
    Equal("test-client", GetQueryValue(handler.LastRequest?.RequestUri, "clientId"), "注销客户端 ID");
    Equal("test-instance", GetQueryValue(handler.LastRequest?.RequestUri, "instanceId"), "注销客户端实例 ID");
}

static string? GetQueryValue(Uri? uri, string key)
{
    if (uri is null) return null;
    foreach (var item in uri.Query.TrimStart('?').Split('&', StringSplitOptions.RemoveEmptyEntries))
    {
        var parts = item.Split('=', 2);
        if (Uri.UnescapeDataString(parts[0]) == key)
            return parts.Length == 2 ? Uri.UnescapeDataString(parts[1]) : "";
    }
    return null;
}

static void TestRuntimeOwnership()
{
    var directory = Path.Combine(Path.GetTempPath(), "cvxkernel-tests", Guid.NewGuid().ToString("N"));
    var lockPath = Path.Combine(directory, ".runtime-owner.lock");
    try
    {
        using (RuntimeOwnershipLease.Acquire(lockPath))
        {
            var blocked = false;
            try
            {
                using var duplicate = RuntimeOwnershipLease.Acquire(lockPath);
            }
            catch (InvalidOperationException)
            {
                blocked = true;
            }
            Equal(true, blocked, "运行时重复所有权");
        }

        using var reacquired = RuntimeOwnershipLease.Acquire(lockPath);
    }
    finally
    {
        if (Directory.Exists(directory)) Directory.Delete(directory, recursive: true);
    }
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
