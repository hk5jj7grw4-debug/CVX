package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/types"
)

// SocialAPI 社交发现相关接口
type SocialAPI struct {
	client Client
}

// NewSocialAPI 创建社交发现API实例
func NewSocialAPI(client Client) *SocialAPI {
	return &SocialAPI{client: client}
}

// GetLbsFriend 获取附近人
//
// 根据经纬度获取附近的微信用户列表。
//
// 参数:
//   - longitude: 经度，例如 "120.24646699999994"
//   - latitude: 纬度，例如 "30.197153999999998"
//
// 返回:
//   - *types.GetLbsFriendResponse: 附近人列表信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Social.GetLbsFriend("120.24646699999994", "30.197153999999998")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("附近人数量: %d\n", resp.ContactCount)
//	for _, contact := range resp.ContactList {
//	    fmt.Printf("- %s (%s)\n", contact.NickName, contact.UserName)
//	    fmt.Printf("  距离: %s\n", contact.Distance)
//	    fmt.Printf("  性别: %d (1-男, 2-女, 0-未知)\n", contact.Sex)
//	    fmt.Printf("  地区: %s %s\n", contact.Province, contact.City)
//	    fmt.Printf("  个性签名: %s\n", contact.Signature)
//	    fmt.Printf("  头像: %s\n", contact.SmallHeadImgUrl)
//	}
//
// 注意:
//   - 需要提供准确的经纬度坐标
//   - 返回的 UserName 可能是加密后的唯一标识
//   - Sex 字段：1-男性，2-女性，0-未知
//   - Distance 字段表示与自己的物理距离
//   - AntispamTicket 用于后续添加好友操作
//   - 需要微信账号开启了"附近的人"功能
func (a *SocialAPI) GetLbsFriend(longitude, latitude string) (*types.GetLbsFriendResponse, error) {
	req := types.GetLbsFriendRequest{
		Longitude: longitude,
		Latitude:  latitude,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_lbs_friend", req)
	if err != nil {
		return nil, fmt.Errorf("get lbs friend: %w", err)
	}

	var resp types.GetLbsFriendResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}
