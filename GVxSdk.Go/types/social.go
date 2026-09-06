package types

// GetLbsFriendRequest 获取附近人请求
type GetLbsFriendRequest struct {
	Longitude string `json:"longitude"` // 经度
	Latitude  string `json:"latitude"`  // 纬度
}

// GetLbsFriendResponse 获取附近人响应
type GetLbsFriendResponse struct {
	BaseResponse    BaseResponse `json:"baseResponse"`
	ContactCount    int          `json:"contactCount"`
	ContactList     []LbsContact `json:"contactList"`
	State           int          `json:"state"`
	FlushTime       int          `json:"flushTime"`
	IsShowRoom      int          `json:"isShowRoom"`
	RoomMemberCount int          `json:"roomMemberCount"`
}

// LbsContact 附近人联系人信息
type LbsContact struct {
	UserName        string            `json:"userName"`
	NickName        string            `json:"nickName"`
	Province        string            `json:"province"`
	City            string            `json:"city"`
	Signature       string            `json:"signature"`
	Distance        string            `json:"distance"`
	Sex             int               `json:"sex"`
	ImgStatus       int               `json:"imgStatus"`
	VerifyFlag      int               `json:"verifyFlag"`
	WeiboFlag       int               `json:"weiboFlag"`
	HeadImgVersion  int               `json:"headImgVersion"`
	SnsUserInfo     LbsSnsUserInfo    `json:"snsUserInfo"`
	Country         string            `json:"country"`
	BigHeadImgUrl   string            `json:"bigHeadImgUrl"`
	SmallHeadImgUrl string            `json:"smallHeadImgUrl"`
	CustomizedInfo  LbsCustomizedInfo `json:"customizedInfo"`
	AntispamTicket  string            `json:"antispamTicket"`
	Flag            int               `json:"flag"`
	FinderFlag      int               `json:"finderFlag"`
}

// LbsSnsUserInfo 附近人SNS用户信息
type LbsSnsUserInfo struct {
	SnsFlag          int    `json:"snsFlag"`
	SnsBgimgId       string `json:"snsBgimgId"`
	SnsBgobjectId    string `json:"snsBgobjectId"`
	SnsFlagEx        int    `json:"snsFlagEx"`
	SnsPrivacyRecent int    `json:"snsPrivacyRecent"`
}

// LbsCustomizedInfo 附近人自定义信息
type LbsCustomizedInfo struct {
	BrandFlag int `json:"brandFlag"`
}
