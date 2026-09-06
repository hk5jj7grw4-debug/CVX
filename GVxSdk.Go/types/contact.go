package types

// GetMyQRCodeRequest 获取好友二维码请求
type GetMyQRCodeRequest struct {
	Wxid   string `json:"wxid"`   // 微信ID或群聊ID
	Opcode string `json:"opcode"` // 操作码，通常为 "0"
	Style  string `json:"style"`  // 风格，1-8 可选
	Info   string `json:"info"`   // 说明信息
}

// GetMyQRCodeResponse 获取好友二维码响应
type GetMyQRCodeResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	QRCode       QRCodeData   `json:"qrcode"`
}

// QRCodeData 二维码数据
type QRCodeData struct {
	ILen               int    `json:"iLen"`               // 数据长度
	Buffer             string `json:"buffer"`             // Base64编码的二维码图片数据
	Style              int    `json:"style"`              // 风格
	DominatorColorSize int    `json:"dominatorColorSize"` // 主色调大小
}

// SearchContactRequest 搜索微信号/手机号请求
type SearchContactRequest struct {
	Search string `json:"search"` // 搜索关键词：微信号或手机号
}

// SearchContactResponse 搜索微信号/手机号响应
type SearchContactResponse struct {
	BaseResponse          BaseResponse          `json:"baseResponse"`
	UserName              StringValue           `json:"userName"`
	NickName              StringValue           `json:"nickName"`
	Initial               StringValue           `json:"initial"`
	QuanPin               StringValue           `json:"quanPin"`
	Sex                   int                   `json:"sex"`
	ImgBuf                BufferData            `json:"imgBuf"`
	Signature             string                `json:"signature"`
	PersonalCard          int                   `json:"personalCard"`
	VerifyFlag            int                   `json:"verifyFlag"`
	WeiboFlag             int                   `json:"weiboFlag"`
	AlbumStyle            int                   `json:"albumStyle"`
	AlbumFlag             int                   `json:"albumFlag"`
	SnsUserInfo           SnsUserInfo           `json:"snsUserInfo"`
	CustomizedInfo        CustomizedInfo        `json:"customizedInfo"`
	ContactCount          int                   `json:"contactCount"`
	BigHeadImgUrl         string                `json:"bigHeadImgUrl"`
	SmallHeadImgUrl       string                `json:"smallHeadImgUrl"`
	ResBuf                BufferData            `json:"resBuf"`
	AntispamTicket        string                `json:"antispamTicket"`
	MatchType             int                   `json:"matchType"`
	ExtFlag               int                   `json:"extFlag"`
	SearchContactJumpInfo SearchContactJumpInfo `json:"searchContactJumpInfo"`
}

// StringValue 字符串值
type StringValue struct {
	String string `json:"String"`
}

// BufferData 缓冲区数据
type BufferData struct {
	ILen int `json:"iLen"`
}

// SnsUserInfo SNS用户信息
type SnsUserInfo struct {
	SnsFlag          int    `json:"snsFlag"`
	SnsBgobjectId    string `json:"snsBgobjectId"`
	SnsFlagEx        int    `json:"snsFlagEx"`
	SnsPrivacyRecent int    `json:"snsPrivacyRecent"`
}

// CustomizedInfo 自定义信息
type CustomizedInfo struct {
	BrandFlag int `json:"brandFlag"`
}

// SearchContactJumpInfo 搜索联系人跳转信息
type SearchContactJumpInfo struct {
	// 可根据实际需要添加字段
}

// GetProfileCacheResponse 获取个人资料缓存响应
type GetProfileCacheResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	UserInfo     UserInfo     `json:"userInfo"`
	UserInfoExt  UserInfoExt  `json:"userInfoExt"`
}

