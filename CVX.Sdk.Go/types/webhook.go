package types

// WebhookMessage 回调消息通用结构
type WebhookMessage struct {
	AccountWxid string        `json:"account_wxid"` // 当前登录的微信ID
	Count       int           `json:"count"`        // 消息数量
	HTTPPort    int           `json:"http_port"`    // HTTP端口
	Hwnd        int           `json:"hwnd"`         // 窗口句柄
	List        []MessageItem `json:"list"`         // 消息列表
	MessageDesc string        `json:"messageDesc"`  // 消息描述
	MessageType int           `json:"messageType"`  // 消息类型
	PID         int           `json:"pid"`          // 进程ID
	SystemCode  int           `json:"systemCode"`   // 系统代码

	// 应用消息（文件/卡片/小程序等）特有字段
	CRC32          int     `json:"crc32,omitempty"`          // CRC32校验
	FileType       int     `json:"fileType,omitempty"`       // 文件类型
	HitMD5         int     `json:"hitMd5,omitempty"`         // MD5命中
	MD5            string  `json:"md5,omitempty"`            // MD5值
	Msg            *AppMsg `json:"msg,omitempty"`            // 应用消息内容
	MsgForwardType int     `json:"msgForwardType,omitempty"` // 消息转发类型
	Signature      string  `json:"signature,omitempty"`      // 签名
}

// MessageItem 消息项
type MessageItem struct {
	ClientMsgID  int           `json:"clientMsgId"`            // 客户端消息ID
	Content      string        `json:"content"`                // 消息内容
	CreateTime   int64         `json:"createTime"`             // 创建时间
	MsgSource    string        `json:"msgSource"`              // 消息源
	ToUserName   ToUserName    `json:"toUserName"`             // 接收人
	Type         int           `json:"type"`                   // 消息类型
	FromUserName *FromUserName `json:"fromUserName,omitempty"` // 发送人（群消息时存在）
	MsgID        int64         `json:"msgId,omitempty"`        // 消息ID
	NewMsgID     string        `json:"newMsgId,omitempty"`     // 新消息ID

	// 图片消息相关字段
	AESKey            string      `json:"aeskey,omitempty"`            // AES密钥
	CDNBigImgSize     int         `json:"cdnbigImgSize,omitempty"`     // CDN大图大小
	CDNBigImgURL      string      `json:"cdnbigImgUrl,omitempty"`      // CDN大图URL
	CDNMidImgSize     int         `json:"cdnmidImgSize,omitempty"`     // CDN中图大小
	CDNMidImgURL      string      `json:"cdnmidImgUrl,omitempty"`      // CDN中图URL
	CDNThumbAESKey    string      `json:"cdnthumbAeskey,omitempty"`    // CDN缩略图AES密钥
	CDNThumbImgHeight int         `json:"cdnthumbImgHeight,omitempty"` // CDN缩略图高度
	CDNThumbImgSize   int         `json:"cdnthumbImgSize,omitempty"`   // CDN缩略图大小
	CDNThumbImgURL    string      `json:"cdnthumbImgUrl,omitempty"`    // CDN缩略图URL
	CDNThumbImgWidth  int         `json:"cdnthumbImgWidth,omitempty"`  // CDN缩略图宽度
	ClientImgID       *ToUserName `json:"clientImgId,omitempty"`       // 客户端图片ID
	CompressType      int         `json:"compressType,omitempty"`      // 压缩类型
	CRC32             int         `json:"crc32,omitempty"`             // CRC32校验
	Data              *ImageData  `json:"data,omitempty"`              // 图片数据
	DataLen           int         `json:"dataLen,omitempty"`           // 数据长度
	EncryVer          int         `json:"encryVer,omitempty"`          // 加密版本
	HevcMidSize       int         `json:"hevcMidSize,omitempty"`       // HEVC中图大小
	HitMD5            int         `json:"hitMd5,omitempty"`            // MD5命中
	ImgType           int         `json:"imgType,omitempty"`           // 图片类型
	MD5               string      `json:"md5,omitempty"`               // MD5值
	MsgType           int         `json:"msgType,omitempty"`           // 消息类型
	NetType           int         `json:"netType,omitempty"`           // 网络类型
	PhotoFrom         int         `json:"photoFrom,omitempty"`         // 图片来源
	StartPos          int         `json:"startPos,omitempty"`          // 起始位置
	TotalLen          int         `json:"totalLen,omitempty"`          // 总长度
}

