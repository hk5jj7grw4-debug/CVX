using System.IO.Compression;
using System.Net;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using CVX.Client;

using var fixture = new Fixture();
await using var kernel = new KernelServer();
var host = new FakeProcessHost(kernel);
var options = new WechatRobotOptions
{
    ComponentDirectory = fixture.Root,
    GvxApiPort = kernel.Port,
    CallbackUrl = $"http://127.0.0.1:{Ports.Free()}/api/recvMsg",
    ConnectTimeout = TimeSpan.FromSeconds(5),
    StartTimeout = TimeSpan.FromSeconds(1),
};
var runtime = new WechatRobotRuntime(options, host);
var messages = 0;
runtime.MessageReceived += (_, _) => Interlocked.Increment(ref messages);
try
{
    var status = await runtime.ConnectAsync();
    Check(status.ApiReady && !status.IsLoggedIn && host.Starts == 1, "A logged-out kernel is ready after local injection");
    await runtime.ConnectAsync();
    Check(host.Starts == 1, "Healthy repeated connection reuses the process");
    await using (var competing = new WechatRobotRuntime(options with { CallbackUrl = $"http://127.0.0.1:{Ports.Free()}/api/recvMsg" }, host))
        await Expect<InvalidOperationException>(() => competing.ConnectAsync(), "A second runtime cannot race injection or replace the callback");

    await SendMessage(status.CallbackUrl);
    await Until(() => messages == 1);
    host.Crash();
    await runtime.ConnectAsync();
    Check(host.Starts == 2, "The same runtime recovers a crashed kernel");
    await runtime.RestartAsync();
    Check(host.Starts == 3 && host.Stops == 1, "Manual restart stops and reinjects the managed process");
    await SendMessage(status.CallbackUrl);
    await Until(() => messages == 2);
    Check(messages == 2, "Message subscriptions survive recovery and restart");

    kernel.Payload = "{}";
    await Expect<InvalidOperationException>(() => runtime.ConnectAsync(), "A listening port with an unrelated JSON response is not ready");
    Check(host.Starts == 3 && host.Stops == 1, "Invalid responses do not trigger destructive recovery");
    // Restart may repair an abnormal response from a verified owned process.
    host.ResetPayloadOnStart = true;
    await runtime.RestartAsync();
    Check(host.Starts == 4, "Explicit restart repairs malformed responses from the owned process");

    host.Crash(); host.FailStart = true;
    await Expect<InvalidOperationException>(() => runtime.ConnectAsync(), "Launch failures are reported");
    host.FailStart = false;
    await runtime.ConnectAsync();
    Check(host.Starts == 5, "Retry after failure reacquires callback and ownership");

    host.Crash(); host.SuppressApi = true;
    await Expect<TimeoutException>(() => runtime.ConnectAsync(), "An injector exiting zero without an API does not count as ready");
    host.SuppressApi = false;
    await runtime.ConnectAsync();
    Check(host.Starts == 7, "A timed-out launch can be recovered in the same runtime");
    var stops = host.Stops;
    await runtime.DisposeAsync(); await runtime.DisposeAsync();
    Check(host.Stops == stops && kernel.Running, "Async disposal leaves the kernel running");
    await Expect<ObjectDisposedException>(() => runtime.ConnectAsync(), "Disposed runtimes reject new work");

    await using (var reopened = new WechatRobotRuntime(options, host))
    {
        await reopened.ConnectAsync();
        Check(host.Starts == 7, "A new runtime reuses a verified persisted process and callback");
    }
    await using (var rebound = new WechatRobotRuntime(options with { CallbackUrl = $"http://127.0.0.1:{Ports.Free()}/api/recvMsg" }, host))
    {
        await rebound.ConnectAsync();
        Check(host.Starts == 8, "A changed callback is applied by reinjecting the managed process");
    }

    host.Crash(); host.SuppressApi = true;
    var pending = new WechatRobotRuntime(options with { StartTimeout = TimeSpan.FromSeconds(30) }, host);
    var connection = pending.ConnectAsync();
    await Until(() => host.Starts == 9);
    var disposal = pending.DisposeAsync().AsTask();
    await Expect<OperationCanceledException>(() => connection, "Disposal cancels an in-flight recovery");
    await disposal.WaitAsync(TimeSpan.FromSeconds(3));
    host.SuppressApi = false; host.Crash();
    kernel.Start();
    await using (var foreign = new WechatRobotRuntime(options, host))
        await Expect<InvalidOperationException>(() => foreign.ConnectAsync(), "Online kernels without a verified owned process are not restarted");
    Check(!kernel.SawAuthorization, "Kernel requests never contain component credentials");
}
finally { await runtime.DisposeAsync(); }