// UserInfo 用户信息
type UserInfo struct {
	BitFlag                                 int                   `json:"bitFlag"`
	UserName                                StringValue           `json:"userName"`
	NickName                                StringValue           `json:"nickName"`
	BindUin                                 int                   `json:"bindUin"`
	BindEmail                               interface{}           `json:"bindEmail"`
	BindMobile                              StringValue           `json:"bindMobile"`
	Status                                  int                   `json:"status"`
	ImgLen                                  int                   `json:"imgLen"`
	Sex                                     int                   `json:"sex"`
	Province                                string                `json:"province"`
	City                                    string                `json:"city"`
	Signature                               string                `json:"signature"`
	PersonalCard                            int                   `json:"personalCard"`
	DisturbSetting                          DisturbSetting        `json:"disturbSetting"`
	PluginFlag                              int                   `json:"pluginFlag"`
	VerifyFlag                              int                   `json:"verifyFlag"`
	Point                                   int                   `json:"point"`
	Experience                              int                   `json:"experience"`
	Level                                   int                   `json:"level"`
	LevelLowExp                             int                   `json:"levelLowExp"`
	LevelHighExp                            int                   `json:"levelHighExp"`
	PluginSwitch                            int                   `json:"pluginSwitch"`
	GmailList                               GmailList             `json:"gmailList"`
	Alias                                   string                `json:"alias"`
	WeiboFlag                               int                   `json:"weiboFlag"`
	FaceBookFlag                            int                   `json:"faceBookFlag"`
	FbuserId                                string                `json:"fbuserId"`
	AlbumStyle                              int                   `json:"albumStyle"`
	AlbumFlag                               int                   `json:"albumFlag"`
	TxnewsCategory                          int                   `json:"txnewsCategory"`
	Country                                 string                `json:"country"`
	UserInfoExt                             UserInfoExt           `json:"userInfoExt"`
	BigHeadImgUrl                           string                `json:"bigHeadImgUrl"`
	SmallHeadImgUrl                         string                `json:"smallHeadImgUrl"`
	MainAcctType                            int                   `json:"mainAcctType"`
	ExtXml                                  interface{}           `json:"extXml"`
	SafeDeviceList                          SafeDeviceList        `json:"safeDeviceList"`
	SafeDevice                              int                   `json:"safeDevice"`
	GrayscaleFlag                           int                   `json:"grayscaleFlag"`
	RegCountry                              string                `json:"regCountry"`
	LinkedinContactItem                     interface{}           `json:"linkedinContactItem"`
	PatternLockInfo                         PatternLockInfo       `json:"patternLockInfo"`
	PayWalletType                           int                   `json:"payWalletType"`
	WalletRegion                            int                   `json:"walletRegion"`
	ExtStatus                               string                `json:"extStatus"`
	UserStatus                              int                   `json:"userStatus"`
	PaySetting                              string                `json:"paySetting"`
	PatSuffix                               string                `json:"patSuffix"`
	PatSuffixVersion                        int                   `json:"patSuffixVersion"`
	TeenagerModeFinderSetting               int                   `json:"teenagerModeFinderSetting"`
	TeenagerModeBizAcctSetting              int                   `json:"teenagerModeBizAcctSetting"`
	TeenagerModeMiniProgramSetting          int                   `json:"teenagerModeMiniProgramSetting"`
	XagreementInfo                          XagreementInfo        `json:"xagreementInfo"`
	Salt                                    string                `json:"salt"`
	FinderSetting                           string                `json:"finderSetting"`
	RingBackSetting                         RingBackSetting       `json:"ringBackSetting"`
	SmcryptoFlag                            int                   `json:"smcryptoFlag"`
	GlobalRingBackSetting                   GlobalRingBackSetting `json:"globalRingBackSetting"`
	NewcomeMsgDefaultVoiceNumber            int                   `json:"newcomeMsgDefaultVoiceNumber"`
	DiscoveryPageCtrlFlag                   string                `json:"discoveryPageCtrlFlag"`
	ExtStatus2                              string                `json:"extStatus2"`
	FinderLiveAliasSync                     FinderLiveAliasSync   `json:"finderLiveAliasSync"`
	SpamFlag                                int                   `json:"spamFlag"`
	DeleteTime                              string                `json:"deleteTime"`
	LiveAliasRoleType                       int                   `json:"liveAliasRoleType"`
	VerifyContentList                       VerifyContentList     `json:"verifyContentList"`
	Lqtversion                              int                   `json:"lqtversion"`
	TeenagerModeEmotionSetting              int                   `json:"teenagerModeEmotionSetting"`
	NotificationBannerDisplayContentSetting int                   `json:"notificationBannerDisplayContentSetting"`
}

