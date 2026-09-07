package types

// GetRoomMembersRequest 获取群成员列表请求
type GetRoomMembersRequest struct {
	RoomID string `json:"room_id"` // 群聊ID，例如 "39259098574@chatroom"
}

// GetChatroomDetailCacheRequest 获取群详情缓存请求
type GetChatroomDetailCacheRequest struct {
	RoomID string `json:"room_id"` // 群聊ID，例如 "18402658081@chatroom"
}

// GetGroupMemberBySQLRequest 获取群成员数据(简要)请求
type GetGroupMemberBySQLRequest struct {
	RoomID string `json:"roomId"` // 群聊ID，例如 "18402658081@chatroom"
}

// GetRoomMembersResponse 获取群成员列表响应
type GetRoomMembersResponse struct {
	BaseResponse          BaseResponse     `json:"baseResponse"`
	ChatroomUserName      string           `json:"chatroomUserName"`
	ServerVersion         int              `json:"serverVersion"`
	NewChatroomData       NewChatroomData  `json:"newChatroomData"`
	ChatRoomOwner         string           `json:"chatRoomOwner"`
	AllMemberCount        int              `json:"allMemberCount"`
	AllMemberUserNameList []MemberUserName `json:"allMemberUserNameList"`
	AdminCount            int              `json:"adminCount"`
}

// BaseResponse 基础响应
type BaseResponse struct {
	Ret    int         `json:"ret"`
	ErrMsg interface{} `json:"errMsg"`
}

// NewChatroomData 新群聊数据
type NewChatroomData struct {
	MemberCount      int              `json:"memberCount"`
	ChatRoomMember   []ChatRoomMember `json:"chatRoomMember"`
	InfoMask         int              `json:"infoMask"`
	ChatRoomUserName interface{}      `json:"chatRoomUserName"`
	WatchMemberCount int              `json:"watchMemberCount"`
}

// ChatRoomMember 群成员信息
type ChatRoomMember struct {
	UserName               string `json:"userName"`                         // 用户微信ID
	NickName               string `json:"nickName"`                         // 昵称
	BigHeadImgUrl          string `json:"bigHeadImgUrl"`                    // 大头像URL
	SmallHeadImgUrl        string `json:"smallHeadImgUrl"`                  // 小头像URL
	ChatroomMemberFlag     int    `json:"chatroomMemberFlag"`               // 群成员标志
	Status                 int    `json:"status"`                           // 状态
	InviterUserName        string `json:"inviterUserName,omitempty"`        // 邀请人微信ID（可选）
	AddChatRoomSceneNewXml string `json:"addChatRoomSceneNewXml,omitempty"` // 加群场景XML（可选）
}

// MemberUserName 成员用户名
type MemberUserName struct {
	String string `json:"String"`
}

// GetChatroomDetailCacheResponse 获取群详情缓存响应
type GetChatroomDetailCacheResponse struct {
	AccountWxid string             `json:"account_wxid"`
	Data        ChatroomDetailData `json:"data"`
	ErrCode     int                `json:"errCode"`
	ErrMsg      string             `json:"errMsg"`
}

// ChatroomDetailData 群详情数据
type ChatroomDetailData struct {
	AdminCount            int                   `json:"adminCount"`
	AdminUserNameList     []MemberUserName      `json:"adminUserNameList"`
	AllMemberCount        int                   `json:"allMemberCount"`
	AllMemberUserNameList []MemberUserName      `json:"allMemberUserNameList"`
	BaseResponse          BaseResponse          `json:"baseResponse"`
	ChatRoomOwner         string                `json:"chatRoomOwner"`
	ChatroomUserName      string                `json:"chatroomUserName"`
	NewChatroomData       ChatroomDetailNewData `json:"newChatroomData"`
	ServerVersion         int                   `json:"serverVersion"`
}

// ChatroomDetailNewData 群详情新数据
type ChatroomDetailNewData struct {
	ChatRoomMember   []ChatroomDetailMember `json:"chatRoomMember"`
	ChatRoomUserName interface{}            `json:"chatRoomUserName"`
	InfoMask         int                    `json:"infoMask"`
	MemberCount      int                    `json:"memberCount"`
	WatchMemberCount int                    `json:"watchMemberCount"`
}

