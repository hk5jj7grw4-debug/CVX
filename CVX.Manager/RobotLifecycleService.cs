using System.Diagnostics;
using System.IO.Compression;
using System.Net;
using System.Net.Http.Headers;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text.Json.Nodes;

namespace CVX.Manager;

public sealed record RobotComponentUpdate(
    bool UpdateAvailable,
    string Version,
    string DownloadUrl,
    string Sha256,
    long Size);

public sealed record RobotRuntimeStatus(
    bool Ok,
    string DesiredState,
    string RuntimeState,
    bool WechatRunning,
    bool GvxApiReady,
    int GvxApiPort,
    IReadOnlyList<int> WechatProcessIds,
    string? WechatVersion,
    string? ComponentVersion,
    string VersionStatus,
    IReadOnlyList<string> SupportedWechatVersions,
    bool HttpCallbackReady,
    string? LastError);

public sealed record RobotOperationProgress(
    bool Running,
    string? Operation,
    string Phase,
    string PhaseText,
    int? Progress,
    DateTimeOffset? StartedAt,
    string? Error);

public sealed class RobotLifecycleService : IDisposable
{
    const string ComponentSlug = "wechat-robot";
    static readonly string[] WechatProcessNames = ["Weixin", "WeChat"];

    readonly RobotManagerSettingsStore _settings;
    readonly RobotClientSessionStore _clients;
    readonly SemaphoreSlim _operation = new(1, 1);
    readonly object _progressGate = new();
    readonly HttpClient _downloads = new(new SocketsHttpHandler { UseProxy = false })
    {
        Timeout = TimeSpan.FromMinutes(10),
    };
    string? _lastError;
    RobotOperationProgress _progress = new(false, null, "idle", "等待操作", null, null, null);

    public RobotLifecycleService(
        RobotManagerSettingsStore settings,
        RobotClientSessionStore clients)
    {
        _settings = settings;
        _clients = clients;
    }

    string Root => _settings.Current.DataDirectory;
    string Versions => Path.Combine(Root, "versions");
    string Downloads => Path.Combine(Root, "downloads");

    public RobotOperationProgress GetOperationProgress()
    {
        lock (_progressGate) return _progress;
    }

    public async Task<RobotRuntimeStatus> GetStatusAsync(CancellationToken ct = default)
    {
        var processes = GetWechatProcesses();
        try
        {
            var apiReady = await ProbeGvxAsync(ct);
            var callbackReady = apiReady
                                && _settings.IsCurrentCallbackApplied
                                && await _clients.IsCallbackReadyAsync(
                                    _settings.Current.HttpCallbackUrl,
                                    ct);
            var installed = LoadInstalledManifest();
            string? wechatVersion = null;
            string? componentVersion = null;
            string versionStatus = "unknown";
            IReadOnlyList<string> supported = [];

            if (installed is not null)
            {
                var (directory, manifest) = installed.Value;
                componentVersion = manifest.Version;
                supported = manifest.SupportedWechatVersions;
                var exe = ResolveBundledWechatExe(directory, manifest) ?? BundledWechatExe(directory, manifest);
                wechatVersion = GetWechatVersion(exe) ?? DetectBundledWechatVersion(directory, manifest);
                versionStatus = IsSupportedWechatVersion(wechatVersion, manifest) ? "compatible" : "incompatible";
            }

            var runtime = apiReady ? "online" : processes.Count > 0 ? "starting" : "stopped";
            return new RobotRuntimeStatus(
                apiReady,
                _settings.Current.DesiredState,
                runtime,
                processes.Count > 0,
                apiReady,
                _settings.Current.GvxApiPort,
                processes.Select(process => process.Id).ToArray(),
                wechatVersion,
                componentVersion,
                versionStatus,
                supported,
                callbackReady,
                _lastError);
        }
        finally
        {
            foreach (var process in processes) process.Dispose();
        }
    }

    public Task<RobotRuntimeStatus> InitializeAsync(CancellationToken ct = default) =>
        ReconcileAsync(ct);

