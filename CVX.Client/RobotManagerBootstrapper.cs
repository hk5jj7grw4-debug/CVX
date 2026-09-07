using System.Diagnostics;
using System.IO.Compression;
using System.Net;
using System.Security.Cryptography;
using System.Text.Json;

namespace CVX.Client;

/// <summary>
/// Installs a missing manager or starts an installed manager. It does not
/// proxy or wrap the GVx business API.
/// </summary>
internal sealed class RobotManagerBootstrapper : IDisposable
{
    const int CurrentProtocolVersion = 2;
    const string ExpectedServiceName = "CVX.Manager";
    const string AppSlug = "wechat-robot-manager";
    const string ExecutableName = "CVX.Manager.exe";
    readonly WechatRobotOptions _options;
    readonly RobotManagerClient _manager;
    readonly HttpClient _downloads;
    readonly JsonSerializerOptions _json = new(JsonSerializerDefaults.Web)
    {
        PropertyNameCaseInsensitive = true,
    };

    string VersionsDirectory => Path.Combine(_options.InstallDirectory, "versions");
    string DownloadsDirectory => Path.Combine(_options.InstallDirectory, "downloads");

    public RobotManagerBootstrapper(WechatRobotOptions options)
    {
        _options = options;
        _manager = new RobotManagerClient(options.ManagerAddress);
        _downloads = new HttpClient(new SocketsHttpHandler { UseProxy = false })
        {
            Timeout = TimeSpan.FromMinutes(10),
        };
    }

    internal RobotManagerClient Client => _manager;

    internal async Task EnsureManagerRunningAsync(
        string? updateServerUrl,
        CancellationToken cancellationToken)
    {
        var health = await TryGetHealthAsync(cancellationToken);
        if (health is not null)
        {
            ValidateManagerIdentity(health);
            ValidateManagerProtocol(health);
            return;
        }

        var installed = FindInstalledVersion();
        if (installed is null)
        {
            if (string.IsNullOrWhiteSpace(updateServerUrl))
                throw new InvalidOperationException("Manager 尚未安装，请配置 UpdateServerUrl");
            var update = await CheckLatestAsync(updateServerUrl, cancellationToken);
            if (string.IsNullOrWhiteSpace(update.Version) || update.Artifact is null)
                throw new InvalidOperationException("更新服务器没有返回可安装的 Manager");
            installed = await DownloadAndInstallAsync(update, cancellationToken);
        }

        StartManager(installed);
        health = await WaitForHealthAsync(cancellationToken);
        ValidateManagerIdentity(health);
        ValidateManagerProtocol(health);
    }

    async Task<RobotManagerHealth?> TryGetHealthAsync(CancellationToken cancellationToken)
    {
        try
        {
            using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            timeout.CancelAfter(TimeSpan.FromSeconds(2));
            return await _manager.GetHealthAsync(timeout.Token);
        }
        catch (Exception exception) when (
            (exception is HttpRequestException { StatusCode: null } or OperationCanceledException)
            && !cancellationToken.IsCancellationRequested)
        {
            return null;
        }
    }

    async Task<LatestUpdateResponse> CheckLatestAsync(
        string serverUrl,
        CancellationToken cancellationToken)
    {
        var baseUrl = NormalizeBase(serverUrl);
        var url = $"{baseUrl}/api/v1/apps/{AppSlug}/updates/latest" +
                  "?current_version=0.0.0&channel=stable&platform=windows&arch=x64&artifact_type=portable";
        using var request = new HttpRequestMessage(HttpMethod.Get, url);
        using var response = await _downloads.SendAsync(request, cancellationToken);
        if (response.StatusCode == HttpStatusCode.Unauthorized)
            throw new UnauthorizedAccessException("Manager 版本查询必须允许匿名访问，请检查下载服务配置");
        return await ReadUpdateAsync(response, cancellationToken);
    }

    async Task<string> DownloadAndInstallAsync(
        LatestUpdateResponse update,
        CancellationToken cancellationToken)
    {
        var artifact = update.Artifact ?? throw new InvalidOperationException("RobotManager 更新缺少产物");
        if (!Uri.TryCreate(artifact.DownloadUrl, UriKind.Absolute, out var downloadUrl)
            || string.IsNullOrWhiteSpace(artifact.Sha256))
            throw new InvalidOperationException("RobotManager 更新缺少有效下载地址或 SHA-256");

        Directory.CreateDirectory(DownloadsDirectory);
        Directory.CreateDirectory(VersionsDirectory);
        var archive = Path.Combine(DownloadsDirectory, $"{SafeFileName(AppSlug)}-{SafeFileName(update.Version)}.zip");

        try
        {
            using (var request = new HttpRequestMessage(HttpMethod.Get, downloadUrl))
            {
                using var response = await _downloads.SendAsync(
                    request,
                    HttpCompletionOption.ResponseHeadersRead,
                    cancellationToken);
                if (response.StatusCode == HttpStatusCode.Unauthorized)
                    throw new UnauthorizedAccessException("Manager 安装包必须允许匿名下载，请检查下载服务配置");
                response.EnsureSuccessStatusCode();
                await using var input = await response.Content.ReadAsStreamAsync(cancellationToken);
                await using var output = new FileStream(
                    archive,
                    FileMode.Create,
                    FileAccess.Write,
                    FileShare.None,
                    81920,
                    useAsync: true);
                await input.CopyToAsync(output, cancellationToken);
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
                if (!File.Exists(Path.Combine(staging, ExecutableName)))
                    throw new InvalidOperationException($"RobotManager 安装包缺少 {ExecutableName}");
                if (Directory.Exists(target)) Directory.Delete(target, recursive: true);
                Directory.Move(staging, target);
            }
            catch
            {
                if (Directory.Exists(staging)) Directory.Delete(staging, recursive: true);
                throw;
            }

            return target;
        }
        finally
        {
            if (File.Exists(archive)) File.Delete(archive);
        }
    }

    void StartManager(string directory)
    {
        if (!OperatingSystem.IsWindows())
            throw new PlatformNotSupportedException("CVX.Manager 只能在 Windows 上启动");
        var executable = Path.Combine(directory, ExecutableName);
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
        var deadline = DateTimeOffset.UtcNow.Add(TimeSpan.FromSeconds(20));
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
                ExpectedServiceName,
                StringComparison.Ordinal))
            throw new InvalidOperationException(
                $"端口上的服务不是预期的 RobotManager：{health.Service}");
    }

    static void ValidateManagerProtocol(RobotManagerHealth health)
    {
        if (health.ProtocolVersion != CurrentProtocolVersion)
            throw new InvalidOperationException(
                $"RobotManager 协议版本不兼容：需要 {CurrentProtocolVersion}，实际 {health.ProtocolVersion}");
    }

    string? FindInstalledVersion()
    {
        if (!Directory.Exists(VersionsDirectory)) return null;
        var best = Directory.GetDirectories(VersionsDirectory)
            .Select(directory => (Directory: directory, Version: Path.GetFileName(directory)))
            .Where(item => Version.TryParse(item.Version, out _)
                           && File.Exists(Path.Combine(item.Directory, ExecutableName)))
            .OrderByDescending(item => Version.Parse(item.Version))
            .FirstOrDefault();
        return best.Directory;
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
    }
}
