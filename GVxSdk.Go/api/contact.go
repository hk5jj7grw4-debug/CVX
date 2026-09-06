package api

import (
	"encoding/json"
	"fmt"

	"github.com/olaria01/gVxSdk/sdk/go/types"
)

// ContactAPI 联系人相关接口
type ContactAPI struct {
	client Client
}

// NewContactAPI 创建联系人API实例
func NewContactAPI(client Client) *ContactAPI {
	return &ContactAPI{client: client}
}

// GetMyQRCode 获取好友二维码
//
// 获取指定微信ID或群聊的二维码图片。
// 支持1-8种不同风格的二维码样式。
//
// 参数:
//   - wxid: 微信ID或群聊ID，例如 "45220347292@chatroom"
//   - opcode: 操作码，通常为 "0"
//   - style: 风格，1-8 可选，不同数字代表不同的二维码样式
//   - info: 说明信息
//
// 返回:
//   - *types.GetMyQRCodeResponse: 二维码数据，包含Base64编码的图片
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.GetMyQRCode("45220347292@chatroom", "0", "7", "说明")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// resp.QRCode.Buffer 包含Base64编码的二维码图片数据
//	// 可以保存为图片文件或直接使用
//	fmt.Printf("二维码长度: %d\n", resp.QRCode.ILen)
//
// 注意:
//   - style 参数范围为 1-8，可以尝试不同风格查看效果
//   - 返回的 Buffer 字段是Base64编码的图片数据
//   - 需要解码后才能保存为图片文件
func (a *ContactAPI) GetMyQRCode(wxid, opcode, style, info string) (*types.GetMyQRCodeResponse, error) {
	req := types.GetMyQRCodeRequest{
		Wxid:   wxid,
		Opcode: opcode,
		Style:  style,
		Info:   info,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_my_qrocde", req)
	if err != nil {
		return nil, fmt.Errorf("get my qrcode: %w", err)
	}

	var resp types.GetMyQRCodeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 && resp.BaseResponse.Ret != -2 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// SearchContact 搜索微信号/手机号
//
// 通过微信号或手机号搜索联系人，返回用户的详细信息。
//
// 参数:
//   - search: 搜索关键词，可以是微信号或手机号
//
// 返回:
//   - *types.SearchContactResponse: 搜索到的联系人详细信息
//   - error: 搜索失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.SearchContact("wxid_hify7vdpvg5d22")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("昵称: %s\n", resp.NickName.String)
//	fmt.Printf("微信号: %s\n", resp.UserName.String)
//	fmt.Printf("个性签名: %s\n", resp.Signature)
//	fmt.Printf("性别: %d\n", resp.Sex)
//	fmt.Printf("头像: %s\n", resp.SmallHeadImgUrl)
//
// 注意:
//   - 搜索关键词可以是微信号或手机号
//   - 返回的 UserName 可能是 v3 开头的陌生人ID，需要添加好友后才能获取真实微信ID
//   - MatchType 字段表示匹配类型
//   - AntispamTicket 用于后续添加好友操作
func (a *ContactAPI) SearchContact(search string) (*types.SearchContactResponse, error) {
	req := types.SearchContactRequest{
		Search: search,
	}

	respBody, err := a.client.DoRequest("POST", "/api/net_scene_search_contact", req)
	if err != nil {
		return nil, fmt.Errorf("search contact: %w", err)
	}

	var resp types.SearchContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// GetProfileCache 获取个人资料缓存
//
// 获取当前登录账号的详细个人资料信息，包括昵称、头像、地区、签名等。
// 此接口无需参数，直接返回当前登录用户的完整资料。
//
// 返回:
//   - *types.GetProfileCacheResponse: 个人资料详细信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.GetProfileCache()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("微信号: %s\n", resp.UserInfo.UserName.String)
//	fmt.Printf("昵称: %s\n", resp.UserInfo.NickName.String)
//	fmt.Printf("手机号: %s\n", resp.UserInfo.BindMobile.String)
//	fmt.Printf("性别: %d (1-男, 2-女)\n", resp.UserInfo.Sex)
//	fmt.Printf("地区: %s %s\n", resp.UserInfo.Province, resp.UserInfo.City)
//	fmt.Printf("个性签名: %s\n", resp.UserInfo.Signature)
//	fmt.Printf("头像: %s\n", resp.UserInfoExt.SmallHeadImgUrl)
//
// 注意:
//   - 此接口返回的是缓存数据，可能不是实时最新的
//   - 包含大量个人信息字段，可根据需要选择使用
//   - Sex 字段：1-男性，2-女性
//   - 包含安全设备列表、支付设置等敏感信息
func (a *ContactAPI) GetProfileCache() (*types.GetProfileCacheResponse, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_profile_cache", nil)
	if err != nil {
		return nil, fmt.Errorf("get profile cache: %w", err)
	}

	var resp types.GetProfileCacheResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// GetContact 网络查询好友资料
//
// 通过网络查询指定微信ID的详细资料信息。
// 可以查询好友、陌生人或群聊的详细信息。
//
// 参数:
//   - wxid: 微信ID，例如 "xiaowen8-9" 或群聊ID
//
// 返回:
//   - *types.GetContactResponse: 联系人详细资料
//   - error: 查询失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.GetContact("xiaowen8-9")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if resp.ContactCount > 0 {
//	    contact := resp.ContactList[0]
//	    fmt.Printf("昵称: %s\n", contact.NickName.String)
//	    fmt.Printf("微信号: %s\n", contact.UserName.String)
//	    fmt.Printf("备注: %s\n", contact.Remark.String)
//	    fmt.Printf("性别: %d (1-男, 2-女)\n", contact.Sex)
//	    fmt.Printf("地区: %s %s\n", contact.Province, contact.City)
//	    fmt.Printf("个性签名: %s\n", contact.Signature)
//	    fmt.Printf("头像: %s\n", contact.SmallHeadImgUrl)
//	}
//
// 注意:
//   - 此接口通过网络实时查询，可能比缓存接口慢
//   - 返回的 ContactList 数组通常只包含一个元素
//   - Sex 字段：1-男性，2-女性
//   - VerifyFlag 表示验证状态（是否需要验证）
func (a *ContactAPI) GetContact(wxid string) (*types.GetContactResponse, error) {
	req := types.GetContactRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_contact", req)
	if err != nil {
		return nil, fmt.Errorf("get contact: %w", err)
	}

	var resp types.GetContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// ModSelfNickName 修改自己昵称
//
// 修改当前登录账号的昵称。
//
// 参数:
//   - newName: 新昵称，支持中文、英文、数字和emoji表情
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.ModSelfNickName("鸭梨🍐大a")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("昵称修改成功")
//
// 注意:
//   - 昵称支持emoji表情
//   - 昵称长度有限制，过长可能会失败
//   - 修改后可能需要一定时间才能在所有地方生效
func (a *ContactAPI) ModSelfNickName(newName string) error {
	req := types.ModSelfNickNameRequest{
		NewName: newName,
	}

	respBody, err := a.client.DoRequest("POST", "/api/mod_self_nick_name", req)
	if err != nil {
		return fmt.Errorf("mod self nick name: %w", err)
	}

	var resp types.ModSelfNickNameResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Ret != 0 {
		return fmt.Errorf("api error: ret=%d", resp.Ret)
	}

	// 检查操作日志返回
	if resp.OplogRet.Count > 0 && len(resp.OplogRet.Ret) > 0 {
		if resp.OplogRet.Ret[0] != 0 {
			return fmt.Errorf("oplog error: ret=%d", resp.OplogRet.Ret[0])
		}
	}

	return nil
}

// AddFriend 添加好友
//
// 向指定用户发送好友申请。
// 需要提供从搜索接口获取的 v3、v4 等参数。
//
// 参数:
//   - v3: V3加密用户名，从 SearchContact 接口的 UserName 字段获取
//   - v4: V4验证票据，从 SearchContact 接口的 AntispamTicket 字段获取
//   - scence: 场景值，通常为 "3"
//   - friendFlg: 好友标志，通常为 "0"
//   - verifyContent: 验证消息内容，向对方展示的申请理由
//
// 返回:
//   - error: 添加失败时返回错误信息
//
// 示例:
//
//	// 先搜索用户
//	searchResp, err := client.Contact.SearchContact("wxid_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 使用搜索结果添加好友
//	err = client.Contact.AddFriend(
//	    searchResp.UserName.String,
//	    searchResp.AntispamTicket,
//	    "3",
//	    "0",
//	    "你好，我想加你为好友",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("好友申请已发送")
//
// 注意:
//   - v3 和 v4 参数必须从 SearchContact 接口获取
//   - 不同的添加场景 scence 值可能不同
//   - verifyContent 是向对方展示的验证消息
//   - 对方需要同意后才能成为好友
func (a *ContactAPI) AddFriend(v3, v4, scence, friendFlg, verifyContent string) error {
	req := types.AddFriendRequest{
		V3:            v3,
		V4:            v4,
		Scence:        scence,
		FriendFlg:     friendFlg,
		VerifyContent: verifyContent,
	}

	respBody, err := a.client.DoRequest("POST", "/api/add_friend", req)
	if err != nil {
		return fmt.Errorf("add friend: %w", err)
	}

	var resp types.AddFriendResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// GetProfileNew 获取个人最新网络资料
//
// 从网络获取当前登录账号的最新个人资料信息。
// 与 GetProfileCache 不同，此接口从网络实时获取，数据更新。
//
// 返回:
//   - *types.GetProfileNewResponse: 个人资料详细信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.GetProfileNew()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("微信号: %s\n", resp.UserInfo.UserName.String)
//	fmt.Printf("昵称: %s\n", resp.UserInfo.NickName.String)
//	fmt.Printf("手机号: %s\n", resp.UserInfo.BindMobile.String)
//	fmt.Printf("性别: %d (1-男, 2-女)\n", resp.UserInfo.Sex)
//	fmt.Printf("地区: %s %s\n", resp.UserInfo.Province, resp.UserInfo.City)
//	fmt.Printf("个性签名: %s\n", resp.UserInfo.Signature)
//	fmt.Printf("头像: %s\n", resp.UserInfoExt.SmallHeadImgUrl)
//
// 注意:
//   - 此接口从网络实时获取，比 GetProfileCache 慢但数据更新
//   - 返回的数据结构与 GetProfileCache 相同
//   - Sex 字段：1-男性，2-女性
//   - 包含安全设备列表、支付设置等敏感信息
func (a *ContactAPI) GetProfileNew() (*types.GetProfileNewResponse, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_profile_new", nil)
	if err != nil {
		return nil, fmt.Errorf("get profile new: %w", err)
	}

	var resp types.GetProfileNewResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// GetContactFast 快速查找好友资料
//
// 快速获取好友资料信息，速度非常快。
// 建议第一次使用时调用 wechat_init 更新好友列表，之后查询速度会非常快。
//
// 参数:
//   - wxid: 微信ID，例如 "wxid_xxx" 或 "filehelper"。如果为空字符串，则更新所有好友资料
//
// 返回:
//   - *types.GetContactFastResponse: 好友资料信息
//   - error: 查询失败时返回错误信息
//
// 示例:
//
//	// 查询单个好友
//	resp, err := client.Contact.GetContactFast("filehelper")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("昵称: %s\n", resp.Contact.NickName.String)
//	fmt.Printf("微信号: %s\n", resp.Contact.UserName.String)
//	fmt.Printf("备注: %s\n", resp.Contact.Remark.String)
//	fmt.Printf("性别: %d (0-未知, 1-男, 2-女)\n", resp.Contact.Sex)
//	fmt.Printf("地区: %s %s %s\n", resp.Contact.Country, resp.Contact.Province, resp.Contact.City)
//	fmt.Printf("头像: %s\n", resp.Contact.SmallHeadImgUrl)
//
//	// 更新所有好友资料（第一次使用时建议调用）
//	_, err = client.Contact.GetContactFast("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("好友资料更新完成")
//
// 注意:
//   - 建议第一次使用时传入空字符串更新所有好友资料
//   - 更新所有好友是长耗时操作，具体时间取决于好友数量
//   - 更新后重新登录查询速度也会很快
//   - Sex 字段：0-未知，1-男性，2-女性
func (a *ContactAPI) GetContactFast(wxid string) (*types.GetContactFastResponse, error) {
	req := types.GetContactFastRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_contact_fast", req)
	if err != nil {
		return nil, fmt.Errorf("get contact fast: %w", err)
	}

	var resp types.GetContactFastResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d", resp.Ret)
	}

	return &resp, nil
}

// UpdateAllFriend 更新好友列表
//
// 向微信服务器获取一次完整的好友列表并更新本地缓存。
// 该接口与 GetContactFast 联动，建议一天更新一次。
// 如果好友长时间没有变化，不更新也可以。
//
// 返回:
//   - *types.UpdateAllFriendResponse: 包含所有好友信息的列表
//   - error: 更新失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.UpdateAllFriend()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("好友总数: %d\n", resp.FriendCount)
//	for _, friend := range resp.Data {
//	    if friend.Ret == 0 {
//	        fmt.Printf("- %s (%s)\n",
//	            friend.Contact.NickName.String,
//	            friend.Contact.UserName.String)
//	        if friend.Contact.Remark.String != "" {
//	            fmt.Printf("  备注: %s\n", friend.Contact.Remark.String)
//	        }
//	    }
//	}
//
// 注意:
//   - 这是一个长耗时操作，具体时间取决于好友数量
//   - 建议一天更新一次即可
//   - 更新后使用 GetContactFast 查询速度会非常快
//   - 如果好友长时间没有变化，可以不频繁更新
//   - 返回的数据包含所有好友的详细信息
func (a *ContactAPI) UpdateAllFriend() (*types.UpdateAllFriendResponse, error) {
	respBody, err := a.client.DoRequest("POST", "/api/update_all_friend", nil)
	if err != nil {
		return nil, fmt.Errorf("update all friend: %w", err)
	}

	var resp types.UpdateAllFriendResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// RemarkContact 修改好友备注
//
// 修改指定好友的备注名称。
//
// 参数:
//   - wxid: 好友的微信ID，例如 "wxid_ozyqateb85un22"
//   - remark: 新的备注名称，例如 "老王"
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.RemarkContact("wxid_ozyqateb85un22", "老王")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("备注修改成功")
//
// 注意:
//   - 只能修改已经是好友的用户备注
//   - 备注名称支持中文、英文、数字和emoji表情
//   - 备注长度有限制，过长可能会失败
//   - 修改后可能需要一定时间才能在所有地方生效
func (a *ContactAPI) RemarkContact(wxid, remark string) error {
	req := types.RemarkContactRequest{
		Wxid:   wxid,
		Remark: remark,
	}

	respBody, err := a.client.DoRequest("POST", "/api/remark_contact", req)
	if err != nil {
		return fmt.Errorf("remark contact: %w", err)
	}

	var resp types.RemarkContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// DelContact 删除好友
//
// 删除指定的好友联系人。
//
// 参数:
//   - wxid: 要删除的好友微信ID，例如 "wxid_8543785438012"
//
// 返回:
//   - error: 删除失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.DelContact("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("好友删除成功")
//
// 注意:
//   - 删除操作不可逆，请谨慎使用
//   - 只能删除已经是好友的用户
//   - 删除后对方仍然可以看到你的朋友圈（如果之前有权限）
//   - 删除后聊天记录会保留，但无法再发送消息
//   - 如果需要重新添加，需要对方同意
func (a *ContactAPI) DelContact(wxid string) error {
	req := types.DelContactRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/del_contact", req)
	if err != nil {
		return fmt.Errorf("del contact: %w", err)
	}

	var resp types.DelContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// UploadHeadImg 修改头像
//
// 修改当前登录账号的头像。
//
// 参数:
//   - filePath: 头像图片文件路径，例如 "D:\\2.png"
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.UploadHeadImg("D:\\2.png")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("头像修改成功")
//
// 注意:
//   - 文件路径必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 支持常见的图片格式（jpg, png 等）
//   - 建议使用正方形图片，避免变形
//   - 修改后可能需要一定时间才能在所有地方生效
func (a *ContactAPI) UploadHeadImg(filePath string) error {
	req := types.UploadHeadImgRequest{
		FilePath: filePath,
	}

	respBody, err := a.client.DoRequest("POST", "/api/upload_head_img", req)
	if err != nil {
		return fmt.Errorf("upload head img: %w", err)
	}

	var resp types.UploadHeadImgResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// ModSelfSignature 修改个人签名
//
// 修改当前登录账号的个性签名。
//
// 参数:
//   - newSignature: 新的个性签名，例如 "666666666666666"
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.ModSelfSignature("生活不止眼前的苟且，还有诗和远方")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("个性签名修改成功")
//
// 注意:
//   - 个性签名支持中文、英文、数字和emoji表情
//   - 签名长度有限制，过长可能会失败
//   - 修改后可能需要一定时间才能在所有地方生效
func (a *ContactAPI) ModSelfSignature(newSignature string) error {
	req := types.ModSelfSignatureRequest{
		NewSignature: newSignature,
	}

	respBody, err := a.client.DoRequest("POST", "/api/mod_self_nick_signature", req)
	if err != nil {
		return fmt.Errorf("mod self signature: %w", err)
	}

	var resp types.ModSelfSignatureResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// GetLabelLists 获取标签列表
//
// 获取当前账号的所有联系人标签列表。
// 标签用于对好友进行分类管理。
//
// 返回:
//   - *types.GetLabelListsResponse: 标签列表信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.GetLabelLists()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("标签总数: %d\n", resp.LabelCount)
//	for _, label := range resp.LabelPairList {
//	    fmt.Printf("- [%d] %s\n", label.LabelId, label.LabelName)
//	}
//
// 注意:
//   - 返回的标签列表包含所有自定义标签
//   - LabelId 可用于后续的标签管理操作
//   - 标签名称支持中文、英文、数字和emoji表情
func (a *ContactAPI) GetLabelLists() (*types.GetLabelListsResponse, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_label_lists", nil)
	if err != nil {
		return nil, fmt.Errorf("get label lists: %w", err)
	}

	var resp types.GetLabelListsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// AddLabel 增加标签
//
// 创建一个新的联系人标签。
// 标签可用于对好友进行分类管理。
//
// 参数:
//   - label: 标签名称，例如 "我的标签"
//
// 返回:
//   - *types.AddLabelResponse: 包含新创建的标签信息
//   - error: 创建失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Contact.AddLabel("我的标签")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("标签创建成功！\n")
//	fmt.Printf("标签总数: %d\n", resp.LabelCount)
//	for _, label := range resp.LabelPairList {
//	    fmt.Printf("- [%d] %s\n", label.LabelId, label.LabelName)
//	}
//
// 注意:
//   - 标签名称支持中文、英文、数字和emoji表情
//   - 标签名称不能重复
//   - 返回的 LabelPairList 包含新创建的标签信息
//   - LabelId 可用于后续的标签管理操作
func (a *ContactAPI) AddLabel(label string) (*types.AddLabelResponse, error) {
	req := types.AddLabelRequest{
		Label: label,
	}

	respBody, err := a.client.DoRequest("POST", "/api/add_label", req)
	if err != nil {
		return nil, fmt.Errorf("add label: %w", err)
	}

	var resp types.AddLabelResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// DelLabel 删除标签
//
// 删除指定的联系人标签。
//
// 参数:
//   - labelID: 标签ID，可以从 GetLabelLists 或 AddLabel 接口获取
//
// 返回:
//   - error: 删除失败时返回错误信息
//
// 示例:
//
//	// 先获取标签列表
//	resp, err := client.Contact.GetLabelLists()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 删除第一个标签
//	if len(resp.LabelPairList) > 0 {
//	    labelID := fmt.Sprintf("%d", resp.LabelPairList[0].LabelId)
//	    err = client.Contact.DelLabel(labelID)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println("标签删除成功")
//	}
//
// 注意:
//   - labelID 参数为字符串类型，需要将整数ID转换为字符串
//   - 删除操作不可逆，请谨慎使用
//   - 删除标签后，该标签下的好友不会被删除，只是移除标签关联
func (a *ContactAPI) DelLabel(labelID string) error {
	req := types.DelLabelRequest{
		LabelID: labelID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/del_label", req)
	if err != nil {
		return fmt.Errorf("del label: %w", err)
	}

	var resp types.DelLabelResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// ModifyContactLabel 修改好友标签
//
// 为指定好友设置标签。可以设置多个标签。
//
// 参数:
//   - wxids: 好友微信ID，例如 "wxid_8543785438012"
//   - labelID: 标签ID列表，多个标签用逗号分隔，例如 "2,6"
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	// 为好友设置单个标签
//	err := client.Contact.ModifyContactLabel("wxid_8543785438012", "2")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("标签设置成功")
//
//	// 为好友设置多个标签
//	err = client.Contact.ModifyContactLabel("wxid_8543785438012", "2,6,8")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("多个标签设置成功")
//
// 注意:
//   - labelID 参数为字符串类型，多个标签用逗号分隔
//   - 此操作会覆盖好友原有的标签设置
//   - 如果要清空标签，可以传入空字符串
//   - 标签ID可以从 GetLabelLists 接口获取
func (a *ContactAPI) ModifyContactLabel(wxids, labelID string) error {
	req := types.ModifyContactLabelRequest{
		Wxids:   wxids,
		LabelID: labelID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/modify_contact_label", req)
	if err != nil {
		return fmt.Errorf("modify contact label: %w", err)
	}

	var resp types.ModifyContactLabelResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// UpdateLabelName 更新标签名字
//
// 修改指定标签的名称。
//
// 参数:
//   - labelID: 标签ID，可以从 GetLabelLists 接口获取
//   - newName: 新的标签名称，例如 "新标签名字"
//
// 返回:
//   - error: 更新失败时返回错误信息
//
// 示例:
//
//	// 先获取标签列表
//	resp, err := client.Contact.GetLabelLists()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 修改第一个标签的名称
//	if len(resp.LabelPairList) > 0 {
//	    err = client.Contact.UpdateLabelName(resp.LabelPairList[0].LabelId, "新标签名字")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println("标签名称更新成功")
//	}
//
// 注意:
//   - labelID 参数为整数类型
//   - 新标签名称支持中文、英文、数字和emoji表情
//   - 新标签名称不能与已有标签重复
//   - 修改后立即生效
func (a *ContactAPI) UpdateLabelName(labelID int, newName string) error {
	req := types.UpdateLabelNameRequest{
		LabelID: labelID,
		NewName: newName,
	}

	respBody, err := a.client.DoRequest("POST", "/api/update_label_name", req)
	if err != nil {
		return fmt.Errorf("update label name: %w", err)
	}

	var resp types.UpdateLabelNameResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// UpdateSingleProfile 更新单个用户资料
//
// 从服务器更新指定用户的最新资料信息到本地缓存
//
// 参数:
//   - wxid: 微信ID，例如 "filehelper"
//
// 返回:
//   - *types.SingleProfileData: 更新后的用户资料信息
//   - error: 更新失败时返回错误信息
//
// 示例:
//
//	profile, err := client.Contact.UpdateSingleProfile("filehelper")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("昵称: %s\n", profile.NickName.String)
//	fmt.Printf("微信号: %s\n", profile.Alias)
//	fmt.Printf("性别: %d\n", profile.Sex)
//	fmt.Printf("地区: %s %s\n", profile.Province, profile.City)
//	fmt.Printf("头像: %s\n", profile.SmallHeadImgUrl)
//
// 注意:
//   - 此接口会从服务器拉取最新的用户资料
//   - 返回的资料包含昵称、头像、地区、性别等详细信息
//   - 可用于更新好友或群成员的最新资料
func (a *ContactAPI) UpdateSingleProfile(wxid string) (*types.SingleProfileData, error) {
	req := types.UpdateSingleProfileRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/update_single_profile", req)
	if err != nil {
		return nil, fmt.Errorf("update single profile: %w", err)
	}

	var resp types.UpdateSingleProfileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	// 如果 data 不为 nil，尝试解析为 SingleProfileData
	if resp.Data != nil {
		dataBytes, err := json.Marshal(resp.Data)
		if err != nil {
			return nil, fmt.Errorf("marshal data: %w", err)
		}

		var profile types.SingleProfileData
		if err := json.Unmarshal(dataBytes, &profile); err != nil {
			return nil, fmt.Errorf("unmarshal profile data: %w", err)
		}

		return &profile, nil
	}

	return nil, fmt.Errorf("no profile data returned")
}

// SetTop 置顶好友
//
// 将指定的好友或群聊置顶到聊天列表顶部
//
// 参数:
//   - wxid: 微信ID或群聊ID
//
// 返回:
//   - error: 置顶失败时返回错误信息
//
// 示例:
//
//	// 置顶好友
//	err := client.Contact.SetTop("wxid_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("好友已置顶")
//
//	// 置顶群聊
//	err = client.Contact.SetTop("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群聊已置顶")
//
// 注意:
//   - 可以置顶好友或群聊
//   - 置顶后会显示在聊天列表顶部
//   - 再次调用可以取消置顶
func (a *ContactAPI) SetTop(wxid string) error {
	req := types.SetTopRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/set_top", req)
	if err != nil {
		return fmt.Errorf("set top: %w", err)
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

// CancelTop 取消置顶
//
// 取消指定好友或群聊的置顶状态
//
// 参数:
//   - wxid: 微信ID或群聊ID
//
// 返回:
//   - error: 取消置顶失败时返回错误信息
//
// 示例:
//
//	// 取消置顶好友
//	err := client.Contact.CancelTop("wxid_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已取消置顶")
//
//	// 取消置顶群聊
//	err = client.Contact.CancelTop("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已取消置顶")
//
// 注意:
//   - 可以取消置顶好友或群聊
//   - 取消后会恢复到正常排序位置
func (a *ContactAPI) CancelTop(wxid string) error {
	req := types.CancelTopRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/cancel_top", req)
	if err != nil {
		return fmt.Errorf("cancel top: %w", err)
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

// SetStar 星标好友
//
// 将指定好友设置为星标朋友
//
// 参数:
//   - wxid: 微信ID
//
// 返回:
//   - error: 设置失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.SetStar("wxid_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已设置为星标朋友")
//
// 注意:
//   - 星标朋友会显示在通讯录的星标朋友分组中
//   - 再次调用可以取消星标
func (a *ContactAPI) SetStar(wxid string) error {
	req := types.SetStarRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/set_start", req)
	if err != nil {
		return fmt.Errorf("set star: %w", err)
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

// DelStar 取消星标
//
// 取消指定好友的星标状态
//
// 参数:
//   - wxid: 微信ID
//
// 返回:
//   - error: 取消失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.DelStar("wxid_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已取消星标")
//
// 注意:
//   - 取消后好友将从星标朋友分组中移除
//   - 恢复到普通好友状态
func (a *ContactAPI) DelStar(wxid string) error {
	req := types.DelStarRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/del_start", req)
	if err != nil {
		return fmt.Errorf("del star: %w", err)
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

// SetMuteUser 开启消息免打扰
//
// 开启指定好友或群聊的消息免打扰功能
//
// 参数:
//   - wxid: 微信ID或群聊ID
//
// 返回:
//   - error: 设置失败时返回错误信息
//
// 示例:
//
//	// 开启好友免打扰
//	err := client.Contact.SetMuteUser("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已开启消息免打扰")
//
//	// 开启群聊免打扰
//	err = client.Contact.SetMuteUser("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已开启群聊免打扰")
//
// 注意:
//   - 开启后将不会收到该好友或群聊的消息提醒
//   - 消息仍会正常接收，只是不会有通知
//   - 可以对好友或群聊设置免打扰
func (a *ContactAPI) SetMuteUser(wxid string) error {
	req := types.SetMuteUserRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/set_mute_user", req)
	if err != nil {
		return fmt.Errorf("set mute user: %w", err)
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

// DelMuteUser 关闭消息免打扰
//
// 关闭指定好友或群聊的消息免打扰功能
//
// 参数:
//   - wxid: 微信ID或群聊ID
//
// 返回:
//   - error: 关闭失败时返回错误信息
//
// 示例:
//
//	// 关闭好友免打扰
//	err := client.Contact.DelMuteUser("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已关闭消息免打扰")
//
//	// 关闭群聊免打扰
//	err = client.Contact.DelMuteUser("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已关闭群聊免打扰")
//
// 注意:
//   - 关闭后将恢复该好友或群聊的消息提醒
//   - 可以对好友或群聊关闭免打扰
func (a *ContactAPI) DelMuteUser(wxid string) error {
	req := types.DelMuteUserRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/del_mute_user", req)
	if err != nil {
		return fmt.Errorf("del mute user: %w", err)
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

// BlackUser 拉黑好友
//
// 将指定好友加入黑名单
//
// 参数:
//   - wxid: 微信ID
//
// 返回:
//   - error: 拉黑失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.BlackUser("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已拉黑该好友")
//
// 注意:
//   - 拉黑后将无法收到对方的消息
//   - 对方也无法给你发送消息
//   - 拉黑后仍保留好友关系
//   - 可以通过取消拉黑恢复正常
func (a *ContactAPI) BlackUser(wxid string) error {
	req := types.BlackUserRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/black_user", req)
	if err != nil {
		return fmt.Errorf("black user: %w", err)
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

// DelBlackUser 移出黑名单
//
// 将指定好友从黑名单中移除
//
// 参数:
//   - wxid: 微信ID
//
// 返回:
//   - error: 移除失败时返回错误信息
//
// 示例:
//
//	err := client.Contact.DelBlackUser("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已移出黑名单")
//
// 注意:
//   - 移出后将恢复正常的消息收发
//   - 双方可以正常通信
func (a *ContactAPI) DelBlackUser(wxid string) error {
	req := types.DelBlackUserRequest{
		Wxid: wxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/del_black_user", req)
	if err != nil {
		return fmt.Errorf("del black user: %w", err)
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

// VerifyFriend 同意好友申请
//
// 同意好友验证申请，可以设置备注名和标签。
// 通常在收到好友申请的回调消息后调用此接口。
//
// 参数:
//   - wxid: 微信ID或V3加密用户名（从好友申请消息中获取）
//   - v4: V4验证票据（从好友申请消息中获取）
//   - remark: 备注名，可以为空
//   - label: 标签ID，可以为空，多个标签用逗号分隔
//   - scene: 来源场景值（从好友申请消息中获取）
//
// 返回:
//   - error: 同意失败时返回错误信息
//
// 示例:
//
//	// 基本用法：同意好友申请并设置备注
//	err := client.Contact.VerifyFriend(
//	    "wxid_8543785438012",
//	    "v4_000b708f0b0400000100000000002c0b06c6e5b066d23eedebc223691000000050ded0b020927e3c97896a09d47e6e9e99a49ffcdc8bbe2894ddc5a9245107f13bb0ac48fe6fcb0345f2da3e01ca2da76c5617ebb83954f66c0b1ac33a3e1958625430adb8ca9c10a47078f33a545ec8b63dc95a456b4bb0fde045d9f6a2d61c4233f94b5c3e7d5ab2f2028ca35feebba88c71f62dd11667@stranger",
//	    "9763",
//	    "13",
//	    3,
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已同意好友申请")
//
//	// 不设置备注和标签
//	err = client.Contact.VerifyFriend(
//	    "wxid_xxx",
//	    "v4_xxx@stranger",
//	    "",
//	    "",
//	    3,
//	)
//
// 注意:
//   - wxid 和 v4 参数通常从好友申请的回调消息中获取
//   - v4 是一个长字符串，通常以 "v4_" 开头，以 "@stranger" 结尾
//   - scene 参数表示添加来源，常见值：3-通过搜索添加，17-通过名片分享
//   - remark 和 label 可以为空字符串
//   - 同意后对方会立即成为好友
func (a *ContactAPI) VerifyFriend(wxid, v4, remark, label string, scene int) error {
	req := types.VerifyFriendRequest{
		Wxid:   wxid,
		V4:     v4,
		Remark: remark,
		Label:  label,
		Scene:  scene,
	}

	respBody, err := a.client.DoRequest("POST", "/api/verify_friend", req)
	if err != nil {
		return fmt.Errorf("verify friend: %w", err)
	}

	var resp types.VerifyFriendResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// GetContactList2 获取好友列表方法2
//
// 使用二叉树方式获取好友列表。
// 注意：群聊只有保存到通讯录里才会显示在列表中。
//
// 返回:
//   - []types.FriendInfo2: 好友信息列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	friends, err := client.Contact.GetContactList2()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个好友\n", len(friends))
//	for _, friend := range friends {
//	    fmt.Printf("- %s (%s)\n", friend.NickName, friend.Wxid)
//	    if friend.Remark != "" {
//	        fmt.Printf("  备注: %s\n", friend.Remark)
//	    }
//	    if friend.Alias != "" {
//	        fmt.Printf("  微信号: %s\n", friend.Alias)
//	    }
//	    fmt.Printf("  地区: %s %s %s\n", friend.Country, friend.Province, friend.City)
//	}
//
// 注意:
//   - 此方法使用二叉树方式获取，性能较好
//   - 群聊只有保存到通讯录里才会显示
//   - 返回的列表包含好友的详细信息（昵称、备注、地区、头像等）
//   - 包含拼音和简拼字段，方便排序和搜索
func (a *ContactAPI) GetContactList2() ([]types.FriendInfo2, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_contact_list2", nil)
	if err != nil {
		return nil, fmt.Errorf("get contact list2: %w", err)
	}

	var resp types.GetContactList2Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return resp.FriendList, nil
}

// GetFriendWxids 获取所有好友的wxid
//
// 从网络获取所有好友的微信ID列表。
// 注意：这是一个长耗时接口，可能需要较长时间完成。
//
// 返回:
//   - []string: 好友微信ID列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	wxids, err := client.Contact.GetFriendWxids()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个好友\n", len(wxids))
//	for i, wxid := range wxids {
//	    fmt.Printf("%d. %s\n", i+1, wxid)
//	}
//
// 注意:
//   - 这是一个长耗时接口，可能需要较长时间完成
//   - 从网络获取最新的好友列表
//   - 建议在后台异步执行，避免阻塞主流程
//   - 适合在启动时或定期更新好友列表时使用
//   - 只返回wxid，不包含其他详细信息
func (a *ContactAPI) GetFriendWxids() ([]string, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_friend_wxids", nil)
	if err != nil {
		return nil, fmt.Errorf("get friend wxids: %w", err)
	}

	var resp types.GetFriendWxidsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Wxids, nil
}

// BatchGetWxids 批量获取wxid信息
//
// 批量获取多个微信ID的详细信息。
// 可以一次性获取多个好友或群成员的资料。
//
// 参数:
//   - wxids: 微信ID列表，用逗号分隔，例如 "wxid_xxx,wxid_yyy,wxid_zzz"
//
// 返回:
//   - []types.BatchWxidInfo: 微信ID信息列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	infos, err := client.Contact.BatchGetWxids(
//	    "wxid_60ow7mbi0gpj22,wxid_8543785438012,wxid_tqyh06fntmo722",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, info := range infos {
//	    fmt.Printf("微信ID: %s\n", info.Wxid)
//	    fmt.Printf("昵称: %s\n", info.NickName)
//	    if info.Remark != "" {
//	        fmt.Printf("备注: %s\n", info.Remark)
//	    }
//	    if info.Alias != "" {
//	        fmt.Printf("微信号: %s\n", info.Alias)
//	    }
//	    fmt.Printf("地区: %s %s %s\n", info.Country, info.Province, info.City)
//	    fmt.Println("---")
//	}
//
// 注意:
//   - 可以一次性查询多个wxid的信息
//   - wxids 参数用逗号分隔，不要有空格
//   - 返回的信息包含昵称、备注、地区、头像等详细资料
//   - 适合批量查询好友或群成员信息的场景
func (a *ContactAPI) BatchGetWxids(wxids string) ([]types.BatchWxidInfo, error) {
	req := types.BatchGetWxidsRequest{
		Wxids: wxids,
	}

	respBody, err := a.client.DoRequest("POST", "/api/batch_get_wxids", req)
	if err != nil {
		return nil, fmt.Errorf("batch get wxids: %w", err)
	}

	var resp types.BatchGetWxidsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// Folding 折叠群聊或个人
//
// 将指定的群聊或个人联系人折叠到聊天列表的折叠区域。
// 折叠后的聊天不会显示在主聊天列表中，但仍可以接收消息。
//
// 参数:
//   - roomID: 群聊ID（格式为 "数字@chatroom"）或个人微信ID
//
// 返回:
//   - error: 折叠失败时返回错误信息
//
// 示例:
//
//	// 折叠群聊
//	err := client.Contact.Folding("18402658081@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群聊已折叠")
//
//	// 折叠个人联系人
//	err = client.Contact.Folding("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("联系人已折叠")
//
// 注意:
//   - 折叠后的聊天会移到折叠区域
//   - 仍然可以接收消息，但不会在主列表显示
//   - 可以折叠群聊或个人联系人
//   - 折叠的聊天可以通过展开操作恢复到主列表
func (a *ContactAPI) Folding(roomID string) error {
	req := types.FoldingRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/folding", req)
	if err != nil {
		return fmt.Errorf("folding: %w", err)
	}

	var resp types.FoldingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// Unfolding 取消折叠群聊或个人
//
// 将已折叠的群聊或个人联系人恢复到主聊天列表。
// 取消折叠后，聊天会重新显示在主聊天列表中。
//
// 参数:
//   - roomID: 群聊ID（格式为 "数字@chatroom"）或个人微信ID
//
// 返回:
//   - error: 取消折叠失败时返回错误信息
//
// 示例:
//
//	// 取消折叠群聊
//	err := client.Contact.Unfolding("18402658081@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群聊已恢复到主列表")
//
//	// 取消折叠个人联系人
//	err = client.Contact.Unfolding("wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("联系人已恢复到主列表")
//
// 注意:
//   - 取消折叠后，聊天会重新显示在主聊天列表中
//   - 可以取消折叠群聊或个人联系人
//   - 与 Folding 方法相对应
func (a *ContactAPI) Unfolding(roomID string) error {
	req := types.UnfoldingRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/unfolding", req)
	if err != nil {
		return fmt.Errorf("unfolding: %w", err)
	}

	var resp types.UnfoldingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}
