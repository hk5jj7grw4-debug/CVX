using System.Diagnostics;
using System.Text.Json;

namespace CVX.Client;

internal sealed record WechatProcess(int Id, long StartedAtUtcTicks, string Executable);
internal sealed record AppliedRuntime(WechatProcess Process, string CallbackUrl, int ApiPort, string ComponentVersion, string? WechatVersion = null, WechatVersionMatch VersionMatch = WechatVersionMatch.Unknown);
internal sealed record KernelHealth(bool IsLoggedIn);

// Isolates Windows process operations so recovery can be tested without injection.
internal interface IWechatProcessHost
{
    int? ListenerProcessId(int port);
    IReadOnlyList<WechatProcess> FindOwned(string componentDirectory);
    Task StopAsync(WechatProcess process, CancellationToken ct);
    Task StopAllAsync(CancellationToken ct);
    IDisposable Start(string executable, string directory, string wechat, string dll, string config);
    int? ExitCode(IDisposable injector);
}

internal sealed class WindowsWechatProcessHost : IWechatProcessHost
{
    public int? ListenerProcessId(int port) => WindowsPortOwner.Find(port);

    public IReadOnlyList<WechatProcess> FindOwned(string componentDirectory)
    {
        if (!OperatingSystem.IsWindows()) return [];
        var result = new List<WechatProcess>();
        foreach (var name in new[] { "Weixin", "WeChat" })
        foreach (var process in Process.GetProcessesByName(name))
        {
            using (process)
            {
                try
                {
                    var exe = process.MainModule?.FileName;
                    if (exe is null || !IsOwnedExecutable(componentDirectory, exe)) continue;
                    result.Add(new(process.Id, process.StartTime.ToUniversalTime().Ticks, Path.GetFullPath(exe)));
                }
                catch (Exception ex) when (ex is InvalidOperationException or System.ComponentModel.Win32Exception) { }
            }
        }
        return result;
    }

    internal static bool IsOwnedExecutable(string componentDirectory, string executable)
    {
        var versions = Path.GetFullPath(Path.Combine(componentDirectory, "versions")) + Path.DirectorySeparatorChar;
        var path = Path.GetFullPath(executable);
        if (!path.StartsWith(versions, StringComparison.OrdinalIgnoreCase)) return false;
        var name = Path.GetFileName(path);
        return name.Equals("Weixin.exe", StringComparison.OrdinalIgnoreCase)
            || name.Equals("WeChat.exe", StringComparison.OrdinalIgnoreCase);
    }

    public async Task StopAllAsync(CancellationToken ct)
    {
        if (!OperatingSystem.IsWindows()) throw new PlatformNotSupportedException("关闭微信仅支持 Windows");
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(ct);
        timeout.CancelAfter(TimeSpan.FromSeconds(10));
        while (true)
        {
            timeout.Token.ThrowIfCancellationRequested();
            var found = false;
            foreach (var name in new[] { "Weixin", "WeChat", "WeixinExt", "WeixinUpdate", "WeChatAppEx" })
            {
                var processes = Process.GetProcessesByName(name);
                try
                {
                    foreach (var process in processes)
                    {
                        try
                        {
                            if (process.HasExited) continue;
                            found = true;
                            var exe = process.MainModule?.FileName
                                ?? throw new InvalidOperationException("无法识别微信进程路径，已停止修复");
                            await StopAsync(new(process.Id, process.StartTime.ToUniversalTime().Ticks, exe), timeout.Token).ConfigureAwait(false);
                        }
                        catch (InvalidOperationException) when (process.HasExited) { }
                        catch (System.ComponentModel.Win32Exception) when (process.HasExited) { }
                    }
                }
                finally { foreach (var process in processes) process.Dispose(); }
            }
            if (!found) return;
            await Task.Delay(100, timeout.Token).ConfigureAwait(false);
        }
    }

    public async Task StopAsync(WechatProcess owned, CancellationToken ct)
    {
        if (!OperatingSystem.IsWindows()) throw new PlatformNotSupportedException("微信启动和重启仅支持 Windows");
        Process process;
        try { process = Process.GetProcessById(owned.Id); }
        catch (ArgumentException) { return; }
        using (process)
        {
            if (process.HasExited) return;
            if (process.StartTime.ToUniversalTime().Ticks != owned.StartedAtUtcTicks ||
                !string.Equals(process.MainModule?.FileName, owned.Executable, StringComparison.OrdinalIgnoreCase))
                throw new InvalidOperationException("微信进程身份已变化，拒绝停止其他进程");
            process.Kill(entireProcessTree: true);
            using var timeout = CancellationTokenSource.CreateLinkedTokenSource(ct);
            timeout.CancelAfter(TimeSpan.FromSeconds(10));
            await process.WaitForExitAsync(timeout.Token).ConfigureAwait(false);
        }
    }

