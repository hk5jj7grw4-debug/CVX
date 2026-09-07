namespace CVX.Client;

/// <summary>Connection settings. Existing Manager settings are preserved when optional values are omitted.</summary>
public sealed record WechatRobotOptions
{
    public Uri ManagerAddress { get; init; } = RobotManagerClient.DefaultBaseAddress;
    public string? UpdateServerUrl { get; init; }
    public string ClientId { get; init; } = System.Reflection.Assembly.GetEntryAssembly()?.GetName().Name ?? "cvx-client";
    public string InstallDirectory { get; init; } = Path.Combine(
        Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "Sao", "WechatRobotManagerHost");
    public string CallbackUrl { get; init; } = WechatCallbackReceiver.DefaultCallbackUrl;
    public TimeSpan ConnectTimeout { get; init; } = TimeSpan.FromMinutes(12);
}

/// <summary>Owns one callback connection. Disposal leaves Manager and WeChat running.</summary>
public sealed class WechatRobotRuntime : IAsyncDisposable
{
    readonly WechatRobotOptions _options;
    readonly RobotManagerBootstrapper _bootstrapper;
    readonly WechatCallbackReceiver _receiver = new();
    readonly SemaphoreSlim _lifecycleGate = new(1, 1);
    readonly CancellationTokenSource _lifetime = new();
    readonly object _disposeGate = new();
    readonly string _instanceId = Guid.NewGuid().ToString("N");
    RuntimeOwnershipLease? _ownershipLease;
    Task? _disposeTask;
    bool _registered;
    bool _connected;
    bool _disposed;

    public WechatRobotRuntime(WechatRobotOptions? options = null)
    {
        _options = options ?? new();
        if (_options.ConnectTimeout <= TimeSpan.Zero || _options.ConnectTimeout.TotalMilliseconds > uint.MaxValue - 1)
            throw new ArgumentOutOfRangeException(nameof(options), "连接超时必须为有效的正时间间隔");
        if (_options.UpdateServerUrl is not null &&
            (!Uri.TryCreate(_options.UpdateServerUrl, UriKind.Absolute, out var updateUri) ||
             updateUri.Scheme is not ("http" or "https")))
            throw new ArgumentException("更新地址必须是 HTTP 或 HTTPS URL", nameof(options));
        if (!_options.ManagerAddress.IsAbsoluteUri)
            throw new ArgumentException("Manager API 地址必须是绝对 URL", nameof(options));
        if (string.IsNullOrWhiteSpace(_options.InstallDirectory))
            throw new ArgumentException("Manager 安装目录不能为空", nameof(options));
        if (string.IsNullOrWhiteSpace(_options.ClientId))
            throw new ArgumentException("客户端 ID 不能为空", nameof(options));
        _bootstrapper = new(_options);
        _receiver.MessageReceived += (_, message) => MessageReceived?.Invoke(this, message);
        _receiver.ReceiveError += (_, error) => ReceiveError?.Invoke(this, error);
    }

    public event EventHandler<WechatCallbackMessage>? MessageReceived;
    public event EventHandler<Exception>? ReceiveError;

    public async Task<RobotManagerStatus> ConnectAsync(CancellationToken cancellationToken = default)
    {
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken, _lifetime.Token);
        timeout.CancelAfter(_options.ConnectTimeout);
        var ct = timeout.Token;
        await _lifecycleGate.WaitAsync(ct).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            if (_connected)
            {
                var current = await _bootstrapper.Client.GetStatusAsync(ct).ConfigureAwait(false);
                if (IsReady(current)) return current;
            }
            _connected = false;
            try
            {
                _ownershipLease ??= RuntimeOwnershipLease.Acquire(Path.Combine(_options.InstallDirectory, ".runtime-owner.lock"));
                var callback = _receiver.CallbackUrl ?? _receiver.Start(new() { CallbackUrl = _options.CallbackUrl });
                await _bootstrapper.EnsureManagerRunningAsync(_options.UpdateServerUrl, ct).ConfigureAwait(false);
                // Registration can succeed remotely even when its response is lost.
                _registered = true;
                var registration = new RobotClientRegistration(_options.ClientId.Trim(), _instanceId, callback.ToString());
                var session = await _bootstrapper.Client.RegisterClientAsync(registration, ct).ConfigureAwait(false);
                if (!session.Ok || session.ClientId != registration.ClientId || session.InstanceId != _instanceId)
                    throw new InvalidOperationException("Manager 没有确认当前客户端实例");
                await _bootstrapper.Client.ConfigureConnectionAsync(_options, callback.ToString(), ct).ConfigureAwait(false);
                var status = await _bootstrapper.Client.StartAsync(ct).ConfigureAwait(false);
                while (!IsReady(status))
                {
                    var operation = await _bootstrapper.Client.GetOperationAsync(ct).ConfigureAwait(false);
                    if (!operation.Running && operation.Phase == "failed")
                        throw new InvalidOperationException("机器人启动失败：" + operation.Error);
                    await Task.Delay(250, ct).ConfigureAwait(false);
                    status = await _bootstrapper.Client.GetStatusAsync(ct).ConfigureAwait(false);
                }
                _connected = true;
                return status;
            }
            catch
            {
                await DisconnectAsync().ConfigureAwait(false);
                throw;
            }
        }
        finally { _lifecycleGate.Release(); }
    }

    static bool IsReady(RobotManagerStatus status) =>
        status.Ok && status.WechatRunning && status.GvxApiReady && status.HttpCallbackReady;

    async Task DisconnectAsync()
    {
        _connected = false;
        try
        {
            if (_registered)
            {
                _registered = false;
                using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(2));
                try
                {
                    await _bootstrapper.Client.UnregisterClientAsync(
                        _options.ClientId.Trim(), _instanceId, timeout.Token).ConfigureAwait(false);
                }
                catch { /* Preserve the connection error; unregister is instance-scoped. */ }
            }
            await _receiver.StopAsync().ConfigureAwait(false);
        }
        finally
        {
            _ownershipLease?.Dispose();
            _ownershipLease = null;
        }
    }

    public ValueTask DisposeAsync()
    {
        lock (_disposeGate) return new ValueTask(_disposeTask ??= DisposeCoreAsync());
    }

    async Task DisposeCoreAsync()
    {
        _lifetime.Cancel();
        await _lifecycleGate.WaitAsync().ConfigureAwait(false);
        try
        {
            _disposed = true;
            await DisconnectAsync().ConfigureAwait(false);
        }
        finally
        {
            _bootstrapper.Dispose();
            _lifecycleGate.Release();
            GC.SuppressFinalize(this);
        }
    }
}

internal sealed class RuntimeOwnershipLease : IDisposable
{
    readonly FileStream _stream;

    RuntimeOwnershipLease(FileStream stream) => _stream = stream;

    public static RuntimeOwnershipLease Acquire(string lockPath)
    {
        var directory = Path.GetDirectoryName(lockPath)
            ?? throw new InvalidOperationException("运行时锁文件缺少目录");
        Directory.CreateDirectory(directory);
        try
        {
            var stream = new FileStream(
                lockPath,
                FileMode.OpenOrCreate,
                FileAccess.ReadWrite,
                FileShare.None);
            return new RuntimeOwnershipLease(stream);
        }
        catch (IOException exception)
        {
            throw new InvalidOperationException(
                "已有其他客户端正在管理微信机器人；当前程序仍可通过 RobotManagerClient 或 CVXClient 调用已运行的服务。",
                exception);
        }
    }

    public void Dispose() => _stream.Dispose();
}
