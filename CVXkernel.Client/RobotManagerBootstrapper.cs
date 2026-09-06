using System.Diagnostics;
using System.IO.Compression;
using System.Net;
using System.Net.Http.Headers;
using System.Reflection;
using System.Security.Cryptography;
using System.Text.Json;

namespace CVXkernel.Client;

public sealed record RobotManagerBootstrapperOptions
{
    public const int CurrentProtocolVersion = 2;
    public const string ExpectedServiceName = "Sao.WechatRobotManager";

    public Uri ManagerApiBaseAddress { get; init; } = RobotManagerClient.DefaultBaseAddress;
    public string RootDirectory { get; init; } = Path.Combine(
        Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
        "Sao",
        "WechatRobotManagerHost");
    public string DownloadToken { get; init; } = "";
    public string ApiToken { get; init; } = "";
    public string AppSlug { get; init; } = "wechat-robot-manager";
    public string ClientId { get; init; } = Assembly.GetEntryAssembly()?.GetName().Name
                                             ?? "cvxkernel-client";
    public string ExecutableName { get; init; } = "Sao.WechatRobotManager.exe";
    public string Channel { get; init; } = "stable";
    public string Platform { get; init; } = "windows";
    public string Architecture { get; init; } = "x64";
    public string ArtifactType { get; init; } = "portable";
    public TimeSpan HealthProbeTimeout { get; init; } = TimeSpan.FromSeconds(2);
    public TimeSpan ManagerStartTimeout { get; init; } = TimeSpan.FromSeconds(20);
    public TimeSpan ClientReadyTimeout { get; init; } = TimeSpan.FromSeconds(45);
    public TimeSpan DownloadTimeout { get; init; } = TimeSpan.FromMinutes(10);
    public bool DisableDownloadProxy { get; init; } = true;

    internal string ResolveDownloadToken()
    {
        if (!string.IsNullOrWhiteSpace(DownloadToken)) return DownloadToken.Trim();
        return Assembly.GetEntryAssembly()?
                   .GetCustomAttributes<AssemblyMetadataAttribute>()
                   .FirstOrDefault(attribute => attribute.Key == "ReleaseHubRobotManagerToken")?
                   .Value?.Trim() ?? "";
    }
}

/// <summary>
/// Installs, updates, starts, and configures the lifecycle manager. It does not
/// proxy or wrap the GVx business API.
/// </summary>
public sealed class RobotManagerBootstrapper : IDisposable
{
    readonly RobotManagerBootstrapperOptions _options;
    readonly RobotManagerClient _manager;
    readonly HttpClient _downloads;
    readonly JsonSerializerOptions _json = new(JsonSerializerDefaults.Web)
    {
        PropertyNameCaseInsensitive = true,
    };
    readonly SemaphoreSlim _ensureGate = new(1, 1);

    string VersionsDirectory => Path.Combine(_options.RootDirectory, "versions");
    string DownloadsDirectory => Path.Combine(_options.RootDirectory, "downloads");

    public RobotManagerBootstrapper(RobotManagerBootstrapperOptions? options = null)
    {
        _options = options ?? new RobotManagerBootstrapperOptions();
        ValidateOptions(_options);
        _manager = new RobotManagerClient(
            _options.ManagerApiBaseAddress,
            apiToken: _options.ApiToken);
        _downloads = _options.DisableDownloadProxy
            ? new HttpClient(new SocketsHttpHandler { UseProxy = false })
            : new HttpClient();
        _downloads.Timeout = _options.DownloadTimeout;
    }

    public IRobotManagerClient Client => _manager;

    public async Task<RobotManagerStatus> EnsureRunningAsync(
        RobotManagerConfig config,
        IProgress<int>? progress = null,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(config);
        await _ensureGate.WaitAsync(cancellationToken);
        try
        {
            await EnsureManagerRunningCoreAsync(config.UpdateServerUrl, progress, cancellationToken);
            var effectiveConfig = config with { ApiToken = _options.ApiToken.Trim() };
            await _manager.ConfigureAsync(effectiveConfig, cancellationToken);
            return await _manager.ReconcileAsync(cancellationToken);
        }
        finally
        {
            _ensureGate.Release();
        }
    }