    public IDisposable Start(string executable, string directory, string wechat, string dll, string config)
    {
        if (!OperatingSystem.IsWindows()) throw new PlatformNotSupportedException("微信启动和注入仅支持 Windows");
        var info = new ProcessStartInfo(executable)
        {
            WorkingDirectory = directory, UseShellExecute = false, CreateNoWindow = true,
            WindowStyle = ProcessWindowStyle.Hidden,
        };
        info.ArgumentList.Add(wechat);
        info.ArgumentList.Add(dll);
        info.ArgumentList.Add(config);
        return Process.Start(info) ?? throw new InvalidOperationException("注入程序启动失败");
    }

    public int? ExitCode(IDisposable injector) => ((Process)injector).HasExited ? ((Process)injector).ExitCode : null;
}

internal sealed class WechatLauncher : IDisposable
{
    internal Action<WechatRuntimePhase, string?, string?, double?>? Report { get; set; }
    readonly WechatRobotOptions _options;
    readonly ComponentInstaller _installer;
    readonly IWechatProcessHost _processes;
    readonly HttpClient _http = new(new SocketsHttpHandler { UseProxy = false });
    static readonly JsonSerializerOptions JsonOptions = new(JsonSerializerDefaults.Web);
    string StatePath => Path.Combine(_options.ComponentDirectory, "client-runtime.json");

    internal WechatLauncher(WechatRobotOptions options, IWechatProcessHost? processes = null)
    {
        _options = options;
        _installer = new(options);
        _installer.Report = (phase, version, progress) => Report?.Invoke(phase, version, null, progress);
        _processes = processes ?? new WindowsWechatProcessHost();
    }

    internal async Task<WechatRobotStatus> ConnectAsync(Uri callback, bool restart, CancellationToken ct)
    {
        var owned = _processes.FindOwned(_options.ComponentDirectory);

        KernelHealth? health;
        try { health = await ProbeAsync(ct).ConfigureAwait(false); }
        catch (InvalidOperationException)
        {
            // Protocol/HTTP failures enter the same disk check as a timeout.
            // Listener ownership is verified below before any process is stopped.
            health = null;
        }
        owned = _processes.FindOwned(_options.ComponentDirectory);
        var applied = await ReadStateAsync(ct).ConfigureAwait(false);
        var listener = _processes.ListenerProcessId(_options.GvxApiPort);
        var managed = owned.FirstOrDefault(p => p.Id == listener);
        if (!restart && health is not null && managed is not null && applied is not null &&
            applied.Process == managed && applied.ApiPort == _options.GvxApiPort &&
            string.Equals(applied.CallbackUrl, callback.AbsoluteUri, StringComparison.Ordinal))
            return new(true, health.IsLoggedIn, applied.ComponentVersion, callback)
            {
                DetectedWechatVersion = applied.WechatVersion,
                VersionMatch = applied.VersionMatch
            };

        var inspection = _installer.InspectVersion();
        Report?.Invoke(WechatRuntimePhase.Checking, inspection.Version, inspection.ComponentVersion, null);
        if (listener is not null && managed is null)
            throw new InvalidOperationException("内核端口被非受管进程占用");
        var repairVersion = inspection.Compatible == false;
        if (!repairVersion && owned.Select(p => p.Executable).Distinct(StringComparer.OrdinalIgnoreCase).Count() > 1)
            throw new InvalidOperationException("组件目录中有不同路径的微信实例，无法确定恢复目标");
        if (health is not null && managed is null)
            throw new InvalidOperationException("内核接口已在线，但无法确认其受管微信进程；请关闭旧实例后重试");
        if (repairVersion)
        {
            Report?.Invoke(WechatRuntimePhase.Repairing, inspection.Version, inspection.ComponentVersion, null);
            await _processes.StopAllAsync(ct).ConfigureAwait(false);
        }
        else foreach (var process in owned) await _processes.StopAsync(process, ct).ConfigureAwait(false);
        if (repairVersion || owned.Count > 0)
        {
            var deadline = DateTimeOffset.UtcNow.AddSeconds(10);
            while (await ProbeAsync(ct).ConfigureAwait(false) is not null)
            {
                if (DateTimeOffset.UtcNow >= deadline) throw new TimeoutException("受管微信停止后内核接口仍在线");
                await Task.Delay(250, ct).ConfigureAwait(false);
            }
        }
        if (File.Exists(StatePath)) File.Delete(StatePath);
        var installed = await _installer.EnsureInstalledAsync(ct).ConfigureAwait(false);
        var manifest = installed.Manifest;
        var wechat = ComponentInstaller.EnsureBundleExtracted(installed.Directory, manifest);
        var verified = _installer.InspectVersion();
        Report?.Invoke(WechatRuntimePhase.Connecting, verified.Version, manifest.Version, null);
        var config = manifest.LaunchArgs.DeepClone().AsObject();
        config["recivemode"] = "http";
        config["http_server_port"] = _options.GvxApiPort;
        config["http_callback_url"] = callback.AbsoluteUri;
        config["usedefault"] = false;
        config["start_server_while_login"] = true;
        using var injector = _processes.Start(
            ComponentInstaller.SafePath(installed.Directory, manifest.Inject.Name), installed.Directory,
            wechat, ComponentInstaller.SafePath(installed.Directory, manifest.Dll.Name), config.ToJsonString());
        var startDeadline = DateTimeOffset.UtcNow.Add(_options.StartTimeout);
        while (DateTimeOffset.UtcNow < startDeadline)
        {
            ct.ThrowIfCancellationRequested();
            if (_processes.ExitCode(injector) is { } code && code != 0)
                throw new InvalidOperationException($"注入程序异常退出，代码 {code}");
            health = await ProbeAsync(ct).ConfigureAwait(false);
            if (health is not null)
            {
                owned = _processes.FindOwned(_options.ComponentDirectory);
                var injected = owned.FirstOrDefault(p => p.Id == _processes.ListenerProcessId(_options.GvxApiPort));
                if (injected is null || !string.Equals(injected.Executable, Path.GetFullPath(wechat), StringComparison.OrdinalIgnoreCase))
                    throw new InvalidOperationException("内核响应正常，但微信进程身份与本次启动不匹配");
                var state = new AppliedRuntime(injected, callback.AbsoluteUri, _options.GvxApiPort, manifest.Version, verified.Version, verified.Version is null ? WechatVersionMatch.Unknown : WechatVersionMatch.Matching);
                var temp = StatePath + ".tmp";
                try
                {
                    await File.WriteAllTextAsync(temp, JsonSerializer.Serialize(state, JsonOptions), ct).ConfigureAwait(false);
                    File.Move(temp, StatePath, overwrite: true);
                }
                finally { if (File.Exists(temp)) File.Delete(temp); }
                return new(true, health.IsLoggedIn, manifest.Version, callback) { DetectedWechatVersion = verified.Version, VersionMatch = state.VersionMatch };
            }
            await Task.Delay(250, ct).ConfigureAwait(false);
        }
        throw new TimeoutException("微信内核未在规定时间内就绪；可再次连接或手动重启");
    }