// DisturbSetting 免打扰设置
type DisturbSetting struct {
	NightSetting  int       `json:"nightSetting"`
	NightTime     TimeRange `json:"nightTime"`
	AllDaySetting int       `json:"allDaySetting"`
	AllDayTime    TimeRange `json:"allDayTime"`
}

// TimeRange 时间范围
type TimeRange struct {
	BeginTime int `json:"beginTime"`
	EndTime   int `json:"endTime"`
}

// GmailList Gmail列表
type GmailList struct {
	Count int `json:"count"`
}

// UserInfoExt 用户信息扩展
type UserInfoExt struct {
	SnsUserInfo       SnsUserInfoExt `json:"snsUserInfo"`
	MyBrandList       string         `json:"myBrandList"`
	BigChatRoomSize   int            `json:"bigChatRoomSize"`
	BigChatRoomQuota  int            `json:"bigChatRoomQuota"`
	BigChatRoomInvite int            `json:"bigChatRoomInvite"`
	BigHeadImgUrl     string         `json:"bigHeadImgUrl"`
	SmallHeadImgUrl   string         `json:"smallHeadImgUrl"`
}

// SnsUserInfoExt SNS用户信息扩展
type SnsUserInfoExt struct {
	SnsFlag          int    `json:"snsFlag"`
	SnsBgimgId       string `json:"snsBgimgId"`
	SnsBgobjectId    string `json:"snsBgobjectId"`
	SnsFlagEx        int    `json:"snsFlagEx"`
	SnsPrivacyRecent int    `json:"snsPrivacyRecent"`
}

// SafeDeviceList 安全设备列表
type SafeDeviceList struct {
	Count int          `json:"count"`
	List  []SafeDevice `json:"list"`
}

// SafeDevice 安全设备
type SafeDevice struct {
	Name       string `json:"name"`
	Uuid       string `json:"uuid"`
	DeviceType string `json:"deviceType"`
	CreateTime int    `json:"createTime"`
}

// PatternLockInfo 图案锁信息
type PatternLockInfo struct {
	PatternVersion int        `json:"patternVersion"`
	Sign           SignBuffer `json:"sign"`
	LockStatus     int        `json:"lockStatus"`
}

// SignBuffer 签名缓冲区
type SignBuffer struct {
	ILen   int    `json:"iLen"`
	Buffer string `json:"buffer"`
}

// XagreementInfo 协议信息
type XagreementInfo struct {
	FuncsSwitch           string `json:"funcsSwitch"`
	FuncsUserChoiceSwitch string `json:"funcsUserChoiceSwitch"`
}

// RingBackSetting 回铃设置
type RingBackSetting struct {
	FinderObjectId string `json:"finderObjectId"`
	StartTs        int    `json:"startTs"`
	EndTs          int    `json:"endTs"`
}

// GlobalRingBackSetting 全局回铃设置
type GlobalRingBackSetting struct {
	Type      int           `json:"type"`
	StartTime int           `json:"startTime"`
	EndTime   int           `json:"endTime"`
	Music     MusicSetting  `json:"music"`
	Finder    FinderSetting `json:"finder"`
}

// MusicSetting 音乐设置
type MusicSetting struct {
	Sid int `json:"sid"`
}

// FinderSetting Finder设置
type FinderSetting struct {
	FinderObjectId string `json:"finderObjectId"`
}

// FinderLiveAliasSync Finder直播别名同步
type FinderLiveAliasSync struct {
	UpdateTime string `json:"updateTime"`
}

// VerifyContentList 验证内容列表
type VerifyContentList struct {
	Count int `json:"count"`
}