    internal async Task<RobotManagerStatus> ConnectClientAsync(
        RobotManagerConfig config,
        RobotClientRegistration registration,
        IProgress<int>? progress,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(config);
        ArgumentNullException.ThrowIfNull(registration);
        await _ensureGate.WaitAsync(cancellationToken);
        var registered = false;
        try
        {
            await EnsureManagerRunningCoreAsync(config.UpdateServerUrl, progress, cancellationToken);
            var session = await _manager.RegisterClientAsync(registration, cancellationToken);
            if (!session.Ok
                || !string.Equals(session.ClientId, registration.ClientId, StringComparison.Ordinal)
                || !string.Equals(session.InstanceId, registration.InstanceId, StringComparison.Ordinal))
                throw new InvalidOperationException("RobotManager 没有确认当前客户端实例");
            registered = true;

            var effectiveConfig = config with
            {
                ApiToken = _options.ApiToken.Trim(),
                HttpCallbackUrl = registration.CallbackUrl,
            };
            await _manager.ConfigureAsync(effectiveConfig, cancellationToken);
            return await _manager.ReconcileAsync(cancellationToken);
        }
        catch
        {
            if (registered)
            {
                using var cleanupTimeout = new CancellationTokenSource(TimeSpan.FromSeconds(2));
                try
                {
                    await _manager.UnregisterClientAsync(
                        registration.ClientId,
                        registration.InstanceId,
                        cleanupTimeout.Token);
                }
                catch
                {
                    // Preserve the original connect error if cleanup cannot reach the manager.
                }
            }
            throw;
        }
        finally
        {
            _ensureGate.Release();
        }
    }

    async Task EnsureManagerRunningCoreAsync(
        string updateServerUrl,
        IProgress<int>? progress,
        CancellationToken cancellationToken)
    {
        var health = await TryGetHealthAsync(cancellationToken);
        if (health is not null) ValidateManagerIdentity(health);

        var installed = FindInstalledVersion();
        var currentVersion = health?.Version ?? installed?.Version ?? "0.0.0";
        var token = _options.ResolveDownloadToken();
        LatestUpdateResponse? update = null;

        if (token.Length > 0)
            update = await CheckLatestAsync(updateServerUrl, currentVersion, token, cancellationToken);

        if (update?.UpdateAvailable == true)
        {
            if (string.IsNullOrWhiteSpace(update.Version) || update.Artifact is null)
                throw new InvalidOperationException("RobotManager 更新响应缺少版本或产物信息");
            installed = await DownloadAndInstallAsync(update, token, progress, cancellationToken);
            if (health is not null)
            {
                await StopManagerHostAsync(cancellationToken);
                health = null;
            }
        }
        else if (health is null && installed is null)
        {
            var reason = token.Length == 0 ? "未提供下载 Token" : "更新服务器没有返回可安装版本";
            throw new InvalidOperationException($"RobotManager 尚未安装，且{reason}");
        }

        if (health is not null) ValidateManagerProtocol(health);

        if (health is null)
        {
            installed ??= FindInstalledVersion()
                ?? throw new InvalidOperationException("RobotManager 尚未安装");
            StartManager(installed.Value.Directory);
            health = await WaitForHealthAsync(cancellationToken);
            ValidateManagerIdentity(health);
            ValidateManagerProtocol(health);
        }
    }

    public string? GetInstalledVersion() => FindInstalledVersion()?.Version;