    public Task<RobotRuntimeStatus> ReconcileAsync(CancellationToken ct = default) =>
        RunExclusiveAsync("reconcile", "正在检查机器人状态", async token =>
        {
            Report("checking", "正在检查组件和微信版本", 5);
            await _settings.SetDesiredStateAsync("running", token);
            var installed = LoadInstalledManifest();
            var installedValid = installed is not null &&
                                 VerifyFiles(installed.Value.Directory, installed.Value.Manifest).Count == 0;
            if (!installedValid)
            {
                if (await ProbeGvxAsync(token))
                {
                    Report("stopping_wechat", "正在停止微信以修复组件", 10);
                    StopWechat();
                }
                await RepairCoreAsync(token);
            }
            else
            {
                RepairBundleIfNeeded(installed!.Value.Directory, installed.Value.Manifest);
                if (!string.IsNullOrWhiteSpace(_settings.Current.ComponentToken))
                {
                    var update = await CheckLatestCoreAsync(installed.Value.Manifest.Version, token);
                    if (update.UpdateAvailable)
                    {
                        if (await ProbeGvxAsync(token) || IsWechatRunning())
                        {
                            Report("stopping_wechat", "正在停止微信以更新组件", 10);
                            StopWechat();
                        }
                        await InstallUpdateAsync(update, token);
                    }
                }
            }

            if (!await ProbeGvxAsync(token) || !_settings.IsCurrentCallbackApplied)
                await StartCoreAsync(token);
            return await GetStatusAsync(token);
        }, ct);

    public Task<RobotRuntimeStatus> StartAsync(CancellationToken ct = default) =>
        RunExclusiveAsync("start", "正在启动微信机器人", async token =>
        {
            Report("checking", "正在检查机器人组件", 10);
            await _settings.SetDesiredStateAsync("running", token);
            if (!await ProbeGvxAsync(token) || !_settings.IsCurrentCallbackApplied)
            {
                await EnsureInstalledAsync(token);
                await StartCoreAsync(token);
            }
            return await GetStatusAsync(token);
        }, ct);

    public Task<RobotRuntimeStatus> StopAsync(CancellationToken ct = default) =>
        RunExclusiveAsync("stop", "正在停止微信机器人", async token =>
        {
            await _settings.SetDesiredStateAsync("stopped", token);
            Report("stopping_wechat", "正在停止微信和机器人组件", 40);
            StopWechat();
            await WaitUntilAsync(async () => !await ProbeGvxAsync(token), TimeSpan.FromSeconds(10), token);
            return await GetStatusAsync(token);
        }, ct);

    public Task<RobotRuntimeStatus> RestartAsync(CancellationToken ct = default) =>
        RunExclusiveAsync("restart", "正在重启微信机器人", async token =>
        {
            await _settings.SetDesiredStateAsync("running", token);
            Report("stopping_wechat", "正在停止微信", 10);
            StopWechat();
            await EnsureInstalledAsync(token);
            await StartCoreAsync(token);
            return await GetStatusAsync(token);
        }, ct);

    public Task<RobotRuntimeStatus> RepairAsync(CancellationToken ct = default) =>
        RunExclusiveAsync("repair", "正在检查并修复机器人组件", async token =>
        {
            Report("checking", "正在检查组件文件", 5);
            var wasRunning = await ProbeGvxAsync(token);
            if (wasRunning)
            {
                Report("stopping_wechat", "正在停止微信以修复组件", 10);
                StopWechat();
            }
            await RepairCoreAsync(token);
            if (wasRunning || _settings.Current.DesiredState == "running")
                await StartCoreAsync(token);
            return await GetStatusAsync(token);
        }, ct);

    public Task<RobotComponentUpdate> CheckUpdateAsync(CancellationToken ct = default) =>
        RunExclusiveAsync(
            "check_update",
            "正在检查机器人组件更新",
            token => CheckLatestCoreAsync(LoadInstalledManifest()?.Manifest.Version ?? "0.0.0", token),
            ct);

