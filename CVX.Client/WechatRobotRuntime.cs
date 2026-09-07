using System.Text.Json;

namespace CVX.Client;

/// <summary>Local kernel connection and component installation settings.</summary>
public sealed record WechatRobotOptions
{
    public string? UpdateServerUrl { get; init; }
    public string? ComponentToken { get; init; }
    public string ComponentDirectory { get; init; } = Path.Combine(
        Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "Sao", "WechatRobot");
    public int GvxApiPort { get; init; } = 19088;
    public string CallbackUrl { get; init; } = WechatCallbackReceiver.DefaultCallbackUrl;
    public TimeSpan ConnectTimeout { get; init; } = TimeSpan.FromMinutes(12);
    public TimeSpan StartTimeout { get; init; } = TimeSpan.FromSeconds(30);
}

/// <summary>Connects or recovers the kernel in-process. Disposal leaves WeChat running.</summary>
public sealed class WechatRobotRuntime : IAsyncDisposable
{
    readonly WechatRobotOptions _options;
    readonly WechatLauncher _launcher;
    readonly WechatCallbackReceiver _receiver = new();
    readonly SemaphoreSlim _lifecycleGate = new(1, 1);
    readonly CancellationTokenSource _lifetime = new();
    readonly object _disposeGate = new();
    RuntimeOwnershipLease? _ownershipLease;
    Task? _disposeTask;

    public WechatRobotRuntime(WechatRobotOptions? options = null) : this(options ?? new(), null) { }

    internal WechatRobotRuntime(WechatRobotOptions options, IWechatProcessHost? processes)
    {
        _options = ResolveOptions(options);
        _launcher = new(_options, processes);
        _receiver.MessageReceived += (_, message) => Publish(MessageReceived, message);
        _receiver.ReceiveError += (_, error) => PublishError(error);
    }

    public event EventHandler<WechatCallbackMessage>? MessageReceived;
    public event EventHandler<Exception>? ReceiveError;

    /// <summary>Reuses a healthy kernel or reinstalls missing components and launches it. Safe to retry after failure.</summary>
    public Task<WechatRobotStatus> ConnectAsync(CancellationToken cancellationToken = default) =>
        RunAsync(restart: false, cancellationToken);

    /// <summary>Stops only the managed WeChat process and injects again, retaining message subscriptions.</summary>
    public Task<WechatRobotStatus> RestartAsync(CancellationToken cancellationToken = default) =>
        RunAsync(restart: true, cancellationToken);

    async Task<WechatRobotStatus> RunAsync(bool restart, CancellationToken cancellationToken)
    {
        lock (_disposeGate) ObjectDisposedException.ThrowIf(_disposeTask is not null, this);
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken, _lifetime.Token);
        timeout.CancelAfter(_options.ConnectTimeout);
        var ct = timeout.Token;
        await _lifecycleGate.WaitAsync(ct).ConfigureAwait(false);
        try
        {
            ct.ThrowIfCancellationRequested();
            try
            {
                _ownershipLease ??= RuntimeOwnershipLease.Acquire(Path.Combine(_options.ComponentDirectory, ".runtime-owner.lock"));
                var callback = _receiver.CallbackUrl ?? _receiver.Start(new() { CallbackUrl = _options.CallbackUrl });
                return await _launcher.ConnectAsync(callback, restart, ct).ConfigureAwait(false);
            }
            catch
            {
                await DisconnectAsync().ConfigureAwait(false);
                throw;
            }
        }
        finally { _lifecycleGate.Release(); }
    }

    async Task DisconnectAsync()
    {
        try { await _receiver.StopAsync().ConfigureAwait(false); }
        finally { _ownershipLease?.Dispose(); _ownershipLease = null; }
    }

    public ValueTask DisposeAsync()
    {
        lock (_disposeGate) return new ValueTask(_disposeTask ??= DisposeCoreAsync());
    }

    async Task DisposeCoreAsync()
    {
        _lifetime.Cancel();
        await _lifecycleGate.WaitAsync().ConfigureAwait(false);
        try { await DisconnectAsync().ConfigureAwait(false); }
        finally { _launcher.Dispose(); _lifecycleGate.Release(); }
    }

    void Publish<T>(EventHandler<T>? handlers, T value)
    {
        if (handlers is null) return;
        foreach (var handler in handlers.GetInvocationList().Cast<EventHandler<T>>())
        {
            try { handler(this, value); }
            catch (Exception ex) { PublishError(ex); }
        }
    }

    void PublishError(Exception error)
    {
        if (ReceiveError is not { } handlers) return;
        foreach (var handler in handlers.GetInvocationList().Cast<EventHandler<Exception>>())
            try { handler(this, error); } catch { }
    }

    static WechatRobotOptions ResolveOptions(WechatRobotOptions options)
    {
        if (string.IsNullOrWhiteSpace(options.ComponentDirectory)) throw new ArgumentException("组件目录不能为空", nameof(options));
        if (options.GvxApiPort is < 1 or > 65535) throw new ArgumentOutOfRangeException(nameof(options), "内核端口无效");
        foreach (var timeout in new[] { options.ConnectTimeout, options.StartTimeout })
            if (timeout <= TimeSpan.Zero || timeout.TotalMilliseconds > uint.MaxValue - 1)
                throw new ArgumentOutOfRangeException(nameof(options), "超时必须是有效的正时间间隔");
        options = options with { ComponentDirectory = Path.GetFullPath(options.ComponentDirectory) };
        // Read only the old component credentials. No daemon state or manager endpoint survives migration.
        var legacy = Path.Combine(options.ComponentDirectory, "manager-settings.json");
        if ((options.UpdateServerUrl is null || options.ComponentToken is null) && File.Exists(legacy))
        {
            using var document = JsonDocument.Parse(File.ReadAllText(legacy));
            var root = document.RootElement;
            options = options with
            {
                UpdateServerUrl = options.UpdateServerUrl ?? (root.TryGetProperty("updateServerUrl", out var url) ? url.GetString() : null),
                ComponentToken = options.ComponentToken ?? (root.TryGetProperty("componentToken", out var token) ? token.GetString() : null),
            };
        }
        if (options.UpdateServerUrl is not null &&
            (!Uri.TryCreate(options.UpdateServerUrl, UriKind.Absolute, out var uri) || uri.Scheme is not ("http" or "https")))
            throw new ArgumentException("组件更新地址必须是 HTTP 或 HTTPS URL", nameof(options));
        return options;
    }
}

internal sealed class RuntimeOwnershipLease : IDisposable
{
    readonly FileStream _stream;
    RuntimeOwnershipLease(FileStream stream) => _stream = stream;
    internal static RuntimeOwnershipLease Acquire(string path)
    {
        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        try { return new(new FileStream(path, FileMode.OpenOrCreate, FileAccess.ReadWrite, FileShare.None)); }
        catch (IOException ex) { throw new InvalidOperationException("已有客户端占用此组件目录，请先关闭其连接", ex); }
    }
    public void Dispose() => _stream.Dispose();
}
