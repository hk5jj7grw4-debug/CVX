using System.Text.Json;

namespace GVxSdk;

/// <summary>Generic access to all current and future middleware endpoints.</summary>
public sealed class RawApi
{
    private readonly GVxClient _client;

    internal RawApi(GVxClient client) => _client = client;

    public Task<JsonElement> PostAsync(
        string path,
        object? request = null,
        CancellationToken cancellationToken = default) =>
        _client.PostAsync<JsonElement>(path, request, cancellationToken);

    public Task<TResponse> PostAsync<TResponse>(
        string path,
        object? request = null,
        CancellationToken cancellationToken = default) =>
        _client.PostAsync<TResponse>(path, request, cancellationToken);
}
