package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/types"
)

// VideoAPI 视频相关接口
type VideoAPI struct {
	client Client
}

// NewVideoAPI 创建视频API实例
func NewVideoAPI(client Client) *VideoAPI {
	return &VideoAPI{client: client}
}

// CDNVideoForward CDN转发视频
func (a *VideoAPI) CDNVideoForward(wxid, cdnVideoURL, aesKey string, videoLength, thumbLength, playLength int) error {
	req := types.CDNVideoForwardRequest{
		Wxid:        wxid,
		CDNVideoURL: cdnVideoURL,
		AESKey:      aesKey,
		VideoLength: videoLength,
		ThumbLength: thumbLength,
		PlayLength:  playLength,
	}

	respBody, err := a.client.DoRequest("POST", "/api/cdn_video_forward", req)
	if err != nil {
		return fmt.Errorf("cdn video forward: %w", err)
	}

	var resp types.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// DownloadVideo 下载视频
//
// 下载视频消息中的视频文件到本地。
// 该接口的所有数据均来自消息回调。
//
// 参数:
//   - totalLen: 视频总长度（字节），从消息回调中获取
//   - newMsgId: 新消息ID，从消息回调中获取
//   - path: 保存路径，例如 "d:\\121.mp4"
//   - msgId: 消息ID，从消息回调中获取
//
// 返回:
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	// 从消息回调中获取视频信息
//	totalLen := 68264760
//	newMsgId := int64(150973212120165920)
//	msgId := 425004325
//	savePath := "d:\\121.mp4"
//
//	err := client.Video.DownloadVideo(totalLen, newMsgId, savePath, msgId)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("视频下载成功")
//
// 注意:
//   - 所有参数必须从消息回调中获取，不能随意填写
//   - path 参数必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 下载是异步操作，接口返回成功表示开始下载
//   - 实际下载完成时间取决于视频大小和网络速度
func (a *VideoAPI) DownloadVideo(totalLen int, newMsgId int64, path string, msgId int) error {
	req := types.DownloadVideoRequest{
		TotalLen: totalLen,
		NewMsgId: newMsgId,
		Path:     path,
		MsgId:    msgId,
	}

	respBody, err := a.client.DoRequest("POST", "/api/download_video", req)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
	}

	var resp types.DownloadVideoResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}
