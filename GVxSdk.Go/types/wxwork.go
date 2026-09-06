package types

// DownloadWxWorkFileRequest 下载企业文件/图片请求
type DownloadWxWorkFileRequest struct {
	URL     string `json:"url"`     // 文件下载URL
	Key     string `json:"key"`     // 解密密钥
	Out     string `json:"out"`     // 输出文件路径
	AuthKey string `json:"authkey"` // 认证密钥
}

// DownloadWxWorkFileResponse 下载企业文件/图片响应
type DownloadWxWorkFileResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
