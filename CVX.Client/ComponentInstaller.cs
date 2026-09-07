using System.IO.Compression;
using System.Net;
using System.Net.Http.Headers;
using System.Security.Cryptography;
using System.Text.Json.Nodes;

namespace CVX.Client;

internal sealed record InstalledComponent(string Directory, RobotManifest Manifest);

internal sealed class ComponentInstaller(WechatRobotOptions options) : IDisposable
{
    internal const string RequiredWechatVersion = "4.1.8.27";
    internal Action<WechatRuntimePhase, string?, double?>? Report { get; set; }
    readonly HttpClient _http = new(new SocketsHttpHandler { UseProxy = false }) { Timeout = TimeSpan.FromMinutes(10) };
    string Versions => Path.Combine(options.ComponentDirectory, "versions");

    internal async Task<InstalledComponent> EnsureInstalledAsync(CancellationToken ct)
        => await Task.Run(() => EnsureCoreAsync(ct), ct).ConfigureAwait(false);

    async Task<InstalledComponent> EnsureCoreAsync(CancellationToken ct)
    {
        var installed = FindInstalled();
        if (installed is not null && VerifyFiles(installed.Directory, installed.Manifest))
        {
            RepairBundleIfNeeded(installed.Directory, installed.Manifest);
            return installed;
        }
        if (string.IsNullOrWhiteSpace(options.ComponentToken))
            throw new InvalidOperationException("组件未安装或已损坏，请配置 ComponentToken 以下载组件");
        if (string.IsNullOrWhiteSpace(options.UpdateServerUrl))
            throw new InvalidOperationException("组件未安装或已损坏，请配置 UpdateServerUrl");
        Report?.Invoke(WechatRuntimePhase.Downloading, null, null);
        using var request = new HttpRequestMessage(HttpMethod.Get,
            options.UpdateServerUrl.TrimEnd('/') + "/api/v1/apps/wechat-robot/updates/latest?current_version=0.0.0&channel=stable&platform=windows&arch=x64");
        request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", options.ComponentToken.Trim());
        using var response = await _http.SendAsync(request, ct).ConfigureAwait(false);
        EnsureSuccess(response);
        var root = JsonNode.Parse(await response.Content.ReadAsStringAsync(ct).ConfigureAwait(false))?.AsObject()
            ?? throw new InvalidOperationException("组件服务返回无效 JSON");
        var version = root["version"]?.GetValue<string>() ?? "";
        if (string.IsNullOrWhiteSpace(version) || version is "." or ".." || version.IndexOfAny(['/', '\\', ':']) >= 0)
            throw new InvalidOperationException("组件版本无效");
        var artifact = root["artifact"]?.AsObject() ?? throw new InvalidOperationException("组件服务没有返回安装包");
        var url = (artifact["download_url"] ?? artifact["downloadUrl"])?.GetValue<string>();
        var sha = artifact["sha256"]?.GetValue<string>();
        var size = artifact["size"]?.GetValue<long>() ?? 0;
        if (!Uri.TryCreate(url, UriKind.Absolute, out var download) || download.Scheme is not ("http" or "https") || string.IsNullOrWhiteSpace(sha))
            throw new InvalidOperationException("组件包缺少下载地址或 SHA-256");
        Directory.CreateDirectory(Versions);
        var downloads = Path.Combine(options.ComponentDirectory, "downloads");
        Directory.CreateDirectory(downloads);
        var archive = Path.Combine(downloads, Guid.NewGuid().ToString("N") + ".zip");
        var staging = SafePath(Versions, version + ".staging");
        try
        {
            using (var downloadRequest = new HttpRequestMessage(HttpMethod.Get, download))
            {
                downloadRequest.Headers.Authorization = new AuthenticationHeaderValue("Bearer", options.ComponentToken.Trim());
                using var package = await _http.SendAsync(downloadRequest, HttpCompletionOption.ResponseHeadersRead, ct).ConfigureAwait(false);
                EnsureSuccess(package);
                await using var source = await package.Content.ReadAsStreamAsync(ct).ConfigureAwait(false);
                await using var target = new FileStream(archive, FileMode.CreateNew, FileAccess.Write, FileShare.None, 81920, true);
                var buffer = new byte[81920];
                long received = 0;
                int count;
                while ((count = await source.ReadAsync(buffer, ct).ConfigureAwait(false)) > 0)
                {
                    await target.WriteAsync(buffer.AsMemory(0, count), ct).ConfigureAwait(false);
                    received += count;
                    Report?.Invoke(WechatRuntimePhase.Downloading, null, size > 0 ? Math.Min(100, received * 100d / size) : null);
                }
            }
            ct.ThrowIfCancellationRequested();
            if ((size > 0 && new FileInfo(archive).Length != size) || !Hash(archive).Equals(sha.Trim(), StringComparison.OrdinalIgnoreCase))
                throw new InvalidOperationException("组件包大小或 SHA-256 校验失败");
            if (Directory.Exists(staging)) Directory.Delete(staging, true);
            ExtractArchive(archive, staging);
            var manifest = RobotManifest.Load(Path.Combine(staging, "manifest.json"));
            if (manifest.Version != version || !VerifyFiles(staging, manifest))
                throw new InvalidOperationException("组件清单版本或文件哈希不匹配");
            RepairBundleIfNeeded(staging, manifest);
            ct.ThrowIfCancellationRequested();
            var destination = SafePath(Versions, version);
            if (Directory.Exists(destination)) Directory.Delete(destination, true);
            Directory.Move(staging, destination);
            return new(destination, manifest);
        }
        finally
        {
            if (File.Exists(archive)) File.Delete(archive);
            if (Directory.Exists(staging)) Directory.Delete(staging, true);
        }
    }

