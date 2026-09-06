using System.Net;
using System.Net.Sockets;
using System.Text;

namespace CVXkernel.Client;

public sealed record WechatCallbackReceiverOptions
{
    public string CallbackUrl { get; init; } = WechatCallbackReceiver.DefaultCallbackUrl;
    public int MaxHeaderBytes { get; init; } = 64 * 1024;
    public int MaxBodyBytes { get; init; } = 1024 * 1024;
}

/// <summary>Loopback-only HTTP receiver for message callbacks from the WeChat component.</summary>
public sealed class WechatCallbackReceiver : IAsyncDisposable, IDisposable
{
    public const string DefaultCallbackUrl = "http://127.0.0.1:5000/api/recvMsg";

    readonly object _gate = new();
    CancellationTokenSource? _cts;
    TcpListener? _listener;
    Task? _acceptLoop;
    string _path = "/api/recvMsg";
    int _maxHeaderBytes;
    int _maxBodyBytes;
    long _receivedCount;

    public Uri? CallbackUrl { get; private set; }
    public long ReceivedCount => Interlocked.Read(ref _receivedCount);
    public bool IsRunning
    {
        get { lock (_gate) return _listener is not null; }
    }

    public event EventHandler<WechatCallbackMessage>? MessageReceived;
    public event EventHandler<Exception>? ReceiveError;

    public Uri Start(WechatCallbackReceiverOptions? options = null)
    {
        options ??= new WechatCallbackReceiverOptions();
        var requested = ValidateOptions(options);

        lock (_gate)
        {
            if (_listener is not null)
                throw new InvalidOperationException("微信消息回调服务已经启动");

            _path = string.IsNullOrWhiteSpace(requested.AbsolutePath) ? "/api/recvMsg" : requested.AbsolutePath;
            _maxHeaderBytes = options.MaxHeaderBytes;
            _maxBodyBytes = options.MaxBodyBytes;
            var listener = new TcpListener(IPAddress.Loopback, requested.Port);
            listener.Start();
            var cancellation = new CancellationTokenSource();

            _listener = listener;
            _cts = cancellation;
            var actualPort = ((IPEndPoint)listener.LocalEndpoint).Port;
            CallbackUrl = new UriBuilder(Uri.UriSchemeHttp, "127.0.0.1", actualPort, _path).Uri;
            _acceptLoop = AcceptLoopAsync(listener, cancellation.Token);
            return CallbackUrl;
        }
    }

    public async Task StopAsync()
    {
        CancellationTokenSource? cancellation;
        TcpListener? listener;
        Task? acceptLoop;
        lock (_gate)
        {
            cancellation = _cts;
            listener = _listener;
            acceptLoop = _acceptLoop;
            _cts = null;
            _listener = null;
            _acceptLoop = null;
            CallbackUrl = null;
        }

        if (listener is null) return;
        try { cancellation?.Cancel(); } catch { }
        try { listener.Stop(); } catch { }
        if (acceptLoop is not null)
        {
            try { await acceptLoop; }
            catch (OperationCanceledException) { }
        }
        cancellation?.Dispose();
    }

    async Task AcceptLoopAsync(TcpListener listener, CancellationToken cancellationToken)
    {
        while (!cancellationToken.IsCancellationRequested)
        {
            try
            {
                var client = await listener.AcceptTcpClientAsync(cancellationToken);
                _ = HandleClientAsync(client, cancellationToken);
            }
            catch (Exception exception) when (
                exception is OperationCanceledException or SocketException
                && cancellationToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception exception)
            {
                PublishError(exception);
            }
        }
    }

    async Task HandleClientAsync(TcpClient client, CancellationToken cancellationToken)
    {
        using (client)
        await using (var stream = client.GetStream())
        {
            try
            {
                var headerBytes = await ReadHeadersAsync(stream, cancellationToken);
                var lines = Encoding.ASCII.GetString(headerBytes)
                    .Split("\r\n", StringSplitOptions.RemoveEmptyEntries);
                var requestParts = lines.FirstOrDefault()?.Split(' ', StringSplitOptions.RemoveEmptyEntries) ?? [];
                var method = requestParts.ElementAtOrDefault(0);
                var target = requestParts.ElementAtOrDefault(1);
                var path = target is null ? null : new Uri(new Uri("http://127.0.0.1"), target).AbsolutePath;
                var contentLength = ParseContentLength(lines);

                if (!string.Equals(path, _path, StringComparison.OrdinalIgnoreCase))
                {
                    await WriteResponseAsync(stream, 404, "{\"ok\":false}", cancellationToken);
                    return;
                }
                if (method == "GET")
                {
                    await WriteResponseAsync(stream, 200, "{\"ok\":true}", cancellationToken);
                    return;
                }
                if (method is not ("PUT" or "POST"))
                {
                    await WriteResponseAsync(stream, 404, "{\"ok\":false}", cancellationToken);
                    return;
                }
                if (contentLength is < 0 || contentLength > _maxBodyBytes)
                {
                    await WriteResponseAsync(stream, 413, "{\"ok\":false}", cancellationToken);
                    return;
                }

                var body = new byte[contentLength];
                await stream.ReadExactlyAsync(body, cancellationToken);
                var message = WechatCallbackMessageParser.Parse(body);

                // Acknowledge the component before user handlers run.
                await WriteResponseAsync(stream, 200, "{\"ok\":true}", cancellationToken);
                Interlocked.Increment(ref _receivedCount);
                _ = Task.Run(() => PublishMessage(message), CancellationToken.None);
            }
            catch (Exception exception) when (!cancellationToken.IsCancellationRequested)
            {
                try { await WriteResponseAsync(stream, 400, "{\"ok\":false}", cancellationToken); }
                catch { }
                PublishError(exception);
            }
        }
    }

