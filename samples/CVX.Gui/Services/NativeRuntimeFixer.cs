using System.IO;

namespace CVX.Gui.Services;

internal static class NativeRuntimeFixer
{
    private static readonly string[] DllNames =
    [
        "msvcp140.dll",
        "vcruntime140.dll",
        "vcruntime140_1.dll"
    ];

    public static int EnsureVcRuntimeBesideWechat()
    {
        var system32 = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "");
        var copied = 0;

        foreach (var directory in EnumerateWechatDirs())
        {
            foreach (var name in DllNames)
            {
                var source = Path.Combine(system32, name);
                if (!File.Exists(source))
                {
                    continue;
                }

                var destination = Path.Combine(directory, name);
                File.Copy(source, destination, overwrite: true);
                copied++;
            }
        }

        return copied;
    }

    private static IEnumerable<string> EnumerateWechatDirs()
    {
        var root = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
            "CVX",
            "versions");

        if (!Directory.Exists(root))
        {
            yield break;
        }

        foreach (var versionDir in Directory.GetDirectories(root))
        {
            yield return versionDir;

            var weixinRoot = Path.Combine(versionDir, "Weixin");
            if (Directory.Exists(weixinRoot))
            {
                yield return weixinRoot;
            }

            var versionName = Path.GetFileName(versionDir);
            var inner = Path.Combine(weixinRoot, versionName);
            if (Directory.Exists(inner))
            {
                yield return inner;
            }
        }
    }
}
