
gVxSdk 开发规范总结
1. 项目结构
gVxSdk/
├── client.go           # 主客户端，统一入口
├── go.mod             # Go 模块定义
├── README.md          # 项目文档
├── api/               # API 实现层
│   ├── client.go      # Client 接口定义
│   ├── message.go     # 消息相关 API
│   ├── emotion.go     # 表情相关 API
│   ├── video.go       # 视频相关 API
│   ├── voice.go       # 语音相关 API
│   ├── cdn.go         # CDN 相关 API
│   ├── sns.go         # 朋友圈相关 API
│   └── system.go      # 系统相关 API
└── types/             # 类型定义层
    ├── message.go     # 消息类型
    ├── emotion.go     # 表情类型
    ├── video.go       # 视频类型
    ├── voice.go       # 语音类型
    ├── cdn.go         # CDN 类型
    ├── sns.go         # 朋友圈类型
    ├── system.go      # 系统类型
    └── webhook.go     # 回调消息类型
2. 设计原则
2.1 职责分离
types/ - 只定义数据结构，不包含业务逻辑
api/ - 实现具体的 API 调用逻辑
client.go - 提供统一的客户端入口
2.2 模块化设计
按功能领域划分模块（Message, Emotion, Video 等）
每个模块独立，互不依赖
通过主客户端统一管理
2.3 最小化原则
SDK 只提供类型定义和 API 调用
不包含 HTTP 服务器、路由等框架性功能
用户保持完全控制权
3. 命名规范
3.1 文件命名
小写字母，使用下划线分隔（Go 惯例）
按功能模块命名：message.go, emotion.go
3.2 类型命名
请求类型：{功能}Request
例：SendTextMsgRequest, SendFavEmotionRequest
响应类型：{功能}Response
例：GetWxBasePathResponse, GetVoiceTransResponse
通用响应：Response（所有接口共用）
3.3 API 方法命名
使用动词开头，驼峰命名
清晰表达功能：SendTextMsg, GetWxBasePath, AutoLogin
3.4 模块命名
使用名词，首字母大写
例：MessageAPI, EmotionAPI, SystemAPI
4. 代码规范
4.1 注释规范
每个公开的函数必须包含：

// FunctionName 功能简述
//
// 详细说明（可选）
//
// 参数:
//   - param1: 参数说明
//   - param2: 参数说明
//
// 返回:
//   - type: 返回值说明
//   - error: 错误说明
//
// 示例:
//
//	code example
//
// 注意:
//   - 注意事项1
//   - 注意事项2
4.2 错误处理
// 统一的错误包装格式
if err != nil {
    return fmt.Errorf("operation name: %w", err)
}

// API 错误检查
if resp.Code != 1 && resp.Code != 0 {
    return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
}
4.3 JSON 标签
type Request struct {
    Field1 string `json:"field1"`           // 必填字段
    Field2 string `json:"field2,omitempty"` // 可选字段
}
5. API 实现模式
5.1 标准 API 方法结构
func (a *ModuleAPI) MethodName(params...) (returnType, error) {
    // 1. 构造请求
    req := types.RequestType{
        Field1: param1,
        Field2: param2,
    }

    // 2. 发送请求
    respBody, err := a.client.DoRequest("POST", "/api/endpoint", req)
    if err != nil {
        return nil, fmt.Errorf("method name: %w", err)
    }

    // 3. 解析响应
    var resp types.ResponseType
    if err := json.Unmarshal(respBody, &resp); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    // 4. 检查错误
    if resp.Code != 1 && resp.Code != 0 {
        return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
    }

    // 5. 返回结果
    return resp.Data, nil
}
5.2 无请求体的 API
func (a *ModuleAPI) MethodName() error {
    respBody, err := a.client.DoRequest("POST", "/api/endpoint", nil)
    // ... 后续处理相同
}
6. 客户端初始化模式
// 主客户端结构
type Client struct {
    baseURL    string
    httpClient *http.Client
    
    // API 模块
    Message *api.MessageAPI
    Emotion *api.EmotionAPI
    // ...
}

// 初始化方法
func NewClient(baseURL string) *Client {
    c := &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
    
    // 初始化所有模块
    c.Message = &api.MessageAPI{client: c}
    c.Emotion = &api.EmotionAPI{client: c}
    // ...
    
    return c
}
7. 回调处理规范
7.1 只提供类型定义
在 
webhook.go
 中定义回调消息结构
不提供 HTTP 服务器实现
用户自己实现回调处理逻辑
7.2 回调类型设计
// 通用回调消息结构
type WebhookMessage struct {
    AccountWxid string        `json:"account_wxid"`
    List        []MessageItem `json:"list"`
    // 通用字段...
    
    // 特定消息类型的可选字段
    Msg *AppMsg `json:"msg,omitempty"`
}

// 消息项（支持多种消息类型）
type MessageItem struct {
    // 基础字段
    Type    int    `json:"type"`
    Content string `json:"content"`
    
    // 图片消息字段（可选）
    MD5 string `json:"md5,omitempty"`
    // ...
}
8. 版本管理
8.1 Git 提交规范
feat: 新功能
fix: 修复bug
docs: 文档更新
refactor: 重构
test: 测试相关
chore: 构建/工具相关
8.2 模块版本
// go.mod
module github.com/yourusername/gvxsdk

go 1.21
9. 使用示例规范
每个模块的 README 应包含：

安装说明
快速开始示例
完整的使用示例
注意事项
10. 测试规范（未来扩展）
// 测试文件命名：*_test.go
func TestSendTextMsg(t *testing.T) {
    client := NewClient("http://test-server")
    err := client.Message.SendTextMsg("test", "hello")
    // 断言...
}
总结
这个 SDK 遵循：

✅ 简洁性 - 只做 API 封装，不做框架
✅ 模块化 - 按功能领域清晰划分
✅ 可扩展 - 易于添加新接口
✅ 类型安全 - 完整的类型定义
✅ 文档完善 - 详细的注释和示例
✅ Go 惯例 - 遵循 Go 语言最佳实践