// ImageData 图片数据
type ImageData struct {
	Buffer string `json:"buffer"` // 图片数据缓冲区
	ILen   int    `json:"iLen"`   // 数据长度
}

// ToUserName 接收人信息
type ToUserName struct {
	String string `json:"String"` // 接收人微信ID
}

// FromUserName 发送人信息
type FromUserName struct {
	String string `json:"String"` // 发送人微信ID
}

// 消息类型常量
const (
	MessageTypeText     = 1     // 文本消息
	MessageTypeImage    = 3     // 图片消息
	MessageTypeVoice    = 34    // 语音消息
	MessageTypeVideo    = 43    // 视频消息
	MessageTypeEmotion  = 47    // 表情消息
	MessageTypeLocation = 48    // 位置消息
	MessageTypeApp      = 49    // 应用消息（链接、小程序等）
	MessageTypeCard     = 42    // 名片消息
	MessageTypeQuote    = 57    // 引用消息
	MessageTypeSystem   = 10000 // 系统消息
)

// AppMsg 应用消息（文件、卡片、小程序等）
type AppMsg struct {
	AppID            string `json:"appId"`            // 应用ID
	ClientMsgID      string `json:"clientMsgId"`      // 客户端消息ID
	Content          string `json:"content"`          // XML内容
	CreateTime       int64  `json:"createTime"`       // 创建时间
	FromUserName     string `json:"fromUserName"`     // 发送人
	JsAppID          string `json:"jsAppId"`          // JS应用ID
	MsgSource        string `json:"msgSource"`        // 消息源
	SDKVersion       int    `json:"sdkVersion"`       // SDK版本
	ShareURLOpen     string `json:"shareUrlOpen"`     // 分享打开URL
	ShareURLOriginal string `json:"shareUrlOriginal"` // 分享原始URL
	Source           int    `json:"source"`           // 来源
	ToUserName       string `json:"toUserName"`       // 接收人
	Type             int    `json:"type"`             // 类型
}

// GroupChatMessage 群聊消息回调
type GroupChatMessage struct {
	Content      StringContent   `json:"content"`      // 原始消息内容（包含发送者 wxid + 冒号 + 实际消息内容）
	CreateTime   string          `json:"createTime"`   // 消息创建时间（Unix 时间戳，秒）
	FromUserName StringContent   `json:"fromUserName"` // 群聊 ID（@chatroom 结尾）
	ImgBuf       GroupImgBufInfo `json:"imgBuf"`       // 图片缓冲区信息
	ImgStatus    string          `json:"imgStatus"`    // 图片状态（1=有图，0=无图）
	MemberInfo   GroupMemberInfo `json:"member_info"`  // 群成员信息
	MessageType  string          `json:"messageType"`  // 消息类型描述（群聊消息 / 私聊消息）
	MsgID        string          `json:"msgId"`        // 消息 ID（本地整型值）
	MsgSeq       string          `json:"msgSeq"`       // 消息序列号
	MsgSource    string          `json:"msgSource"`    // 消息源的 XML（包含群信息，如成员数、签名等）
	MsgType      string          `json:"msgType"`      // 消息类型数值（1=文本，3=图片，43=视频 等）
	NewMsgID     string          `json:"newMsgId"`     // 消息唯一 ID（服务器下发，字符串类型）
	PushContent  string          `json:"pushContent"`  // 推送文案（昵称 + 消息内容，用于通知栏展示）
	RealContent  string          `json:"real_content"` // 实际消息内容（去掉 wxid 和换行符之后的文本）
	SenderNick   string          `json:"sender_nick"`  // 发送者昵称（可能为空）
	Status       string          `json:"status"`       // 消息状态（3=已送达 等）
	ToUserName   StringContent   `json:"toUserName"`   // 消息接收方 wxid（通常是自己）
}

// StringContent 字符串内容
type StringContent struct {
	String string `json:"String"`
}

// GroupImgBufInfo 群聊图片缓冲区信息
type GroupImgBufInfo struct {
	ILen string `json:"iLen"` // 图片数据长度（0 表示无图片）
}