    internal async Task<KernelHealth?> ProbeAsync(CancellationToken ct)
    {
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(ct);
        timeout.CancelAfter(TimeSpan.FromSeconds(2));
        try
        {
            using var response = await _http.PostAsync($"http://127.0.0.1:{_options.GvxApiPort}/api/check_login", null, timeout.Token).ConfigureAwait(false);
            if (!response.IsSuccessStatusCode) throw new InvalidOperationException($"内核探测返回 HTTP {(int)response.StatusCode}，未执行重启");
            using var document = await JsonDocument.ParseAsync(await response.Content.ReadAsStreamAsync(timeout.Token).ConfigureAwait(false), cancellationToken: timeout.Token).ConfigureAwait(false);
            var root = document.RootElement;
            if (root.ValueKind != JsonValueKind.Object || !root.TryGetProperty("errCode", out var error) ||
                !int.TryParse(error.ToString(), out var code) || code is not (0 or 1) ||
                !root.TryGetProperty("data", out var data) || data.ValueKind != JsonValueKind.Object ||
                !data.TryGetProperty("status", out var status) || status.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
                throw new InvalidOperationException("端口响应不符合内核协议，未执行重启");
            return new(status.GetBoolean());
        }
        catch (JsonException ex) { throw new InvalidOperationException("端口响应不是有效的内核 JSON，未执行重启", ex); }
        catch (HttpRequestException ex) when (ex.StatusCode is null) { return null; }
        catch (OperationCanceledException) when (!ct.IsCancellationRequested) { return null; }
    }

    async Task<AppliedRuntime?> ReadStateAsync(CancellationToken ct)
    {
        if (!File.Exists(StatePath)) return null;
        try { return JsonSerializer.Deserialize<AppliedRuntime>(await File.ReadAllTextAsync(StatePath, ct).ConfigureAwait(false), JsonOptions); }
        catch (JsonException) { return null; }
    }

    public void Dispose() { _http.Dispose(); _installer.Dispose(); }
}
