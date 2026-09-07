namespace CVX.Client;

/// <summary>API readiness is independent of whether the account has logged in.</summary>
public enum WechatRuntimePhase { Idle, Checking, Downloading, Repairing, Connecting, Connected, Failed, Disposed }
public enum WechatVersionMatch { Unknown, Matching, Mismatched }

public sealed record WechatRobotStatus(bool ApiReady, bool IsLoggedIn, string? ComponentVersion, Uri CallbackUrl)
{
    public string RequiredWechatVersion => ComponentInstaller.RequiredWechatVersion;
    public string? DetectedWechatVersion { get; init; }
    public WechatVersionMatch VersionMatch { get; init; }
    public WechatRuntimePhase Phase { get; init; }
    public DateTimeOffset? CheckedAt { get; init; }
    public string? Error { get; init; }
    public double? DownloadProgress { get; init; }
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