    public Task<RobotRuntimeStatus> UpdateAsync(CancellationToken ct = default) =>
        RunExclusiveAsync("update", "正在更新机器人组件", async token =>
        {
            var update = await CheckLatestCoreAsync(LoadInstalledManifest()?.Manifest.Version ?? "0.0.0", token);
            if (!update.UpdateAvailable)
                return await GetStatusAsync(token);

            var shouldRun = _settings.Current.DesiredState == "running" || await ProbeGvxAsync(token);
            if (await ProbeGvxAsync(token))
            {
                Report("stopping_wechat", "正在停止微信以更新组件", 10);
                StopWechat();
            }
            await InstallUpdateAsync(update, token);
            if (shouldRun) await StartCoreAsync(token);
            return await GetStatusAsync(token);
        }, ct);

    async Task<T> RunExclusiveAsync<T>(
        string operation,
        string initialText,
        Func<CancellationToken, Task<T>> action,
        CancellationToken ct)
    {
        await _operation.WaitAsync(ct);
        BeginOperation(operation, initialText);
        try
        {
            var result = await action(ct);
            _lastError = null;
            CompleteOperation();
            return result;
        }
        catch (Exception ex)
        {
            _lastError = ex.Message;
            FailOperation(ex.Message);
            throw;
        }
        finally
        {
            _operation.Release();
        }
    }

    void BeginOperation(string operation, string text)
    {
        lock (_progressGate)
            _progress = new(true, operation, "starting", text, 0, DateTimeOffset.UtcNow, null);
    }

    void Report(string phase, string text, int? progress = null)
    {
        lock (_progressGate)
            _progress = _progress with
            {
                Running = true,
                Phase = phase,
                PhaseText = text,
                Progress = progress is null ? null : Math.Clamp(progress.Value, 0, 100),
                Error = null,
            };
    }

    void CompleteOperation()
    {
        lock (_progressGate)
            _progress = _progress with
            {
                Running = false,
                Phase = "completed",
                PhaseText = "操作完成",
                Progress = 100,
                Error = null,
            };
    }

    void FailOperation(string error)
    {
        lock (_progressGate)
            _progress = _progress with
            {
                Running = false,
                Phase = "failed",
                PhaseText = "操作失败",
                Progress = null,
                Error = error,
            };
    }

    async Task EnsureInstalledAsync(CancellationToken ct)
    {
        var installed = LoadInstalledManifest();
        if (installed is not null && VerifyFiles(installed.Value.Directory, installed.Value.Manifest).Count == 0)
        {
            RepairBundleIfNeeded(installed.Value.Directory, installed.Value.Manifest);
            return;
        }
        await RepairCoreAsync(ct);
    }

    async Task RepairCoreAsync(CancellationToken ct)
    {
        Report("checking", "正在校验机器人组件文件", 10);
        var installed = LoadInstalledManifest();
        if (installed is not null && VerifyFiles(installed.Value.Directory, installed.Value.Manifest).Count == 0)
        {
            RepairBundleIfNeeded(installed.Value.Directory, installed.Value.Manifest);
            return;
        }

        var update = await CheckLatestCoreAsync("0.0.0", ct);
        if (string.IsNullOrWhiteSpace(update.Version) || string.IsNullOrWhiteSpace(update.DownloadUrl))
            throw new InvalidOperationException("组件服务器没有返回可安装版本");
        await InstallUpdateAsync(update, ct);
    }

