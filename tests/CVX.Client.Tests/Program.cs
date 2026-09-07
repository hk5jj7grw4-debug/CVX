using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using CVX.Client;

await using var server = new FakeManager();
var root = Path.Combine(Path.GetTempPath(), "cvx-client-tests-" + Guid.NewGuid().ToString("N"));
var options = new WechatRobotOptions
{
    ManagerAddress = server.Address,
    InstallDirectory = root,
    CallbackUrl = "http://127.0.0.1:0/api/recvMsg",
    ConnectTimeout = TimeSpan.FromSeconds(5),
    // An existing Manager must not contact this deliberately unreachable update server.
    UpdateServerUrl = "http://127.0.0.1:1",
};
try
{
    var runtime = new WechatRobotRuntime(options);
    var received = new TaskCompletionSource<WechatCallbackMessage>(TaskCreationOptions.RunContinuationsAsynchronously);
    runtime.MessageReceived += (_, message) => received.TrySetResult(message);
    await runtime.ConnectAsync();
    await runtime.ConnectAsync();
    Check(server.Registrations == 1 && server.Starts == 1, "Repeated connect reuses the connection");
    Check(server.Config is { } config && !config.TryGetProperty("apiToken", out _) &&
        !config.TryGetProperty("componentToken", out _) && !config.TryGetProperty("autoStart", out _),
        "Connection preserves omitted persistent settings");
    using var http = new HttpClient();
    using var response = await http.PostAsync(server.CallbackUrl, new StringContent("{\"content\":\"hello\"}", Encoding.UTF8, "application/json"));
    response.EnsureSuccessStatusCode();
    Check((await received.Task.WaitAsync(TimeSpan.FromSeconds(3))).Text == "hello", "Public message event works");
    await runtime.DisposeAsync();
    await runtime.DisposeAsync();
    Check(server.Unregistrations == 1 && server.Stops == 0, "Dispose unregisters once without stopping robot");

    server.Unauthorized = true;
    await using (var denied = new WechatRobotRuntime(options))
    {
        try { await denied.ConnectAsync(); throw new Exception("Expected 401"); }
        catch (RobotManagerException ex) when (ex.StatusCode == HttpStatusCode.Unauthorized) { }
    }
    Check(server.Registrations == 1, "401 remains an authentication error");
    server.Unauthorized = false;

    server.FailStart = true;
    await using (var retry = new WechatRobotRuntime(options))
    {
        try { await retry.ConnectAsync(); throw new Exception("Expected start failure"); }
        catch (RobotManagerException ex) when (ex.StatusCode == HttpStatusCode.Conflict) { }
        Check(server.Unregistrations == 2, "Failed connection unregisters");
        server.FailStart = false;
        await retry.ConnectAsync();
    }
    Check(server.Unregistrations == 3, "Retry reacquires callback and ownership");

    server.HoldStart = true;
    var pending = new WechatRobotRuntime(options);
    var connection = pending.ConnectAsync();
    await server.StartEntered.Task.WaitAsync(TimeSpan.FromSeconds(3));
    var disposal = pending.DisposeAsync().AsTask();
    try { await connection; throw new Exception("Expected cancellation"); }
    catch (OperationCanceledException) { }
    server.ReleaseStart.TrySetResult();
    await disposal.WaitAsync(TimeSpan.FromSeconds(4));
    Check(server.Stops == 0, "Disposal cancels an in-flight connection without stopping robot");
    Console.WriteLine("All Client integration checks passed.");
}
finally { if (Directory.Exists(root)) Directory.Delete(root, true); }

static void Check(bool condition, string name)
{
    if (!condition) throw new Exception(name);
    Console.WriteLine("PASS: " + name);
}

sealed class FakeManager : IAsyncDisposable
{
    readonly HttpListener _listener = new();
    readonly Task _loop;
    public Uri Address { get; }
    public int Registrations, Unregistrations, Starts, Stops;
    public bool Unauthorized, FailStart, HoldStart;
    public string CallbackUrl = "";
    public JsonElement? Config;
    public TaskCompletionSource StartEntered = new(TaskCreationOptions.RunContinuationsAsynchronously);
    public TaskCompletionSource ReleaseStart = new(TaskCreationOptions.RunContinuationsAsynchronously);

    public FakeManager()
    {
        var port = new TcpListener(IPAddress.Loopback, 0);
        port.Start();
        Address = new Uri($"http://127.0.0.1:{((IPEndPoint)port.LocalEndpoint).Port}/");
        port.Stop();
        _listener.Prefixes.Add(Address.ToString());
        _listener.Start();
        _loop = RunAsync();
    }

    async Task RunAsync()
    {
        while (_listener.IsListening)
        {
            HttpListenerContext context;
            try { context = await _listener.GetContextAsync(); }
            catch (Exception ex) when (ex is HttpListenerException or ObjectDisposedException) { break; }
            await HandleAsync(context);
        }
    }

    async Task HandleAsync(HttpListenerContext ctx)
    {
        if (ctx.Request.Headers["X-Robot-Token"] is not null || ctx.Request.Headers["Authorization"] is not null)
            throw new Exception("Client must not send authentication headers");
        object body = new { ok = true };
        if (Unauthorized) ctx.Response.StatusCode = 401;
        else switch (ctx.Request.Url!.AbsolutePath)
        {
            case "/v1/health":
                body = new { ok = true, service = "CVX.Manager", version = "0.2.0", protocolVersion = 2, processId = 123 };
                break;
            case "/v1/client" when ctx.Request.HttpMethod == "PUT":
                Registrations++;
                using (var doc = await JsonDocument.ParseAsync(ctx.Request.InputStream))
                {
                    var r = doc.RootElement;
                    CallbackUrl = r.GetProperty("callbackUrl").GetString()!;
                    body = new { ok = true, clientId = r.GetProperty("clientId").GetString(), instanceId = r.GetProperty("instanceId").GetString(), callbackReady = true };
                }
                break;
            case "/v1/client": Unregistrations++; break;
            case "/v1/config":
                using (var doc = await JsonDocument.ParseAsync(ctx.Request.InputStream)) Config = doc.RootElement.Clone();
                break;
            case "/v1/start":
                Starts++;
                if (HoldStart) { StartEntered.TrySetResult(); await ReleaseStart.Task; }
                if (FailStart) { ctx.Response.StatusCode = 409; break; }
                goto case "/v1/status";
            case "/v1/status":
                body = new { ok = true, wechatRunning = true, gvxApiReady = true, httpCallbackReady = true, runtimeState = "online" };
                break;
            case "/v1/stop": case "/v1/host/shutdown": Stops++; break;
            default: throw new Exception("Unexpected request: " + ctx.Request.Url);
        }
        try
        {
            var bytes = JsonSerializer.SerializeToUtf8Bytes(body);
            ctx.Response.ContentType = "application/json";
            ctx.Response.ContentLength64 = bytes.Length;
            await ctx.Response.OutputStream.WriteAsync(bytes);
        }
        catch (HttpListenerException) { /* The cancellation check intentionally disconnects. */ }
        finally { ctx.Response.Close(); }
    }

    public async ValueTask DisposeAsync()
    {
        ReleaseStart.TrySetResult();
        _listener.Close();
        await _loop;
    }
}
