package types

import (
	"encoding/json"
	"testing"
)

func TestGetProfileCacheResponseAvatarFields(t *testing.T) {
	const payload = `{
		"baseResponse":{"ret":0},
		"userInfo":{
			"userName":{"String":"wxid_test"},
			"nickName":{"String":"test user"}
		},
		"userInfoExt":{
			"smallHeadImgUrl":"https://mmhead.c2c.wechat.com/avatar/132",
			"bigHeadImgUrl":"https://mmhead.c2c.wechat.com/avatar/0"
		}
	}`

	var response GetProfileCacheResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("unmarshal profile response: %v", err)
	}

	if response.UserInfo.UserName.String != "wxid_test" {
		t.Fatalf("unexpected wxid: %q", response.UserInfo.UserName.String)
	}
	if response.UserInfoExt.SmallHeadImgUrl != "https://mmhead.c2c.wechat.com/avatar/132" {
		t.Fatalf("unexpected small avatar URL: %q", response.UserInfoExt.SmallHeadImgUrl)
	}
	if response.UserInfoExt.BigHeadImgUrl != "https://mmhead.c2c.wechat.com/avatar/0" {
		t.Fatalf("unexpected big avatar URL: %q", response.UserInfoExt.BigHeadImgUrl)
	}
}