// ChatroomDetailMember 群详情成员信息
type ChatroomDetailMember struct {
	BigHeadImgUrl          string `json:"bigHeadImgUrl"`
	ChatroomMemberFlag     int    `json:"chatroomMemberFlag"`
	DisplayName            string `json:"displayName"`
	NickName               string `json:"nickName"`
	SmallHeadImgUrl        string `json:"smallHeadImgUrl"`
	Status                 int    `json:"status"`
	UserName               string `json:"userName"`
	AddChatRoomSceneNewXml string `json:"addChatRoomSceneNewXml"`
	InviterUserName        string `json:"inviterUserName"`
}

// GetGroupMemberBySQLResponse 获取群成员数据(简要)响应
type GetGroupMemberBySQLResponse struct {
	Members []SimpleMember `json:"members"`
}

// SimpleMember 简要成员信息（不包含头像）
type SimpleMember struct {
	UserName        string `json:"userName"`        // 用户微信ID
	DisplayName     string `json:"displayName"`     // 显示名称（群昵称）
	MemberFlag      int    `json:"memberFlag"`      // 成员标志
	InviterUserName string `json:"inviterUserName"` // 邀请人微信ID
}

// GetChatroomListResponse 获取群聊列表响应
type GetChatroomListResponse struct {
	Code  int            `json:"code"`
	Data  []ChatroomItem `json:"data"`
	Msg   string         `json:"msg"`
	Total int            `json:"total"`
}

// ChatroomItem 群聊项
type ChatroomItem struct {
	BigHeadUrl   string `json:"big_head_url"`
	NickName     string `json:"nick_name"`
	Remark       string `json:"remark"`
	SmallHeadUrl string `json:"small_head_url"`
	Username     string `json:"username"`
}

// GetGroupMemberContactRequest 查询群成员信息请求
type GetGroupMemberContactRequest struct {
	Wxid   string `json:"wxid"`   // 群成员微信ID
	RoomID string `json:"roomId"` // 群聊ID
}

// GetGroupMemberContactResponse 查询群成员信息响应
type GetGroupMemberContactResponse struct {
	BaseResponse              BaseResponse              `json:"baseResponse"`
	ContactCount              int                       `json:"contactCount"`
	ContactList               []GroupMemberContactInfo  `json:"contactList"`
	Ret                       []int                     `json:"ret"`
	VerifyUserValidTicketList VerifyUserValidTicketList `json:"verifyUserValidTicketList"`
}

