package api

import (
	"encoding/json"
	"fmt"

	"github.com/olaria01/gVxSdk/sdk/go/types"
)

// EmotionAPI 表情相关接口
type EmotionAPI struct {
	client Client
}

// NewEmotionAPI 创建表情API实例
func NewEmotionAPI(client Client) *EmotionAPI {
	return &EmotionAPI{client: client}
}

// SendFavEmotion 发送收藏表情
//
// 发送已收藏的表情到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - md5: 表情的MD5值
//   - length: 表情长度（可选，传0表示不设置）
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Emotion.SendFavEmotion("wxid_xxx", "abc123def456", 1024)
func (a *EmotionAPI) SendFavEmotion(wxid, md5 string, length int) error {
	req := types.SendFavEmotionRequest{
		Wxid:   wxid,
		MD5:    md5,
		Length: length,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_fav_emotion", req)
	if err != nil {
		return fmt.Errorf("send fav emotion: %w", err)
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

// SendEmotionMsg 发送本地GIF表情
//
// 发送本地GIF文件作为表情到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - filePath: 本地GIF文件的完整路径，例如 "D:\\bqb\\3.gif"
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Emotion.SendEmotionMsg("wxid_xxx", "D:\\bqb\\3.gif")
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
func (a *EmotionAPI) SendEmotionMsg(wxid, filePath string) error {
	req := types.SendEmotionMsgRequest{
		Wxid:     wxid,
		FilePath: filePath,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_emotion_msg", req)
	if err != nil {
		return fmt.Errorf("send emotion msg: %w", err)
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