    async Task StartCoreAsync(CancellationToken ct)
    {
        if (await ProbeGvxAsync(ct))
        {
            if (_settings.IsCurrentCallbackApplied) return;
            Report("rebinding_callback", "正在重新绑定客户端回调", 82);
            StopWechat();
            await WaitUntilAsync(
                async () => !await ProbeGvxAsync(ct),
                TimeSpan.FromSeconds(10),
                ct);
        }

        Report("checking", "正在准备微信运行环境", 82);
        var installed = LoadInstalledManifest()
            ?? throw new InvalidOperationException("机器人组件尚未安装");
        var (directory, manifest) = installed;
        var problems = VerifyFiles(directory, manifest);
        if (problems.Count > 0)
            throw new InvalidOperationException("组件文件校验失败：" + string.Join("；", problems));

        RepairBundleIfNeeded(directory, manifest);
        var wechat = EnsureBundleExtracted(directory, manifest);
        var inject = Path.Combine(directory, manifest.Inject.Name);
        var dll = Path.Combine(directory, manifest.Dll.Name);
        if (!File.Exists(wechat)) throw new InvalidOperationException("组件中没有找到 Weixin.exe");

        // A stale WeChat process without the local API cannot accept a second injection reliably.
        if (IsWechatRunning())
        {
            Report("stopping_wechat", "正在清理未就绪的微信进程", 84);
            StopWechat();
        }

        var info = new ProcessStartInfo
        {
            FileName = inject,
            WorkingDirectory = directory,
            UseShellExecute = false,
            CreateNoWindow = true,
            WindowStyle = ProcessWindowStyle.Hidden,
        };
        info.ArgumentList.Add(wechat);
        info.ArgumentList.Add(dll);
        info.ArgumentList.Add(BuildLaunchConfig(manifest));

        Report("starting_wechat", "正在启动微信", 88);
        using var injector = Process.Start(info) ?? throw new InvalidOperationException("注入程序启动失败");
        Report("waiting_gvx_api", "正在等待机器人接口就绪", 94);
        var deadline = DateTimeOffset.UtcNow.AddSeconds(_settings.Current.StartTimeoutSeconds);
        while (DateTimeOffset.UtcNow < deadline)
        {
            await Task.Delay(500, ct);
            if (await ProbeGvxAsync(ct))
            {
                await _settings.MarkCurrentCallbackAppliedAsync(ct);
                return;
            }
            if (injector.HasExited && injector.ExitCode != 0)
                throw new InvalidOperationException($"注入程序异常退出，代码 {injector.ExitCode}");
        }
        throw new TimeoutException($"微信已启动，但 gVx API {_settings.Current.GvxApiPort} 未在规定时间内就绪");
    }

    string BuildLaunchConfig(RobotManifest manifest)
    {
        var args = manifest.LaunchArgs.DeepClone().AsObject();
        args["recivemode"] = "http";
        args["http_server_port"] = _settings.Current.GvxApiPort;
        args["http_callback_url"] = _settings.Current.HttpCallbackUrl;
        args["usedefault"] = false;
        args["start_server_while_login"] = true;
        return args.ToJsonString();
    }

    async Task<RobotComponentUpdate> CheckLatestCoreAsync(string currentVersion, CancellationToken ct)
    {
        Report("checking", "正在检查机器人组件更新", 8);
        var settings = _settings.Current;
        var baseUrl = NormalizeBase(settings.UpdateServerUrl);
        if (baseUrl.Length == 0) throw new InvalidOperationException("组件更新服务器地址为空");
        if (string.IsNullOrWhiteSpace(settings.ComponentToken))
            throw new InvalidOperationException("组件下载 Token 为空");

        using var request = new HttpRequestMessage(HttpMethod.Get,
            $"{baseUrl}/api/v1/apps/{ComponentSlug}/updates/latest" +
            $"?current_version={Uri.EscapeDataString(currentVersion)}&channel=stable&platform=windows&arch=x64");
        request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", settings.ComponentToken.Trim());
        using var response = await _downloads.SendAsync(request, ct);
        var body = await response.Content.ReadAsStringAsync(ct);
        if (response.StatusCode == HttpStatusCode.Unauthorized)
            throw new UnauthorizedAccessException("组件 Token 无效、过期或已撤销");
        if (!response.IsSuccessStatusCode)
            throw new HttpRequestException($"组件更新检查失败 HTTP {(int)response.StatusCode}: {Shorten(body)}");

        var root = JsonNode.Parse(body)?.AsObject() ?? throw new InvalidOperationException("组件更新接口返回无效 JSON");
        var artifact = root["artifact"]?.AsObject();
        return new RobotComponentUpdate(
            Bool(root, "update_available", "updateAvailable"),
            Text(root, "version") ?? "",
            Text(artifact, "download_url", "downloadUrl") ?? "",
            Text(artifact, "sha256") ?? "",
            Long(artifact, "size"));
    }

