# CVX Go SDK

微信中间件 SDK，提供 Go 客户端和可供 WinForms/WPF 使用的 .NET 托管 DLL。

## 安装

```bash
go get github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go
```

## 快速开始

### 发送消息

```go
package main

import (
    "log"
    "github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go"
)

func main() {
    // 初始化客户端
    client := cvxsdk.NewClient("http://127.0.0.1:19088")
    
    // 发送文本消息
    err := client.Message.SendTextMsg("filehelper", "Hello, World!")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 接收回调消息

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    
    "github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/types"
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
    var msg types.WebhookMessage
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 处理消息
    for _, item := range msg.List {
        switch item.Type {
        case types.MessageTypeText:
            fmt.Printf("收到文本消息: %s\n", item.Content)
        case types.MessageTypeImage:
            fmt.Printf("收到图片消息: %s\n", item.MD5)
        }
    }
    
    w.WriteHeader(http.StatusOK)
}

func main() {
    http.HandleFunc("/api/recvMsg", handleWebhook)
    log.Fatal(http.ListenAndServe(":8081", nil))
}
```

## API 模块

- **Message** - 消息发送（文本、图片、文件、AT、引用等）
- **Emotion** - 表情相关（收藏表情、GIF）
- **Video** - 视频相关
- **Voice** - 语音相关
- **CDN** - CDN 相关
- **SNS** - 朋友圈相关
- **System** - 系统相关（获取缓存目录、自动登录、解密数据库等）
- **Room** - 群聊相关（获取群成员列表等）

## Windows / .NET DLL

`.NET 8+` 客户端可以引用 GitHub Actions 构建出的 `CVX.Sdk.DotNet.dll`：

```csharp
using CVX.Sdk;

using var vx = new CVXClient("http://127.0.0.1:19088");
await vx.Message.SendTextAsync("filehelper", "Hello from .NET");
```

本地中间件只监听 `127.0.0.1` 时可以不传 Token。若中间件启用了鉴权，可在创建 `CVXClient` 时传入可选的 `authToken`。

对应的 .NET 客户端源码见 [`CVX.Sdk.DotNet`](../CVX.Sdk.DotNet)。

## 系统功能示例

### 解密数据库

```go
// 解密微信加密的数据库文件
err := client.System.DecryptDB(
    "E:\\xwechat_files\\wxid_xxx\\db_storage\\contact\\contact.db",
    "E:\\xwechat_files\\wxid_xxx\\db_storage\\contact\\contact_decrypted.db",
    "910e9301f30a4250aefaf9c51fb8e1646103c228ae1b4cc7899a8456762cdb16",
)
if err != nil {
    log.Fatal(err)
}
fmt.Println("数据库解密成功！")
```

### 获取微信缓存目录

```go
basePath, err := client.System.GetWxBasePath()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("微信缓存目录: %s\n", basePath)
```

### 自动登录

```go
err := client.System.AutoLogin()
if err != nil {
    log.Fatal(err)
}
fmt.Println("自动登录成功！")
```

## 群聊功能示例

### 获取群成员列表

```go
// 获取指定群聊的所有成员信息
resp, err := client.Room.GetRoomMembers("39259098574@chatroom")
if err != nil {
    log.Fatal(err)
}

// 显示群信息
fmt.Printf("群主: %s\n", resp.ChatRoomOwner)
fmt.Printf("成员总数: %d\n", resp.AllMemberCount)

// 遍历成员列表
for _, member := range resp.NewChatroomData.ChatRoomMember {
    fmt.Printf("- %s (%s)\n", member.NickName, member.UserName)
    if member.InviterUserName != "" {
        fmt.Printf("  邀请人: %s\n", member.InviterUserName)
    }
}
```

## 更多示例

查看 [examples](./examples) 目录获取更多使用示例。

## 文档

详细 API 文档请查看代码注释或使用 `go doc` 命令：

```bash
go doc github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/api MessageAPI.SendTextMsg
```
