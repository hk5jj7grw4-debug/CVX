package types

// GetFavsResponse 获取收藏列表响应
type GetFavsResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data []FavItem `json:"data"`
}

// FavItem 收藏项
type FavItem struct {
	FavId      int64  `json:"favId"`      // 收藏ID
	Type       int    `json:"type"`       // 收藏类型
	SourceId   string `json:"sourceId"`   // 来源ID
	SourceType int    `json:"sourceType"` // 来源类型
	UpdateTime int64  `json:"updateTime"` // 更新时间
	Object     string `json:"object"`     // 收藏对象内容
	Tags       string `json:"tags"`       // 标签
	Flag       int    `json:"flag"`       // 标志
	LocalId    int    `json:"localId"`    // 本地ID
}