// GetContactRequest 网络查询好友资料请求
type GetContactRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// GetContactResponse 网络查询好友资料响应
type GetContactResponse struct {
	BaseResponse              BaseResponse              `json:"baseResponse"`
	ContactCount              int                       `json:"contactCount"`
	ContactList               []ContactInfo             `json:"contactList"`
	AdditionalContactList     AdditionalContactList     `json:"additionalContactList"`
	ChatroomVersion           int                       `json:"chatroomVersion"`
	ChatroomMaxCount          int                       `json:"chatroomMaxCount"`
	ChatroomAccessType        int                       `json:"chatroomAccessType"`
	NewChatroomData           NewChatroomDataSimple     `json:"newChatroomData"`
	DeleteFlag                int                       `json:"deleteFlag"`
	PhoneNumListInfo          PhoneNumListInfo          `json:"phoneNumListInfo"`
	ChatroomInfoVersion       int                       `json:"chatroomInfoVersion"`
	DeleteContactScene        int                       `json:"deleteContactScene"`
	ChatroomStatus            int                       `json:"chatroomStatus"`
	ExtFlag                   int                       `json:"extFlag"`
	ChatRoomBusinessType      string                    `json:"chatRoomBusinessType"`
	FriendUserName            string                    `json:"friendUserName"`
	TextStatusFlag            int                       `json:"textStatusFlag"`
	RingBackSetting           RingBackSetting           `json:"ringBackSetting"`
	BitMask2                  string                    `json:"bitMask2"`
	BitValue2                 string                    `json:"bitValue2"`
	ContactExtraInfoBuf       BufferData                `json:"contactExtraInfoBuf"`
	IsInChatRoom              int                       `json:"isInChatRoom"`
	EraseChatRoomMemberData   int                       `json:"eraseChatRoomMemberData"`
	Ret                       []int                     `json:"ret"`
	VerifyUserValidTicketList VerifyUserValidTicketList `json:"verifyUserValidTicketList"`
}

// ContactInfo 联系人信息
type ContactInfo struct {
	UserName           StringValue    `json:"userName"`
	NickName           StringValue    `json:"nickName"`
	Pyinitial          StringValue    `json:"pyinitial"`
	QuanPin            StringValue    `json:"quanPin"`
	Sex                int            `json:"sex"`
	ImgBuf             BufferData     `json:"imgBuf"`
	BitMask            int64          `json:"bitMask"`
	BitVal             int64          `json:"bitVal"`
	ImgFlag            int            `json:"imgFlag"`
	Remark             StringValue    `json:"remark"`
	RemarkPyinitial    StringValue    `json:"remarkPyinitial"`
	RemarkQuanPin      StringValue    `json:"remarkQuanPin"`
	ContactType        int            `json:"contactType"`
	RoomInfoCount      int            `json:"roomInfoCount"`
	DomainList         interface{}    `json:"domainList"`
	ChatRoomNotify     int            `json:"chatRoomNotify"`
	AddContactScene    int            `json:"addContactScene"`
	Province           string         `json:"province"`
	City               string         `json:"city"`
	Signature          string         `json:"signature"`
	PersonalCard       int            `json:"personalCard"`
	HasWeiXinHdHeadImg int            `json:"hasWeiXinHdHeadImg"`
	VerifyFlag         int            `json:"verifyFlag"`
	Level              int            `json:"level"`
	Source             int            `json:"source"`
	Alias              string         `json:"alias"`
	WeiboFlag          int            `json:"weiboFlag"`
	AlbumStyle         int            `json:"albumStyle"`
	AlbumFlag          int            `json:"albumFlag"`
	SnsUserInfo        SnsUserInfo    `json:"snsUserInfo"`
	Country            string         `json:"country"`
	BigHeadImgUrl      string         `json:"bigHeadImgUrl"`
	SmallHeadImgUrl    string         `json:"smallHeadImgUrl"`
	MyBrandList        string         `json:"myBrandList"`
	CustomizedInfo     CustomizedInfo `json:"customizedInfo"`
	HeadImgMd5         string         `json:"headImgMd5"`
	EncryptUserName    string         `json:"encryptUserName"`
}