// GroupMemberInfo 群成员信息
type GroupMemberInfo struct {
	AddChatRoomSceneNewXml string `json:"addChatRoomSceneNewXml"` // 入群场景 XML（包含邀请人信息）
	BigHeadImgUrl          string `json:"bigHeadImgUrl"`          // 群成员头像大图 URL
	ChatroomMemberFlag     string `json:"chatroomMemberFlag"`     // 群成员标志位（内部字段）
	InviterUserName        string `json:"inviterUserName"`        // 邀请者的 wxid
	NickName               string `json:"nickName"`               // 群成员昵称
	SmallHeadImgUrl        string `json:"smallHeadImgUrl"`        // 群成员头像小图 URL
	Status                 string `json:"status"`                 // 群成员状态（0=正常，1=禁言 等）
	UserName               string `json:"userName"`               // 群成员的 wxid
}

// PrivateChatMessage 私聊消息回调
type PrivateChatMessage struct {
	Content       StringContent   `json:"content"`        // 消息内容（文本）
	CreateTime    string          `json:"createTime"`     // 消息创建时间（Unix 时间戳，秒）
	From          string          `json:"from"`           // 来自登录的哪个 wxid
	FromUserName  StringContent   `json:"fromUserName"`   // 发送者账号（wxid 或群号@chatroom）
	HTTPPort      string          `json:"http_port"`      // HTTP 服务端口（调试用）
	ImgBuf        GroupImgBufInfo `json:"imgBuf"`         // 图片缓冲区信息
	ImgStatus     string          `json:"imgStatus"`      // 图片状态（1=有图，0=无图）
	MessageType   string          `json:"messageType"`    // 消息类型描述（如 私聊消息 / 群聊消息）
	MsgID         string          `json:"msgId"`          // 消息 ID（本地生成的整型值）
	MsgSeq        string          `json:"msgSeq"`         // 消息序列号
	MsgSource     string          `json:"msgSource"`      // 消息源的 XML 描述（包含群成员数/签名等）
	MsgType       string          `json:"msgType"`        // 消息类型数值（1=文本，3=图片，43=视频 等）
	NewMsgID      string          `json:"newMsgId"`       // 消息唯一 ID（服务器下发，字符串类型）
	PID           string          `json:"pid"`            // 进程 ID（调试信息）
	PushContent   string          `json:"pushContent"`    // 推送内容（带有发送者昵称 + 消息预览）
	SenderNick    string          `json:"sender_nick"`    // 发送者昵称（可能为空）
	SenderProfile SenderProfile   `json:"sender_profile"` // 发送者资料
	Status        string          `json:"status"`         // 消息状态（3=已送达 等）
	ToUserName    StringContent   `json:"toUserName"`     // 接收者 wxid（通常是自己）
}

// SenderProfile 发送者资料
type SenderProfile struct {
	Alias              string              `json:"alias"`              // 用户别名
	BigHeadImgUrl      string              `json:"bigHeadImgUrl"`      // 用户头像大图 URL
	BitMask            string              `json:"bitMask"`            // 标志位（内部用途）
	BitVal             string              `json:"bitVal"`             // 标志值（内部用途）
	City               string              `json:"city"`               // 用户所在城市
	Country            string              `json:"country"`            // 用户所在国家
	Description        string              `json:"description"`        // 个性签名
	EncryptUserName    string              `json:"encryptUserName"`    // 加密后的用户名
	HasWeiXinHdHeadImg string              `json:"hasWeiXinHdHeadImg"` // 是否有高清头像（1=有，0=无）
	ImgBuf             ProfileImgBuf       `json:"imgBuf"`             // 头像缓冲区
	ImgFlag            string              `json:"imgFlag"`            // 头像标志
	LabelIdlist        string              `json:"labelIdlist"`        // 标签 ID 列表
	NickName           StringContent       `json:"nickName"`           // 用户昵称
	PhoneNumListInfo   ProfilePhoneNumInfo `json:"phoneNumListInfo"`   // 手机号信息
	Province           string              `json:"province"`           // 省份
	Pyinitial          StringContent       `json:"pyinitial"`          // 昵称拼音首字母
	QuanPin            StringContent       `json:"quanPin"`            // 昵称全拼
	Remark             StringContent       `json:"remark"`             // 备注名
	RemarkPyinitial    StringContent       `json:"remarkPyinitial"`    // 备注名拼音首字母
	RemarkQuanPin      StringContent       `json:"remarkQuanPin"`      // 备注名全拼
	Sex                string              `json:"sex"`                // 性别（1=男，2=女，0=未知）
	SmallHeadImgUrl    string              `json:"smallHeadImgUrl"`    // 用户头像小图 URL
	SnsUserInfo        ProfileSnsUserInfo  `json:"snsUserInfo"`        // 朋友圈信息
	TextStatusFlag     string              `json:"textStatusFlag"`     // 状态标志（内部用途）
	UserName           StringContent       `json:"userName"`           // 用户的 wxid
	VerifyFlag         string              `json:"verifyFlag"`         // 认证标志（大 V 等）
}

