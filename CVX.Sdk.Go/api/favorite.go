package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/types"
)

// FavoriteAPI 收藏相关接口
type FavoriteAPI struct {
	client Client
}

// NewFavoriteAPI 创建收藏API实例
func NewFavoriteAPI(client Client) *FavoriteAPI {
	return &FavoriteAPI{client: client}
}

// GetFavs 获取收藏列表
//
// 获取当前账号的所有收藏内容列表
//
// 返回:
//   - []types.FavItem: 收藏列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	favs, err := client.Favorite.GetFavs()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个收藏\n", len(favs))
//	for _, fav := range favs {
//	    fmt.Printf("- 收藏ID: %d, 类型: %d\n", fav.FavId, fav.Type)
//	}
//
// 注意:
//   - 返回的列表包含所有收藏的内容
//   - 包括文本、图片、链接、文件等各种类型的收藏
//   - Type 字段表示收藏的类型
func (a *FavoriteAPI) GetFavs() ([]types.FavItem, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_favs", nil)
	if err != nil {
		return nil, fmt.Errorf("get favs: %w", err)
	}

	var resp types.GetFavsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}
