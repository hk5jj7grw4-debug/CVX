namespace CVXkernel.Client;

/// <summary>
/// Coordinates the application-owned callback receiver with the independently
/// running lifecycle manager. GVx operations remain the application's concern.
/// </summary>
public sealed class WechatRobotRuntime : IAsyncDisposable, IDisposable
{
    readonly RobotManagerBootstrapper _bootstrapper;
    readonly WechatCallbackReceiver _receiver;
    bool _started;

    public WechatRobotRuntime(RobotManagerBootstrapperOptions? bootstrapperOptions = null)
    {
        _bootstrapper = new RobotManagerBootstrapper(bootstrapperOptions);
        _receiver = new WechatCallbackReceiver();
    }

    public IRobotManagerClient Manager => _bootstrapper.Client;
    public WechatCallbackReceiver Messages => _receiver;

    public async Task<RobotManagerStatus> EnsureRunningAsync(
        RobotManagerConfig config,
        WechatCallbackReceiverOptions? callbackOptions = null,
        IProgress<int>? progress = null,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(config);
        var startedHere = !_receiver.IsRunning;
        var callbackUrl = _receiver.CallbackUrl ?? _receiver.Start(callbackOptions);
        try
        {
            var effectiveConfig = config with { HttpCallbackUrl = callbackUrl.ToString() };
            var status = await _bootstrapper.EnsureRunningAsync(effectiveConfig, progress, cancellationToken);
            _started = true;
            return status;
        }
        catch
        {
            if (startedHere) await _receiver.StopAsync();
            throw;
        }
    }

    public async ValueTask DisposeAsync()
    {
        await _receiver.DisposeAsync();
        _bootstrapper.Dispose();
        GC.SuppressFinalize(this);
    }

    public void Dispose()
    {
        if (_started || _receiver.IsRunning) _receiver.Dispose();
        _bootstrapper.Dispose();
        GC.SuppressFinalize(this);
    }
}