// ProfileImgBuf 资料头像缓冲区
type ProfileImgBuf struct {
	Buffer string `json:"buffer"` // 头像 buffer（一般为空）
	ILen   string `json:"iLen"`   // 头像 buffer 长度
}

// ProfilePhoneNumInfo 资料手机号信息
type ProfilePhoneNumInfo struct {
	Count string `json:"count"` // 绑定的手机号数量
}

// ProfileSnsUserInfo 资料朋友圈信息
type ProfileSnsUserInfo struct {
	SnsFlag string `json:"snsFlag"` // 朋友圈标志位（1=可见）
}

// MomentsMessage 朋友圈消息回调
type MomentsMessage struct {
	BlackContactTagIdListCount int               `json:"blackContactTagIdListCount"` // 黑名单联系人标签ID列表数量
	BlackListCount             int               `json:"blackListCount"`             // 黑名单数量
	CommentCount               int               `json:"commentCount"`               // 评论数量
	CommentUserListCount       int               `json:"commentUserListCount"`       // 评论用户列表数量
	CreateTime                 int               `json:"createTime"`                 // 创建时间（Unix 时间戳，秒）
	DeleteFlag                 int               `json:"deleteFlag"`                 // 删除标志
	ExtFlag                    int               `json:"extFlag"`                    // 扩展标志
	FlowerFlag                 int               `json:"flowerFlag"`                 // 花标志
	FlowerUserListCount        int               `json:"flowerUserListCount"`        // 花用户列表数量
	GroupContactTagIdListCount int               `json:"groupContactTagIdListCount"` // 群组联系人标签ID列表数量
	GroupCount                 int               `json:"groupCount"`                 // 群组数量
	GroupUserCount             int               `json:"groupUserCount"`             // 群组用户数量
	GuideFinder                int               `json:"guideFinder"`                // 引导发现
	GuideQw                    int               `json:"guideQw"`                    // 引导企业微信
	GuideTop                   int               `json:"guideTop"`                   // 引导置顶
	ID                         string            `json:"id"`                         // 朋友圈ID
	InTopList                  int               `json:"inTopList"`                  // 是否在置顶列表
	IsNotRichText              int               `json:"isNotRichText"`              // 是否非富文本
	LikeCount                  int               `json:"likeCount"`                  // 点赞数量
	LikeFlag                   int               `json:"likeFlag"`                   // 点赞标志
	LikeUserListCount          int               `json:"likeUserListCount"`          // 点赞用户列表数量
	MsgType                    string            `json:"msgType"`                    // 消息类型（朋友圈刷新）
	NewWithTaListCount         int               `json:"newWithTaListCount"`         // 新的与TA列表数量
	Nickname                   string            `json:"nickname"`                   // 昵称
	NoChange                   int               `json:"noChange"`                   // 无变化标志
	ObjectDesc                 MomentsObjectDesc `json:"objectDesc"`                 // 对象描述
	ObjectOperations           MomentsBuffer     `json:"objectOperations"`           // 对象操作
	ObjectType                 int               `json:"objectType"`                 // 对象类型
	PreDownloadInfo            PreDownloadInfo   `json:"preDownloadInfo"`            // 预下载信息
	ReferID                    string            `json:"referId"`                    // 引用ID
	SnsRedEnvelops             SnsRedEnvelops    `json:"snsRedEnvelops"`             // 朋友圈红包
	Username                   string            `json:"username"`                   // 用户名（wxid）
	WeAppInfo                  WeAppInfo         `json:"weAppInfo"`                  // 小程序信息
	WeiShangFeedType           int               `json:"weiShangFeedType"`           // 微商Feed类型
	WeiShangSellerType         int               `json:"weiShangSellerType"`         // 微商卖家类型
	WeiShangVideoSourceType    int               `json:"weiShangVideoSourceType"`    // 微商视频源类型
	WithChatroomListCount      int               `json:"withChatroomListCount"`      // 与群聊列表数量
	WithTaHasOther             int               `json:"withTaHasOther"`             // 与TA有其他
	WithTaListCount            int               `json:"withTaListCount"`            // 与TA列表数量
	WithUserCount              int               `json:"withUserCount"`              // 与用户数量
	WithUserListCount          int               `json:"withUserListCount"`          // 与用户列表数量
}