// AdditionalContactList 附加联系人列表
type AdditionalContactList struct {
	LinkedinContactItem interface{} `json:"linkedinContactItem"`
}

// NewChatroomDataSimple 简单的新群聊数据
type NewChatroomDataSimple struct {
	MemberCount      int         `json:"memberCount"`
	InfoMask         int         `json:"infoMask"`
	ChatRoomUserName interface{} `json:"chatRoomUserName"`
	WatchMemberCount int         `json:"watchMemberCount"`
}

// PhoneNumListInfo 手机号列表信息
type PhoneNumListInfo struct {
	Count int `json:"count"`
}

// VerifyUserValidTicketList 验证用户有效票据列表
type VerifyUserValidTicketList struct {
	Username       string `json:"username"`
	Antispamticket string `json:"antispamticket"`
}

// ModSelfNickNameRequest 修改自己昵称请求
type ModSelfNickNameRequest struct {
	NewName string `json:"newName"` // 新昵称
}

// ModSelfNickNameResponse 修改自己昵称响应
type ModSelfNickNameResponse struct {
	Ret      int      `json:"ret"`
	OplogRet OplogRet `json:"oplogRet"`
	NewName  string   `json:"newName,omitempty"`
}

// OplogRet 操作日志返回
type OplogRet struct {
	Count  int           `json:"count"`
	Ret    []int         `json:"ret"`
	ErrMsg []interface{} `json:"errMsg"`
}

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	V3            string `json:"v3"`            // V3加密用户名
	V4            string `json:"v4"`            // V4验证票据
	Scence        string `json:"scence"`        // 场景值，例如 "3"
	FriendFlg     string `json:"friendFlg"`     // 好友标志，例如 "0"
	VerifyContent string `json:"verifyContent"` // 验证消息内容
}

// AddFriendResponse 添加好友响应
type AddFriendResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// GetProfileNewResponse 获取个人最新网络响应
// 注意：此类型与 GetProfileCacheResponse 结构相同，复用 UserInfo 类型
type GetProfileNewResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	UserInfo     UserInfo     `json:"userInfo"`
	UserInfoExt  UserInfoExt  `json:"userInfoExt"`
}

// GetContactFastRequest 快速查找好友资料请求
type GetContactFastRequest struct {
	Wxid string `json:"wxid"` // 微信ID，可以为空表示更新所有好友
}

// GetContactFastResponse 快速查找好友资料响应
type GetContactFastResponse struct {
	Contact  ContactFastInfo `json:"contact"`
	Ret      int             `json:"ret"`
	Username string          `json:"username"`
}

// ContactFastInfo 快速查询联系人信息
type ContactFastInfo struct {
	Alias              string          `json:"alias"`
	BigHeadImgUrl      string          `json:"bigHeadImgUrl"`
	BitMask            int64           `json:"bitMask"`
	BitVal             int64           `json:"bitVal"`
	City               string          `json:"city"`
	Country            string          `json:"country"`
	EncryptUserName    string          `json:"encryptUserName"`
	HasWeiXinHdHeadImg int             `json:"hasWeiXinHdHeadImg"`
	ImgBuf             ImgBufInfo      `json:"imgBuf"`
	ImgFlag            int             `json:"imgFlag"`
	NickName           StringValue     `json:"nickName"`
	Province           string          `json:"province"`
	Pyinitial          StringValue     `json:"pyinitial"`
	QuanPin            StringValue     `json:"quanPin"`
	Remark             StringValue     `json:"remark"`
	RemarkPyinitial    StringValue     `json:"remarkPyinitial"`
	RemarkQuanPin      StringValue     `json:"remarkQuanPin"`
	Sex                int             `json:"sex"`
	SmallHeadImgUrl    string          `json:"smallHeadImgUrl"`
	SnsUserInfo        SnsUserInfoFast `json:"snsUserInfo"`
	TextStatusExtInfo  string          `json:"textStatusExtInfo"`
	TextStatusFlag     int             `json:"textStatusFlag"`
	TextStatusId       string          `json:"textStatusId"`
	UserName           StringValue     `json:"userName"`
	VerifyFlag         int             `json:"verifyFlag"`
}