    async Task InstallUpdateAsync(RobotComponentUpdate update, CancellationToken ct)
    {
        if (string.IsNullOrWhiteSpace(update.Sha256))
            throw new InvalidOperationException("组件更新响应缺少 SHA-256");

        Directory.CreateDirectory(Downloads);
        Directory.CreateDirectory(Versions);
        var archivePath = Path.Combine(Downloads, $"{ComponentSlug}-{update.Version}.zip");
        using (var request = new HttpRequestMessage(HttpMethod.Get, update.DownloadUrl))
        {
            Report("downloading", "正在下载机器人组件", 10);
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", _settings.Current.ComponentToken.Trim());
            using var response = await _downloads.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, ct);
            if (response.StatusCode == HttpStatusCode.Unauthorized)
                throw new UnauthorizedAccessException("组件下载 Token 无效、过期或已撤销");
            response.EnsureSuccessStatusCode();
            await using var source = await response.Content.ReadAsStreamAsync(ct);
            await using var target = new FileStream(archivePath, FileMode.Create, FileAccess.Write, FileShare.None, 81920, true);
            var total = response.Content.Headers.ContentLength ?? update.Size;
            var buffer = new byte[81920];
            long received = 0;
            while (true)
            {
                var count = await source.ReadAsync(buffer, ct);
                if (count == 0) break;
                await target.WriteAsync(buffer.AsMemory(0, count), ct);
                received += count;
                if (total > 0)
                {
                    var percent = (int)Math.Clamp(received * 100 / total, 0, 100);
                    Report("downloading", $"正在下载机器人组件 {percent}%", 10 + percent * 55 / 100);
                }
            }
        }

        Report("verifying", "正在校验机器人组件", 70);
        if (update.Size > 0 && new FileInfo(archivePath).Length != update.Size)
            throw new InvalidOperationException("组件包大小校验失败");
        if (!Hash(archivePath).Equals(update.Sha256.Trim(), StringComparison.OrdinalIgnoreCase))
            throw new InvalidOperationException("组件包 SHA-256 校验失败");

        var staging = Path.Combine(Versions, update.Version + ".staging");
        var targetDirectory = Path.Combine(Versions, update.Version);
        Report("extracting", "正在安装机器人组件", 78);
        if (Directory.Exists(staging)) Directory.Delete(staging, true);
        Directory.CreateDirectory(staging);
        ZipFile.ExtractToDirectory(archivePath, staging);

        var manifestPath = Path.Combine(staging, "manifest.json");
        if (!File.Exists(manifestPath)) throw new InvalidOperationException("组件包缺少 manifest.json");
        var manifest = RobotManifest.Load(manifestPath);
        var problems = VerifyFiles(staging, manifest);
        if (problems.Count > 0) throw new InvalidOperationException("组件文件校验失败：" + string.Join("；", problems));
        ExtractBundle(staging, manifest);