// GroupMemberContactInfo 群成员联系人信息
type GroupMemberContactInfo struct {
	UserName                StringValue                 `json:"userName"`
	NickName                StringValue                 `json:"nickName"`
	Pyinitial               interface{}                 `json:"pyinitial"`
	QuanPin                 interface{}                 `json:"quanPin"`
	Sex                     int                         `json:"sex"`
	ImgBuf                  MemberBufferData            `json:"imgBuf"`
	BitMask                 int64                       `json:"bitMask"`
	BitVal                  int64                       `json:"bitVal"`
	ImgFlag                 int                         `json:"imgFlag"`
	Remark                  interface{}                 `json:"remark"`
	RemarkPyinitial         interface{}                 `json:"remarkPyinitial"`
	RemarkQuanPin           interface{}                 `json:"remarkQuanPin"`
	ContactType             int                         `json:"contactType"`
	RoomInfoCount           int                         `json:"roomInfoCount"`
	DomainList              interface{}                 `json:"domainList"`
	ChatRoomNotify          int                         `json:"chatRoomNotify"`
	AddContactScene         int                         `json:"addContactScene"`
	Province                string                      `json:"province"`
	City                    string                      `json:"city"`
	Signature               string                      `json:"signature"`
	PersonalCard            int                         `json:"personalCard"`
	HasWeiXinHdHeadImg      int                         `json:"hasWeiXinHdHeadImg"`
	VerifyFlag              int                         `json:"verifyFlag"`
	Level                   int                         `json:"level"`
	Source                  int                         `json:"source"`
	Alias                   string                      `json:"alias"`
	WeiboFlag               int                         `json:"weiboFlag"`
	AlbumStyle              int                         `json:"albumStyle"`
	AlbumFlag               int                         `json:"albumFlag"`
	SnsUserInfo             SnsUserInfoSimple           `json:"snsUserInfo"`
	Country                 string                      `json:"country"`
	BigHeadImgUrl           string                      `json:"bigHeadImgUrl"`
	SmallHeadImgUrl         string                      `json:"smallHeadImgUrl"`
	CustomizedInfo          CustomizedInfoSimple        `json:"customizedInfo"`
	HeadImgMd5              string                      `json:"headImgMd5"`
	EncryptUserName         string                      `json:"encryptUserName"`
	AdditionalContactList   AdditionalContactListSimple `json:"additionalContactList"`
	ChatroomVersion         int                         `json:"chatroomVersion"`
	ChatroomMaxCount        int                         `json:"chatroomMaxCount"`
	ChatroomAccessType      int                         `json:"chatroomAccessType"`
	NewChatroomData         NewChatroomDataWithUserName `json:"newChatroomData"`
	DeleteFlag              int                         `json:"deleteFlag"`
	PhoneNumListInfo        PhoneNumListInfoSimple      `json:"phoneNumListInfo"`
	ChatroomInfoVersion     int                         `json:"chatroomInfoVersion"`
	DeleteContactScene      int                         `json:"deleteContactScene"`
	ChatroomStatus          int                         `json:"chatroomStatus"`
	ExtFlag                 int                         `json:"extFlag"`
	ChatRoomBusinessType    string                      `json:"chatRoomBusinessType"`
	TextStatusFlag          int                         `json:"textStatusFlag"`
	RingBackSetting         RingBackSettingSimple       `json:"ringBackSetting"`
	BitMask2                string                      `json:"bitMask2"`
	BitValue2               string                      `json:"bitValue2"`
	ContactExtraInfoBuf     MemberBufferData            `json:"contactExtraInfoBuf"`
	IsInChatRoom            int                         `json:"isInChatRoom"`
	EraseChatRoomMemberData int                         `json:"eraseChatRoomMemberData"`
}

// MemberBufferData 成员缓冲区数据
type MemberBufferData struct {
	ILen int `json:"iLen"`
}

// SnsUserInfoSimple 简单SNS用户信息
type SnsUserInfoSimple struct {
	SnsFlag          int    `json:"snsFlag"`
	SnsBgimgId       string `json:"snsBgimgId"`
	SnsBgobjectId    string `json:"snsBgobjectId"`
	SnsFlagEx        int    `json:"snsFlagEx"`
	SnsPrivacyRecent int    `json:"snsPrivacyRecent"`
}

// CustomizedInfoSimple 简单自定义信息
type CustomizedInfoSimple struct {
	BrandFlag int `json:"brandFlag"`
}

// AdditionalContactListSimple 简单附加联系人列表
type AdditionalContactListSimple struct {
	LinkedinContactItem interface{} `json:"linkedinContactItem"`
}

// NewChatroomDataWithUserName 带用户名的新群聊数据
type NewChatroomDataWithUserName struct {
	MemberCount      int         `json:"memberCount"`
	InfoMask         int         `json:"infoMask"`
	ChatRoomUserName StringValue `json:"chatRoomUserName"`
	WatchMemberCount int         `json:"watchMemberCount"`
}

// PhoneNumListInfoSimple 简单手机号列表信息
type PhoneNumListInfoSimple struct {
	Count int `json:"count"`
}

// RingBackSettingSimple 简单回铃设置
type RingBackSettingSimple struct {
	FinderObjectId string `json:"finderObjectId"`
	StartTs        int    `json:"startTs"`
	EndTs          int    `json:"endTs"`
}

// CreateChatRoomRequest 创建群聊请求
type CreateChatRoomRequest struct {
	Wxids string `json:"wxids"` // 成员微信ID列表，用逗号分隔
}

// CreateChatRoomResponse 创建群聊响应
type CreateChatRoomResponse struct {
	BaseResponse BaseResponse         `json:"baseResponse"`
	Topic        interface{}          `json:"topic"`
	Pyinitial    interface{}          `json:"pyinitial"`
	QuanPin      interface{}          `json:"quanPin"`
	MemberCount  int                  `json:"memberCount"`
	MemberList   []ChatRoomMemberInfo `json:"memberList,omitempty"`
	ChatRoomName StringValue          `json:"chatRoomName"`
	ImgBuf       ImgBuffer            `json:"imgBuf"`
}

