package types

// SendAppMsgRequest 发送卡片/XML消息请求
type SendAppMsgRequest struct {
	Wxid    string `json:"wxid"`    // 接收人微信ID
	Content string `json:"content"` // XML内容
	Type    string `json:"type"`    // 消息类型
}

// SendXMLRequest 发送链接消息请求
type SendXMLRequest struct {
	Wxid        string `json:"wxid"`        // 接收人微信ID
	Title       string `json:"title"`       // 链接标题
	Description string `json:"description"` // 链接描述
	ThumbURL    string `json:"thumbUrl"`    // 缩略图URL
	URL         string `json:"url"`         // 链接URL
}

// SendTextMsgRequest 发送文本消息请求
type SendTextMsgRequest struct {
	Wxid string `json:"wxid"` // 接收人微信ID
	Msg  string `json:"msg"`  // 文本消息内容
}

// SendAtTextRequest 发送AT消息请求
type SendAtTextRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
	Msg    string `json:"msg"`    // 消息内容（需包含@昵称）
	Wxids  string `json:"wxids"`  // 被@的微信ID，多个用逗号分隔
}

// SendCardMsgRequest 发送名片消息请求
type SendCardMsgRequest struct {
	Wxid     string `json:"wxid"`     // 发送给谁
	CardWxid string `json:"cardWxid"` // 谁的名片
}

// SendImageMsgRequest 发送图片消息请求
type SendImageMsgRequest struct {
	Wxid     string `json:"wxid"`     // 接收人微信ID
	FilePath string `json:"filepath"` // 图片文件路径（注意：字段名是 filepath 不是 image_path）
}

// SendFileMsgRequest 发送文件消息请求
type SendFileMsgRequest struct {
	Wxid     string `json:"wxid"`     // 接收人微信ID
	FilePath string `json:"filepath"` // 文件路径
}

// SendPatRequest 发送拍一拍请求
type SendPatRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
	Wxid   string `json:"wxid"`   // 被拍的微信ID
}

// SendQuoteRequest 发送引用消息请求
type SendQuoteRequest struct {
	Reply        string `json:"reply"`               // 回复内容
	ReferContent string `json:"referContent"`        // 引用的原消息内容
	FromUsr      string `json:"fromUsr"`             // 原消息发送者微信ID
	NewMsgID     string `json:"newmsgid"`            // 原消息ID
	MsgSource    string `json:"msgSource,omitempty"` // 消息源（可选）
	CreateTime   int    `json:"createTime"`          // 创建时间
	SendTo       string `json:"sendto"`              // 发送给谁（群聊ID或个人微信ID）
}

// SendAppletMsgRequest 发送小程序消息请求
type SendAppletMsgRequest struct {
	Wxid    string `json:"wxid"`    // 接收人微信ID
	Content string `json:"content"` // 小程序XML内容
	Type    string `json:"type"`    // 消息类型（通常为"33"）
}

// SendLocationMsgRequest 发送位置消息请求
type SendLocationMsgRequest struct {
	Wxid    string `json:"wxid"`    // 接收人微信ID
	X       string `json:"x"`       // 经度
	Y       string `json:"y"`       // 纬度
	Label   string `json:"lable"`   // 标签名字（注意：接口拼写为lable）
	PoiName string `json:"poiname"` // 地点名称
}

// RevokeAnyRequest 撤回任何消息请求
type RevokeAnyRequest struct {
	NewMsgID   float64 `json:"newMsgId"`   // 消息ID
	CreateTime int     `json:"createTime"` // 创建时间
	ToUserName string  `json:"toUserName"` // 接收人微信ID或群聊ID
}

// RevokeAnyResponse 撤回任何消息响应
type RevokeAnyResponse struct {
	AccountWxid string        `json:"account_wxid"`
	Data        RevokeAnyData `json:"data"`
	ErrCode     int           `json:"errCode"`
	ErrMsg      string        `json:"errMsg"`
}

// RevokeAnyData 撤回消息数据
type RevokeAnyData struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	SysWording   string       `json:"sysWording"`
}