// MomentsObjectDesc 朋友圈对象描述
type MomentsObjectDesc struct {
	Buffer string `json:"buffer"` // Base64编码的XML内容
	ILen   int    `json:"iLen"`   // 内容长度
}

// MomentsBuffer 朋友圈缓冲区
type MomentsBuffer struct {
	Buffer string `json:"buffer"` // Base64编码的内容
	ILen   int    `json:"iLen"`   // 内容长度
}

// PreDownloadInfo 预下载信息
type PreDownloadInfo struct {
	PreDownloadNetType int `json:"preDownloadNetType"` // 预下载网络类型
	PreDownloadPercent int `json:"preDownloadPercent"` // 预下载百分比
}

// SnsRedEnvelops 朋友圈红包
type SnsRedEnvelops struct {
	ReportID    int `json:"reportId"`    // 报告ID
	ReportKey   int `json:"reportKey"`   // 报告Key
	ResourceID  int `json:"resourceId"`  // 资源ID
	RewardCount int `json:"rewardCount"` // 奖励数量
}

// WeAppInfo 小程序信息
type WeAppInfo struct {
	AppID    int `json:"appId"`    // 小程序ID
	Score    int `json:"score"`    // 评分
	ShowType int `json:"showType"` // 显示类型
}

// ChatWindowSwitchEvent 聊天对象切换回调事件
type ChatWindowSwitchEvent struct {
	AccountWxid string                `json:"account_wxid"` // 本人微信ID
	EventDesc   string                `json:"event_desc"`   // 事件描述，例如 "打开聊天窗口"
	EventType   int                   `json:"event_type"`   // 事件类型，1005 表示窗口切换
	HTTPPort    int                   `json:"http_port"`    // HTTP端口
	Hwnd        int                   `json:"hwnd"`         // 窗口句柄
	NewWxid     string                `json:"new_wxid"`     // 切换到哪个微信ID
	PID         int                   `json:"pid"`          // 进程ID
	UserProfile ChatWindowUserProfile `json:"user_profile"` // 用户资料
}

// ChatWindowUserProfile 聊天窗口用户资料
type ChatWindowUserProfile struct {
	AddContactScene         int                         `json:"addContactScene"`
	AdditionalContactList   ChatWindowAdditionalContact `json:"additionalContactList"`
	AlbumFlag               int                         `json:"albumFlag"`
	AlbumStyle              int                         `json:"albumStyle"`
	Alias                   string                      `json:"alias"`
	BigHeadImgUrl           string                      `json:"bigHeadImgUrl"`
	BitMask                 int64                       `json:"bitMask"`
	BitMask2                string                      `json:"bitMask2"`
	BitVal                  int64                       `json:"bitVal"`
	BitValue2               string                      `json:"bitValue2"`
	ChatRoomBusinessType    string                      `json:"chatRoomBusinessType"`
	ChatRoomNotify          int                         `json:"chatRoomNotify"`
	ChatroomAccessType      int                         `json:"chatroomAccessType"`
	ChatroomInfoVersion     int                         `json:"chatroomInfoVersion"`
	ChatroomMaxCount        int                         `json:"chatroomMaxCount"`
	ChatroomStatus          int                         `json:"chatroomStatus"`
	ChatroomVersion         int                         `json:"chatroomVersion"`
	ContactExtraInfoBuf     ChatWindowBufferInfo        `json:"contactExtraInfoBuf"`
	ContactType             int                         `json:"contactType"`
	CustomizedInfo          ChatWindowCustomizedInfo    `json:"customizedInfo"`
	DeleteContactScene      int                         `json:"deleteContactScene"`
	DeleteFlag              int                         `json:"deleteFlag"`
	DomainList              interface{}                 `json:"domainList"`
	EncryptUserName         string                      `json:"encryptUserName"`
	EraseChatRoomMemberData int                         `json:"eraseChatRoomMemberData"`
	ExtFlag                 int                         `json:"extFlag"`
	FriendUserName          string                      `json:"friendUserName"`
	HasWeiXinHdHeadImg      int                         `json:"hasWeiXinHdHeadImg"`
	HeadImgMd5              string                      `json:"headImgMd5"`
	ImgBuf                  ChatWindowBufferInfo        `json:"imgBuf"`
	ImgFlag                 int                         `json:"imgFlag"`
	IsInChatRoom            int                         `json:"isInChatRoom"`
	Level                   int                         `json:"level"`
	MyBrandList             string                      `json:"myBrandList"`
	NewChatroomData         ChatWindowChatroomData      `json:"newChatroomData"`
	NickName                StringValue                 `json:"nickName"`
	PersonalCard            int                         `json:"personalCard"`
	PhoneNumListInfo        ChatWindowPhoneNumInfo      `json:"phoneNumListInfo"`
	Pyinitial               StringValue                 `json:"pyinitial"`
	QuanPin                 StringValue                 `json:"quanPin"`
	Remark                  interface{}                 `json:"remark"`
	RemarkPyinitial         interface{}                 `json:"remarkPyinitial"`
	RemarkQuanPin           interface{}                 `json:"remarkQuanPin"`
	RingBackSetting         ChatWindowRingBackSetting   `json:"ringBackSetting"`
	RoomInfoCount           int                         `json:"roomInfoCount"`
	Sex                     int                         `json:"sex"`
	Signature               string                      `json:"signature"`
	SmallHeadImgUrl         string                      `json:"smallHeadImgUrl"`
	SnsUserInfo             ChatWindowSnsUserInfo       `json:"snsUserInfo"`
	Source                  int                         `json:"source"`
	TextStatusFlag          int                         `json:"textStatusFlag"`
	UserName                StringValue                 `json:"userName"`
	VerifyFlag              int                         `json:"verifyFlag"`
	WeiboFlag               int                         `json:"weiboFlag"`
}