    void PublishMessage(WechatCallbackMessage message)
    {
        var handlers = MessageReceived?.GetInvocationList();
        if (handlers is null) return;
        foreach (var handler in handlers.Cast<EventHandler<WechatCallbackMessage>>())
        {
            try { handler(this, message); }
            catch (Exception exception) { PublishError(exception); }
        }
    }

    void PublishError(Exception exception)
    {
        var handlers = ReceiveError?.GetInvocationList();
        if (handlers is null) return;
        foreach (var handler in handlers.Cast<EventHandler<Exception>>())
        {
            try { handler(this, exception); }
            catch { }
        }
    }

    async Task<byte[]> ReadHeadersAsync(NetworkStream stream, CancellationToken cancellationToken)
    {
        var bytes = new List<byte>(1024);
        var one = new byte[1];
        while (bytes.Count < _maxHeaderBytes)
        {
            if (await stream.ReadAsync(one, cancellationToken) == 0)
                throw new IOException("HTTP request ended before headers completed.");
            bytes.Add(one[0]);
            var count = bytes.Count;
            if (count >= 4 && bytes[count - 4] == 13 && bytes[count - 3] == 10
                           && bytes[count - 2] == 13 && bytes[count - 1] == 10)
                return bytes.ToArray();
        }
        throw new IOException("HTTP request headers are too large.");
    }

    static int ParseContentLength(IEnumerable<string> lines)
    {
        foreach (var line in lines)
        {
            if (line.StartsWith("Content-Length:", StringComparison.OrdinalIgnoreCase)
                && int.TryParse(line[(line.IndexOf(':') + 1)..].Trim(), out var length))
                return length;
        }
        return 0;
    }

    static async Task WriteResponseAsync(
        NetworkStream stream,
        int status,
        string body,
        CancellationToken cancellationToken)
    {
        var payload = Encoding.UTF8.GetBytes(body);
        var reason = status switch
        {
            200 => "OK",
            400 => "Bad Request",
            404 => "Not Found",
            413 => "Payload Too Large",
            _ => "Error",
        };
        var headers = Encoding.ASCII.GetBytes(
            $"HTTP/1.1 {status} {reason}\r\nContent-Type: application/json; charset=utf-8\r\n" +
            $"Content-Length: {payload.Length}\r\nConnection: close\r\n\r\n");
        await stream.WriteAsync(headers, cancellationToken);
        await stream.WriteAsync(payload, cancellationToken);
    }

    static Uri ValidateOptions(WechatCallbackReceiverOptions options)
    {
        if (!Uri.TryCreate(options.CallbackUrl, UriKind.Absolute, out var callback)
            || callback.Scheme != Uri.UriSchemeHttp
            || !IsLoopback(callback.Host))
            throw new ArgumentException("回调地址必须是本机 HTTP 地址", nameof(options));
        if (options.MaxHeaderBytes is < 1024 or > 1024 * 1024)
            throw new ArgumentOutOfRangeException(nameof(options), "最大请求头必须在 1 KiB-1 MiB 之间");
        if (options.MaxBodyBytes is < 1 or > 16 * 1024 * 1024)
            throw new ArgumentOutOfRangeException(nameof(options), "最大请求体必须在 1 B-16 MiB 之间");
        return callback;
    }

    static bool IsLoopback(string host) =>
        host.Equals("localhost", StringComparison.OrdinalIgnoreCase)
        || (IPAddress.TryParse(host, out var address) && IPAddress.IsLoopback(address));

    public void Dispose() => StopAsync().GetAwaiter().GetResult();

    public async ValueTask DisposeAsync()
    {
        await StopAsync();
        GC.SuppressFinalize(this);
    }
}
