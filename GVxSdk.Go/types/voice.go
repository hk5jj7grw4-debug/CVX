package types

// SendMP3VoiceRequest 发送MP3语音请求
type SendMP3VoiceRequest struct {
	Wxid    string `json:"wxid"`    // 接收人微信ID
	MP3Path string `json:"mp3Path"` // MP3文件路径
}

// GetVoiceTransRequest 语音转文本请求
type GetVoiceTransRequest struct {
	ClientMsgID string `json:"clientMsgId"` // 客户端消息ID
	NewMsgID    string `json:"newMsgId"`    // 新消息ID
	Length      string `json:"length"`      // 语音长度
}

// GetVoiceTransResponse 语音转文本响应
type GetVoiceTransResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"` // 转换后的文本
}

// DownloadVoiceRequest 下载语音请求
type DownloadVoiceRequest struct {
	NewMsgID string `json:"newMsgId"` // 新消息ID
	Length   string `json:"length"`   // 语音长度
	MsgID    string `json:"MsgId"`    // 消息ID
	Path     string `json:"path"`     // 保存路径
}
