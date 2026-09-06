package types

// SNSPostRequest 发送朋友圈请求
type SNSPostRequest struct {
	Content      string `json:"content"`       // 朋友圈文本内容
	BlackList    string `json:"blackList"`     // 黑名单列表（不给谁看）
	WithUserList string `json:"withauserList"` // 白名单列表（仅给谁看）
}

// SNSSendImgRequest 发送图片朋友圈请求
type SNSSendImgRequest struct {
	FileList string `json:"filelist"` // 图片文件列表（多个用逗号分隔）
	Content  string `json:"content"`  // 文本内容
}

// DownloadSNSMediaRequest 朋友圈图片/视频下载请求
type DownloadSNSMediaRequest struct {
	URL string `json:"url"` // 朋友圈图片或视频链接
	Out string `json:"out"` // 保存路径
}

// SNSDelCommentRequest 删除朋友圈评论请求
type SNSDelCommentRequest struct {
	SnsID     string `json:"sns_id"`    // 朋友圈ID
	CommentID string `json:"commentId"` // 评论ID
}

// SNSDelCommentResponse 删除朋友圈评论响应
type SNSDelCommentResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SNSCommentReplyRequest 朋友圈回复请求
type SNSCommentReplyRequest struct {
	Content   string `json:"content"`    // 回复内容
	SnsID     string `json:"sns_id"`     // 朋友圈ID
	CommentID int    `json:"comment_id"` // 评论ID
}

// SNSCommentReplyResponse 朋友圈回复响应
type SNSCommentReplyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SNSDelRequest 删除朋友圈请求
type SNSDelRequest struct {
	SnsID string `json:"sns_id"` // 朋友圈ID
}

// SNSDelResponse 删除朋友圈响应
type SNSDelResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SNSGetDetailRequest 获取朋友圈详情请求
type SNSGetDetailRequest struct {
	SnsID int64 `json:"sns_id"` // 朋友圈ID（数字类型）
}

// SNSGetDetailResponse 获取朋友圈详情响应
type SNSGetDetailResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // 朋友圈详情数据，结构根据实际返回定义
}

// SNSGetFirstPageRequest 获取朋友圈首页请求
type SNSGetFirstPageRequest struct {
	FirstPageMd5 string `json:"firstPageMd5"` // 第一页的MD5，如果没有可以为空
	MaxID        string `json:"maxId"`        // 最大ID，用于分页
}

// SNSGetFirstPageResponse 获取朋友圈首页响应
type SNSGetFirstPageResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // 朋友圈列表数据，结构根据实际返回定义
}

// SNSUploadRequest 朋友圈图片上传请求
type SNSUploadRequest struct {
	FilePath string `json:"filePath"` // 图片文件路径
}

// SNSUploadResponse 朋友圈图片上传响应
type SNSUploadResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // 上传后的图片信息（URL等），可作为图库使用
}

// SNSGetNextPageRequest 获取朋友圈下一页请求
type SNSGetNextPageRequest struct {
	LastItemID string `json:"lastItemid"` // 最后一条朋友圈的ID
}

// SNSGetNextPageResponse 获取朋友圈下一页响应
type SNSGetNextPageResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // 朋友圈列表数据
}
