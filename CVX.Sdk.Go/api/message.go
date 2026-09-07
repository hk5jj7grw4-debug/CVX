package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/types"
)

// MessageAPI 消息相关接口
type MessageAPI struct {
	client Client
}

// NewMessageAPI 创建消息API实例
func NewMessageAPI(client Client) *MessageAPI {
	return &MessageAPI{client: client}
}

// SendAppMsg 发送卡片/XML消息
//
// 该接口可以发送所有类型的卡片消息，包括但不限于：
//   - 小程序卡片
//   - 位置卡片
//   - 音乐卡片
//   - 聊天记录卡片
//   - 其他XML格式的卡片消息
//
// 参数:
//   - wxid: 接收人微信ID，例如 "wxid_xxx" 或 "filehelper"（文件传输助手）
//   - content: XML格式的消息内容，需要符合微信的appmsg格式
//   - msgType: 消息类型，例如 "19" 表示聊天记录卡片
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	xmlContent := `<appmsg appid="" sdkver="">
//	  <title>标题</title>
//	  <des>描述</des>
//	  <type>19</type>
//	  ...
//	</appmsg>`
//	err := client.Message.SendAppMsg("filehelper", xmlContent, "19")
func (a *MessageAPI) SendAppMsg(wxid, content, msgType string) error {
	req := types.SendAppMsgRequest{
		Wxid:    wxid,
		Content: content,
		Type:    msgType,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_app_msg", req)
	if err != nil {
		return fmt.Errorf("send app msg: %w", err)
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

// SendXML 发送链接消息
//
// 发送一个带有标题、描述、缩略图和链接的消息卡片
//
// 参数:
//   - wxid: 接收人微信ID
//   - title: 链接标题
//   - description: 链接描述
//   - thumbURL: 缩略图URL
//   - url: 链接URL
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Message.SendXML(
//	    "wxid_xxx",
//	    "文章标题",
//	    "这是一篇很有趣的文章",
//	    "https://example.com/thumb.jpg",
//	    "https://example.com/article",
//	)
func (a *MessageAPI) SendXML(wxid, title, description, thumbURL, url string) error {
	req := types.SendXMLRequest{
		Wxid:        wxid,
		Title:       title,
		Description: description,
		ThumbURL:    thumbURL,
		URL:         url,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_xml", req)
	if err != nil {
		return fmt.Errorf("send xml: %w", err)
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

// SendTextMsg 发送文本消息
//
// 发送纯文本消息到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID，例如 "wxid_xxx" 或 "filehelper"（文件传输助手）
//   - msg: 文本消息内容
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Message.SendTextMsg("filehelper", "Hello, World!")
func (a *MessageAPI) SendTextMsg(wxid, msg string) error {
	req := types.SendTextMsgRequest{
		Wxid: wxid,
		Msg:  msg,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_text_msg", req)
	if err != nil {
		return fmt.Errorf("send text msg: %w", err)
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

// SendAtText 发送AT消息
//
// 在群聊中发送@某人的消息
//
// 参数:
//   - roomID: 群聊ID，格式如 "39259098574@chatroom"
//   - msg: 消息内容，需要包含@的昵称，例如 "@好名字 在干嘛呢"
//   - wxids: 被@的微信ID，多个用逗号分隔，例如 "wxid_xxx" 或 "wxid_xxx,wxid_yyy"
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	// @单个人
//	err := client.Message.SendAtText(
//	    "39259098574@chatroom",
//	    "@张三 在干嘛呢",
//	    "wxid_xxx",
//	)
//
//	// @多个人
//	err := client.Message.SendAtText(
//	    "39259098574@chatroom",
//	    "@张三 @李四 大家好",
//	    "wxid_xxx,wxid_yyy",
//	)
//
// 注意:
//   - 消息内容中的@昵称必须与实际群成员昵称一致
//   - wxids 参数中的微信ID顺序应与消息中@的顺序对应
func (a *MessageAPI) SendAtText(roomID, msg, wxids string) error {
	req := types.SendAtTextRequest{
		RoomID: roomID,
		Msg:    msg,
		Wxids:  wxids,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_at_text", req)
	if err != nil {
		return fmt.Errorf("send at text: %w", err)
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

// SendCardMsg 发送名片消息
//
// 发送微信名片给指定用户
//
// 参数:
//   - wxid: 接收人微信ID（发送给谁）
//   - cardWxid: 名片的微信ID（谁的名片）
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	// 把 wxid_yyy 的名片发送给 wxid_xxx
//	err := client.Message.SendCardMsg("wxid_xxx", "wxid_yyy")
//
//	// 把某人的名片发送给文件传输助手
//	err := client.Message.SendCardMsg("filehelper", "wxid_yyy")
func (a *MessageAPI) SendCardMsg(wxid, cardWxid string) error {
	req := types.SendCardMsgRequest{
		Wxid:     wxid,
		CardWxid: cardWxid,
	}

	respBody, err := a.client.DoRequest("POST", "/wx41614/send_card_msg", req)
	if err != nil {
		return fmt.Errorf("send card msg: %w", err)
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

// SendImageMsg 发送图片消息
//
// 发送本地图片文件到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - imagePath: 本地图片文件的完整路径，例如 "C:\\images\\photo.jpg"
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Message.SendImageMsg("wxid_xxx", "C:\\images\\photo.jpg")
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
//   - 支持常见图片格式：jpg, png, gif, bmp等
func (a *MessageAPI) SendImageMsg(wxid, imagePath string) error {
	req := types.SendImageMsgRequest{
		Wxid:     wxid,
		FilePath: imagePath,
	}

	// 添加调试日志
	fmt.Printf("[DEBUG] 发送图片请求 - wxid: %s, path: %s\n", wxid, imagePath)

	respBody, err := a.client.DoRequest("POST", "/api/send_image_msg", req)
	if err != nil {
		return fmt.Errorf("send image msg: %w", err)
	}

	var resp types.Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	// 添加响应日志
	fmt.Printf("[DEBUG] 发送图片响应 - code: %d, msg: %s\n", resp.Code, resp.Msg)

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// SendFileMsg 发送文件消息
//
// 发送本地文件到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - filePath: 本地文件的完整路径，例如 "C:\\files\\document.pdf"
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Message.SendFileMsg("filehelper", "C:\\files\\document.pdf")
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows路径需要使用双反斜杠或单正斜杠
//   - 支持各种文件类型：pdf, doc, txt, zip等
func (a *MessageAPI) SendFileMsg(wxid, filePath string) error {
	req := types.SendFileMsgRequest{
		Wxid:     wxid,
		FilePath: filePath,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_file_msg", req)
	if err != nil {
		return fmt.Errorf("send file msg: %w", err)
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

// SendPat 发送拍一拍
//
// 在群聊中拍一拍指定用户
//
// 参数:
//   - roomID: 群聊ID，格式如 "39259098574@chatroom"
//   - wxid: 被拍的微信ID
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	// 在群聊中拍一拍某人
//	err := client.Message.SendPat("39259098574@chatroom", "wxid_xxx")
//
// 注意:
//   - 只能在群聊中使用拍一拍功能
//   - 被拍的用户必须是群成员
func (a *MessageAPI) SendPat(roomID, wxid string) error {
	req := types.SendPatRequest{
		RoomID: roomID,
		Wxid:   wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_pat", req)
	if err != nil {
		return fmt.Errorf("send pat: %w", err)
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

// SendQuote 发送引用消息
//
// 引用某条消息并回复
//
// 参数:
//   - reply: 回复的内容
//   - referContent: 被引用的原消息内容
//   - fromUsr: 原消息发送者的微信ID
//   - newMsgID: 原消息的消息ID
//   - createTime: 原消息的创建时间（Unix时间戳）
//   - sendTo: 发送目标（群聊ID或个人微信ID）
//   - msgSource: 消息源（可选，传空字符串表示不设置）
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	// 引用群聊中的消息并回复
//	err := client.Message.SendQuote(
//	    "我也这么认为",                    // 回复内容
//	    "今天天气真好",                     // 原消息内容
//	    "wxid_xxx",                       // 原消息发送者
//	    "5217518642639526576",            // 原消息ID
//	    1700000000,                       // 原消息时间
//	    "49767299448@chatroom",           // 发送到群聊
//	    "",                               // msgSource可选
//	)
//
// 注意:
//   - 需要从接收到的消息中获取相关参数（消息ID、发送者、时间等）
//   - createTime 通常设置为0，系统会自动处理
func (a *MessageAPI) SendQuote(reply, referContent, fromUsr, newMsgID string, createTime int, sendTo, msgSource string) error {
	req := types.SendQuoteRequest{
		Reply:        reply,
		ReferContent: referContent,
		FromUsr:      fromUsr,
		NewMsgID:     newMsgID,
		MsgSource:    msgSource,
		CreateTime:   createTime,
		SendTo:       sendTo,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_quote", req)
	if err != nil {
		return fmt.Errorf("send quote: %w", err)
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

// SendAppletMsg 发送小程序消息
//
// 发送小程序卡片到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - content: 小程序的XML内容，需要符合微信小程序消息格式
//   - msgType: 消息类型，小程序通常为 "33"
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	xmlContent := `<appmsg appid="" sdkver="0">
//	  <title>小程序标题</title>
//	  <des>小程序描述</des>
//	  <type>33</type>
//	  <weappinfo>
//	    <username>gh_xxx@app</username>
//	    <appid>wx_appid</appid>
//	    <pagepath>pages/index/index</pagepath>
//	  </weappinfo>
//	  ...
//	</appmsg>`
//	err := client.Message.SendAppletMsg("filehelper", xmlContent, "33")
//
// 注意:
//   - 小程序消息的 XML 格式较为复杂，需要包含完整的小程序信息
//   - 通常从转发的小程序消息中获取 XML 内容
func (a *MessageAPI) SendAppletMsg(wxid, content, msgType string) error {
	req := types.SendAppletMsgRequest{
		Wxid:    wxid,
		Content: content,
		Type:    msgType,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_applet_msg", req)
	if err != nil {
		return fmt.Errorf("send applet msg: %w", err)
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

// SendLocationMsg 发送位置消息
//
// 发送地理位置信息到指定微信用户
//
// 参数:
//   - wxid: 接收人微信ID
//   - longitude: 经度（字符串格式）
//   - latitude: 纬度（字符串格式）
//   - label: 位置标签/地址描述
//   - poiName: 地点名称
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.Message.SendLocationMsg(
//	    "wxid_xxx",
//	    "116.397128",      // 经度
//	    "39.916527",       // 纬度
//	    "北京市东城区",     // 标签
//	    "天安门广场",       // 地点名称
//	)
//
// 注意:
//   - 经纬度需要使用字符串格式
//   - 经度范围：-180 到 180
//   - 纬度范围：-90 到 90
func (a *MessageAPI) SendLocationMsg(wxid, longitude, latitude, label, poiName string) error {
	req := types.SendLocationMsgRequest{
		Wxid:    wxid,
		X:       longitude,
		Y:       latitude,
		Label:   label,
		PoiName: poiName,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_location_msg", req)
	if err != nil {
		return fmt.Errorf("send location msg: %w", err)
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

// RevokeAny 撤回任何消息
//
// 撤回指定的消息，可以撤回任何人的消息（需要有权限）
//
// 参数:
//   - newMsgID: 消息ID（从回调消息中获取）
//   - createTime: 消息创建时间（从回调消息中获取）
//   - toUserName: 接收人微信ID或群聊ID
//
// 返回:
//   - error: 撤回失败时返回错误信息
//
// 示例:
//
//	err := client.Message.RevokeAny(
//	    2050044161371926300,
//	    1761391928,
//	    "49767299448@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("消息已撤回")
//
// 注意:
//   - 需要从接收到的消息回调中获取 newMsgID 和 createTime
//   - toUserName 可以是个人微信ID或群聊ID
//   - 撤回消息需要有相应的权限
//   - 消息撤回有时间限制（通常为2分钟内）
func (a *MessageAPI) RevokeAny(newMsgID float64, createTime int, toUserName string) error {
	req := types.RevokeAnyRequest{
		NewMsgID:   newMsgID,
		CreateTime: createTime,
		ToUserName: toUserName,
	}

	respBody, err := a.client.DoRequest("POST", "/api/revoke_any", req)
	if err != nil {
		return fmt.Errorf("revoke any: %w", err)
	}

	var resp types.RevokeAnyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.ErrCode != 1 && resp.ErrCode != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	if resp.Data.BaseResponse.Ret != 0 {
		return fmt.Errorf("base response error: ret=%d, msg=%v",
			resp.Data.BaseResponse.Ret, resp.Data.BaseResponse.ErrMsg)
	}

	return nil
}
