using System.Diagnostics;
using System.IO.Compression;
using System.Net;
using System.Net.Http.Headers;
using System.Security.Cryptography;
using System.Text.Json.Nodes;

namespace CVX.Client;

internal sealed record InstalledComponent(string Directory, RobotManifest Manifest);

internal sealed class ComponentInstaller(WechatRobotOptions options) : IDisposable
{
    readonly HttpClient _http = new(new SocketsHttpHandler { UseProxy = false }) { Timeout = TimeSpan.FromMinutes(10) };
    string Versions => Path.Combine(options.ComponentDirectory, "versions");

    internal async Task<InstalledComponent> EnsureInstalledAsync(CancellationToken ct)
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
                await source.CopyToAsync(target, ct).ConfigureAwait(false);
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
        var target = Path.Combine(directory, "weixin");
        if (Directory.Exists(target)) Directory.Delete(target, true);
        Directory.CreateDirectory(target);
        ExtractArchive(archive, target);
    }

    internal static string EnsureBundleExtracted(string directory, RobotManifest manifest)
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
        SafePath(Path.Combine(directory, "weixin"), manifest.WechatExePath);

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
            var version = (info.FileVersion ?? info.ProductVersion)?.Trim();
            return string.IsNullOrWhiteSpace(version) ? null : version;
        }
        catch { return null; }
    }

    static bool IsSupportedWechatVersion(string? version, RobotManifest manifest) =>
        string.IsNullOrWhiteSpace(version)
            ? manifest.SupportedWechatVersions.Count == 0
            : manifest.SupportedWechatVersions.Count == 0 ||
              manifest.SupportedWechatVersions.Contains(version, StringComparer.OrdinalIgnoreCase);

}
