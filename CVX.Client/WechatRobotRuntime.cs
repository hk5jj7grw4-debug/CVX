using System.Text.Json;

namespace CVX.Client;

/// <summary>Local kernel connection and component installation settings.</summary>
public sealed record WechatRobotOptions
{
    public string? UpdateServerUrl { get; init; }
    public string? ComponentToken { get; init; }
    public string ComponentDirectory { get; init; } = Path.Combine(
        Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "CVX");
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
    Task? _monitor;
    WechatRobotStatus _status;
    public WechatRobotStatus Status => Volatile.Read(ref _status);
    public event EventHandler<WechatRobotStatus>? StatusChanged;

    void SetStatus(WechatRobotStatus status)
    {
        Volatile.Write(ref _status, status);
        Publish(StatusChanged, status);
    }

    void Report(WechatRuntimePhase phase, string? version, string? component, double? progress)
    {
        SetStatus(Status with
        {
            Phase = phase, ApiReady = false, IsLoggedIn = false, Error = null,
            DetectedWechatVersion = version, ComponentVersion = component ?? Status.ComponentVersion,
            VersionMatch = version is null ? WechatVersionMatch.Unknown :
                version == ComponentInstaller.RequiredWechatVersion ? WechatVersionMatch.Matching : WechatVersionMatch.Mismatched,
            DownloadProgress = progress, CheckedAt = DateTimeOffset.UtcNow
        });
    }

    public WechatRobotRuntime(WechatRobotOptions? options = null) : this(options ?? new(), null) { }

    internal WechatRobotRuntime(WechatRobotOptions options, IWechatProcessHost? processes)
    {
        _options = ResolveOptions(options);
        _status = new(false, false, null, new Uri(_options.CallbackUrl));
        _launcher = new(_options, processes);
        _launcher.Report = Report;
        _receiver.MessageReceived += (_, message) => Publish(MessageReceived, message);
        _receiver.ReceiveError += (_, error) => PublishError(error);
    }

    public event EventHandler<WechatCallbackMessage>? MessageReceived;
    public event EventHandler<Exception>? ReceiveError;

    /// <summary>Reuses a healthy kernel or reinstalls missing components and launches it. Safe to retry after failure.</summary>
    public Task<WechatRobotStatus> ConnectAsync(CancellationToken cancellationToken = default)
    {
        lock (_disposeGate)
        {
            ObjectDisposedException.ThrowIf(_disposeTask is not null, this);
            _monitor ??= Task.Run(MonitorAsync);
        }
        return RunAsync(restart: false, cancellationToken);
    }

    async Task MonitorAsync()
    {
        using var timer = new PeriodicTimer(TimeSpan.FromSeconds(3));
        try
        {
            while (await timer.WaitForNextTickAsync(_lifetime.Token).ConfigureAwait(false))
            {
                try { await RunAsync(false, _lifetime.Token, background: true).ConfigureAwait(false); }
                catch (OperationCanceledException) when (_lifetime.IsCancellationRequested) { break; }
                catch (ObjectDisposedException) { break; }
                catch { /* RunAsync publishes the failure; the next tick can retry. */ }
            }
        }
        catch (OperationCanceledException) when (_lifetime.IsCancellationRequested) { }
    }

    /// <summary>Stops only the managed WeChat process and injects again, retaining message subscriptions.</summary>
    public Task<WechatRobotStatus> RestartAsync(CancellationToken cancellationToken = default) =>
        RunAsync(restart: true, cancellationToken);

    async Task<WechatRobotStatus> RunAsync(bool restart, CancellationToken cancellationToken, bool background = false)
    {
        lock (_disposeGate) ObjectDisposedException.ThrowIf(_disposeTask is not null, this);
        using var timeout = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken, _lifetime.Token);
        timeout.CancelAfter(_options.ConnectTimeout);
        var ct = timeout.Token;
        if (background)
        {
            if (!await _lifecycleGate.WaitAsync(0, ct).ConfigureAwait(false)) return Status;
        }
        else await _lifecycleGate.WaitAsync(ct).ConfigureAwait(false);
        try
        {
            ct.ThrowIfCancellationRequested();
            try
            {
                _ownershipLease ??= RuntimeOwnershipLease.Acquire(Path.Combine(_options.ComponentDirectory, ".runtime-owner.lock"));
                var callback = _receiver.CallbackUrl ?? _receiver.Start(new() { CallbackUrl = _options.CallbackUrl });
                // File hashing and archive work must never run on the caller's UI thread.
                var result = await Task.Run(() => _launcher.ConnectAsync(callback, restart, ct), ct).ConfigureAwait(false);
                result = result with
                {
                    Phase = WechatRuntimePhase.Connected, CheckedAt = DateTimeOffset.UtcNow
                };
                SetStatus(result);
                return result;
            }
            catch (Exception ex)
            {
                SetStatus(Status with { ApiReady = false, IsLoggedIn = false, Phase = WechatRuntimePhase.Failed,
                    Error = ex.Message, DownloadProgress = null, CheckedAt = DateTimeOffset.UtcNow });
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
        if (_monitor is not null) await _monitor.ConfigureAwait(false);
        await _lifecycleGate.WaitAsync().ConfigureAwait(false);
        try { await DisconnectAsync().ConfigureAwait(false); }
        finally
        {
            _launcher.Dispose(); _lifecycleGate.Release();
            SetStatus(Status with { ApiReady = false, IsLoggedIn = false, Phase = WechatRuntimePhase.Disposed, DownloadProgress = null });
        }
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
