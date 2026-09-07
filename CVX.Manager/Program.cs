using System.Reflection;
using CVX.Manager;

var bootstrap = WebApplication.CreateBuilder(args);
var initial = bootstrap.Configuration.GetSection("RobotManager").Get<RobotManagerSettings>() ?? new();
if (!Uri.TryCreate(initial.ListenUrl, UriKind.Absolute, out var listenUri)
    || !listenUri.IsLoopback || listenUri.Scheme is not ("http" or "https"))
    throw new InvalidOperationException("Manager 仅允许监听本机回环地址");
bootstrap.WebHost.UseUrls(initial.ListenUrl);

bootstrap.Services.AddSingleton<RobotManagerSettingsStore>();
bootstrap.Services.AddSingleton<RobotClientSessionStore>();
bootstrap.Services.AddSingleton<RobotLifecycleService>();
bootstrap.Services.AddHostedService<RobotAutoStartService>();

var app = bootstrap.Build();

app.Use(async (context, next) =>
{
    try
    {
        await next();
    }
    catch (OperationCanceledException) when (context.RequestAborted.IsCancellationRequested)
    {
    }
    catch (Exception ex)
    {
        context.Response.StatusCode = ex switch
        {
            UnauthorizedAccessException => StatusCodes.Status401Unauthorized,
            InvalidOperationException => StatusCodes.Status409Conflict,
            TimeoutException => StatusCodes.Status504GatewayTimeout,
            _ => StatusCodes.Status500InternalServerError,
        };
        await context.Response.WriteAsJsonAsync(new { ok = false, error = ex.Message });
    }
});

app.MapGet("/v1/health", () => Results.Ok(new
{
    ok = true,
    service = "CVX.Manager",
    version = ProductVersion(),
    protocolVersion = 2,
    processId = Environment.ProcessId,
}));
app.MapGet("/v1/status", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.GetStatusAsync(ct)));
app.MapGet("/v1/operation", (RobotLifecycleService robot) =>
    Results.Ok(robot.GetOperationProgress()));

app.MapPut("/v1/client", async (
    RobotClientRegistration registration,
    RobotClientSessionStore clients,
    CancellationToken ct) => Results.Ok(await clients.RegisterAsync(registration, ct)));
app.MapDelete("/v1/client", (
    string clientId,
    string instanceId,
    RobotClientSessionStore clients) => Results.Ok(new
    {
        ok = true,
        removed = clients.Unregister(clientId, instanceId),
    }));

app.MapPost("/v1/initialize", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.InitializeAsync(ct)));
app.MapPost("/v1/reconcile", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.ReconcileAsync(ct)));
app.MapPost("/v1/start", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.StartAsync(ct)));
app.MapPost("/v1/stop", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.StopAsync(ct)));
app.MapPost("/v1/restart", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.RestartAsync(ct)));
app.MapPost("/v1/repair", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.RepairAsync(ct)));

app.MapGet("/v1/component/update", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.CheckUpdateAsync(ct)));
app.MapPost("/v1/component/update", async (RobotLifecycleService robot, CancellationToken ct) =>
    Results.Ok(await robot.UpdateAsync(ct)));

app.MapGet("/v1/config", (RobotManagerSettingsStore store) => Results.Ok(store.GetPublicSettings()));
app.MapPut("/v1/config", async (
    RobotManagerSettingsUpdate update,
    RobotManagerSettingsStore store,
    CancellationToken ct) => Results.Ok(await store.UpdateAsync(update, ct)));

app.MapPost("/v1/host/shutdown", (IHostApplicationLifetime lifetime) =>
{
    _ = Task.Run(async () =>
    {
        await Task.Delay(250);
        lifetime.StopApplication();
    });
    return Results.Ok(new { ok = true, stopping = true, wechatUntouched = true });
});

await app.RunAsync();

static string ProductVersion()
{
    var version = Assembly.GetEntryAssembly()?.GetName().Version;
    return version is null ? "0.0.0" : $"{version.Major}.{version.Minor}.{Math.Max(version.Build, 0)}";
}
