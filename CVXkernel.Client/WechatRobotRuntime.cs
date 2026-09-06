namespace CVXkernel.Client;

/// <summary>
/// Coordinates the application-owned callback receiver with the independently
/// running lifecycle manager. GVx operations remain the application's concern.
/// </summary>
public sealed class WechatRobotRuntime : IAsyncDisposable, IDisposable
{
    readonly RobotManagerBootstrapper _bootstrapper;
    readonly WechatCallbackReceiver _receiver;
    readonly SemaphoreSlim _lifecycleGate = new(1, 1);
    readonly string _ownershipLockPath;
    readonly string _clientId;
    readonly string _instanceId = Guid.NewGuid().ToString("N");
    readonly TimeSpan _clientReadyTimeout;
    RuntimeOwnershipLease? _ownershipLease;
    bool _registered;
    bool _started;
    bool _disposed;

    public WechatRobotRuntime(RobotManagerBootstrapperOptions? bootstrapperOptions = null)
    {
        var options = bootstrapperOptions ?? new RobotManagerBootstrapperOptions();
        _bootstrapper = new RobotManagerBootstrapper(options);
        _receiver = new WechatCallbackReceiver();
        _ownershipLockPath = Path.Combine(options.RootDirectory, ".runtime-owner.lock");
        _clientId = options.ClientId.Trim();
        _clientReadyTimeout = options.ClientReadyTimeout;
    }

    public IRobotManagerClient Manager => _bootstrapper.Client;
    public WechatCallbackReceiver Messages => _receiver;

    public async Task<RobotManagerStatus> EnsureRunningAsync(
        RobotManagerConfig config,
        WechatCallbackReceiverOptions? callbackOptions = null,
        IProgress<int>? progress = null,
        CancellationToken cancellationToken = default) =>
        await ConnectAsync(config, callbackOptions, progress, cancellationToken);

    /// <summary>
    /// Connects this client instance to the persistent manager and returns only
    /// after WeChat, GVx, and the callback route are ready.
    /// </summary>
    public async Task<RobotManagerStatus> ConnectAsync(
        RobotManagerConfig config,
        WechatCallbackReceiverOptions? callbackOptions = null,
        IProgress<int>? progress = null,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(config);
        await _lifecycleGate.WaitAsync(cancellationToken);
        var acquiredHere = false;
        var startedHere = false;
        try
        {
            ObjectDisposedException.ThrowIf(_disposed, this);
            if (_ownershipLease is null)
            {
                _ownershipLease = RuntimeOwnershipLease.Acquire(_ownershipLockPath);
                acquiredHere = true;
            }

            startedHere = !_receiver.IsRunning;
            var callbackUrl = _receiver.CallbackUrl ?? _receiver.Start(callbackOptions);
            var effectiveConfig = config with { HttpCallbackUrl = callbackUrl.ToString() };
            var registration = new RobotClientRegistration(
                _clientId,
                _instanceId,
                callbackUrl.ToString());
            var status = await _bootstrapper.ConnectClientAsync(
                effectiveConfig,
                registration,
                progress,
                cancellationToken);
            _registered = true;
            status = await WaitForReadyAsync(status, cancellationToken);
            _started = true;
            return status;
        }
        catch
        {
            try
            {
                await UnregisterClientBestEffortAsync();
                if (startedHere) await _receiver.StopAsync();
            }
            finally
            {
                if (acquiredHere) ReleaseOwnership();
            }
            throw;
        }
        finally
        {
            _lifecycleGate.Release();
        }
    }

    async Task<RobotManagerStatus> WaitForReadyAsync(
        RobotManagerStatus status,
        CancellationToken cancellationToken)
    {
        var deadline = DateTimeOffset.UtcNow.Add(_clientReadyTimeout);
        while (!IsReady(status) && DateTimeOffset.UtcNow < deadline)
        {
            if (!string.IsNullOrWhiteSpace(status.LastError)
                && string.Equals(status.RuntimeState, "failed", StringComparison.OrdinalIgnoreCase))
                throw new InvalidOperationException("RobotManager 恢复失败：" + status.LastError);
            await Task.Delay(250, cancellationToken);
            status = await _bootstrapper.Client.GetStatusAsync(cancellationToken);
        }

        if (!IsReady(status))
            throw new TimeoutException(
                $"RobotManager 恢复超时：state={status.RuntimeState}, "
                + $"wechat={status.WechatRunning}, gvx={status.GvxApiReady}, "
                + $"callback={status.HttpCallbackReady}");
        return status;
    }

    static bool IsReady(RobotManagerStatus status) =>
        status.Ok
        && status.WechatRunning
        && status.GvxApiReady
        && status.HttpCallbackReady;

    async Task UnregisterClientBestEffortAsync()
    {
        if (!_registered) return;
        _registered = false;
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(2));
        try
        {
            await _bootstrapper.Client.UnregisterClientAsync(_clientId, _instanceId, timeout.Token);
        }
        catch
        {
            // The manager may already be unavailable; instance matching prevents stale cleanup.
        }
    }

    public async ValueTask DisposeAsync()
    {
        await _lifecycleGate.WaitAsync();
        try
        {
            if (_disposed) return;
            _disposed = true;
            try
            {
                await UnregisterClientBestEffortAsync();
                await _receiver.DisposeAsync();
            }
            finally
            {
                try
                {
                    _bootstrapper.Dispose();
                }
                finally
                {
                    ReleaseOwnership();
                    GC.SuppressFinalize(this);
                }
            }
        }
        finally
        {
            _lifecycleGate.Release();
        }
    }

    public void Dispose()
    {
        _lifecycleGate.Wait();
        try
        {
            if (_disposed) return;
            _disposed = true;
            try
            {
                UnregisterClientBestEffortAsync().GetAwaiter().GetResult();
                if (_started || _receiver.IsRunning) _receiver.Dispose();
            }
            finally
            {
                try
                {
                    _bootstrapper.Dispose();
                }
                finally
                {
                    ReleaseOwnership();
                    GC.SuppressFinalize(this);
                }
            }
        }
        finally
        {
            _lifecycleGate.Release();
        }
    }

    void ReleaseOwnership()
    {
        _ownershipLease?.Dispose();
        _ownershipLease = null;
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
                "已有其他客户端正在管理微信机器人；当前程序仍可通过 RobotManagerClient 或 GVxClient 调用已运行的服务。",
                exception);
        }
    }

    public void Dispose() => _stream.Dispose();
}
