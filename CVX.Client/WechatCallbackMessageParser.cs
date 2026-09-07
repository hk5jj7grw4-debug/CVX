using System.Text;
using System.Text.Json;

namespace CVX.Client;

/// <summary>Normalizes GVx callback JSON without wrapping the GVx client API.</summary>
internal static class WechatCallbackMessageParser
{
    public static WechatCallbackMessage Parse(ReadOnlySpan<byte> body)
    {
        var raw = Encoding.UTF8.GetString(body);
        var type = "raw";
        var groupId = "";
        var senderId = "";
        var senderName = "";
        var text = "";
        var messageId = "";

        try
        {
            using var document = JsonDocument.Parse(raw);
            var root = document.RootElement;
            type = ReadStringDeep(root, "msgType", "MsgType", "messageType", "MessageType", "type", "Type") ?? "json";
            groupId = FirstChatroom(
                ReadWrappedStringDeep(root, "fromUserName", "FromUserName"),
                ReadWrappedStringDeep(root, "roomWxid", "RoomWxid"),
                ReadWrappedStringDeep(root, "groupId", "GroupId")) ?? "";
            senderId = ReadWrappedStringDeep(root, "senderWxid", "SenderWxid", "userName", "UserName") ?? "";
            senderName = ReadWrappedStringDeep(root, "senderNick", "SenderNick", "nickName", "NickName") ?? "";
            text = ReadWrappedStringDeep(root, "commandText", "CommandText", "content", "Content", "text", "Text", "msg", "Msg") ?? "";
            messageId = ReadWrappedStringDeep(root, "newMsgId", "NewMsgId", "messageId", "MessageId") ?? "";

            if (text.Contains(":\n", StringComparison.Ordinal))
            {
                var parts = text.Split([":\n"], 2, StringSplitOptions.None);
                if (string.IsNullOrWhiteSpace(senderId)) senderId = parts[0].Trim();
                text = parts[1].Trim();
            }
        }
        catch (JsonException)
        {
            text = raw;
        }

        return new WechatCallbackMessage(
            DateTimeOffset.Now,
            raw,
            type,
            groupId,
            senderId,
            senderName,
            text,
            messageId);
    }

    static string? ReadWrappedStringDeep(JsonElement root, params string[] names)
    {
        var value = ReadElementDeep(root, names);
        if (value is null) return null;
        var element = value.Value;
        if (element.ValueKind == JsonValueKind.String) return element.GetString();
        if (element.ValueKind is JsonValueKind.Number or JsonValueKind.True or JsonValueKind.False)
            return element.ToString();
        return ReadStringDeep(element, "String", "string", "Value", "value");
    }

    static string? ReadStringDeep(JsonElement root, params string[] names)
    {
        var value = ReadElementDeep(root, names);
        if (value is null) return null;
        var element = value.Value;
        return element.ValueKind == JsonValueKind.String ? element.GetString() : element.ToString();
    }

    static JsonElement? ReadElementDeep(JsonElement root, params string[] names)
    {
        if (root.ValueKind == JsonValueKind.Object)
        {
            foreach (var property in root.EnumerateObject())
            {
                if (names.Any(name => property.Name.Equals(name, StringComparison.OrdinalIgnoreCase)))
                    return property.Value;
                var child = ReadElementDeep(property.Value, names);
                if (child is not null) return child;
            }
        }
        else if (root.ValueKind == JsonValueKind.Array)
        {
            foreach (var item in root.EnumerateArray())
            {
                var child = ReadElementDeep(item, names);
                if (child is not null) return child;
            }
        }
        return null;
    }

    static string? FirstChatroom(params string?[] values) =>
        values.FirstOrDefault(value =>
            !string.IsNullOrWhiteSpace(value)
            && value.Contains("@chatroom", StringComparison.OrdinalIgnoreCase));
}