// ImgBufInfo 图片缓冲区信息
type ImgBufInfo struct {
	Buffer string `json:"buffer"`
	ILen   int    `json:"iLen"`
}

// SnsUserInfoFast 快速SNS用户信息
type SnsUserInfoFast struct {
	SnsFlag int `json:"snsFlag"`
}

// UpdateAllFriendResponse 更新好友列表响应
type UpdateAllFriendResponse struct {
	Data        []FriendUpdateInfo `json:"data"`
	FriendCount int                `json:"friend_count"`
}

// FriendUpdateInfo 好友更新信息
type FriendUpdateInfo struct {
	Contact           FriendContactInfo `json:"contact"`
	Ret               int               `json:"ret"`
	Username          string            `json:"username"`
	CustomizedInfo    FriendCustomInfo  `json:"customizedInfo,omitempty"`
	Description       string            `json:"description,omitempty"`
	LabelIdlist       string            `json:"labelIdlist,omitempty"`
	PhoneNumListInfo  PhoneNumInfo      `json:"phoneNumListInfo,omitempty"`
	TextStatusExtInfo string            `json:"textStatusExtInfo,omitempty"`
	TextStatusId      string            `json:"textStatusId,omitempty"`
	ContactType       int               `json:"contactType,omitempty"`
	DeleteFlag        int               `json:"deleteFlag,omitempty"`
	ChatroomVersion   int               `json:"chatroomVersion,omitempty"`
}

// FriendContactInfo 好友联系人信息
type FriendContactInfo struct {
	Alias              string          `json:"alias"`
	BigHeadImgUrl      string          `json:"bigHeadImgUrl"`
	BitMask            int64           `json:"bitMask"`
	BitVal             int64           `json:"bitVal"`
	City               string          `json:"city"`
	Country            string          `json:"country"`
	EncryptUserName    string          `json:"encryptUserName"`
	HasWeiXinHdHeadImg int             `json:"hasWeiXinHdHeadImg"`
	ImgBuf             ImgBufInfo      `json:"imgBuf"`
	ImgFlag            int             `json:"imgFlag"`
	NickName           StringValue     `json:"nickName"`
	Province           string          `json:"province"`
	Pyinitial          StringValue     `json:"pyinitial"`
	QuanPin            StringValue     `json:"quanPin"`
	Remark             StringValue     `json:"remark"`
	RemarkPyinitial    StringValue     `json:"remarkPyinitial"`
	RemarkQuanPin      StringValue     `json:"remarkQuanPin"`
	Sex                int             `json:"sex"`
	SmallHeadImgUrl    string          `json:"smallHeadImgUrl"`
	SnsUserInfo        SnsUserInfoFast `json:"snsUserInfo"`
	TextStatusFlag     int             `json:"textStatusFlag"`
	UserName           StringValue     `json:"userName"`
	VerifyFlag         int             `json:"verifyFlag"`
}

// FriendCustomInfo 好友自定义信息
type FriendCustomInfo struct {
	BrandFlag    int    `json:"brandFlag"`
	BrandIconUrl string `json:"brandIconUrl"`
	ExternalInfo string `json:"externalInfo"`
}

// PhoneNumInfo 手机号信息
type PhoneNumInfo struct {
	Count int `json:"count"`
}

// RemarkContactRequest 修改好友备注请求
type RemarkContactRequest struct {
	Wxid   string `json:"wxid"`   // 微信ID
	Remark string `json:"remark"` // 新备注名
}

// RemarkContactResponse 修改好友备注响应
type RemarkContactResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// DelContactRequest 删除好友请求
type DelContactRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// DelContactResponse 删除好友响应
type DelContactResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// UploadHeadImgRequest 修改头像请求
type UploadHeadImgRequest struct {
	FilePath string `json:"filepath"` // 头像图片文件路径
}

// UploadHeadImgResponse 修改头像响应
type UploadHeadImgResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// ModSelfSignatureRequest 修改个人签名请求
type ModSelfSignatureRequest struct {
	NewSignature string `json:"newSignature"` // 新的个人签名
}