// ChatWindowAdditionalContact 附加联系人信息
type ChatWindowAdditionalContact struct {
	LinkedinContactItem interface{} `json:"linkedinContactItem"`
}

// ChatWindowBufferInfo 缓冲区信息
type ChatWindowBufferInfo struct {
	ILen int `json:"iLen"`
}

// ChatWindowCustomizedInfo 自定义信息
type ChatWindowCustomizedInfo struct {
	BrandFlag int `json:"brandFlag"`
}

// ChatWindowChatroomData 群聊数据
type ChatWindowChatroomData struct {
	ChatRoomUserName interface{} `json:"chatRoomUserName"`
	InfoMask         int         `json:"infoMask"`
	MemberCount      int         `json:"memberCount"`
	WatchMemberCount int         `json:"watchMemberCount"`
}

// ChatWindowPhoneNumInfo 手机号信息
type ChatWindowPhoneNumInfo struct {
	Count int `json:"count"`
}

// ChatWindowRingBackSetting 回铃设置
type ChatWindowRingBackSetting struct {
	EndTs          int    `json:"endTs"`
	FinderObjectId string `json:"finderObjectId"`
	StartTs        int    `json:"startTs"`
}

// ChatWindowSnsUserInfo SNS用户信息
type ChatWindowSnsUserInfo struct {
	SnsBgimgId       string `json:"snsBgimgId"`
	SnsBgobjectId    string `json:"snsBgobjectId"`
	SnsFlag          int    `json:"snsFlag"`
	SnsFlagEx        int    `json:"snsFlagEx"`
	SnsPrivacyRecent int    `json:"snsPrivacyRecent"`
}

// RoomMemberNicknameChangeEvent 群成员修改昵称回调事件
type RoomMemberNicknameChangeEvent struct {
	AccountWxid string                       `json:"account_wxid"` // 本人微信ID
	Data        RoomMemberNicknameChangeData `json:"data"`         // 事件数据
	EventDesc   string                       `json:"event_desc"`   // 事件描述，例如 "群员昵称修改通知"
	EventType   int                          `json:"event_type"`   // 事件类型，1012 表示群员昵称修改
	HTTPPort    int                          `json:"http_port"`    // HTTP端口
	Hwnd        int                          `json:"hwnd"`         // 窗口句柄
	PID         int                          `json:"pid"`          // 进程ID
}

// RoomMemberNicknameChangeData 群成员昵称修改数据
type RoomMemberNicknameChangeData struct {
	CreateTime  int                    `json:"createtime"`  // 创建时间
	MemberCount int                    `json:"membercount"` // 成员数量
	MemberList  RoomMemberNicknameInfo `json:"memberlist"`  // 成员信息
	RoomID      string                 `json:"roomid"`      // 群聊ID
	RoomName    string                 `json:"roomname"`    // 群聊名称
}