    async Task<RobotManagerHealth?> TryGetHealthAsync(CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            timeout.CancelAfter(_options.HealthProbeTimeout);
            return await _manager.GetHealthAsync(timeout.Token);
        }
        catch (Exception exception) when (
            exception is HttpRequestException or OperationCanceledException
            && !cancellationToken.IsCancellationRequested)
        {
            return null;
        }
    }

    async Task<LatestUpdateResponse> CheckLatestAsync(
        string serverUrl,
        string currentVersion,
        string token,
        CancellationToken cancellationToken)
    {
        var baseUrl = NormalizeBase(serverUrl);
        var url = $"{baseUrl}/api/v1/apps/{Uri.EscapeDataString(_options.AppSlug)}/updates/latest" +
                  $"?current_version={Uri.EscapeDataString(currentVersion)}" +
                  $"&channel={Uri.EscapeDataString(_options.Channel)}" +
                  $"&platform={Uri.EscapeDataString(_options.Platform)}" +
                  $"&arch={Uri.EscapeDataString(_options.Architecture)}" +
                  $"&artifact_type={Uri.EscapeDataString(_options.ArtifactType)}";
        using var request = new HttpRequestMessage(HttpMethod.Get, url);
        request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", token);
        using var response = await _downloads.SendAsync(request, cancellationToken);
        if (response.StatusCode == HttpStatusCode.Unauthorized)
            throw new UnauthorizedAccessException("RobotManager 下载 Token 无效、过期或已撤销");
        return await ReadUpdateAsync(response, cancellationToken);
    }

    async Task<(string Directory, string Version)> DownloadAndInstallAsync(
        LatestUpdateResponse update,
        string token,
        IProgress<int>? progress,
        CancellationToken cancellationToken)
    {
        var artifact = update.Artifact ?? throw new InvalidOperationException("RobotManager 更新缺少产物");
        if (!Uri.TryCreate(artifact.DownloadUrl, UriKind.Absolute, out var downloadUrl)
            || string.IsNullOrWhiteSpace(artifact.Sha256))
            throw new InvalidOperationException("RobotManager 更新缺少有效下载地址或 SHA-256");

        Directory.CreateDirectory(DownloadsDirectory);
        Directory.CreateDirectory(VersionsDirectory);
        var archive = Path.Combine(DownloadsDirectory, $"{SafeFileName(_options.AppSlug)}-{SafeFileName(update.Version)}.zip");

        try
        {
            using (var request = new HttpRequestMessage(HttpMethod.Get, downloadUrl))
            {
                request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", token);
                using var response = await _downloads.SendAsync(
                    request,
                    HttpCompletionOption.ResponseHeadersRead,
                    cancellationToken);
                if (response.StatusCode == HttpStatusCode.Unauthorized)
                    throw new UnauthorizedAccessException("RobotManager 安装包下载 Token 无效、过期或已撤销");
                response.EnsureSuccessStatusCode();
                var total = response.Content.Headers.ContentLength ?? artifact.Size;
                await using var input = await response.Content.ReadAsStreamAsync(cancellationToken);
                await using var output = new FileStream(
                    archive,
                    FileMode.Create,
                    FileAccess.Write,
                    FileShare.None,
                    81920,
                    useAsync: true);
                var buffer = new byte[81920];
                long received = 0;
                while (true)
                {
                    var count = await input.ReadAsync(buffer, cancellationToken);
                    if (count == 0) break;
                    await output.WriteAsync(buffer.AsMemory(0, count), cancellationToken);
                    received += count;
                    if (total > 0)
                        progress?.Report((int)Math.Clamp(received * 100 / total, 0, 99));
                }
                await output.FlushAsync(cancellationToken);
            }

            if (artifact.Size > 0 && new FileInfo(archive).Length != artifact.Size)
                throw new InvalidOperationException("RobotManager 安装包大小校验失败");
            if (!Hash(archive).Equals(artifact.Sha256.Trim(), StringComparison.OrdinalIgnoreCase))
                throw new InvalidOperationException("RobotManager 安装包 SHA-256 校验失败");

            var staging = Path.Combine(VersionsDirectory, SafeFileName(update.Version) + ".staging");
            var target = Path.Combine(VersionsDirectory, SafeFileName(update.Version));
            if (Directory.Exists(staging)) Directory.Delete(staging, recursive: true);
            Directory.CreateDirectory(staging);
            try
            {
                ZipFile.ExtractToDirectory(archive, staging);
                if (!File.Exists(Path.Combine(staging, _options.ExecutableName)))
                    throw new InvalidOperationException($"RobotManager 安装包缺少 {_options.ExecutableName}");
                if (Directory.Exists(target)) Directory.Delete(target, recursive: true);
                Directory.Move(staging, target);
            }
            catch
            {
                if (Directory.Exists(staging)) Directory.Delete(staging, recursive: true);
                throw;
            }

            progress?.Report(100);
            return (target, update.Version);
        }
        finally
        {
            if (File.Exists(archive)) File.Delete(archive);
        }
    }

    async Task StopManagerHostAsync(CancellationToken cancellationToken)
    {
        try
        {
            await _manager.ShutdownHostAsync(cancellationToken);
        }
        catch (HttpRequestException)
        {
            KillManagerProcesses();
        }

        var deadline = DateTimeOffset.UtcNow.AddSeconds(10);
        while (DateTimeOffset.UtcNow < deadline)
        {
            if (await TryGetHealthAsync(cancellationToken) is null) return;
            await Task.Delay(250, cancellationToken);
        }
        throw new TimeoutException("旧版 RobotManager 退出超时");
    }

    void KillManagerProcesses()
    {
        var processName = Path.GetFileNameWithoutExtension(_options.ExecutableName);
        foreach (var process in Process.GetProcessesByName(processName))
        {
            try
            {
                process.Kill(entireProcessTree: true);
                process.WaitForExit(5000);
            }
            catch
            {
                // A subsequent health check decides whether shutdown succeeded.
            }
            finally
            {
                process.Dispose();
            }
        }
    }

    void StartManager(string directory)
    {
        if (!OperatingSystem.IsWindows())
            throw new PlatformNotSupportedException("Sao.WechatRobotManager 只能在 Windows 上启动");
        var executable = Path.Combine(directory, _options.ExecutableName);
        _ = Process.Start(new ProcessStartInfo
        {
            FileName = executable,
            WorkingDirectory = directory,
            UseShellExecute = true,
            WindowStyle = ProcessWindowStyle.Hidden,
        }) ?? throw new InvalidOperationException("RobotManager 启动失败");
    }

    async Task<RobotManagerHealth> WaitForHealthAsync(CancellationToken cancellationToken)
    {
        var deadline = DateTimeOffset.UtcNow.Add(_options.ManagerStartTimeout);
        while (DateTimeOffset.UtcNow < deadline)
        {
            await Task.Delay(250, cancellationToken);
            if (await TryGetHealthAsync(cancellationToken) is { } health) return health;
        }
        throw new TimeoutException("RobotManager 启动超时");
    }

    static void ValidateManagerIdentity(RobotManagerHealth health)
    {
        if (!health.Ok)
            throw new InvalidOperationException("RobotManager 健康检查未通过");
        if (!string.Equals(
                health.Service,
                RobotManagerBootstrapperOptions.ExpectedServiceName,
                StringComparison.Ordinal))
            throw new InvalidOperationException(
                $"端口上的服务不是预期的 RobotManager：{health.Service}");
    }

    static void ValidateManagerProtocol(RobotManagerHealth health)
    {
        if (health.ProtocolVersion != RobotManagerBootstrapperOptions.CurrentProtocolVersion)
            throw new InvalidOperationException(
                $"RobotManager 协议版本不兼容：需要 {RobotManagerBootstrapperOptions.CurrentProtocolVersion}，实际 {health.ProtocolVersion}");
    }

    (string Directory, string Version)? FindInstalledVersion()
    {
        if (!Directory.Exists(VersionsDirectory)) return null;
        var best = Directory.GetDirectories(VersionsDirectory)
            .Select(directory => (Directory: directory, Version: Path.GetFileName(directory)))
            .Where(item => Version.TryParse(item.Version, out _)
                           && File.Exists(Path.Combine(item.Directory, _options.ExecutableName)))
            .OrderByDescending(item => Version.Parse(item.Version))
            .FirstOrDefault();
        return best.Directory is null ? null : best;
    }

    async Task<LatestUpdateResponse> ReadUpdateAsync(
        HttpResponseMessage response,
        CancellationToken cancellationToken)
    {
        var body = await response.Content.ReadAsStringAsync(cancellationToken);
        if (!response.IsSuccessStatusCode)
            throw new HttpRequestException($"RobotManager 更新检查失败 HTTP {(int)response.StatusCode}: {Shorten(body)}");
        return JsonSerializer.Deserialize<LatestUpdateResponse>(body, _json)
               ?? throw new InvalidOperationException("RobotManager 更新接口返回空 JSON");
    }

    static void ValidateOptions(RobotManagerBootstrapperOptions options)
    {
        if (!options.ManagerApiBaseAddress.IsAbsoluteUri)
            throw new ArgumentException("Manager API 地址必须是绝对 URL", nameof(options));
        if (string.IsNullOrWhiteSpace(options.RootDirectory))
            throw new ArgumentException("Manager 安装目录不能为空", nameof(options));
        if (string.IsNullOrWhiteSpace(options.ExecutableName)
            || Path.GetFileName(options.ExecutableName) != options.ExecutableName)
            throw new ArgumentException("Manager 可执行文件名无效", nameof(options));
        if (options.ClientReadyTimeout <= TimeSpan.Zero)
            throw new ArgumentOutOfRangeException(nameof(options), "客户端 Ready 超时必须大于零");
        if (string.IsNullOrWhiteSpace(options.ClientId))
            throw new ArgumentException("客户端 ID 不能为空", nameof(options));
    }

    static string NormalizeBase(string value)
    {
        var result = (value ?? "").Trim();
        if (result.Length == 0) throw new InvalidOperationException("更新服务器地址为空");
        if (!result.StartsWith("http://", StringComparison.OrdinalIgnoreCase)
            && !result.StartsWith("https://", StringComparison.OrdinalIgnoreCase))
            result = "http://" + result;
        return result.TrimEnd('/');
    }

    static string SafeFileName(string value)
    {
        if (string.IsNullOrWhiteSpace(value))
            throw new InvalidOperationException("版本或应用名称为空");
        var invalid = Path.GetInvalidFileNameChars();
        return string.Concat(value.Trim().Select(character => invalid.Contains(character) ? '_' : character));
    }

    static string Hash(string path)
    {
        using var stream = File.OpenRead(path);
        return Convert.ToHexString(SHA256.HashData(stream)).ToLowerInvariant();
    }

    static string Shorten(string value)
    {
        var trimmed = string.IsNullOrWhiteSpace(value) ? "(empty)" : value.Trim();
        return trimmed[..Math.Min(512, trimmed.Length)];
    }

    public void Dispose()
    {
        _manager.Dispose();
        _downloads.Dispose();
        _ensureGate.Dispose();
    }
}
