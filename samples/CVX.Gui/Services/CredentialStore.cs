using System.IO;
using System.Text.Json;

namespace CVX.Gui.Services;

internal sealed class AppCredentials
{
    public string UpdateServer { get; set; } = "";
    public string ComponentToken { get; set; } = "";
    public int GvxApiPort { get; set; } = 19088;
}

internal static class CredentialStore
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        WriteIndented = true
    };

    public static string FilePath { get; } = Path.Combine(
        Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
        "CVX",
        "gui-credentials.json");

    public static AppCredentials Load()
    {
        var credentials = new AppCredentials();

        if (File.Exists(FilePath))
        {
            try
            {
                credentials = JsonSerializer.Deserialize<AppCredentials>(
                    File.ReadAllText(FilePath)) ?? credentials;
            }
            catch (JsonException)
            {
            }
        }

        OverrideFromEnvironment(credentials);
        if (credentials.GvxApiPort <= 0)
        {
            credentials.GvxApiPort = 19088;
        }

        return credentials;
    }

    public static void Save(AppCredentials credentials)
    {
        var directory = Path.GetDirectoryName(FilePath);
        if (!string.IsNullOrEmpty(directory))
        {
            Directory.CreateDirectory(directory);
        }

        File.WriteAllText(FilePath, JsonSerializer.Serialize(credentials, JsonOptions));
    }

    private static void OverrideFromEnvironment(AppCredentials credentials)
    {
        credentials.UpdateServer = Env("CVX_UPDATE_SERVER") ?? credentials.UpdateServer;
        credentials.ComponentToken = Env("CVX_COMPONENT_TOKEN") ?? credentials.ComponentToken;
    }

    private static string? Env(string name)
    {
        var value = Environment.GetEnvironmentVariable(name);
        return string.IsNullOrWhiteSpace(value) ? null : value.Trim();
    }
}