// ChatRoomMemberInfo 群聊成员信息
type ChatRoomMemberInfo struct {
	MemberName      StringValue `json:"memberName"`
	MemberStatus    int         `json:"memberStatus"`
	NickName        StringValue `json:"nickName"`
	Pyinitial       StringValue `json:"pyinitial"`
	QuanPin         StringValue `json:"quanPin"`
	Sex             int         `json:"sex"`
	Remark          interface{} `json:"remark"`
	RemarkPyinitial interface{} `json:"remarkPyinitial"`
	RemarkQuanPin   interface{} `json:"remarkQuanPin"`
	ContactType     int         `json:"contactType"`
	Province        string      `json:"province,omitempty"`
	City            string      `json:"city,omitempty"`
	Signature       string      `json:"signature"`
	PersonalCard    int         `json:"personalCard"`
	VerifyFlag      int         `json:"verifyFlag"`
	Country         string      `json:"country,omitempty"`
}

// ImgBuffer 图片缓冲区
type ImgBuffer struct {
	ILen   int    `json:"iLen"`
	Buffer string `json:"buffer,omitempty"`
}

// InviteMemberToChatRoomRequest 邀请进入群聊请求
type InviteMemberToChatRoomRequest struct {
	WxidList string `json:"wxid_list"` // 成员微信ID列表，用逗号分隔
	RoomID   string `json:"room_id"`   // 群聊ID
}

// InviteMemberToChatRoomResponse 邀请进入群聊响应
type InviteMemberToChatRoomResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// AddMemberToChatRoomRequest 添加群成员请求（40人以内）
type AddMemberToChatRoomRequest struct {
	WxidList string `json:"wxid_list"` // 成员微信ID列表，用逗号分隔
	RoomID   string `json:"room_id"`   // 群聊ID
}

// AddMemberToChatRoomResponse 添加群成员响应
type AddMemberToChatRoomResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// SetRoomAdminRequest 添加群管理请求
type SetRoomAdminRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
	Admin  string `json:"admin"`  // 管理员微信ID
}

// SetRoomAdminResponse 添加群管理响应
type SetRoomAdminResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// DelRoomAdminRequest 删除群管理请求
type DelRoomAdminRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
	Admin  string `json:"admin"`  // 管理员微信ID
}

// DelRoomAdminResponse 删除群管理响应
type DelRoomAdminResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// SetRoomAnnouncementRequest 设置群公告请求
type SetRoomAnnouncementRequest struct {
	RoomID       string `json:"roomId"`       // 群聊ID
	Announcement string `json:"announcement"` // 群公告内容
}

// SetRoomAnnouncementResponse 设置群公告响应
type SetRoomAnnouncementResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// DelMemberFromChatRoomRequest 踢出群成员请求
type DelMemberFromChatRoomRequest struct {
	WxidList string `json:"wxid_list"` // 成员微信ID列表，用逗号分隔
	RoomID   string `json:"room_id"`   // 群聊ID
}

// DelMemberFromChatRoomResponse 踢出群成员响应
type DelMemberFromChatRoomResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
}

// QuitAndDelChatRoomRequest 退出群聊请求
type QuitAndDelChatRoomRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
}

// QuitAndDelChatRoomResponse 退出群聊响应
type QuitAndDelChatRoomResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// EnterRoomRequest 同意群聊邀请请求
type EnterRoomRequest struct {
	URL string `json:"url"` // 群聊邀请链接（需要用 a8key 转换后的 URL）
}

// EnterRoomResponse 同意群聊邀请响应
type EnterRoomResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// GetChatroomInfoRequest 获取群详情请求
type GetChatroomInfoRequest struct {
	RoomID string `json:"roomId"` // 群聊ID，例如 "45220347292@chatroom"
}

