package types

// SendCDNImgMsgRequest CDN发送图片请求（用于转发消息）
type SendCDNImgMsgRequest struct {
	ToWxid          string `json:"toWxid"`          // 接收人微信ID
	TotalLen        string `json:"totalLen"`        // 图片总长度
	FileID          string `json:"fileId"`          // 文件ID
	AESKey          string `json:"aesky"`           // AES密钥
	CDNMidImgSize   string `json:"cdnmidImgSize"`   // CDN中等图片大小
	CDNThumbImgSize string `json:"cdnthumbImgSize"` // CDN缩略图大小
	EncryVer        string `json:"encryVer"`        // 加密版本
}

// DownloadImgRequest 下载图片请求
type DownloadImgRequest struct {
	ToUser       string `json:"to_user"`       // 自己的 wxid
	FromUser     string `json:"from_user"`     // 来自谁发的图片
	StartPos     int    `json:"start_pos"`     // 起始位置
	TotalLen     int    `json:"total_len"`     // 图片总长度（cdnmidimgurl_size）
	DataLen      int    `json:"data_len"`      // 数据长度（cdnmidimgurl_size）
	CompressType int    `json:"compress_type"` // 压缩类型，默认为 0
	MsgID        int64  `json:"MsgId"`         // 消息 ID
	Path         string `json:"path"`          // 保存路径
}

// DownloadImgResponse 下载图片响应
type DownloadImgResponse struct {
	CompressType    int    `json:"CompressType"`     // 压缩类型，默认为 0
	FromUserName    string `json:"FromUserName"`     // 来自谁发的图片
	MsgID           int64  `json:"MsgId"`            // 消息 ID
	ToUserName      string `json:"ToUserName"`       // 自己的 wxid
	TotalLen        int    `json:"TotalLen"`         // 图片总长度（cdnmidimgurl_size）
	DownloadedBytes int    `json:"downloaded_bytes"` // 已下载字节数（cdnmidimgurl_size）
	Path            string `json:"path"`             // 保存路径
	Status          string `json:"status"`           // 下载状态（success 表示成功）
}

// DownloadFileRequest 下载文件请求
type DownloadFileRequest struct {
	FromUser string  `json:"from_user"` // 发送文件的人的 wxid
	TotalLen string  `json:"total_len"` // 文件总长度
	MsgID    float64 `json:"MsgId"`     // 消息 ID（数字类型）
	Path     string  `json:"path"`      // 保存路径
	AttachID string  `json:"attachid"`  // 附件 ID
	Type     string  `json:"type"`      // 文件类型
}

// DownloadFileResponse 下载文件响应
type DownloadFileResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// CDNDownloadRequest CDN下载请求
type CDNDownloadRequest struct {
	FileID  string `json:"fileid"`  // 文件ID
	AESKey  string `json:"asekey"`  // AES密钥
	ImgType int    `json:"imgType"` // 文件类型：1-高清图片 2-普通图片 3-缩略图 4-视频 5-文件
	Out     string `json:"out"`     // 输出路径
}

// GetCDNInfoResponse 获取CDN信息响应
type GetCDNInfoResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"` // CDN信息数据
}