// ModSelfSignatureResponse 修改个人签名响应
type ModSelfSignatureResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// GetLabelListsResponse 获取标签列表响应
type GetLabelListsResponse struct {
	BaseResponse  BaseResponse `json:"baseResponse"`
	LabelCount    int          `json:"labelCount"`
	LabelPairList []LabelPair  `json:"labelPairList"`
}

// LabelPair 标签对
type LabelPair struct {
	LabelName string `json:"labelName"` // 标签名称
	LabelId   int    `json:"labelId"`   // 标签ID
}

// AddLabelRequest 增加标签请求
type AddLabelRequest struct {
	Label string `json:"label"` // 标签名称
}

// AddLabelResponse 增加标签响应
type AddLabelResponse struct {
	BaseResponse  BaseResponse `json:"baseResponse"`
	LabelCount    int          `json:"labelCount"`
	LabelPairList []LabelPair  `json:"labelPairList"`
}

// DelLabelRequest 删除标签请求
type DelLabelRequest struct {
	LabelID string `json:"label_id"` // 标签ID
}

// DelLabelResponse 删除标签响应
type DelLabelResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// ModifyContactLabelRequest 修改好友标签请求
type ModifyContactLabelRequest struct {
	Wxids   string `json:"wxids"`   // 好友微信ID
	LabelID string `json:"labelId"` // 标签ID列表，多个用逗号分隔
}

// ModifyContactLabelResponse 修改好友标签响应
type ModifyContactLabelResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// UpdateLabelNameRequest 更新标签名字请求
type UpdateLabelNameRequest struct {
	LabelID int    `json:"labelId"` // 标签ID
	NewName string `json:"newName"` // 新标签名字
}

// UpdateLabelNameResponse 更新标签名字响应
type UpdateLabelNameResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// UpdateSingleProfileRequest 更新单个用户资料请求
type UpdateSingleProfileRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// UpdateSingleProfileResponse 更新单个用户资料响应
type UpdateSingleProfileResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
	// 实际返回的用户资料数据（当 code 为 0 时）
	Profile *SingleProfileData `json:"-"` // 不参与JSON序列化，由程序手动解析
}

// SingleProfileData 单个用户资料数据
type SingleProfileData struct {
	Alias              string          `json:"alias"`
	BigHeadImgUrl      string          `json:"bigHeadImgUrl"`
	BitMask            int64           `json:"bitMask"`
	BitVal             int64           `json:"bitVal"`
	City               string          `json:"city"`
	Country            string          `json:"country"`
	EncryptUserName    string          `json:"encryptUserName"`
	HasWeiXinHdHeadImg int             `json:"hasWeiXinHdHeadImg"`
	ImgBuf             ImgBufInfo      `json:"imgBuf"`
	ImgFlag            int             `json:"imgFlag"`
	NickName           StringValue     `json:"nickName"`
	Province           string          `json:"province"`
	Pyinitial          StringValue     `json:"pyinitial"`
	QuanPin            StringValue     `json:"quanPin"`
	Remark             StringValue     `json:"remark"`
	RemarkPyinitial    StringValue     `json:"remarkPyinitial"`
	RemarkQuanPin      StringValue     `json:"remarkQuanPin"`
	Sex                int             `json:"sex"`
	SmallHeadImgUrl    string          `json:"smallHeadImgUrl"`
	SnsUserInfo        SnsUserInfoFast `json:"snsUserInfo"`
	TextStatusExtInfo  string          `json:"textStatusExtInfo"`
	TextStatusFlag     int             `json:"textStatusFlag"`
	TextStatusId       string          `json:"textStatusId"`
	UserName           StringValue     `json:"userName"`
	VerifyFlag         int             `json:"verifyFlag"`
}

// SetTopRequest 置顶好友请求
type SetTopRequest struct {
	Wxid string `json:"wxid"` // 微信ID或群聊ID
}

// CancelTopRequest 取消置顶请求
type CancelTopRequest struct {
	Wxid string `json:"wxid"` // 微信ID或群聊ID
}