    internal InstalledComponent? FindInstalled()
    {
        if (!Directory.Exists(Versions)) return null;
        InstalledComponent? best = null;
        foreach (var directory in Directory.GetDirectories(Versions))
        {
            if (directory.EndsWith(".staging", StringComparison.OrdinalIgnoreCase)) continue;
            var path = Path.Combine(directory, "manifest.json");
            if (!File.Exists(path)) continue;
            try
            {
                var manifest = RobotManifest.Load(path);
                if (best is null || CompareVersions(manifest.Version, best.Manifest.Version) > 0)
                    best = new(directory, manifest);
            }
            catch (Exception ex) when (ex is System.Text.Json.JsonException or InvalidOperationException or IOException) { }
        }
        return best;
    }

    static bool VerifyFiles(string directory, RobotManifest manifest) => manifest.Files.All(file =>
    {
        var path = SafePath(directory, file.Name);
        return File.Exists(path) && Hash(path).Equals(file.Sha256, StringComparison.OrdinalIgnoreCase);
    });

    internal static string SafePath(string root, string relative)
    {
        var prefix = Path.GetFullPath(root).TrimEnd(Path.DirectorySeparatorChar) + Path.DirectorySeparatorChar;
        var path = Path.GetFullPath(Path.Combine(prefix, relative.Replace('\\', Path.DirectorySeparatorChar)));
        if (!path.StartsWith(prefix, OperatingSystem.IsWindows() ? StringComparison.OrdinalIgnoreCase : StringComparison.Ordinal))
            throw new InvalidOperationException("组件路径超出安装目录");
        return path;
    }

    static void ExtractArchive(string archive, string destination)
    {
        using var zip = ZipFile.OpenRead(archive);
        foreach (var entry in zip.Entries)
        {
            _ = SafePath(destination, entry.FullName);
            if (((entry.ExternalAttributes >> 16) & 0xF000) == 0xA000)
                throw new InvalidOperationException("组件包不能包含符号链接");
        }
        zip.ExtractToDirectory(destination);
    }

    static void EnsureSuccess(HttpResponseMessage response)
    {
        if (response.StatusCode == HttpStatusCode.Unauthorized)
            throw new UnauthorizedAccessException("组件 Token 无效、过期或已撤销");
        response.EnsureSuccessStatusCode();
    }

    static string Hash(string path)
    {
        using var stream = File.OpenRead(path);
        return Convert.ToHexString(SHA256.HashData(stream));
    }

    static int CompareVersions(string left, string right) =>
        Version.TryParse(left, out var l) && Version.TryParse(right, out var r)
            ? l.CompareTo(r) : string.Compare(left, right, StringComparison.OrdinalIgnoreCase);

