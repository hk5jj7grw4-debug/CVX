using System.Net.Http.Json;
using System.Text.Json;

namespace CVXkernel.Client;

public interface IRobotManagerClient
{
    Task<RobotManagerHealth> GetHealthAsync(CancellationToken cancellationToken = default);
    Task<RobotManagerStatus> GetStatusAsync(CancellationToken cancellationToken = default);
    Task<RobotManagerOperation> GetOperationAsync(CancellationToken cancellationToken = default);
    Task ConfigureAsync(RobotManagerConfig config, CancellationToken cancellationToken = default);
    Task<RobotManagerStatus> ReconcileAsync(CancellationToken cancellationToken = default);
    Task<RobotManagerStatus> StartAsync(CancellationToken cancellationToken = default);
    Task<RobotManagerStatus> StopAsync(CancellationToken cancellationToken = default);
    Task<RobotManagerStatus> RestartAsync(CancellationToken cancellationToken = default);
    Task<RobotManagerStatus> RepairAsync(CancellationToken cancellationToken = default);
    Task ShutdownHostAsync(CancellationToken cancellationToken = default);
}

/// <summary>Typed HTTP client for the lifecycle-only Robot Manager service.</summary>
public sealed class RobotManagerClient : IRobotManagerClient, IDisposable
{
    public static readonly Uri DefaultBaseAddress = new("http://127.0.0.1:19087");

    readonly HttpClient _http;
    readonly bool _ownsHttpClient;
    readonly JsonSerializerOptions _json = new(JsonSerializerDefaults.Web)
    {
        PropertyNameCaseInsensitive = true,
    };

    public RobotManagerClient(Uri? baseAddress = null, TimeSpan? timeout = null, string? apiToken = null)
        : this(new HttpClient(), baseAddress, ownsHttpClient: true, timeout, apiToken)
    {
    }

    public RobotManagerClient(
        HttpClient httpClient,
        Uri? baseAddress = null,
        bool ownsHttpClient = false,
        TimeSpan? timeout = null,
        string? apiToken = null)
    {
        _http = httpClient ?? throw new ArgumentNullException(nameof(httpClient));
        _ownsHttpClient = ownsHttpClient;
        _http.BaseAddress = baseAddress ?? _http.BaseAddress ?? DefaultBaseAddress;
        _http.Timeout = timeout ?? TimeSpan.FromMinutes(12);
        if (!string.IsNullOrWhiteSpace(apiToken))
        {
            _http.DefaultRequestHeaders.Remove("X-Robot-Token");
            _http.DefaultRequestHeaders.TryAddWithoutValidation("X-Robot-Token", apiToken.Trim());
        }
    }

    public Task<RobotManagerHealth> GetHealthAsync(CancellationToken cancellationToken = default) =>
        GetAsync<RobotManagerHealth>("/v1/health", cancellationToken);

    public Task<RobotManagerStatus> GetStatusAsync(CancellationToken cancellationToken = default) =>
        GetAsync<RobotManagerStatus>("/v1/status", cancellationToken);

    public Task<RobotManagerOperation> GetOperationAsync(CancellationToken cancellationToken = default) =>
        GetAsync<RobotManagerOperation>("/v1/operation", cancellationToken);

    public async Task ConfigureAsync(RobotManagerConfig config, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(config);
        ValidateConfig(config);
        var body = new
        {
            updateServerUrl = config.UpdateServerUrl,
            componentToken = config.ComponentToken,
            apiToken = config.ApiToken,
            autoStart = config.AutoStart,
            gvxApiPort = config.GvxApiPort,
            httpCallbackUrl = config.HttpCallbackUrl,
            startTimeoutSeconds = config.StartTimeoutSeconds,
        };
        using var response = await _http.PutAsJsonAsync("/v1/config", body, _json, cancellationToken);
        await EnsureSuccessAsync(response, cancellationToken);
    }

    public Task<RobotManagerStatus> ReconcileAsync(CancellationToken cancellationToken = default) =>
        PostStatusAsync("/v1/reconcile", cancellationToken);

    public Task<RobotManagerStatus> StartAsync(CancellationToken cancellationToken = default) =>
        PostStatusAsync("/v1/start", cancellationToken);

    public Task<RobotManagerStatus> StopAsync(CancellationToken cancellationToken = default) =>
        PostStatusAsync("/v1/stop", cancellationToken);

    public Task<RobotManagerStatus> RestartAsync(CancellationToken cancellationToken = default) =>
        PostStatusAsync("/v1/restart", cancellationToken);

    public Task<RobotManagerStatus> RepairAsync(CancellationToken cancellationToken = default) =>
        PostStatusAsync("/v1/repair", cancellationToken);

    public async Task ShutdownHostAsync(CancellationToken cancellationToken = default)
    {
        using var response = await _http.PostAsync("/v1/host/shutdown", null, cancellationToken);
        await EnsureSuccessAsync(response, cancellationToken);
    }

    async Task<T> GetAsync<T>(string path, CancellationToken cancellationToken)
    {
        using var response = await _http.GetAsync(path, cancellationToken);
        return await ReadAsync<T>(response, cancellationToken);
    }

    async Task<RobotManagerStatus> PostStatusAsync(string path, CancellationToken cancellationToken)
    {
        using var response = await _http.PostAsync(path, null, cancellationToken);
        return await ReadAsync<RobotManagerStatus>(response, cancellationToken);
    }

    async Task<T> ReadAsync<T>(HttpResponseMessage response, CancellationToken cancellationToken)
    {
        var body = await response.Content.ReadAsStringAsync(cancellationToken);
        if (!response.IsSuccessStatusCode)
            throw new RobotManagerException(response.StatusCode, body);
        return JsonSerializer.Deserialize<T>(body, _json)
               ?? throw new InvalidOperationException("RobotManager 返回空 JSON");
    }

    static async Task EnsureSuccessAsync(HttpResponseMessage response, CancellationToken cancellationToken)
    {
        if (response.IsSuccessStatusCode) return;
        var body = await response.Content.ReadAsStringAsync(cancellationToken);
        throw new RobotManagerException(response.StatusCode, body);
    }

    static void ValidateConfig(RobotManagerConfig config)
    {
        if (!Uri.TryCreate(config.UpdateServerUrl, UriKind.Absolute, out _))
            throw new ArgumentException("更新服务器地址必须是绝对 URL", nameof(config));
        if (config.GvxApiPort is < 1 or > 65535)
            throw new ArgumentOutOfRangeException(nameof(config), "GVx API 端口必须在 1-65535 之间");
        if (!Uri.TryCreate(config.HttpCallbackUrl, UriKind.Absolute, out var callback)
            || callback.Scheme != Uri.UriSchemeHttp)
            throw new ArgumentException("消息回调地址必须是绝对 HTTP URL", nameof(config));
        if (config.StartTimeoutSeconds is < 1 or > 600)
            throw new ArgumentOutOfRangeException(nameof(config), "启动超时必须在 1-600 秒之间");
    }

    public void Dispose()
    {
        if (_ownsHttpClient) _http.Dispose();
    }
}

public sealed class RobotManagerException : HttpRequestException
{
    public RobotManagerException(System.Net.HttpStatusCode statusCode, string responseBody)
        : base($"RobotManager HTTP {(int)statusCode}: {Shorten(responseBody)}", null, statusCode)
    {
        ResponseBody = responseBody;
    }

    public string ResponseBody { get; }

    static string Shorten(string value)
    {
        var trimmed = string.IsNullOrWhiteSpace(value) ? "(empty)" : value.Trim();
        return trimmed[..Math.Min(512, trimmed.Length)];
    }
}
