using System.Net;

namespace CVX.Manager;

public sealed record RobotClientRegistration(
    string ClientId,
    string InstanceId,
    string CallbackUrl);

public sealed record RobotClientSession(
    bool Ok,
    string ClientId,
    string InstanceId,
    bool CallbackReady);

public sealed class RobotClientSessionStore : IDisposable
{
    readonly object _gate = new();
    readonly HttpClient _http = new(new SocketsHttpHandler { UseProxy = false })
    {
        Timeout = TimeSpan.FromSeconds(2),
    };
    RobotClientRegistration? _current;

    public async Task<RobotClientSession> RegisterAsync(
        RobotClientRegistration registration,
        CancellationToken cancellationToken)
    {
        var normalized = Normalize(registration);
        if (!await ProbeCallbackAsync(normalized.CallbackUrl, cancellationToken))
            throw new InvalidOperationException("客户端回调地址不可达");

        lock (_gate) _current = normalized;
        return new RobotClientSession(
            true,
            normalized.ClientId,
            normalized.InstanceId,
            CallbackReady: true);
    }

    public bool Unregister(string clientId, string instanceId)
    {
        lock (_gate)
        {
            if (_current is null
                || !string.Equals(_current.ClientId, clientId?.Trim(), StringComparison.Ordinal)
                || !string.Equals(_current.InstanceId, instanceId?.Trim(), StringComparison.Ordinal))
                return false;
            _current = null;
            return true;
        }
    }

    public async Task<bool> IsCallbackReadyAsync(
        string configuredCallbackUrl,
        CancellationToken cancellationToken)
    {
        RobotClientRegistration? current;
        lock (_gate) current = _current;
        return current is not null
               && string.Equals(current.CallbackUrl, configuredCallbackUrl, StringComparison.OrdinalIgnoreCase)
               && await ProbeCallbackAsync(current.CallbackUrl, cancellationToken);
    }

    async Task<bool> ProbeCallbackAsync(string callbackUrl, CancellationToken cancellationToken)
    {
        try
        {
            using var response = await _http.GetAsync(callbackUrl, cancellationToken);
            return response.IsSuccessStatusCode;
        }
        catch (Exception exception) when (
            exception is HttpRequestException or OperationCanceledException
            && !cancellationToken.IsCancellationRequested)
        {
            return false;
        }
    }

    static RobotClientRegistration Normalize(RobotClientRegistration registration)
    {
        ArgumentNullException.ThrowIfNull(registration);
        var clientId = registration.ClientId?.Trim();
        var instanceId = registration.InstanceId?.Trim();
        var callbackUrl = registration.CallbackUrl?.Trim();
        if (string.IsNullOrWhiteSpace(clientId))
            throw new InvalidOperationException("客户端 ID 不能为空");
        if (string.IsNullOrWhiteSpace(instanceId))
            throw new InvalidOperationException("客户端实例 ID 不能为空");
        if (!Uri.TryCreate(callbackUrl, UriKind.Absolute, out var callback)
            || callback.Scheme != Uri.UriSchemeHttp
            || !IsLoopback(callback.Host))
            throw new InvalidOperationException("客户端回调地址必须是本机 HTTP 地址");
        return new RobotClientRegistration(clientId, instanceId, callback.AbsoluteUri);
    }

    static bool IsLoopback(string host) =>
        host.Equals("localhost", StringComparison.OrdinalIgnoreCase)
        || (IPAddress.TryParse(host, out var address) && IPAddress.IsLoopback(address));

    public void Dispose() => _http.Dispose();
}
