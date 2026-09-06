package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVXkernel/GVxSdk.Go/types"
)

// VoiceAPI 语音相关接口
type VoiceAPI struct {
	client Client
}

// NewVoiceAPI 创建语音API实例
func NewVoiceAPI(client Client) *VoiceAPI {
	return &VoiceAPI{client: client}
}

// SendMP3Voice 发送MP3语音消息
//
// 发送本地MP3文件作为语音消息到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - mp3Path: 本地MP3文件的完整路径，例如 "C:\\audio\\voice.mp3"
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Voice.SendMP3Voice("wxid_xxx", "C:\\audio\\voice.mp3")
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
//   - 文件必须是MP3格式
func (a *VoiceAPI) SendMP3Voice(wxid, mp3Path string) error {
	req := types.SendMP3VoiceRequest{
		Wxid:    wxid,
		MP3Path: mp3Path,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_mp3_voice", req)
	if err != nil {
		return fmt.Errorf("send mp3 voice: %w", err)
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

// GetVoiceTrans 语音转文本
//
// 将语音消息转换为文本
//
// 参数:
//   - clientMsgID: 客户端消息ID（从回调消息中获取）
//   - newMsgID: 新消息ID（从回调消息中获取）
//   - length: 语音长度（从回调消息中获取）
//
// 返回:
//   - string: 转换后的文本内容
//   - error: 转换失败时返回错误信息
//
// 示例:
//
//	// 从回调消息中获取语音信息
//	text, err := client.Voice.GetVoiceTrans(
//	    "8111825985001399988",
//	    "211095990",
//	    "11832",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("语音内容: %s\n", text)
//
// 注意:
//   - 需要从接收到的语音消息回调中获取相关参数
//   - 转换可能需要一定时间
func (a *VoiceAPI) GetVoiceTrans(clientMsgID, newMsgID, length string) (string, error) {
	req := types.GetVoiceTransRequest{
		ClientMsgID: clientMsgID,
		NewMsgID:    newMsgID,
		Length:      length,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_voice_trans", req)
	if err != nil {
		return "", fmt.Errorf("get voice trans: %w", err)
	}

	var resp types.GetVoiceTransResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return "", fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// DownloadVoice 下载语音
//
// 下载语音消息到本地文件
//
// 参数:
//   - newMsgID: 新消息ID（从回调消息中获取）
//   - length: 语音长度（从回调消息中获取）
//   - msgID: 消息ID（从回调消息中获取）
//   - path: 保存路径，例如 "d:\\7054.slik"
//
// 返回:
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	// 从回调消息中获取语音信息
//	err := client.Voice.DownloadVoice(
//	    "3823231950088215754",
//	    "70541",
//	    "1669230954",
//	    "d:\\voice\\7054.slik",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("语音下载成功")
//
// 注意:
//   - 需要从接收到的语音消息回调中获取相关参数
//   - path 必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
//   - 保存路径的目录必须存在
func (a *VoiceAPI) DownloadVoice(newMsgID, length, msgID, path string) error {
	req := types.DownloadVoiceRequest{
		NewMsgID: newMsgID,
		Length:   length,
		MsgID:    msgID,
		Path:     path,
	}

	respBody, err := a.client.DoRequest("POST", "/api/download_voice", req)
	if err != nil {
		return fmt.Errorf("download voice: %w", err)
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