// GetChatroomInfoResponse 获取群详情响应
type GetChatroomInfoResponse struct {
	AccountWxid string           `json:"account_wxid,omitempty"`
	Data        ChatroomInfoData `json:"data,omitempty"`
	ErrCode     int              `json:"errCode,omitempty"`
	ErrMsg      string           `json:"errMsg,omitempty"`
	// 异常情况下的字段
	BaseResponse            BaseResponse `json:"baseResponse,omitempty"`
	ChatRoomInfoVersion     int          `json:"chatRoomInfoVersion,omitempty"`
	AnnouncementPublishTime int          `json:"announcementPublishTime,omitempty"`
	ChatRoomStatus          int          `json:"chatRoomStatus,omitempty"`
	ChatRoomBusinessType    string       `json:"chatRoomBusinessType,omitempty"`
	RoomTools               RoomTools    `json:"roomTools,omitempty"`
	RoomBindAppList         BindAppList  `json:"roomBindAppList,omitempty"`
	SpamStatus              int          `json:"spamStatus,omitempty"`
	FinderInfo              BufferInfo   `json:"finderInfo,omitempty"`
	TopMsgInfo              BufferInfo   `json:"topMsgInfo,omitempty"`
}

// ChatroomInfoData 群聊详情数据
type ChatroomInfoData struct {
	Announcement            string       `json:"announcement"`
	AnnouncementEditor      string       `json:"announcementEditor"`
	AnnouncementPublishTime int          `json:"announcementPublishTime"`
	BaseResponse            BaseResponse `json:"baseResponse"`
	ChatRoomBusinessType    string       `json:"chatRoomBusinessType"`
	ChatRoomInfoVersion     int          `json:"chatRoomInfoVersion"`
	ChatRoomStatus          int          `json:"chatRoomStatus"`
	FinderInfo              BufferInfo   `json:"finderInfo"`
	RoomBindAppList         BindAppList  `json:"roomBindAppList"`
	RoomTools               RoomTools    `json:"roomTools"`
	SpamStatus              int          `json:"spamStatus"`
	TopMsgInfo              BufferInfo   `json:"topMsgInfo"`
	XmlAnnouncement         string       `json:"xmlAnnouncement"`
}

// RoomTools 群工具
type RoomTools struct {
	RoomToolsWxAppCount int `json:"roomToolsWxAppCount"`
}

// BindAppList 绑定应用列表
type BindAppList struct {
	RoomBindAppListCount int `json:"roomBindAppListCount"`
}

// BufferInfo 缓冲区信息
type BufferInfo struct {
	ILen int `json:"iLen"`
}

// RemovChatroomToContactRequest 移除群聊通讯录请求
type RemovChatroomToContactRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
}

// RemovChatroomToContactResponse 移除群聊通讯录响应
type RemovChatroomToContactResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SaveChatroomToContactRequest 保存群聊到通讯录请求
type SaveChatroomToContactRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
}

// SaveChatroomToContactResponse 保存群聊到通讯录响应
type SaveChatroomToContactResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// GetGroupMemberInfoRequest 获取群成员数据请求
type GetGroupMemberInfoRequest struct {
	RoomID   string `json:"roomId"`    // 群聊ID
	MemberID string `json:"memeberId"` // 成员微信ID（注意：字段名拼写为 memeberId）
}

// GetGroupMemberInfoResponse 获取群成员数据响应
type GetGroupMemberInfoResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // 成员数据
}

// ModChatroomTopicRequest 修改群名称请求
type ModChatroomTopicRequest struct {
	Wxid  string `json:"wxid"`  // 群聊ID
	Topic string `json:"topic"` // 新的群名称
}

// BatchGetRoomCacheResponse 获取所有群资料缓存响应
type BatchGetRoomCacheResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"` // 使用 interface{} 避免复杂的类型定义
}

// GetMemberNickRequest 获取群成员简要信息请求
type GetMemberNickRequest struct {
	Wxid   string `json:"wxid"`   // 群成员微信ID
	RoomID string `json:"roomId"` // 群聊ID
}

// GetMemberNickResponse 获取群成员简要信息响应
type GetMemberNickResponse struct {
	AccountWxid string         `json:"account_wxid"`
	Data        MemberNickData `json:"data"`
	ErrCode     int            `json:"errCode"`
	ErrMsg      string         `json:"errMsg"`
}

