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
    var versionRoot = Path.Combine(fixture.Root, "versions", "1.0.0");
    Check(File.Exists(Path.Combine(versionRoot, "Weixin", "Weixin.exe")) &&
        File.Exists(Path.Combine(versionRoot, "Weixin", "4.1.8.27", "runtime.dll")) &&
        !Directory.Exists(Path.Combine(versionRoot, "weixin", "Weixin")),
        "The official launcher and version directory are extracted directly into the component version");
    Check(host.FindOwned(fixture.Root)[0].Executable == Path.Combine(versionRoot, "Weixin", "Weixin.exe"),
        "Injection uses the manifest launcher instead of an executable in the official version directory");
    Check(status.ApiReady && !status.IsLoggedIn && host.Starts == 1, "A logged-out kernel is ready after local injection");
    host.Helpers = true;
    await runtime.ConnectAsync();
    Check(host.Starts == 1, "Healthy repeated connection reuses the process");
    await using (var competing = new WechatRobotRuntime(options with { CallbackUrl = $"http://127.0.0.1:{Ports.Free()}/api/recvMsg" }, host))
        await Expect<InvalidOperationException>(() => competing.ConnectAsync(), "A second runtime cannot race injection or replace the callback");

    host.Helpers = false;
    await SendMessage(status.CallbackUrl);
    await Until(() => messages == 1);
    host.Crash();
    await runtime.ConnectAsync();
    Check(host.Starts == 2 && host.StopAllCalls == 0, "A compatible crashed kernel recovers without closing all WeChat processes");
    await runtime.RestartAsync();
    Check(host.Starts == 3 && host.Stops == 1, "Manual restart stops and reinjects the managed process");
    await SendMessage(status.CallbackUrl);
    await Until(() => messages == 2);
    Check(messages == 2, "Message subscriptions survive recovery and restart");

    kernel.Payload = "{}";
    host.ResetPayloadOnStart = true;
    await runtime.ConnectAsync();
    Check(host.Starts == 4 && host.Stops == 2,
        "Malformed responses from a verified owned listener trigger automatic recovery");

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
        host.Helpers = true;
        var restored = await reopened.ConnectAsync();
        Check(restored.DetectedWechatVersion == "4.1.8.27", "Persisted version is available after reconnect");
        host.Helpers = false;
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
    kernel.Payload = "{}";
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
var expectedDefault = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "CVX");
Check(new WechatRobotOptions().ComponentDirectory == expectedDefault, "The default root is CVX, not Sao");
Check(WindowsWechatProcessHost.IsOwnedExecutable(fixture.Root, Path.Combine(fixture.Root, "versions", "1.0.0", "Portable", "Weixin.exe")) &&
    WindowsWechatProcessHost.IsOwnedExecutable(fixture.Root, Path.Combine(fixture.Root, "versions", "1.0.0", "Weixin", "WeChat.exe")),
    "Process ownership allows supported executable names without a hardcoded wrapper directory");
Check(!WindowsWechatProcessHost.IsOwnedExecutable(fixture.Root, Path.Combine(fixture.Root + "-old", "versions", "1.0.0", "Weixin", "Weixin.exe")) &&
    !WindowsWechatProcessHost.IsOwnedExecutable(fixture.Root, Path.Combine(fixture.Root, "versions", "1.0.0", "inject.exe")),
    "Process ownership excludes other roots and non-WeChat executables");