await using var updates = new UpdateServer(fixture.Package);
var downloadRoot = Path.Combine(fixture.Root, "download-case");
var downloadOptions = options with { ComponentDirectory = downloadRoot, UpdateServerUrl = updates.Address.ToString() };
using (var installer = new ComponentInstaller(downloadOptions))
    await Expect<InvalidOperationException>(() => installer.EnsureInstalledAsync(default), "Missing component token blocks downloads");
Check(updates.Requests == 0, "Missing credentials are rejected before contacting the component service");
using (var installer = new ComponentInstaller(downloadOptions with { ComponentToken = "component-test" }))
{
    var installed = await installer.EnsureInstalledAsync(default);
    Check(installed.Manifest.Version == "1.0.0" && updates.Requests == 2 && updates.AllAuthorized,
        "Component metadata and archive use Bearer auth and install a verified package");
}
updates.InvalidHash = true;
using (var installer = new ComponentInstaller(downloadOptions with { ComponentDirectory = Path.Combine(fixture.Root, "bad-hash"), ComponentToken = "component-test" }))
    await Expect<InvalidOperationException>(() => installer.EnsureInstalledAsync(default), "Hash mismatch rejects installation");
using (var installer = new ComponentInstaller(downloadOptions))
{
    var requests = updates.Requests;
    await installer.EnsureInstalledAsync(default);
    Check(updates.Requests == requests, "Installed components are reused offline without a token or update check");
}
await Expect<InvalidOperationException>(() => Task.Run(() => ComponentInstaller.SafePath(fixture.Root, "../escape")),
    "Manifest paths cannot escape the component directory");
Console.WriteLine("All direct Client integration checks passed.");

