using System.Text.Json;
using System.Text.Json.Serialization;

namespace CVX.Sdk;

/// <summary>Login, local database and middleware system APIs.</summary>
public sealed class SystemApi
{
    private readonly CVXClient _client;

    internal SystemApi(CVXClient client) => _client = client;

    public async Task<string> GetWechatBasePathAsync(CancellationToken cancellationToken = default)
    {
        var response = await _client.PostCommandAsync("/api/getwxbasepath", cancellationToken: cancellationToken).ConfigureAwait(false);
        return CVXClient.GetRequiredProperty(response, "data").GetString() ?? string.Empty;
    }

    public Task AutoLoginAsync(CancellationToken cancellationToken = default) =>
        CommandAsync("/api/auto_login", cancellationToken: cancellationToken);

    public Task DecryptDatabaseAsync(
        string databasePath,
        string outputPath,
        string key,
        CancellationToken cancellationToken = default) =>
        CommandAsync(
            "/api/decrypt_db",
            new { dp_path = databasePath, out_path = outputPath, key },
            cancellationToken);

    public Task<JsonElement> RefreshQrCodeAsync(CancellationToken cancellationToken = default) =>
        _client.PostAsync<JsonElement>("/api/reflash_qrcode", cancellationToken: cancellationToken);

    public Task<JsonElement> GetA8KeyAsync(
        string url,
        string urlType = "0",
        string scene = "0",
        CancellationToken cancellationToken = default) =>
        _client.PostAsync<JsonElement>("/api/get_a8key", new { url, urlType, scene }, cancellationToken);

    public Task<JsonElement> JsLoginAsync(
        string appId,
        CancellationToken cancellationToken = default) =>
        _client.PostAsync<JsonElement>("/api/js_login", new { waId = appId }, cancellationToken);

    public Task<JsonElement> GetDatabaseHandlesAsync(CancellationToken cancellationToken = default) =>
        _client.PostAsync<JsonElement>("/api/get_db_handle", cancellationToken: cancellationToken);

    public Task<JsonElement> ExecuteSqliteAsync(
        string databaseName,
        string sql,
        CancellationToken cancellationToken = default) =>
        _client.PostAsync<JsonElement>(
            "/api/sqlite3_exec",
            new { db_name = databaseName, sql_fmt = sql },
            cancellationToken);

    public async Task<string> GetConfigPathAsync(CancellationToken cancellationToken = default)
    {
        var response = await _client.PostCommandAsync("/api/get_config_path", cancellationToken: cancellationToken).ConfigureAwait(false);
        return CVXClient.GetRequiredProperty(response, "configPath").GetString() ?? string.Empty;
    }

    public Task SetAntiRevokeAsync(bool enabled, CancellationToken cancellationToken = default) =>
        CommandAsync("/api/anti_revoke", new { swtich = enabled ? "true" : "false" }, cancellationToken);

    public Task InitializeWechatAsync(CancellationToken cancellationToken = default) =>
        CommandAsync("/api/wechat_init", cancellationToken: cancellationToken);

    public Task BackupDatabaseAsync(
        string outputDirectory,
        string databaseName,
        CancellationToken cancellationToken = default) =>
        CommandAsync(
            "/api/backup_database",
            new { outputDir = outputDirectory, name = databaseName },
            cancellationToken);

    public async Task<LoginStatusResponse> GetLoginStatusAsync(CancellationToken cancellationToken = default)
    {
        var response = await _client.PostAsync<LoginStatusResponse>(
            "/api/check_login",
            cancellationToken: cancellationToken).ConfigureAwait(false);
        if (response.ErrCode is not (0 or 1))
        {
            throw new CVXApiException($"gVx login API error: errCode={response.ErrCode}");
        }

        return response;
    }

    public async Task<bool> CheckLoginAsync(CancellationToken cancellationToken = default) =>
        (await GetLoginStatusAsync(cancellationToken).ConfigureAwait(false)).Data.Status;

    public async Task<string> ScanQrCodeAsync(
        string imagePath,
        CancellationToken cancellationToken = default)
    {
        var response = await _client.PostCommandAsync(
            "/api/qrscan",
            new { path = imagePath },
            cancellationToken).ConfigureAwait(false);
        var data = CVXClient.GetRequiredProperty(response, "data");
        return CVXClient.GetRequiredProperty(data, "scan_res").GetString() ?? string.Empty;
    }

    public Task LogoutAsync(CancellationToken cancellationToken = default) =>
        CommandAsync("/api/logout", cancellationToken: cancellationToken);

    private async Task CommandAsync(
        string path,
        object? request = null,
        CancellationToken cancellationToken = default)
    {
        _ = await _client.PostCommandAsync(path, request, cancellationToken).ConfigureAwait(false);
    }
}

public sealed class LoginStatusResponse
{
    [JsonPropertyName("account_wxid")]
    public string? AccountWxid { get; init; }

    public LoginStatusData Data { get; init; } = new();
    public int ErrCode { get; init; }
}

public sealed class LoginStatusData
{
    public bool Status { get; init; }
}