        if (Directory.Exists(targetDirectory)) Directory.Delete(targetDirectory, true);
        Directory.Move(staging, targetDirectory);
        File.Delete(archivePath);
    }

    (string Directory, RobotManifest Manifest)? LoadInstalledManifest()
    {
        if (!Directory.Exists(Versions)) return null;
        (string Directory, RobotManifest Manifest)? best = null;
        foreach (var directory in Directory.GetDirectories(Versions))
        {
            var path = Path.Combine(directory, "manifest.json");
            if (!File.Exists(path)) continue;
            try
            {
                var manifest = RobotManifest.Load(path);
                if (best is null || CompareVersions(manifest.Version, best.Value.Manifest.Version) > 0)
                    best = (directory, manifest);
            }
            catch { }
        }
        return best;
    }

    static IReadOnlyList<string> VerifyFiles(string directory, RobotManifest manifest)
    {
        var problems = new List<string>();
        foreach (var file in manifest.Files)
        {
            var path = Path.Combine(directory, file.Name);
            if (!File.Exists(path)) problems.Add($"缺失 {file.Name}");
            else if (!Hash(path).Equals(file.Sha256, StringComparison.OrdinalIgnoreCase))
                problems.Add($"{file.Name} 哈希不符");
        }
        return problems;
    }

    async Task<bool> ProbeGvxAsync(CancellationToken ct)
    {
        try
        {
            using var timeout = CancellationTokenSource.CreateLinkedTokenSource(ct);
            timeout.CancelAfter(TimeSpan.FromSeconds(1));
            using var tcp = new TcpClient();
            await tcp.ConnectAsync(IPAddress.Loopback, _settings.Current.GvxApiPort, timeout.Token);
            return true;
        }
        catch (Exception ex) when (ex is SocketException or OperationCanceledException)
        {
            return false;
        }
    }

    static List<Process> GetWechatProcesses() => WechatProcessNames
        .SelectMany(Process.GetProcessesByName)
        .GroupBy(process => process.Id)
        .Select(group => group.First())
        .ToList();

    static bool IsWechatRunning()
    {
        var processes = GetWechatProcesses();
        try { return processes.Count > 0; }
        finally
        {
            foreach (var process in processes) process.Dispose();
        }
    }

    static void StopWechat()
    {
        foreach (var process in GetWechatProcesses())
        {
            try
            {
                process.Kill(true);
                process.WaitForExit(5000);
            }
            catch { }
            finally { process.Dispose(); }
        }
    }

    static void ExtractBundle(string directory, RobotManifest manifest)
    {
        var archive = Path.Combine(directory, manifest.WechatBundle.Name);
        if (!File.Exists(archive)) throw new InvalidOperationException($"缺少便携微信包 {manifest.WechatBundle.Name}");
        var target = Path.Combine(directory, "weixin");
        if (Directory.Exists(target)) Directory.Delete(target, true);
        Directory.CreateDirectory(target);
        ZipFile.ExtractToDirectory(archive, target);
    }

    static string EnsureBundleExtracted(string directory, RobotManifest manifest)
    {
        var exe = ResolveBundledWechatExe(directory, manifest) ?? BundledWechatExe(directory, manifest);
        if (File.Exists(exe)) return exe;
        ExtractBundle(directory, manifest);
        return ResolveBundledWechatExe(directory, manifest) ?? BundledWechatExe(directory, manifest);
    }

    static void RepairBundleIfNeeded(string directory, RobotManifest manifest)
    {
        var exe = EnsureBundleExtracted(directory, manifest);
        var version = GetWechatVersion(exe) ?? DetectBundledWechatVersion(directory, manifest);
        if (IsSupportedWechatVersion(version, manifest)) return;
        ExtractBundle(directory, manifest);
        exe = EnsureBundleExtracted(directory, manifest);
        version = GetWechatVersion(exe) ?? DetectBundledWechatVersion(directory, manifest);
        if (!IsSupportedWechatVersion(version, manifest))
            throw new InvalidOperationException($"便携微信版本 {version ?? "未知"} 与组件不兼容");
    }

    static string BundledWechatExe(string directory, RobotManifest manifest) =>
        Path.Combine(directory, "weixin", manifest.WechatExePath.Replace('/', Path.DirectorySeparatorChar));

    static string? ResolveBundledWechatExe(string directory, RobotManifest manifest)
    {
        var parts = manifest.WechatExePath.Split(['/', '\\'], StringSplitOptions.RemoveEmptyEntries);
        if (parts.Length >= 2)
        {
            var root = Path.Combine(directory, "weixin", parts[0]);
            if (Directory.Exists(root))
            {
                return Directory.GetDirectories(root)
                    .Select(path => (Path: path, Version: Path.GetFileName(path)))
                    .Where(item => Version.TryParse(item.Version, out _))
                    .OrderByDescending(item => Version.Parse(item.Version))
                    .Select(item => Path.Combine(item.Path, Path.GetFileName(manifest.WechatExePath)))
                    .FirstOrDefault(File.Exists);
            }
        }
        var exact = BundledWechatExe(directory, manifest);
        return File.Exists(exact) ? exact : null;
    }

    static string? DetectBundledWechatVersion(string directory, RobotManifest manifest)
    {
        var first = manifest.WechatExePath.Split('/', '\\')[0];
        var root = Path.Combine(directory, "weixin", first);
        if (!Directory.Exists(root)) return null;
        return Directory.GetDirectories(root)
            .Select(Path.GetFileName)
            .Where(value => Version.TryParse(value, out _))
            .OrderByDescending(value => Version.Parse(value!))
            .FirstOrDefault();
    }

    static string? GetWechatVersion(string path)
    {
        if (!File.Exists(path)) return null;
        try
        {
            var info = FileVersionInfo.GetVersionInfo(path);
            return (info.FileVersion ?? info.ProductVersion)?.Trim();
        }
        catch { return null; }
    }

    static bool IsSupportedWechatVersion(string? version, RobotManifest manifest) =>
        string.IsNullOrWhiteSpace(version)
            ? manifest.SupportedWechatVersions.Count == 0
            : manifest.SupportedWechatVersions.Count == 0 ||
              manifest.SupportedWechatVersions.Contains(version, StringComparer.OrdinalIgnoreCase) ||
              version.Equals(manifest.Version, StringComparison.OrdinalIgnoreCase);

    static async Task WaitUntilAsync(Func<Task<bool>> condition, TimeSpan timeout, CancellationToken ct)
    {
        var deadline = DateTimeOffset.UtcNow + timeout;
        while (DateTimeOffset.UtcNow < deadline)
        {
            if (await condition()) return;
            await Task.Delay(250, ct);
        }
    }

    static string Hash(string path)
    {
        using var stream = File.OpenRead(path);
        return Convert.ToHexString(SHA256.HashData(stream)).ToLowerInvariant();
    }

    static int CompareVersions(string left, string right) =>
        Version.TryParse(left, out var l) && Version.TryParse(right, out var r)
            ? l.CompareTo(r)
            : string.Compare(left, right, StringComparison.OrdinalIgnoreCase);

    static string NormalizeBase(string value)
    {
        var result = value.Trim();
        if (result.Length > 0 && !result.StartsWith("http://", StringComparison.OrdinalIgnoreCase) &&
            !result.StartsWith("https://", StringComparison.OrdinalIgnoreCase)) result = "http://" + result;
        return result.TrimEnd('/');
    }

    static string Shorten(string value) => string.IsNullOrWhiteSpace(value)
        ? "(empty)"
        : value.Trim()[..Math.Min(180, value.Trim().Length)];

    static string? Text(JsonObject? node, params string[] names)
    {
        if (node is null) return null;
        foreach (var name in names)
            if (node[name] is JsonValue value && value.TryGetValue<string>(out var text)) return text;
        return null;
    }

    static bool Bool(JsonObject node, params string[] names)
    {
        foreach (var name in names)
            if (node[name] is JsonValue value && value.TryGetValue<bool>(out var result)) return result;
        return false;
    }

    static long Long(JsonObject? node, string name) =>
        node?[name] is JsonValue value && value.TryGetValue<long>(out var result) ? result : 0;

    public void Dispose()
    {
        _downloads.Dispose();
        _operation.Dispose();
    }
}

