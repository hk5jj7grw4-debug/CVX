using System.Text.Json;

namespace CVX.Manager;

public sealed class RobotManagerSettings
{
    public string ListenUrl { get; set; } = "http://127.0.0.1:19087";
    public string UpdateServerUrl { get; set; } = "http://203.195.196.180:38081";
    public string ComponentToken { get; set; } = "";
    public bool AutoStart { get; set; } = true;
    public string DesiredState { get; set; } = "running";
    public int GvxApiPort { get; set; } = 19088;
    public string HttpCallbackUrl { get; set; } = "http://127.0.0.1:5000/api/recvMsg";
    public string AppliedHttpCallbackUrl { get; set; } = "";
    public int StartTimeoutSeconds { get; set; } = 30;
    public string DataDirectory { get; set; } = "";
}

public sealed record RobotManagerSettingsUpdate(
    string? UpdateServerUrl,
    string? ComponentToken,
    bool? AutoStart,
    int? GvxApiPort,
    string? HttpCallbackUrl,
    int? StartTimeoutSeconds);

public sealed record RobotManagerPublicSettings(
    string ListenUrl,
    string UpdateServerUrl,
    bool ComponentTokenConfigured,
    bool AutoStart,
    string DesiredState,
    int GvxApiPort,
    string HttpCallbackUrl,
    int StartTimeoutSeconds,
    string DataDirectory);

public sealed class RobotManagerSettingsStore
{
    static readonly JsonSerializerOptions JsonOptions = new(JsonSerializerDefaults.Web) { WriteIndented = true };
    readonly SemaphoreSlim _gate = new(1, 1);
    readonly string _settingsPath;

    public RobotManagerSettingsStore(IConfiguration configuration)
    {
        Current = configuration.GetSection("RobotManager").Get<RobotManagerSettings>() ?? new();
        if (string.IsNullOrWhiteSpace(Current.DataDirectory))
        {
            Current.DataDirectory = Path.Combine(
                Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
                "Sao",
                "WechatRobot");
        }

        Directory.CreateDirectory(Current.DataDirectory);
        _settingsPath = Path.Combine(Current.DataDirectory, "manager-settings.json");
        if (File.Exists(_settingsPath))
        {
            var saved = JsonSerializer.Deserialize<RobotManagerSettings>(File.ReadAllText(_settingsPath), JsonOptions);
            if (saved is not null)
            {
                saved.ListenUrl = Current.ListenUrl;
                saved.DataDirectory = Current.DataDirectory;
                saved.DesiredState = saved.DesiredState is "running" or "stopped"
                    ? saved.DesiredState
                    : "running";
                Current = saved;
            }
        }
    }

    public RobotManagerSettings Current { get; private set; }

    public bool IsCurrentCallbackApplied =>
        !string.IsNullOrWhiteSpace(Current.AppliedHttpCallbackUrl)
        && string.Equals(
            Current.AppliedHttpCallbackUrl,
            Current.HttpCallbackUrl,
            StringComparison.OrdinalIgnoreCase);

    public RobotManagerPublicSettings GetPublicSettings() => new(
        Current.ListenUrl,
        Current.UpdateServerUrl,
        !string.IsNullOrWhiteSpace(Current.ComponentToken),
        Current.AutoStart,
        Current.DesiredState,
        Current.GvxApiPort,
        Current.HttpCallbackUrl,
        Current.StartTimeoutSeconds,
        Current.DataDirectory);

    public async Task<RobotManagerPublicSettings> UpdateAsync(RobotManagerSettingsUpdate update, CancellationToken ct)
    {
        await _gate.WaitAsync(ct);
        try
        {
            var next = Current;
            if (update.UpdateServerUrl is not null) next.UpdateServerUrl = update.UpdateServerUrl.Trim();
            if (update.ComponentToken is not null) next.ComponentToken = update.ComponentToken.Trim();
            if (update.AutoStart.HasValue) next.AutoStart = update.AutoStart.Value;
            if (update.GvxApiPort.HasValue) next.GvxApiPort = ValidatePort(update.GvxApiPort.Value);
            if (update.HttpCallbackUrl is not null) next.HttpCallbackUrl = ValidateCallback(update.HttpCallbackUrl);
            if (update.StartTimeoutSeconds.HasValue)
                next.StartTimeoutSeconds = Math.Clamp(update.StartTimeoutSeconds.Value, 5, 300);

            await File.WriteAllTextAsync(_settingsPath, JsonSerializer.Serialize(next, JsonOptions), ct);
            Current = next;
            return GetPublicSettings();
        }
        finally
        {
            _gate.Release();
        }
    }

    public async Task SetDesiredStateAsync(string state, CancellationToken ct)
    {
        await _gate.WaitAsync(ct);
        try
        {
            Current.DesiredState = state;
            await File.WriteAllTextAsync(_settingsPath, JsonSerializer.Serialize(Current, JsonOptions), ct);
        }
        finally
        {
            _gate.Release();
        }
    }

    public async Task MarkCurrentCallbackAppliedAsync(CancellationToken ct)
    {
        await _gate.WaitAsync(ct);
        try
        {
            Current.AppliedHttpCallbackUrl = Current.HttpCallbackUrl;
            await File.WriteAllTextAsync(_settingsPath, JsonSerializer.Serialize(Current, JsonOptions), ct);
        }
        finally
        {
            _gate.Release();
        }
    }

    static int ValidatePort(int port) => port is >= 1 and <= 65535
        ? port
        : throw new ArgumentOutOfRangeException(nameof(port), "端口必须在 1 到 65535 之间");

    static string ValidateCallback(string value)
    {
        var url = value.Trim();
        if (!Uri.TryCreate(url, UriKind.Absolute, out var uri) || uri.Scheme is not ("http" or "https"))
            throw new InvalidOperationException("回调地址必须是有效的 HTTP 或 HTTPS URL");
        return url;
    }
}
