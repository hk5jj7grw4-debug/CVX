package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVXkernel/GVxSdk.Go/types"
)

// SNSAPI 朋友圈相关接口
type SNSAPI struct {
	client Client
}

// NewSNSAPI 创建朋友圈API实例
func NewSNSAPI(client Client) *SNSAPI {
	return &SNSAPI{client: client}
}

// Post 发送朋友圈
//
// 发布朋友圈动态
//
// 参数:
//   - content: 朋友圈文本内容
//   - blackList: 黑名单列表（不给谁看），多个微信ID用逗号分隔，传空字符串表示所有人可见
//   - withUserList: 白名单列表（仅给谁看），多个微信ID用逗号分隔，传空字符串表示不限制
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	// 发送给所有人
//	err := client.SNS.Post("今天天气真好！", "", "")
//
//	// 不给某些人看
//	err := client.SNS.Post("今天天气真好！", "wxid_xxx,wxid_yyy", "")
//
//	// 仅给某些人看
//	err := client.SNS.Post("今天天气真好！", "", "wxid_aaa,wxid_bbb")
//
// 警告:
//   - 此接口容易导致账号掉线，不建议频繁使用
//   - 建议谨慎使用，避免影响账号稳定性
func (a *SNSAPI) Post(content, blackList, withUserList string) error {
	req := types.SNSPostRequest{
		Content:      content,
		BlackList:    blackList,
		WithUserList: withUserList,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_post", req)
	if err != nil {
		return fmt.Errorf("sns post: %w", err)
	}

	var resp types.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// SendImg 发送图片朋友圈
//
// 发布带图片的朋友圈动态
//
// 参数:
//   - fileList: 图片文件路径列表，多个文件用逗号分隔，例如 "D:\\pic1.jpg,D:\\pic2.jpg"
//   - content: 朋友圈文本内容
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	// 发送单张图片
//	err := client.SNS.SendImg("D:\\photo.jpg", "今天天气真好！")
//
//	// 发送多张图片
//	err := client.SNS.SendImg(
//	    "D:\\pic1.jpg,D:\\pic2.jpg,D:\\pic3.jpg",
//	    "今天拍的照片",
//	)
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
//   - 多个图片路径用逗号分隔
//   - 此接口可能影响账号稳定性，建议谨慎使用
func (a *SNSAPI) SendImg(fileList, content string) error {
	req := types.SNSSendImgRequest{
		FileList: fileList,
		Content:  content,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_send_img", req)
	if err != nil {
		return fmt.Errorf("sns send img: %w", err)
	}

	var resp types.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// DownloadMedia 下载朋友圈图片/视频（测试中）
//
// 下载朋友圈中的图片或视频到本地
//
// 参数:
//   - url: 朋友圈图片或视频的链接
//   - outPath: 保存路径，例如 "D:\\sns.jpg"
//
// 返回:
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	err := client.SNS.DownloadMedia(
//	    "http://szmmsns.qpic.cn/mmsns/xxx/0?idx=1&token=xxx",
//	    "D:\\sns.jpg",
//	)
//
// 注意:
//   - 此接口处于测试阶段，可能不稳定
//   - 保存路径必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
func (a *SNSAPI) DownloadMedia(url, outPath string) error {
	req := types.DownloadSNSMediaRequest{
		URL: url,
		Out: outPath,
	}

	respBody, err := a.client.DoRequest("POST", "/api/download_sns_media", req)
	if err != nil {
		return fmt.Errorf("download sns media: %w", err)
	}

	var resp types.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// DelComment 删除朋友圈评论
//
// 删除指定朋友圈动态下的评论。
//
// 参数:
//   - snsID: 朋友圈ID，例如 "14661929784229180031"
//   - commentID: 评论ID，例如 "3"
//
// 返回:
//   - error: 删除失败时返回错误信息
//
// 示例:
//
//	err := client.SNS.DelComment("14661929784229180031", "3")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("评论删除成功")
//
// 注意:
//   - 只能删除自己朋友圈下的评论
//   - snsID 和 commentID 需要从朋友圈消息或接口中获取
//   - 删除操作不可撤销
//   - 删除后评论将从朋友圈中消失
func (a *SNSAPI) DelComment(snsID, commentID string) error {
	req := types.SNSDelCommentRequest{
		SnsID:     snsID,
		CommentID: commentID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_del_comment", req)
	if err != nil {
		return fmt.Errorf("sns del comment: %w", err)
	}

	var resp types.SNSDelCommentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// CommentReply 朋友圈回复
//
// 回复朋友圈动态下的评论。
//
// 参数:
//   - content: 回复内容，例如 "66666666666"
//   - snsID: 朋友圈ID，例如 "14667428703163265648"
//   - commentID: 评论ID，例如 3
//
// 返回:
//   - error: 回复失败时返回错误信息
//
// 示例:
//
//	err := client.SNS.CommentReply("赞同！", "14667428703163265648", 3)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("回复成功")
//
// 注意:
//   - snsID 和 commentID 需要从朋友圈消息或接口中获取
//   - 回复内容支持文本和emoji表情
//   - 回复会显示在原评论下方
//   - 可以回复自己或他人朋友圈的评论
func (a *SNSAPI) CommentReply(content, snsID string, commentID int) error {
	req := types.SNSCommentReplyRequest{
		Content:   content,
		SnsID:     snsID,
		CommentID: commentID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_comment_reply", req)
	if err != nil {
		return fmt.Errorf("sns comment reply: %w", err)
	}

	var resp types.SNSCommentReplyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// Del 删除朋友圈
//
// 删除指定的朋友圈动态。
//
// 参数:
//   - snsID: 朋友圈ID，例如 "14667428703163265648"
//
// 返回:
//   - error: 删除失败时返回错误信息
//
// 示例:
//
//	err := client.SNS.Del("14667428703163265648")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("朋友圈删除成功")
//
// 注意:
//   - 只能删除自己发布的朋友圈
//   - snsID 需要从朋友圈消息或接口中获取
//   - 删除操作不可撤销
//   - 删除后朋友圈动态及其所有评论、点赞都会消失
func (a *SNSAPI) Del(snsID string) error {
	req := types.SNSDelRequest{
		SnsID: snsID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_del", req)
	if err != nil {
		return fmt.Errorf("sns del: %w", err)
	}

	var resp types.SNSDelResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// GetDetail 获取朋友圈详情
//
// 获取指定朋友圈动态的详细信息。
//
// 参数:
//   - snsID: 朋友圈ID（数字类型），例如 14420282581074719000
//
// 返回:
//   - *types.SNSGetDetailResponse: 朋友圈详情信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.SNS.GetDetail(14420282581074719000)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("朋友圈详情: %+v\n", resp.Data)
//
// 注意:
//   - snsID 必须是数字类型
//   - 返回的 Data 字段包含朋友圈的详细信息
//   - 可以获取朋友圈的内容、图片、评论、点赞等信息
func (a *SNSAPI) GetDetail(snsID int64) (*types.SNSGetDetailResponse, error) {
	req := types.SNSGetDetailRequest{
		SnsID: snsID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_get_detail", req)
	if err != nil {
		return nil, fmt.Errorf("sns get detail: %w", err)
	}

	var resp types.SNSGetDetailResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp, nil
}

// GetFirstPage 获取朋友圈首页
//
// 获取朋友圈首页的动态列表。
//
// 参数:
//   - firstPageMd5: 第一页的MD5值，如果是首次获取可以传空字符串
//   - maxID: 最大ID，用于分页加载，首次获取可以传空字符串
//
// 返回:
//   - *types.SNSGetFirstPageResponse: 朋友圈首页数据
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	// 首次获取朋友圈首页
//	resp, err := client.SNS.GetFirstPage("", "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("朋友圈首页: %+v\n", resp.Data)
//
//	// 分页加载更多（使用上次返回的 maxID）
//	resp, err = client.SNS.GetFirstPage("", "14420282581074719000")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 注意:
//   - 首次获取时两个参数都传空字符串
//   - 分页加载时使用上次返回的 maxID
//   - firstPageMd5 用于缓存优化，可选参数
//   - 返回的 Data 字段包含朋友圈动态列表
func (a *SNSAPI) GetFirstPage(firstPageMd5, maxID string) (*types.SNSGetFirstPageResponse, error) {
	req := types.SNSGetFirstPageRequest{
		FirstPageMd5: firstPageMd5,
		MaxID:        maxID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_get_first_page", req)
	if err != nil {
		return nil, fmt.Errorf("sns get first page: %w", err)
	}

	var resp types.SNSGetFirstPageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp, nil
}

// Upload 朋友圈图片上传
//
// 上传图片到朋友圈服务器，获取的 URL 可以作为图库使用。
//
// 参数:
//   - filePath: 图片文件路径，例如 "D:\\QQ20251211-100920.png"
//
// 返回:
//   - *types.SNSUploadResponse: 上传结果，包含图片 URL 等信息
//   - error: 上传失败时返回错误信息
//
// 示例:
//
//	resp, err := client.SNS.Upload("D:\\QQ20251211-100920.png")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("图片上传成功: %+v\n", resp.Data)
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 上传后的 URL 可以用于发布朋友圈或作为图库使用
//   - 支持常见的图片格式（jpg, png, gif 等）
func (a *SNSAPI) Upload(filePath string) (*types.SNSUploadResponse, error) {
	req := types.SNSUploadRequest{
		FilePath: filePath,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_upload", req)
	if err != nil {
		return nil, fmt.Errorf("sns upload: %w", err)
	}

	var resp types.SNSUploadResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp, nil
}

// GetNextPage 获取朋友圈下一页
//
// 获取朋友圈的下一页动态列表，用于分页加载。
//
// 参数:
//   - lastItemID: 最后一条朋友圈的ID，例如 "14689529228577936097"
//
// 返回:
//   - *types.SNSGetNextPageResponse: 下一页朋友圈数据
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	// 先获取首页
//	firstPage, err := client.SNS.GetFirstPage("", "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取下一页（使用首页最后一条的ID）
//	nextPage, err := client.SNS.GetNextPage("14689529228577936097")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("下一页朋友圈: %+v\n", nextPage.Data)
//
// 注意:
//   - lastItemID 必须是上一页最后一条朋友圈的ID
//   - 通常与 GetFirstPage 配合使用实现分页加载
//   - 返回的 Data 字段包含下一页的朋友圈动态列表
//   - 如果没有更多数据，返回的列表可能为空
func (a *SNSAPI) GetNextPage(lastItemID string) (*types.SNSGetNextPageResponse, error) {
	req := types.SNSGetNextPageRequest{
		LastItemID: lastItemID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sns_get_next_page", req)
	if err != nil {
		return nil, fmt.Errorf("sns get next page: %w", err)
	}

	var resp types.SNSGetNextPageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp, nil
}