// SetStarRequest 星标好友请求
type SetStarRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// DelStarRequest 取消星标请求
type DelStarRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// SetMuteUserRequest 开启消息免打扰请求
type SetMuteUserRequest struct {
	Wxid string `json:"wxid"` // 微信ID或群聊ID
}

// DelMuteUserRequest 关闭消息免打扰请求
type DelMuteUserRequest struct {
	Wxid string `json:"wxid"` // 微信ID或群聊ID
}

// BlackUserRequest 拉黑好友请求
type BlackUserRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// DelBlackUserRequest 移出黑名单请求
type DelBlackUserRequest struct {
	Wxid string `json:"wxid"` // 微信ID
}

// VerifyFriendRequest 同意好友申请请求
type VerifyFriendRequest struct {
	Wxid   string `json:"wxid"`   // 微信ID或V3加密用户名
	V4     string `json:"v4"`     // V4验证票据
	Remark string `json:"remark"` // 备注名
	Label  string `json:"label"`  // 标签ID
	Scene  int    `json:"scene"`  // 来源场景值
}

// VerifyFriendResponse 同意好友申请响应
type VerifyFriendResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// GetContactList2Response 获取好友列表方法2响应
type GetContactList2Response struct {
	FriendCount int           `json:"friend_count"`
	FriendList  []FriendInfo2 `json:"friend_list"`
}

// FriendInfo2 好友信息（方法2）
type FriendInfo2 struct {
	Alias         string `json:"alias"`          // 微信号
	BigHeadUrl    string `json:"big_head_url"`   // 大头像URL
	City          string `json:"city"`           // 城市
	Country       string `json:"country"`        // 国家
	Description   string `json:"description"`    // 描述/个性签名
	Jianping      string `json:"jianping"`       // 简拼
	NickName      string `json:"nick_name"`      // 昵称
	Pinying       string `json:"pinying"`        // 拼音
	Province      string `json:"province"`       // 省份
	Remark        string `json:"remark"`         // 备注名
	RemarkPinying string `json:"remarkpinying"`  // 备注拼音
	SmallHeadUrl  string `json:"small_head_url"` // 小头像URL
	Wxid          string `json:"wxid"`           // 微信ID
}

// GetFriendWxidsResponse 获取所有好友wxid响应
type GetFriendWxidsResponse struct {
	Code  int      `json:"code"`
	Msg   string   `json:"msg"`
	Wxids []string `json:"wxids,omitempty"` // wxid列表
	Count int      `json:"count,omitempty"` // 好友数量
}

// BatchGetWxidsRequest 批量获取wxid信息请求
type BatchGetWxidsRequest struct {
	Wxids string `json:"wxids"` // 微信ID列表，用逗号分隔
}

// BatchGetWxidsResponse 批量获取wxid信息响应
type BatchGetWxidsResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data []BatchWxidInfo `json:"data,omitempty"`
}

// BatchWxidInfo 批量wxid信息
type BatchWxidInfo struct {
	Wxid            string `json:"wxid"`
	NickName        string `json:"nickName"`
	Alias           string `json:"alias"`
	Remark          string `json:"remark"`
	BigHeadImgUrl   string `json:"bigHeadImgUrl"`
	SmallHeadImgUrl string `json:"smallHeadImgUrl"`
	Sex             int    `json:"sex"`
	Province        string `json:"province"`
	City            string `json:"city"`
	Country         string `json:"country"`
	Signature       string `json:"signature"`
	// 可根据实际返回添加更多字段
}

// FoldingRequest 折叠群聊或个人请求
type FoldingRequest struct {
	RoomID string `json:"roomId"` // 群聊ID或个人微信ID
}

// FoldingResponse 折叠群聊或个人响应
type FoldingResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// UnfoldingRequest 取消折叠群聊或个人请求
type UnfoldingRequest struct {
	RoomID string `json:"roomId"` // 群聊ID或个人微信ID
}

// UnfoldingResponse 取消折叠群聊或个人响应
type UnfoldingResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
