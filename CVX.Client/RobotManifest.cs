using System.Text.Json.Nodes;

namespace CVX.Client;

internal sealed record RobotFile(string Name, string Sha256);

internal sealed record RobotManifest(
    string Version,
    IReadOnlyList<string> SupportedWechatVersions,
    RobotFile Inject,
    RobotFile Dll,
    RobotFile WechatBundle,
    string WechatExePath,
    JsonObject LaunchArgs)
{
    public IEnumerable<RobotFile> Files => [Inject, Dll, WechatBundle];

    public static RobotManifest Load(string path)
    {
        var root = JsonNode.Parse(File.ReadAllText(path))?.AsObject()
            ?? throw new InvalidOperationException("组件清单 manifest.json 为空或无效");
        var files = root["files"]?.AsObject()
            ?? throw new InvalidOperationException("组件清单缺少 files 节");
        var bundle = files["wechatBundle"]?.AsObject()
            ?? throw new InvalidOperationException("组件清单缺少 files.wechatBundle");
        var supported = (root["supportedWechatVersions"] as JsonArray)?
            .Select(x => x?.GetValue<string>()?.Trim())
            .Where(x => !string.IsNullOrWhiteSpace(x))
            .Select(x => x!)
            .ToArray() ?? [];

        return new RobotManifest(
            Text(root, "version") ?? "0.0.0",
            supported,
            ReadFile(files, "inject"),
            ReadFile(files, "dll"),
            ReadFile(files, "wechatBundle"),
            (Text(bundle, "exePath") ?? throw new InvalidOperationException("组件清单缺少微信 exePath"))
                .Replace('\\', '/'),
            root["launchArgs"]?.AsObject()?.DeepClone().AsObject() ?? new JsonObject());
    }

    static RobotFile ReadFile(JsonObject files, string name)
    {
        var item = files[name]?.AsObject() ?? throw new InvalidOperationException($"组件清单缺少 files.{name}");
        return new RobotFile(
            Text(item, "name") ?? throw new InvalidOperationException($"files.{name} 缺少 name"),
            Text(item, "sha256") ?? throw new InvalidOperationException($"files.{name} 缺少 sha256"));
    }

    static string? Text(JsonObject node, string name) =>
        node.TryGetPropertyValue(name, out var value) && value is not null ? value.GetValue<string>().Trim() : null;
}