// MemberNickData 群成员简要信息数据
type MemberNickData struct {
	AddChatRoomSceneNewXml string `json:"addChatRoomSceneNewXml"` // 加群场景XML
	BigHeadImgUrl          string `json:"bigHeadImgUrl"`          // 大头像URL
	ChatroomMemberFlag     int    `json:"chatroomMemberFlag"`     // 群成员标志
	InviterUserName        string `json:"inviterUserName"`        // 邀请人微信ID
	NickName               string `json:"nickName"`               // 昵称
	SmallHeadImgUrl        string `json:"smallHeadImgUrl"`        // 小头像URL
	Status                 int    `json:"status"`                 // 状态
	UserName               string `json:"userName"`               // 微信ID
}

// TransferChatroomOwnerRequest 转让群主请求
type TransferChatroomOwnerRequest struct {
	RoomID string `json:"roomId"`  // 群聊ID
	ToWxid string `json:"to_wxid"` // 新群主微信ID
}

// ModChatRoomSelfNickNameRequest 修改自己在群里的昵称请求
type ModChatRoomSelfNickNameRequest struct {
	RoomID   string `json:"roomId"`   // 群聊ID
	NickName string `json:"nickName"` // 新的群昵称
}

// ModChatRoomSelfNickNameResponse 修改自己在群里的昵称响应
type ModChatRoomSelfNickNameResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// GetRoomWxidsResponse 获取所有群wxid响应
type GetRoomWxidsResponse struct {
	Code  int      `json:"code"`
	Msg   string   `json:"msg"`
	Wxids []string `json:"wxids,omitempty"` // 群wxid列表
	Count int      `json:"count,omitempty"` // 群数量
}

// GetRoomsInfoResponse 获取群信息响应
type GetRoomsInfoResponse struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg"`
	Rooms []RoomInfo `json:"rooms,omitempty"`
}

// RoomInfo 群信息
type RoomInfo struct {
	BigHeadURL   string `json:"big_head_url"`   // 大头像URL
	MembersCount int    `json:"members_count"`  // 成员数量
	NickName     string `json:"nick_name"`      // 群昵称
	Owner        string `json:"owner"`          // 群主微信ID
	Remark       string `json:"remark"`         // 备注
	RoomWxid     string `json:"room_wxid"`      // 群聊微信ID
	SmallHeadURL string `json:"small_head_url"` // 小头像URL
	Username     string `json:"username"`       // 用户名
}

// GetChatroomAnnouncementRequest 获取群公告请求
type GetChatroomAnnouncementRequest struct {
	RoomID string `json:"roomId"` // 群聊ID
}

// GetChatroomAnnouncementResponse 获取群公告响应
type GetChatroomAnnouncementResponse struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	Announcement string `json:"announcement,omitempty"` // 群公告内容
	Editor       string `json:"editor,omitempty"`       // 公告编辑者
	PublishTime  int64  `json:"publishTime,omitempty"`  // 发布时间戳
}

// GetAllRoomDetailResponse 获取所有群详情响应
type GetAllRoomDetailResponse struct {
	Code  int              `json:"code"`
	Msg   string           `json:"msg"`
	Rooms []RoomDetailInfo `json:"rooms,omitempty"`
}

// RoomDetailInfo 群详细信息
type RoomDetailInfo struct {
	RoomID          string             `json:"roomId"`          // 群聊ID
	NickName        string             `json:"nickName"`        // 群昵称
	BigHeadImgUrl   string             `json:"bigHeadImgUrl"`   // 大头像URL
	SmallHeadImgUrl string             `json:"smallHeadImgUrl"` // 小头像URL
	MemberCount     int                `json:"memberCount"`     // 成员数量
	Members         []RoomMemberSimple `json:"members"`         // 成员列表
	Owner           string             `json:"owner"`           // 群主微信ID
	Announcement    string             `json:"announcement"`    // 群公告
}

// RoomMemberSimple 群成员简要信息
type RoomMemberSimple struct {
	Wxid        string `json:"wxid"`        // 微信ID
	NickName    string `json:"nickName"`    // 昵称
	DisplayName string `json:"displayName"` // 群昵称
}