using (var installer = new ComponentInstaller(downloadOptions))
{
    var installed = await installer.EnsureInstalledAsync(default);
    var protectedFiles = installed.Manifest.Files.Select(file => Path.Combine(installed.Directory, file.Name))
        .Append(Path.Combine(installed.Directory, "manifest.json"))
        .ToDictionary(path => path, path => Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(path))));
    File.Delete(Path.Combine(installed.Directory, "Weixin", "Weixin.exe"));
    await installer.EnsureInstalledAsync(default);
    Check(protectedFiles.All(item => Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(item.Key))) == item.Value) &&
        File.ReadAllText(Path.Combine(installed.Directory, "Weixin", "4.1.8.27", "runtime.dll")) == "official runtime",
        "Bundle repair preserves the injector, DLL, manifest, ZIP and official runtime layout");
    Check(File.ReadAllText(Path.Combine(installed.Directory, "Weixin", "Weixin.exe")) == "official launcher",
        "A missing manifest launcher is repaired rather than replaced by a nested executable");
}
// Exercise the background lifecycle without calling ConnectAsync again.
using (var monitoredFixture = new Fixture())
await using (var monitoredKernel = new KernelServer())
{
    var monitoredHost = new FakeProcessHost(monitoredKernel) { Helpers = true };
    var monitored = new WechatRobotRuntime(options with { ComponentDirectory = monitoredFixture.Root,
        GvxApiPort = monitoredKernel.Port, CallbackUrl = $"http://127.0.0.1:{Ports.Free()}/api/recvMsg" }, monitoredHost);
    var phases = new System.Collections.Concurrent.ConcurrentQueue<WechatRuntimePhase>();
    monitored.StatusChanged += (_, value) => phases.Enqueue(value.Phase);
    try
    {
        await monitored.ConnectAsync();
        var snapshot = JsonDocument.Parse(File.ReadAllText(Path.Combine(monitoredFixture.Root, "client-runtime.json")));
        Check(snapshot.RootElement.GetProperty("process").GetProperty("id").GetInt32() == monitoredHost.ListenerProcessId(monitoredKernel.Port),
            "The listener PID is persisted even when an older helper is enumerated first");
        monitoredHost.Helpers = false;
        var portable = Path.Combine(monitoredFixture.Root, "versions", "1.0.0", "Weixin");
        Directory.Move(Path.Combine(portable, "4.1.8.27"), Path.Combine(portable, "4.1.9.0"));
        phases.Clear();
        await monitored.ConnectAsync();
        Check(monitoredHost.Starts == 1 && monitoredHost.Stops == 0 &&
            Directory.Exists(Path.Combine(portable, "4.1.9.0")) &&
            !phases.Contains(WechatRuntimePhase.Checking) && !phases.Contains(WechatRuntimePhase.Repairing),
            "Healthy communication ignores disk version changes without scanning or repairing");
        monitoredHost.Crash();
        monitoredHost.FailStopAll = true;
        await Expect<InvalidOperationException>(() => monitored.ConnectAsync(), "Failure to stop all WeChat processes aborts repair");
        Check(Directory.Exists(Path.Combine(portable, "4.1.9.0")) && monitoredHost.Starts == 1,
            "A failed stop cannot overwrite the installed tree or inject again");
        monitoredHost.FailStopAll = false;
        await Until(() => monitoredHost.Starts == 2 && monitored.Status.ApiReady, 8);
        Check(monitoredHost.StopAllCalls >= 2, "Version incompatibility closes all WeChat processes before recovery");
        Check(monitored.Status.VersionMatch == WechatVersionMatch.Matching &&
            phases.Contains(WechatRuntimePhase.Repairing) && !Directory.Exists(Path.Combine(portable, "4.1.9.0")),
            "Background checks detect version drift and repair from the local archive without credentials");
        await monitored.DisposeAsync();
        monitoredHost.Crash();
        await Task.Delay(3300);
        Check(monitoredHost.Starts == 2 && monitored.Status.Phase == WechatRuntimePhase.Disposed,
            "Disposal stops automatic recovery and publishes disposed status");
    }
    finally { await monitored.DisposeAsync(); }
}
updates.InvalidHash = false;
using (var installer = new ComponentInstaller(downloadOptions with { ComponentToken = "component-test" }))
{
    var phases = new List<WechatRuntimePhase>();
    installer.Report = (phase, _, _) => phases.Add(phase);
    File.Delete(Path.Combine(downloadRoot, "versions", "1.0.0", "wechat.zip"));
    var requests = updates.Requests;
    await installer.EnsureInstalledAsync(default);
    Check(updates.Requests == requests + 2 && phases.Contains(WechatRuntimePhase.Downloading),
        "A missing local archive is downloaded from latest and reports progress");
}
await using (var slow = new KernelServer { DelayResponse = true })
using (var launcher = new WechatLauncher(options with { GvxApiPort = slow.Port }, new FakeProcessHost(slow)))
{
    slow.Start();
    var watch = System.Diagnostics.Stopwatch.StartNew();
    Check(await launcher.ProbeAsync(default) is null && watch.Elapsed.TotalSeconds is >= 1.5 and < 3.5,
        "An unresponsive API times out asynchronously after two seconds");
    if (OperatingSystem.IsWindows())
    {
        using var tcp = new TcpListener(IPAddress.Loopback, 0);
        tcp.Start();
        Check(WindowsPortOwner.Find(((IPEndPoint)tcp.LocalEndpoint).Port) == Environment.ProcessId,
            "Windows resolves the actual loopback listener PID");
    }
}
using (var versionFixture = new Fixture())
{
    var root = Path.Combine(versionFixture.Root, "version-detection", "nested", "Weixin");
    Directory.CreateDirectory(root);
    var exe = Path.Combine(root, "Weixin.exe");
    File.WriteAllText(exe, "launcher");
    var fixedVersions = new Dictionary<string, string>();
    string? ReadVersion(string path) => fixedVersions.GetValueOrDefault(path);
    string Add(string name, string? version)
    {
        var dir = Path.Combine(root, name); Directory.CreateDirectory(dir);
        var dll = Path.Combine(dir, "Weixin.dll");
        if (version is not null) { File.WriteAllText(dll, "module"); fixedVersions[dll] = version; }
        return dll;
    }
    fixedVersions[exe] = "4.1.8.27";
    Add("4.1.8.27", "4.1.8.27");
    var newer = Add("4.1.12.55", "4.1.12.55");
    Add("4.1.99.0", null);
    Add("9.9", "9.9.0.0");
    Check(WechatVersionDetector.Detect(exe, ReadVersion) == "4.1.12.55",
        "Complete runtime DLL wins over old launcher, partial update and two-part directory");
    fixedVersions.Remove(newer);
    Check(WechatVersionDetector.Detect(exe, ReadVersion) == "4.1.8.27",
        "A complete older installation wins over unreadable newer candidates");
    fixedVersions[Path.Combine(root, "4.1.8.27", "Weixin.dll")] = "4.1.10.0";
    Check(WechatVersionDetector.Detect(exe, ReadVersion) == "4.1.10.0",
        "With no complete candidate the DLL version wins over its directory label");
    fixedVersions.Clear();
    Check(WechatVersionDetector.Detect(exe, ReadVersion) == "4.1.99.0",
        "Unreadable DLLs fall back to the highest four-part directory");
    foreach (var dir in Directory.GetDirectories(root)) Directory.Delete(dir, true);
    fixedVersions[exe] = "4.1.8.27";
    Check(WechatVersionDetector.Detect(exe, ReadVersion) == "4.1.8.27" &&
        WechatVersionDetector.Normalize(" 4.1.8.27 ") == "4.1.8.27" &&
        WechatVersionDetector.Normalize("4.1") is null,
        "Launcher fallback and whitelist normalization require four-part versions");
}
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
static async Task Until(Func<bool> condition, int seconds = 3)
{
    using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(seconds));
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
        {
            using (var writer = new StreamWriter(zip.CreateEntry("Weixin/Weixin.exe").Open())) writer.Write("official launcher");
            using (var writer = new StreamWriter(zip.CreateEntry("Weixin/4.1.8.27/runtime.dll").Open())) writer.Write("official runtime");
            // A nested executable must not override the manifest's launcher.
            using (var writer = new StreamWriter(zip.CreateEntry("Weixin/4.1.8.27/Weixin.exe").Open())) writer.Write("nested executable");
        }
        object FileRecord(string name) => new { name, sha256 = Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(Path.Combine(directory, name)))) };
        File.WriteAllText(Path.Combine(directory, "manifest.json"), JsonSerializer.Serialize(new
        {
            version = "1.0.0", supportedWechatVersions = new[] { "4.1.8.27" },
            files = new { inject = FileRecord("inject.exe"), dll = FileRecord("kernel.dll"),
                wechatBundle = new { name = "wechat.zip", sha256 = Convert.ToHexString(SHA256.HashData(File.ReadAllBytes(Path.Combine(directory, "wechat.zip")))), exePath = "Weixin/Weixin.exe" } },
            launchArgs = new { },
        }));
        var archive = Path.Combine(Root, "package.zip"); ZipFile.CreateFromDirectory(directory, archive); Package = File.ReadAllBytes(archive);
    }
    public void Dispose() => Directory.Delete(Root, true);
}
sealed class FakeProcessHost(KernelServer server) : IWechatProcessHost
{
    WechatProcess? _owned;
    public int Starts, Stops, StopAllCalls;
    public bool FailStopAll;
    public Task StopAllAsync(CancellationToken ct)
    {
        ct.ThrowIfCancellationRequested();
        StopAllCalls++;
        if (FailStopAll) throw new InvalidOperationException("test stop failure");
        Crash(); return Task.CompletedTask;
    }
    public int? ListenerProcessId(int port) => server.Running ? _owned?.Id ?? 99999 : null;
    public bool FailStart, SuppressApi, ResetPayloadOnStart, Helpers;
    public IReadOnlyList<WechatProcess> FindOwned(string root) => _owned is null ? [] : Helpers
        ? [new(_owned.Id + 1000, 0, _owned.Executable), _owned] : [_owned];
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
    internal bool DelayResponse;
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
            if (DelayResponse) await Task.Delay(3000);
            if (!listener.IsListening) break;
            var bytes = Encoding.UTF8.GetBytes(Payload);
            try
            {
                ctx.Response.ContentType = "application/json"; ctx.Response.ContentLength64 = bytes.Length;
                await ctx.Response.OutputStream.WriteAsync(bytes);
            }
            catch (Exception ex) when (ex is HttpListenerException or IOException or ObjectDisposedException) { }
            finally
            {
                try { ctx.Response.Close(); }
                catch (InvalidOperationException) when (DelayResponse || !listener.IsListening)
                {
                    try { ctx.Response.Abort(); }
                    catch (ObjectDisposedException) { /* Listener shutdown already released the request queue. */ }
                }
            }
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
