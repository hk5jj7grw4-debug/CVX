package types

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// SendFavEmotionRequest 发送收藏表情请求
type SendFavEmotionRequest struct {
	Wxid   string `json:"wxid"`             // 接收人微信ID
	MD5    string `json:"md5"`              // 表情MD5
	Length int    `json:"length,omitempty"` // 表情长度（可选）
}

// SendEmotionMsgRequest 发送本地GIF表情请求
type SendEmotionMsgRequest struct {
	Wxid     string `json:"wxid"`     // 接收人微信ID
	FilePath string `json:"filepath"` // 本地GIF文件路径
}
