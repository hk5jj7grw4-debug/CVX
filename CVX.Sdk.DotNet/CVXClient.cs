using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace CVX.Sdk;

/// <summary>Client for the gVx middleware HTTP API.</summary>
public sealed class CVXClient : IDisposable
{
    private static readonly JsonSerializerOptions DefaultJsonOptions = new(JsonSerializerDefaults.Web)
    {
        PropertyNameCaseInsensitive = true,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
        NumberHandling = JsonNumberHandling.AllowReadingFromString,
    };

    private readonly HttpClient _httpClient;
    private readonly bool _ownsHttpClient;
    private readonly string? _authToken;

    public CVXClient(
        string baseUrl = "http://127.0.0.1:19088",
        string? authToken = null,
        HttpClient? httpClient = null)
    {
        if (!Uri.TryCreate(EnsureTrailingSlash(baseUrl), UriKind.Absolute, out var baseAddress))
        {
            throw new ArgumentException("baseUrl must be an absolute HTTP or HTTPS URL.", nameof(baseUrl));
        }

        if (baseAddress.Scheme is not ("http" or "https"))
        {
            throw new ArgumentException("baseUrl must use HTTP or HTTPS.", nameof(baseUrl));
        }

        _ownsHttpClient = httpClient is null;
        _httpClient = httpClient ?? new HttpClient { Timeout = TimeSpan.FromSeconds(30) };
        _httpClient.BaseAddress = baseAddress;
        _authToken = string.IsNullOrWhiteSpace(authToken) ? null : authToken;

        Message = new MessageApi(this);
        System = new SystemApi(this);
        Contact = new ContactApi(this);
        Room = new RoomApi(this);
        Raw = new RawApi(this);
    }

    public MessageApi Message { get; }

    public SystemApi System { get; }

    public ContactApi Contact { get; }

    public RoomApi Room { get; }

    /// <summary>Access to every endpoint, including APIs without a typed wrapper.</summary>
    public RawApi Raw { get; }

    public async Task<TResponse> PostAsync<TResponse>(
        string path,
        object? request = null,
        CancellationToken cancellationToken = default)
    {
        using var message = new HttpRequestMessage(HttpMethod.Post, NormalizePath(path));
        if (request is not null)
        {
            message.Content = JsonContent.Create(request, options: DefaultJsonOptions);
        }

        if (_authToken is not null)
        {
            message.Headers.TryAddWithoutValidation("X-Auth-Token", _authToken);
        }

        HttpResponseMessage response;
        try
        {
            response = await _httpClient.SendAsync(message, cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex) when (ex is HttpRequestException or TaskCanceledException)
        {
            throw new CVXApiException("Unable to call the local gVx middleware.", innerException: ex);
        }

        using (response)
        {
            var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
            if (!response.IsSuccessStatusCode)
            {
                throw new CVXApiException(
                    $"gVx middleware returned HTTP {(int)response.StatusCode}.",
                    response.StatusCode,
                    body);
            }

            try
            {
                var result = JsonSerializer.Deserialize<TResponse>(body, DefaultJsonOptions);
                return result ?? throw new JsonException("Response body is null.");
            }
            catch (JsonException ex)
            {
                throw new CVXApiException("Unable to parse the gVx middleware response.", response.StatusCode, body, ex);
            }
        }
    }

    internal async Task<JsonElement> PostCommandAsync(
        string path,
        object? request = null,
        CancellationToken cancellationToken = default)
    {
        var response = await PostAsync<JsonElement>(path, request, cancellationToken).ConfigureAwait(false);
        EnsureBusinessSuccess(response);
        return response;
    }

    internal static JsonElement GetRequiredProperty(JsonElement response, string propertyName)
    {
        if (!response.TryGetProperty(propertyName, out var value))
        {
            throw new CVXApiException($"gVx response does not contain '{propertyName}'.", responseBody: response.GetRawText());
        }

        return value;
    }

    internal static void EnsureBusinessSuccess(JsonElement response)
    {
        if (TryGetInt(response, "code", out var code) && code is not (0 or 1))
        {
            throw BusinessException(response, code, "msg");
        }

        if (TryGetInt(response, "errCode", out var errCode) && errCode is not (0 or 1))
        {
            throw BusinessException(response, errCode, "errMsg");
        }
    }

    private static CVXApiException BusinessException(JsonElement response, int code, string messageProperty)
    {
        var message = response.TryGetProperty(messageProperty, out var value) ? value.ToString() : "Unknown API error";
        return new CVXApiException($"gVx API error: code={code}, message={message}", responseBody: response.GetRawText());
    }

    private static bool TryGetInt(JsonElement response, string propertyName, out int value)
    {
        value = 0;
        if (!response.TryGetProperty(propertyName, out var property))
        {
            return false;
        }

        return property.ValueKind switch
        {
            JsonValueKind.Number => property.TryGetInt32(out value),
            JsonValueKind.String => int.TryParse(property.GetString(), out value),
            _ => false,
        };
    }

    private static string EnsureTrailingSlash(string value) => value.EndsWith('/') ? value : value + "/";

    private static string NormalizePath(string value) => value.TrimStart('/');

    public void Dispose()
    {
        if (_ownsHttpClient)
        {
            _httpClient.Dispose();
        }
    }
}
