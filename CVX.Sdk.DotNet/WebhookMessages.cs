using System.Text.Json;
using System.Text.Json.Serialization;

namespace CVX.Sdk;

/// <summary>Group message pushed by the local middleware HTTP callback.</summary>
public sealed class GroupChatMessage
{
    public WrappedString Content { get; init; } = new();
    public WrappedString FromUserName { get; init; } = new();
    public string? MessageType { get; init; }

    [JsonConverter(typeof(FlexibleStringJsonConverter))]
    public string? MsgType { get; init; }

    [JsonConverter(typeof(FlexibleStringJsonConverter))]
    public string? NewMsgId { get; init; }

    [JsonPropertyName("real_content")]
    public string? RealContent { get; init; }

    [JsonPropertyName("sender_nick")]
    public string? SenderNick { get; init; }

    [JsonPropertyName("member_info")]
    public GroupMessageMember MemberInfo { get; init; } = new();

    public string CommandText
    {
        get
        {
            if (!string.IsNullOrWhiteSpace(RealContent))
                return RealContent.Trim();

            var content = Content.Value?.Trim() ?? string.Empty;
            var separator = content.IndexOf(":\n", StringComparison.Ordinal);
            return separator >= 0 ? content[(separator + 2)..].Trim() : content;
        }
    }
}

public sealed class FlexibleStringJsonConverter : JsonConverter<string?>
{
    public override string? Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options) =>
        reader.TokenType switch
        {
            JsonTokenType.String => reader.GetString(),
            JsonTokenType.Number => ReadNumber(ref reader),
            JsonTokenType.Null => null,
            _ => throw new JsonException($"Cannot convert {reader.TokenType} to string."),
        };

    public override void Write(Utf8JsonWriter writer, string? value, JsonSerializerOptions options) =>
        writer.WriteStringValue(value);

    private static string ReadNumber(ref Utf8JsonReader reader)
    {
        using var value = JsonDocument.ParseValue(ref reader);
        return value.RootElement.GetRawText();
    }
}

public sealed class GroupMessageMember
{
    public string? UserName { get; init; }
    public string? NickName { get; init; }
}

public static class WebhookMessageParser
{
    private static readonly JsonSerializerOptions Options = new(JsonSerializerDefaults.Web)
    {
        PropertyNameCaseInsensitive = true,
        NumberHandling = JsonNumberHandling.AllowReadingFromString,
    };

    public static bool TryParseGroupMessage(ReadOnlySpan<byte> json, out GroupChatMessage? message)
    {
        try
        {
            message = JsonSerializer.Deserialize<GroupChatMessage>(json, Options);
            return message?.FromUserName.Value?.EndsWith("@chatroom", StringComparison.OrdinalIgnoreCase) == true;
        }
        catch (JsonException)
        {
            message = null;
            return false;
        }
    }
}
