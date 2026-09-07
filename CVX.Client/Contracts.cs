using System.Text.Json.Serialization;

namespace CVX.Client;

public sealed record RobotManagerHealth(
    bool Ok,
    string Service,
    string Version,
    int ProtocolVersion,
    int ProcessId);

public sealed record RobotManagerStatus(
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

public sealed record RobotClientRegistration(
    string ClientId,
    string InstanceId,
    string CallbackUrl);

public sealed record RobotClientSession(
    bool Ok,
    string ClientId,
    string InstanceId,
    bool CallbackReady);

public sealed record RobotManagerOperation(
    bool Running,
    string? Operation,
    string Phase,
    string PhaseText,
    int? Progress,
    DateTimeOffset? StartedAt,
    string? Error);

/// <summary>Configuration understood by CVX.Manager.</summary>
public sealed record RobotManagerConfig
{
    public required string UpdateServerUrl { get; init; }
    public string ComponentToken { get; init; } = "";
    public bool AutoStart { get; init; } = true;
    public int GvxApiPort { get; init; } = 19088;
    public string HttpCallbackUrl { get; init; } = WechatCallbackReceiver.DefaultCallbackUrl;
    public int StartTimeoutSeconds { get; init; } = 30;
}

public sealed record WechatCallbackMessage(
    DateTimeOffset ReceivedAt,
    string Body,
    string MessageType,
    string GroupId,
    string SenderId,
    string SenderName,
    string Text,
    string MessageId)
{
    public bool IsGroupMessage => !string.IsNullOrWhiteSpace(GroupId);
}

internal sealed class LatestUpdateResponse
{
    [JsonPropertyName("version")]
    public string Version { get; set; } = "";

    [JsonPropertyName("artifact")]
    public UpdateArtifact? Artifact { get; set; }
}

internal sealed class UpdateArtifact
{
    [JsonPropertyName("size")]
    public long Size { get; set; }

    [JsonPropertyName("sha256")]
    public string Sha256 { get; set; } = "";

    [JsonPropertyName("download_url")]
    public string DownloadUrl { get; set; } = "";
}
