namespace GVxSdk;

/// <summary>Message sending APIs.</summary>
public sealed class MessageApi
{
    private readonly GVxClient _client;

    internal MessageApi(GVxClient client) => _client = client;

    public Task SendTextAsync(string wxid, string message, CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_text_msg", new { wxid, msg = message }, cancellationToken);

    public Task SendImageAsync(string wxid, string filePath, CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_image_msg", new { wxid, filepath = filePath }, cancellationToken);

    public Task SendFileAsync(string wxid, string filePath, CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_file_msg", new { wxid, filepath = filePath }, cancellationToken);

    public Task SendAppMessageAsync(
        string wxid,
        string content,
        string messageType,
        CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_app_msg", new { wxid, content, type = messageType }, cancellationToken);

    public Task SendAppletAsync(
        string wxid,
        string content,
        string messageType = "33",
        CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_applet_msg", new { wxid, content, type = messageType }, cancellationToken);

    public Task SendLinkAsync(
        string wxid,
        string title,
        string description,
        string thumbUrl,
        string url,
        CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_xml", new { wxid, title, description, thumbUrl, url }, cancellationToken);

    public Task SendAtTextAsync(
        string roomId,
        string message,
        string wxids,
        CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_at_text", new { roomId, msg = message, wxids }, cancellationToken);

    public Task SendCardAsync(
        string wxid,
        string cardWxid,
        CancellationToken cancellationToken = default) =>
        CommandAsync("/wx41614/send_card_msg", new { wxid, cardWxid }, cancellationToken);

    public Task SendPatAsync(
        string roomId,
        string wxid,
        CancellationToken cancellationToken = default) =>
        CommandAsync("/api/send_pat", new { roomId, wxid }, cancellationToken);

    public Task SendQuoteAsync(
        string reply,
        string referContent,
        string fromUser,
        string newMessageId,
        int createTime,
        string sendTo,
        string? messageSource = null,
        CancellationToken cancellationToken = default) =>
        CommandAsync(
            "/api/send_quote",
            new
            {
                reply,
                referContent,
                fromUsr = fromUser,
                newmsgid = newMessageId,
                msgSource = messageSource,
                createTime,
                sendto = sendTo,
            },
            cancellationToken);

    public Task SendLocationAsync(
        string wxid,
        string longitude,
        string latitude,
        string label,
        string poiName,
        CancellationToken cancellationToken = default) =>
        CommandAsync(
            "/api/send_location_msg",
            new { wxid, x = longitude, y = latitude, lable = label, poiname = poiName },
            cancellationToken);

    public Task RevokeAsync(
        long newMessageId,
        int createTime,
        string toUserName,
        CancellationToken cancellationToken = default) =>
        CommandAsync(
            "/api/revoke_any",
            new { newMsgId = newMessageId, createTime, toUserName },
            cancellationToken);

    private async Task CommandAsync(string path, object request, CancellationToken cancellationToken)
    {
        _ = await _client.PostCommandAsync(path, request, cancellationToken).ConfigureAwait(false);
    }
}