    public void Dispose() => _http.Dispose();
    static void ExtractBundle(string directory, RobotManifest manifest)
    {
        var archive = SafePath(directory, manifest.WechatBundle.Name);
        if (!File.Exists(archive)) throw new InvalidOperationException($"缺少便携微信包 {manifest.WechatBundle.Name}");
        // Stage first so repair replaces only bundle content, never the version directory itself.
        var staging = Path.Combine(directory, ".wechat-" + Guid.NewGuid().ToString("N") + ".staging");
        try
        {
            ExtractArchive(archive, staging);
            if (ResolveBundledWechatExe(staging, manifest) is null)
                throw new InvalidOperationException("便携微信包缺少清单指定的可执行文件");
            var entries = Directory.GetFileSystemEntries(staging);
            var protectedPaths = manifest.Files.Select(file => SafePath(directory, file.Name))
                .Append(Path.Combine(directory, "manifest.json")).ToArray();
            foreach (var entry in entries)
            {
                var destination = SafePath(directory, Path.GetFileName(entry));
                if (protectedPaths.Any(path => path.Equals(destination, StringComparison.OrdinalIgnoreCase)
                    || path.StartsWith(destination + Path.DirectorySeparatorChar, StringComparison.OrdinalIgnoreCase)))
                    throw new InvalidOperationException("便携微信包不能覆盖组件清单、注入器、DLL 或压缩包");
            }
            foreach (var entry in entries)
            {
                var destination = SafePath(directory, Path.GetFileName(entry));
                if (Directory.Exists(destination)) Directory.Delete(destination, true);
                else if (File.Exists(destination)) File.Delete(destination);
                if (Directory.Exists(entry)) Directory.Move(entry, destination);
                else File.Move(entry, destination);
            }
        }
        finally { if (Directory.Exists(staging)) Directory.Delete(staging, true); }
    }

    internal static string EnsureBundleExtracted(string directory, RobotManifest manifest)
    {
        var exe = ResolveBundledWechatExe(directory, manifest) ?? BundledWechatExe(directory, manifest);
        if (File.Exists(exe)) return exe;
        ExtractBundle(directory, manifest);
        return ResolveBundledWechatExe(directory, manifest) ?? BundledWechatExe(directory, manifest);
    }

    void RepairBundleIfNeeded(string directory, RobotManifest manifest)
    {
        if (ResolveBundledWechatExe(directory, manifest) is null)
            Report?.Invoke(WechatRuntimePhase.Repairing, null, null);
        var exe = EnsureBundleExtracted(directory, manifest);
        var version = DetectWechatVersion(directory, manifest, exe);
        if (IsSupportedWechatVersion(version, manifest)) return;
        Report?.Invoke(WechatRuntimePhase.Repairing, version, null);
        ExtractBundle(directory, manifest);
        exe = EnsureBundleExtracted(directory, manifest);
        version = DetectWechatVersion(directory, manifest, exe);
        if (!IsSupportedWechatVersion(version, manifest))
            throw new InvalidOperationException($"便携微信版本 {version ?? "未知"} 与组件不兼容");
    }

    static string BundledWechatExe(string directory, RobotManifest manifest) =>
        SafePath(directory, manifest.WechatExePath);

    static string? ResolveBundledWechatExe(string directory, RobotManifest manifest)
    {
        var exact = BundledWechatExe(directory, manifest);
        return File.Exists(exact) ? exact : null;
    }

    internal static string? DetectWechatVersion(string directory, RobotManifest manifest, string exe) =>
        WechatVersionDetector.Detect(exe);

    internal (string? Version, string? ComponentVersion, bool? Compatible) InspectVersion()
    {
        var installed = FindInstalled();
        if (installed is null) return (null, null, null);
        var exe = BundledWechatExe(installed.Directory, installed.Manifest);
        var version = File.Exists(exe) ? DetectWechatVersion(installed.Directory, installed.Manifest, exe) : null;
        return (version, installed.Manifest.Version, version is null ? null : IsSupportedWechatVersion(version, installed.Manifest));
    }

    static bool IsSupportedWechatVersion(string? version, RobotManifest manifest) =>
        manifest.SupportedWechatVersions.Count == 0 ||
        (WechatVersionDetector.Normalize(version) is { } normalized &&
         manifest.SupportedWechatVersions.Any(value => WechatVersionDetector.Normalize(value) == normalized));
}