public sealed class RobotAutoStartService : BackgroundService
{
    readonly RobotManagerSettingsStore _settings;
    readonly RobotLifecycleService _robot;
    readonly ILogger<RobotAutoStartService> _logger;

    public RobotAutoStartService(
        RobotManagerSettingsStore settings,
        RobotLifecycleService robot,
        ILogger<RobotAutoStartService> logger)
    {
        _settings = settings;
        _robot = robot;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        if (!_settings.Current.AutoStart || _settings.Current.DesiredState != "running") return;
        await TryInitializeAsync(stoppingToken);

        using var timer = new PeriodicTimer(TimeSpan.FromSeconds(15));
        while (await timer.WaitForNextTickAsync(stoppingToken))
        {
            if (!_settings.Current.AutoStart || _settings.Current.DesiredState != "running") continue;
            var status = await _robot.GetStatusAsync(stoppingToken);
            if (!status.WechatRunning && !status.GvxApiReady)
                await TryInitializeAsync(stoppingToken);
        }
    }

    async Task TryInitializeAsync(CancellationToken ct)
    {
        try
        {
            await _robot.InitializeAsync(ct);
        }
        catch (Exception ex) when (!ct.IsCancellationRequested)
        {
            _logger.LogError(ex, "机器人自动初始化失败");
        }
    }
}