// RoomMemberNicknameInfo 群成员昵称信息
type RoomMemberNicknameInfo struct {
	BigHeadImgUrl      string `json:"bigHeadImgUrl"`      // 大头像URL
	ChatroomMemberFlag int    `json:"chatroomMemberFlag"` // 群成员标志
	DisplayName        string `json:"displayName"`        // 显示名称（群昵称）
	NickName           string `json:"nickName"`           // 昵称
	SmallHeadImgUrl    string `json:"smallHeadImgUrl"`    // 小头像URL
	Status             int    `json:"status"`             // 状态
	UserName           string `json:"userName"`           // 用户微信ID
}

// RoomMemberJoinEvent 群成员进群回调事件
type RoomMemberJoinEvent struct {
	AccountWxid string             `json:"account_wxid"` // 当前登录账号ID
	Data        RoomMemberJoinData `json:"data"`         // 事件数据
	EventDesc   string             `json:"event_desc"`   // 事件描述
	EventType   string             `json:"event_type"`   // 事件类型
	HTTPPort    string             `json:"http_port"`    // HTTP服务端口
	Hwnd        string             `json:"hwnd"`         // 窗口句柄
	PID         string             `json:"pid"`          // 进程ID
}

// RoomMemberJoinData 群成员进群数据
type RoomMemberJoinData struct {
	CreateTime  string             `json:"createtime"`  // 事件创建时间
	MemberCount string             `json:"membercount"` // 群成员数量
	MemberList  RoomMemberJoinInfo `json:"memberlist"`  // 成员信息
	RoomID      string             `json:"roomid"`      // 群ID
	RoomName    string             `json:"roomname"`    // 群名称
}

// RoomMemberJoinInfo 群成员进群信息
type RoomMemberJoinInfo struct {
	AddChatRoomSceneNewXml string `json:"addChatRoomSceneNewXml"` // 添加群成员场景XML
	BigHeadImgUrl          string `json:"bigHeadImgUrl"`          // 大头像URL
	ChatroomMemberFlag     string `json:"chatroomMemberFlag"`     // 群成员标志
	InviterUserName        string `json:"inviterUserName"`        // 邀请者用户名
	NickName               string `json:"nickName"`               // 昵称
	SmallHeadImgUrl        string `json:"smallHeadImgUrl"`        // 小头像URL
	Status                 string `json:"status"`                 // 状态
	UserName               string `json:"userName"`               // 用户名
}

// RoomMemberLeaveEvent 群成员退群回调事件
type RoomMemberLeaveEvent struct {
	AccountWxid string              `json:"account_wxid"` // 当前登录账号ID
	Data        RoomMemberLeaveData `json:"data"`         // 事件数据
	EventDesc   string              `json:"event_desc"`   // 事件描述
	EventType   string              `json:"event_type"`   // 事件类型
	HTTPPort    string              `json:"http_port"`    // HTTP服务端口
	Hwnd        string              `json:"hwnd"`         // 窗口句柄
	PID         string              `json:"pid"`          // 进程ID
}

// RoomMemberLeaveData 群成员退群数据
type RoomMemberLeaveData struct {
	CreateTime  string              `json:"createtime"`  // 事件创建时间
	MemberCount string              `json:"membercount"` // 群成员数量
	MemberList  RoomMemberLeaveInfo `json:"memberlist"`  // 成员信息
	RoomID      string              `json:"roomid"`      // 群ID
	RoomName    string              `json:"roomname"`    // 群名称
}

// RoomMemberLeaveInfo 群成员退群信息
type RoomMemberLeaveInfo struct {
	AddChatRoomSceneNewXml string `json:"addChatRoomSceneNewXml"` // 群成员场景XML
	BigHeadImgUrl          string `json:"bigHeadImgUrl"`          // 大头像URL
	ChatroomMemberFlag     string `json:"chatroomMemberFlag"`     // 群成员标志
	InviterUserName        string `json:"inviterUserName"`        // 邀请者用户名
	NickName               string `json:"nickName"`               // 昵称
	SmallHeadImgUrl        string `json:"smallHeadImgUrl"`        // 小头像URL
	Status                 string `json:"status"`                 // 状态
	UserName               string `json:"userName"`               // 用户名
}