static void Check(bool value, string text)
{
    if (!value) throw new Exception(text);
    Console.WriteLine("PASS: " + text);
}
static async Task Expect<T>(Func<Task> action, string text) where T : Exception
{
    try { await action(); }
    catch (T) { Console.WriteLine("PASS: " + text); return; }
    throw new Exception("Expected " + typeof(T).Name + ": " + text);
}
static async Task Until(Func<bool> condition)
{
    using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(3));
    while (!condition()) await Task.Delay(10, timeout.Token);
}
static async Task SendMessage(Uri callback)
{
    using var http = new HttpClient();
    using var response = await http.PostAsync(callback, new StringContent("{\"content\":\"hello\"}", Encoding.UTF8, "application/json"));
    response.EnsureSuccessStatusCode();
}
static class Ports
{
    internal static int Free()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0); listener.Start();
        var port = ((IPEndPoint)listener.LocalEndpoint).Port; listener.Stop(); return port;
    }
}
sealed class Fixture : IDisposable
{
    internal string Root { get; } = Path.Combine(Path.GetTempPath(), "cvx-direct-" + Guid.NewGuid().ToString("N"));
    internal byte[] Package { get; }
    internal Fixture()
    {
        var directory = Path.Combine(Root, "versions", "1.0.0"); Directory.CreateDirectory(directory);
        File.WriteAllText(Path.Combine(directory, "inject.exe"), "fake injector");
        File.WriteAllText(Path.Combine(directory, "kernel.dll"), "fake kernel");
        using (var zip = ZipFile.Open(Path.Combine(directory, "wechat.zip"), ZipArchiveMode.Create))
        using (var writer = new StreamWriter(zip.CreateEntry("Weixin/4.1.0/Weixin.exe").Open())) writer.Write("fake wechat");
        object FileRecord(string name) => new { name, sha256 = Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(Path.Combine(directory, name)))) };
        File.WriteAllText(Path.Combine(directory, "manifest.json"), JsonSerializer.Serialize(new
        {
            version = "1.0.0", supportedWechatVersions = new[] { "4.1.0" },
            files = new { inject = FileRecord("inject.exe"), dll = FileRecord("kernel.dll"),
                wechatBundle = new { name = "wechat.zip", sha256 = Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(Path.Combine(directory, "wechat.zip")))), exePath = "Weixin/4.1.0/Weixin.exe" } },
            launchArgs = new { },
        }));
        var archive = Path.Combine(Root, "package.zip"); ZipFile.CreateFromDirectory(directory, archive); Package = File.ReadAllBytes(archive);
    }
    public void Dispose() => Directory.Delete(Root, true);
}
sealed class FakeProcessHost(KernelServer server) : IWechatProcessHost
{
    WechatProcess? _owned;
    public int Starts, Stops;
    public bool FailStart, SuppressApi, ResetPayloadOnStart;
    public IReadOnlyList<WechatProcess> FindOwned(string root) => _owned is null ? [] : [_owned];
    public Task StopAsync(WechatProcess process, CancellationToken ct)
    {
        ct.ThrowIfCancellationRequested();
        if (process != _owned) throw new Exception("Wrong process stopped");
        Stops++; Crash(); return Task.CompletedTask;
    }
    public void Crash() { server.Stop(); _owned = null; }
    public IDisposable Start(string executable, string directory, string wechat, string dll, string config)
    {
        if (FailStart) throw new InvalidOperationException("test failure");
        using var document = JsonDocument.Parse(config);
        if (document.RootElement.GetProperty("http_server_port").GetInt32() != server.Port) throw new Exception("Wrong launch port");
        Starts++; _owned = new(100 + Starts, Starts, Path.GetFullPath(wechat));
        if (ResetPayloadOnStart) server.Payload = KernelServer.ValidPayload;
        if (!SuppressApi) server.Start();
        return new MemoryStream();
    }
    public int? ExitCode(IDisposable injector) => 0;
}
sealed class KernelServer : IAsyncDisposable
{
    internal const string ValidPayload = "{\"errCode\":1,\"data\":{\"status\":false}}";
    internal int Port { get; } = Ports.Free();
    internal string Payload = ValidPayload;
    internal bool SawAuthorization;
    HttpListener? _listener;
    readonly List<Task> _loops = [];
    internal bool Running => _listener is not null;
    internal void Start()
    {
        if (_listener is not null) return;
        var listener = new HttpListener(); listener.Prefixes.Add($"http://127.0.0.1:{Port}/"); listener.Start();
        _listener = listener; _loops.Add(RunAsync(listener));
    }
    async Task RunAsync(HttpListener listener)
    {
        while (listener.IsListening)
        {
            HttpListenerContext ctx;
            try { ctx = await listener.GetContextAsync(); }
            catch (Exception ex) when (ex is HttpListenerException or ObjectDisposedException) { break; }
            SawAuthorization |= ctx.Request.Headers["Authorization"] is not null;
            if (ctx.Request.Url!.AbsolutePath != "/api/check_login" || ctx.Request.HttpMethod != "POST") throw new Exception("Wrong kernel probe");
            var bytes = Encoding.UTF8.GetBytes(Payload); ctx.Response.ContentType = "application/json"; ctx.Response.ContentLength64 = bytes.Length;
            try { await ctx.Response.OutputStream.WriteAsync(bytes); }
            catch (Exception ex) when (ex is HttpListenerException or IOException or ObjectDisposedException) { }
            finally { ctx.Response.Close(); }
        }
    }
    internal void Stop() { _listener?.Close(); _listener = null; }
    public async ValueTask DisposeAsync() { Stop(); await Task.WhenAll(_loops); }
}
sealed class UpdateServer : IAsyncDisposable
{
    readonly HttpListener _listener = new();
    readonly Task _loop;
    readonly byte[] _package;
    internal Uri Address { get; } = new($"http://127.0.0.1:{Ports.Free()}/");
    internal int Requests;
    internal bool AllAuthorized = true, InvalidHash;
    internal UpdateServer(byte[] package)
    {
        _package = package; _listener.Prefixes.Add(Address.ToString()); _listener.Start(); _loop = RunAsync();
    }
    async Task RunAsync()
    {
        while (_listener.IsListening)
        {
            HttpListenerContext ctx;
            try { ctx = await _listener.GetContextAsync(); }
            catch (Exception ex) when (ex is HttpListenerException or ObjectDisposedException) { break; }
            Requests++; AllAuthorized &= ctx.Request.Headers["Authorization"] == "Bearer component-test";
            var bytes = ctx.Request.Url!.AbsolutePath == "/package.zip" ? _package : JsonSerializer.SerializeToUtf8Bytes(new
            {
                version = "1.0.0", artifact = new { download_url = new Uri(Address, "package.zip").ToString(), size = _package.Length,
                    sha256 = InvalidHash ? "bad-hash" : Convert.ToHexString(SHA256.HashData(_package)) },
            });
            ctx.Response.ContentLength64 = bytes.Length; await ctx.Response.OutputStream.WriteAsync(bytes); ctx.Response.Close();
        }
    }
    public async ValueTask DisposeAsync() { _listener.Close(); await _loop; }
}
