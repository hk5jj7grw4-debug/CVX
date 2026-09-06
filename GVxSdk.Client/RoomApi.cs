using System.Text.Json.Serialization;

namespace GVxSdk;

/// <summary>Chatroom information APIs.</summary>
public sealed class RoomApi
{
    private readonly GVxClient _client;

    internal RoomApi(GVxClient client) => _client = client;

    /// <summary>Gets all chatrooms for the current account.</summary>
    public async Task<IReadOnlyList<ChatroomInfo>> GetChatroomListAsync(
        CancellationToken cancellationToken = default)
    {
        var response = await _client.PostAsync<ChatroomListResponse>(
            "/api/get_chatroom_list",
            cancellationToken: cancellationToken).ConfigureAwait(false);
        if (response.Code is not (0 or 1))
        {
            throw new GVxApiException(
                $"gVx chatroom list API error: code={response.Code}, message={response.Message}");
        }

        return response.Data;
    }
}

public sealed class ChatroomListResponse
{
    public int Code { get; init; }
    public IReadOnlyList<ChatroomInfo> Data { get; init; } = [];

    [JsonPropertyName("msg")]
    public string? Message { get; init; }

    public int Total { get; init; }
}

public sealed class ChatroomInfo
{
    [JsonPropertyName("username")]
    public string GroupId { get; init; } = "";

    [JsonPropertyName("nick_name")]
    public string GroupName { get; init; } = "";

    [JsonPropertyName("remark")]
    public string Remark { get; init; } = "";

    [JsonPropertyName("small_head_url")]
    public string SmallHeadImageUrl { get; init; } = "";

    [JsonPropertyName("big_head_url")]
    public string BigHeadImageUrl { get; init; } = "";
}
