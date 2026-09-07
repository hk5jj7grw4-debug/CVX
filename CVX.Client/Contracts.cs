namespace CVX.Client;

/// <summary>API readiness is independent of whether the account has logged in.</summary>
public sealed record WechatRobotStatus(bool ApiReady, bool IsLoggedIn, string? ComponentVersion, Uri CallbackUrl);

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
