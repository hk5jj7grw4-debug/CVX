package types

// CDNVideoForwardRequest CDN转发视频请求
type CDNVideoForwardRequest struct {
	Wxid        string `json:"wxid"`        // 接收人微信ID
	CDNVideoURL string `json:"cdnVideoUrl"` // CDN视频URL
	AESKey      string `json:"aesKey"`      // AES密钥
	VideoLength int    `json:"videoLength"` // 视频长度
	ThumbLength int    `json:"thumbLength"` // 缩略图长度
	PlayLength  int    `json:"playLength"`  // 播放时长
}

// DownloadVideoRequest 下载视频请求
type DownloadVideoRequest struct {
	TotalLen int    `json:"total_len"` // 视频总长度（字节）
	NewMsgId int64  `json:"NewMsgId"`  // 新消息ID
	Path     string `json:"path"`      // 保存路径
	MsgId    int    `json:"MsgId"`     // 消息ID
}

// DownloadVideoResponse 下载视频响应
type DownloadVideoResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
