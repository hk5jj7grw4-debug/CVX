using System.Diagnostics;

namespace CVX.Client;

internal static class WechatVersionDetector
{
    internal static string? Normalize(string? value) =>
        Version.TryParse(value?.Trim(), out var parsed) && parsed.Build >= 0 && parsed.Revision >= 0
            ? parsed.ToString(4) : null;

    // This inspects the installed tree, not the modules currently mapped in a process.
    internal static string? Detect(string executable, Func<string, string?>? readVersion = null)
    {
        if (!File.Exists(executable)) return null;
        readVersion ??= ReadFixedVersion;
        var root = Path.GetDirectoryName(Path.GetFullPath(executable))!;
        var candidates = Directory.GetDirectories(root)
            .Select(path => (Path: path, Version: Normalize(Path.GetFileName(path))))
            .Where(item => item.Version is not null)
            .Select(item => (item.Path, Version: Version.Parse(item.Version!),
                DllVersion: Normalize(readVersion(Path.Combine(item.Path, "Weixin.dll")))))
            .OrderByDescending(item => item.Version)
            .ToArray();
        // Prefer complete installations over a higher, partially written update directory.
        var complete = candidates.FirstOrDefault(item => item.DllVersion == item.Version.ToString(4));
        if (complete.DllVersion is not null) return complete.DllVersion;
        // If no directory agrees with its DLL, retain the DLL's actual version rather than its label.
        var dll = candidates.Where(item => item.DllVersion is not null)
            .OrderByDescending(item => Version.Parse(item.DllVersion!)).FirstOrDefault();
        return dll.DllVersion ?? candidates.FirstOrDefault().Version?.ToString(4)
            ?? Normalize(readVersion(executable));
    }

    static string? ReadFixedVersion(string path)
    {
        if (!File.Exists(path)) return null;
        try
        {
            var info = FileVersionInfo.GetVersionInfo(path);
            if (info.FileMajorPart == 0 && info.FileMinorPart == 0 && info.FileBuildPart == 0 && info.FilePrivatePart == 0)
                return null;
            return $"{info.FileMajorPart}.{info.FileMinorPart}.{info.FileBuildPart}.{info.FilePrivatePart}";
        }
        catch (Exception ex) when (ex is IOException or UnauthorizedAccessException or System.ComponentModel.Win32Exception) { return null; }
    }
}
