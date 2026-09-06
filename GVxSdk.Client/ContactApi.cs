using System.Text.Json;
using System.Text.Json.Serialization;

namespace GVxSdk;

/// <summary>Current account and contact profile APIs.</summary>
public sealed class ContactApi
{
    private readonly GVxClient _client;

    internal ContactApi(GVxClient client) => _client = client;

    public Task<ProfileResponse> GetProfileCacheAsync(CancellationToken cancellationToken = default) =>
        GetProfileAsync("/api/get_profile_cache", cancellationToken);

    public Task<ProfileResponse> GetProfileNewAsync(CancellationToken cancellationToken = default) =>
        GetProfileAsync("/api/get_profile_new", cancellationToken);

    public Task<ContactFastResponse> GetContactFastAsync(
        string wxid,
        CancellationToken cancellationToken = default) =>
        _client.PostAsync<ContactFastResponse>("/api/get_contact_fast", new { wxid }, cancellationToken);

    private async Task<ProfileResponse> GetProfileAsync(
        string path,
        CancellationToken cancellationToken)
    {
        var response = await _client.PostAsync<ProfileResponse>(path, cancellationToken: cancellationToken)
            .ConfigureAwait(false);
        if (response.BaseResponse.Ret != 0)
        {
            throw new GVxApiException(
                $"gVx profile API error: ret={response.BaseResponse.Ret}, message={response.BaseResponse.ErrMsg}");
        }

        return response;
    }
}

public sealed class ProfileResponse
{
    public ProfileBaseResponse BaseResponse { get; init; } = new();
    public ProfileUserInfo UserInfo { get; init; } = new();
    public ProfileUserInfoExt UserInfoExt { get; init; } = new();
}

public sealed class ProfileBaseResponse
{
    public int Ret { get; init; }
    public JsonElement ErrMsg { get; init; }
}

public sealed class ProfileUserInfo
{
    public WrappedString UserName { get; init; } = new();
    public WrappedString NickName { get; init; } = new();
}

public sealed class ProfileUserInfoExt
{
    public string? SmallHeadImgUrl { get; init; }
    public string? BigHeadImgUrl { get; init; }
}

public sealed class WrappedString
{
    [JsonPropertyName("String")]
    public string? Value { get; init; }
}

public sealed class ContactFastResponse
{
    public ContactFastInfo Contact { get; init; } = new();
    public int Ret { get; init; }
    public string? Username { get; init; }
}

public sealed class ContactFastInfo
{
    public WrappedString NickName { get; init; } = new();
    public string? SmallHeadImgUrl { get; init; }
    public string? BigHeadImgUrl { get; init; }
    public ImageBuffer ImgBuf { get; init; } = new();
}

public sealed class ImageBuffer
{
    public string? Buffer { get; init; }
    public int ILen { get; init; }
}
