package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVXkernel/GVxSdk.Go/types"
)

// RoomAPI 群聊相关接口
type RoomAPI struct {
	client Client
}

// NewRoomAPI 创建群聊API实例
func NewRoomAPI(client Client) *RoomAPI {
	return &RoomAPI{client: client}
}

// GetRoomMembers 获取群成员列表
//
// 获取指定群聊的所有成员信息，包括成员昵称、头像、邀请人等详细资料。
// 如果消息里面没有群成员资料的话，可以调用该接口进行缓存。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "39259098574@chatroom"
//
// 返回:
//   - *types.GetRoomMembersResponse: 群成员列表详细信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.GetRoomMembers("39259098574@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("群主: %s\n", resp.ChatRoomOwner)
//	fmt.Printf("成员总数: %d\n", resp.AllMemberCount)
//	for _, member := range resp.NewChatroomData.ChatRoomMember {
//	    fmt.Printf("- %s (%s)\n", member.NickName, member.UserName)
//	}
//
// 注意:
//   - roomID 必须是有效的群聊ID，格式为 "数字@chatroom"
//   - 返回的成员列表包含详细的头像URL和邀请信息
//   - 可以通过 ChatRoomOwner 字段获取群主微信ID
func (a *RoomAPI) GetRoomMembers(roomID string) (*types.GetRoomMembersResponse, error) {
	req := types.GetRoomMembersRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_room_members", req)
	if err != nil {
		return nil, fmt.Errorf("get room members: %w", err)
	}

	var resp types.GetRoomMembersResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v", resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// GetChatroomDetailCache 获取群详情缓存
//
// 获取指定群聊的详细信息缓存，适用于需要频繁获取群资料的场景。
// 相比 GetRoomMembers，此接口返回的数据包含更多字段（如 displayName）。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "18402658081@chatroom"
//
// 返回:
//   - *types.GetChatroomDetailCacheResponse: 群详情缓存信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.GetChatroomDetailCache("18402658081@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("群主: %s\n", resp.Data.ChatRoomOwner)
//	fmt.Printf("成员总数: %d\n", resp.Data.AllMemberCount)
//	for _, member := range resp.Data.NewChatroomData.ChatRoomMember {
//	    fmt.Printf("- %s (%s) - 显示名: %s\n",
//	        member.NickName, member.UserName, member.DisplayName)
//	}
//
// 注意:
//   - 此接口适用于需要频繁获取群资料的场景
//   - 返回的成员信息包含 displayName 字段（群昵称）
//   - 响应格式与 GetRoomMembers 略有不同，包含外层的 account_wxid 等字段
func (a *RoomAPI) GetChatroomDetailCache(roomID string) (*types.GetChatroomDetailCacheResponse, error) {
	req := types.GetChatroomDetailCacheRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_chatroom_detail_cache", req)
	if err != nil {
		return nil, fmt.Errorf("get chatroom detail cache: %w", err)
	}

	var resp types.GetChatroomDetailCacheResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.ErrCode != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	if resp.Data.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("base response error: ret=%d, msg=%v",
			resp.Data.BaseResponse.Ret, resp.Data.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// GetGroupMemberContact 查询群成员信息
//
// 查询指定群聊中某个成员的详细信息。
// 可以获取群成员的昵称、头像、地区、签名等详细资料。
//
// 参数:
//   - wxid: 群成员的微信ID
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - *types.GetGroupMemberContactResponse: 群成员详细信息
//   - error: 查询失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.GetGroupMemberContact("wxid_8zggbw1yo5ib22", "18402658081@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if resp.ContactCount > 0 {
//	    member := resp.ContactList[0]
//	    fmt.Printf("昵称: %s\n", member.NickName.String)
//	    fmt.Printf("微信号: %s\n", member.UserName.String)
//	    fmt.Printf("性别: %d (1-男, 2-女)\n", member.Sex)
//	    fmt.Printf("地区: %s %s\n", member.Province, member.City)
//	    fmt.Printf("个性签名: %s\n", member.Signature)
//	    fmt.Printf("头像: %s\n", member.SmallHeadImgUrl)
//	}
//
// 注意:
//   - 需要同时提供群成员微信ID和群聊ID
//   - 返回的 ContactList 数组通常只包含一个元素
//   - 可以获取群成员在该群中的详细信息
//   - Sex 字段：1-男性，2-女性
func (a *RoomAPI) GetGroupMemberContact(wxid, roomID string) (*types.GetGroupMemberContactResponse, error) {
	req := types.GetGroupMemberContactRequest{
		Wxid:   wxid,
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_group_member_contact", req)
	if err != nil {
		return nil, fmt.Errorf("get group member contact: %w", err)
	}

	var resp types.GetGroupMemberContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// CreateChatRoom 创建群聊
//
// 创建一个新的群聊，并邀请指定的成员加入。
//
// 参数:
//   - wxids: 成员微信ID列表，用逗号分隔，例如 "wxid_3e9mll0g0fad21,wxid_8543785438012"
//
// 返回:
//   - *types.CreateChatRoomResponse: 创建的群聊信息，包含群ID和成员列表
//   - error: 创建失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.CreateChatRoom("wxid_3e9mll0g0fad21,wxid_8543785438012")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("群聊创建成功\n")
//	fmt.Printf("群ID: %s\n", resp.ChatRoomName.String)
//	fmt.Printf("成员数: %d\n", resp.MemberCount)
//	for _, member := range resp.MemberList {
//	    fmt.Printf("- %s (%s)\n", member.NickName.String, member.MemberName.String)
//	}
//
// 注意:
//   - 至少需要2个成员才能创建群聊
//   - wxids 参数中的微信ID必须是好友关系
//   - 创建成功后返回群ID（格式为 数字@chatroom）
//   - 如果创建失败，BaseResponse.Ret 会返回 -2
func (a *RoomAPI) CreateChatRoom(wxids string) (*types.CreateChatRoomResponse, error) {
	req := types.CreateChatRoomRequest{
		Wxids: wxids,
	}

	respBody, err := a.client.DoRequest("POST", "/api/creat_chat_room", req)
	if err != nil {
		return nil, fmt.Errorf("create chat room: %w", err)
	}

	var resp types.CreateChatRoomResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		// 提取错误消息
		errMsg := "unknown error"
		if strVal, ok := resp.BaseResponse.ErrMsg.(map[string]interface{}); ok {
			if str, ok := strVal["String"].(string); ok {
				errMsg = str
			}
		}
		return nil, fmt.Errorf("api error: ret=%d, msg=%s",
			resp.BaseResponse.Ret, errMsg)
	}

	return &resp, nil
}

// InviteMemberToChatRoom 邀请进入群聊
//
// 邀请一个或多个好友加入指定的群聊。
//
// 参数:
//   - wxidList: 成员微信ID列表，用逗号分隔，例如 "wxid_xxx,wxid_yyy"
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - error: 邀请失败时返回错误信息
//
// 示例:
//
//	err := client.Room.InviteMemberToChatRoom(
//	    "wxid_8543785438012",
//	    "45220347292@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("邀请成功")
//
//	// 邀请多个成员
//	err = client.Room.InviteMemberToChatRoom(
//	    "wxid_aaa,wxid_bbb,wxid_ccc",
//	    "45220347292@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("批量邀请成功")
//
// 注意:
//   - 被邀请的用户必须是好友关系
//   - 需要有邀请权限（群主或管理员，或群设置允许成员邀请）
//   - 可以一次邀请多个成员，用逗号分隔微信ID
//   - 群聊ID格式必须正确（数字@chatroom）
func (a *RoomAPI) InviteMemberToChatRoom(wxidList, roomID string) error {
	req := types.InviteMemberToChatRoomRequest{
		WxidList: wxidList,
		RoomID:   roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/invite_member_to_chat_room", req)
	if err != nil {
		return fmt.Errorf("invite member to chat room: %w", err)
	}

	var resp types.InviteMemberToChatRoomResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// AddMemberToChatRoom 添加群成员（40人以内）
//
// 直接添加一个或多个好友加入指定的群聊，无需对方同意。
// 适用于群成员数量在40人以内的群聊。
//
// 参数:
//   - wxidList: 成员微信ID列表，用逗号分隔，例如 "wxid_xxx,wxid_yyy"
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - error: 添加失败时返回错误信息
//
// 示例:
//
//	err := client.Room.AddMemberToChatRoom(
//	    "wxid_3e9mll0g0fad21",
//	    "45220347292@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("添加成功")
//
//	// 添加多个成员
//	err = client.Room.AddMemberToChatRoom(
//	    "wxid_aaa,wxid_bbb,wxid_ccc",
//	    "45220347292@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("批量添加成功")
//
// 注意:
//   - 仅适用于群成员数量在40人以内的群聊
//   - 被添加的用户必须是好友关系
//   - 直接添加，无需对方同意
//   - 超过40人的群聊请使用 InviteMemberToChatRoom（需要对方同意）
//   - 可以一次添加多个成员，用逗号分隔微信ID
//   - 需要有添加权限（群主或管理员，或群设置允许成员添加）
func (a *RoomAPI) AddMemberToChatRoom(wxidList, roomID string) error {
	req := types.AddMemberToChatRoomRequest{
		WxidList: wxidList,
		RoomID:   roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/add_member_to_chat_room", req)
	if err != nil {
		return fmt.Errorf("add member to chat room: %w", err)
	}

	var resp types.AddMemberToChatRoomResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// SetRoomAdmin 添加群管理
//
// 将指定的群成员设置为群管理员。
// 只有群主才能设置管理员。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//   - admin: 要设置为管理员的成员微信ID
//
// 返回:
//   - error: 设置失败时返回错误信息
//
// 示例:
//
//	err := client.Room.SetRoomAdmin(
//	    "49767299448@chatroom",
//	    "wxid_8543785438012",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("设置管理员成功")
//
// 注意:
//   - 只有群主才能设置管理员
//   - 被设置的成员必须已经在群内
//   - 一个群可以有多个管理员
//   - 管理员拥有部分群管理权限（如踢人、修改群公告等）
func (a *RoomAPI) SetRoomAdmin(roomID, admin string) error {
	req := types.SetRoomAdminRequest{
		RoomID: roomID,
		Admin:  admin,
	}

	respBody, err := a.client.DoRequest("POST", "/api/api/set_room_admin", req)
	if err != nil {
		return fmt.Errorf("set room admin: %w", err)
	}

	var resp types.SetRoomAdminResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// DelRoomAdmin 删除群管理
//
// 将指定的群管理员移除管理员身份。
// 只有群主才能删除管理员。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//   - admin: 要移除管理员身份的成员微信ID
//
// 返回:
//   - error: 删除失败时返回错误信息
//
// 示例:
//
//	err := client.Room.DelRoomAdmin(
//	    "49767299448@chatroom",
//	    "wxid_8543785438012",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("删除管理员成功")
//
// 注意:
//   - 只有群主才能删除管理员
//   - 被删除的成员必须是当前的管理员
//   - 删除后该成员变为普通群成员
//   - 不会将成员踢出群聊，只是移除管理员权限
func (a *RoomAPI) DelRoomAdmin(roomID, admin string) error {
	req := types.DelRoomAdminRequest{
		RoomID: roomID,
		Admin:  admin,
	}

	respBody, err := a.client.DoRequest("POST", "/api/api/del_room_admin", req)
	if err != nil {
		return fmt.Errorf("del room admin: %w", err)
	}

	var resp types.DelRoomAdminResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// SetRoomAnnouncement 设置群公告
//
// 设置或修改群聊的公告内容。
// 群主和管理员可以设置群公告。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//   - announcement: 群公告内容
//
// 返回:
//   - error: 设置失败时返回错误信息
//
// 示例:
//
//	err := client.Room.SetRoomAnnouncement(
//	    "51687237616@chatroom",
//	    "通知一下 下次别用之前的群公告版本了",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("设置群公告成功")
//
// 注意:
//   - 群主和管理员可以设置群公告
//   - 公告内容会通知所有群成员
//   - 可以设置空字符串来清空群公告
//   - 公告内容长度有限制
func (a *RoomAPI) SetRoomAnnouncement(roomID, announcement string) error {
	req := types.SetRoomAnnouncementRequest{
		RoomID:       roomID,
		Announcement: announcement,
	}

	respBody, err := a.client.DoRequest("POST", "/api/set_room_announcement_pb", req)
	if err != nil {
		return fmt.Errorf("set room announcement: %w", err)
	}

	var resp types.SetRoomAnnouncementResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// DelMemberFromChatRoom 踢出群成员
//
// 将一个或多个成员从群聊中移除。
// 群主和管理员可以踢出群成员。
//
// 参数:
//   - wxidList: 成员微信ID列表，用逗号分隔，例如 "wxid_xxx,wxid_yyy"
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - error: 踢出失败时返回错误信息
//
// 示例:
//
//	err := client.Room.DelMemberFromChatRoom(
//	    "wxid_8543785438012",
//	    "49767299448@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("踢出成员成功")
//
//	// 踢出多个成员
//	err = client.Room.DelMemberFromChatRoom(
//	    "wxid_aaa,wxid_bbb,wxid_ccc",
//	    "49767299448@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("批量踢出成功")
//
// 注意:
//   - 群主和管理员可以踢出群成员
//   - 管理员不能踢出群主和其他管理员
//   - 群主可以踢出任何成员（包括管理员）
//   - 可以一次踢出多个成员，用逗号分隔微信ID
//   - 被踢出的成员会收到通知
func (a *RoomAPI) DelMemberFromChatRoom(wxidList, roomID string) error {
	req := types.DelMemberFromChatRoomRequest{
		WxidList: wxidList,
		RoomID:   roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/del_member_from_chat_room", req)
	if err != nil {
		return fmt.Errorf("del member from chat room: %w", err)
	}

	var resp types.DelMemberFromChatRoomResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return nil
}

// QuitAndDelChatRoom 退出群聊
//
// 退出指定的群聊。退出后将无法再接收该群的消息。
//
// 参数:
//   - roomID: 群聊ID，例如 "xxxxxxxxxxxx@chatroom"
//
// 返回:
//   - error: 退出失败时返回错误信息
//
// 示例:
//
//	err := client.Room.QuitAndDelChatRoom("xxxxxxxxxxxx@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已退出群聊")
//
// 注意:
//   - 退出后将无法再接收该群的消息
//   - 如果是群主，需要先转让群主身份才能退出
//   - 退出操作不可撤销
//   - 退出后可以被重新邀请进群
func (a *RoomAPI) QuitAndDelChatRoom(roomID string) error {
	req := types.QuitAndDelChatRoomRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/quit_and_del_chat_room", req)
	if err != nil {
		return fmt.Errorf("quit and del chat room: %w", err)
	}

	var resp types.QuitAndDelChatRoomResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// EnterRoom 同意群聊邀请
//
// 通过群聊邀请链接加入群聊。
//
// 参数:
//   - url: 群聊邀请链接，必须是用 a8key 转换后的 URL
//
// 返回:
//   - error: 加入失败时返回错误信息
//
// 示例:
//
//	// 使用 a8key 转换后的 URL
//	url := "https://weixin.qq.com/g/AQYAAHTnjQ-tHLCRwz7OEsG40PUGTTWtGIYAaE9-09DtKqsJ-_icjSkR72_N_D2P"
//	err := client.Room.EnterRoom(url)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已加入群聊")
//
// 注意:
//   - URL 必须是用 a8key 转换后的链接
//   - 原始的群聊邀请链接需要先通过 a8key 接口转换
//   - 转换后的链接格式通常为 https://weixin.qq.com/g/...
//   - 邀请链接可能有时效性，过期后无法加入
func (a *RoomAPI) EnterRoom(url string) error {
	req := types.EnterRoomRequest{
		URL: url,
	}

	respBody, err := a.client.DoRequest("POST", "/api/enter_room", req)
	if err != nil {
		return fmt.Errorf("enter room: %w", err)
	}

	var resp types.EnterRoomResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// GetChatroomInfo 获取群详情
//
// 获取指定群聊的详细信息，包括群公告、群状态等。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "45220347292@chatroom"
//
// 返回:
//   - *types.GetChatroomInfoResponse: 群详情信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.GetChatroomInfo("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if resp.ErrCode == 1 {
//	    fmt.Printf("群公告: %s\n", resp.Data.Announcement)
//	    fmt.Printf("公告编辑者: %s\n", resp.Data.AnnouncementEditor)
//	    fmt.Printf("公告发布时间: %d\n", resp.Data.AnnouncementPublishTime)
//	    fmt.Printf("群状态: %d\n", resp.Data.ChatRoomStatus)
//	} else {
//	    fmt.Println("群信息获取异常")
//	}
//
// 注意:
//   - 返回的响应格式根据成功或失败有所不同
//   - 成功时 ErrCode 为 1，Data 字段包含详细信息
//   - 失败时 BaseResponse.Ret 为 -2 或其他错误码
//   - Announcement 字段包含群公告文本内容
//   - XmlAnnouncement 字段包含群公告的完整XML格式数据
func (a *RoomAPI) GetChatroomInfo(roomID string) (*types.GetChatroomInfoResponse, error) {
	req := types.GetChatroomInfoRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_chatroom_info", req)
	if err != nil {
		return nil, fmt.Errorf("get chatroom info: %w", err)
	}

	var resp types.GetChatroomInfoResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 检查是否是异常响应（BaseResponse.Ret 存在且不为0）
	if resp.BaseResponse.Ret != 0 && resp.ErrCode == 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	// 正常响应检查 ErrCode（成功时为1）
	if resp.ErrCode != 0 && resp.ErrCode != 1 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return &resp, nil
}

// RemovChatroomToContact 移除群聊通讯录
//
// 将指定群聊从通讯录中移除。
// 移除后群聊不会显示在通讯录列表中，但聊天记录仍然保留。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - error: 移除失败时返回错误信息
//
// 示例:
//
//	err := client.Room.RemovChatroomToContact("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群聊已从通讯录移除")
//
// 注意:
//   - 移除后群聊不会显示在通讯录列表中
//   - 聊天记录仍然保留
//   - 如果有新消息，群聊会重新出现在聊天列表中
//   - 此操作不同于退出群聊，只是从通讯录隐藏
func (a *RoomAPI) RemovChatroomToContact(roomID string) error {
	req := types.RemovChatroomToContactRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/remov_chatroom_to_contact", req)
	if err != nil {
		return fmt.Errorf("remov chatroom to contact: %w", err)
	}

	var resp types.RemovChatroomToContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// SaveChatroomToContact 保存群聊到通讯录
//
// 将指定群聊保存到通讯录中。
// 保存后群聊会显示在通讯录的群聊列表中。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - error: 保存失败时返回错误信息
//
// 示例:
//
//	err := client.Room.SaveChatroomToContact("45220347292@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群聊已保存到通讯录")
//
// 注意:
//   - 保存后群聊会显示在通讯录的群聊列表中
//   - 此操作与 RemovChatroomToContact 相反
//   - 通常用于恢复之前从通讯录移除的群聊
func (a *RoomAPI) SaveChatroomToContact(roomID string) error {
	req := types.SaveChatroomToContactRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/save_chatroom_to_contact", req)
	if err != nil {
		return fmt.Errorf("save chatroom to contact: %w", err)
	}

	var resp types.SaveChatroomToContactResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// GetGroupMemberInfo 获取群成员数据
//
// 获取指定群聊中某个成员的详细信息。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "49767299448@chatroom"
//   - memberID: 成员微信ID，例如 "wxid_bktzp6cv7wxe12"
//
// 返回:
//   - *types.GetGroupMemberInfoResponse: 群成员详细信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.GetGroupMemberInfo("49767299448@chatroom", "wxid_bktzp6cv7wxe12")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("成员信息: %+v\n", resp.Data)
//
// 注意:
//   - roomID 必须是有效的群聊ID
//   - memberID 必须是群内成员的微信ID
//   - 返回的数据结构可能因接口版本而异
func (a *RoomAPI) GetGroupMemberInfo(roomID, memberID string) (*types.GetGroupMemberInfoResponse, error) {
	req := types.GetGroupMemberInfoRequest{
		RoomID:   roomID,
		MemberID: memberID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_group_memeber_info", req)
	if err != nil {
		return nil, fmt.Errorf("get group member info: %w", err)
	}

	var resp types.GetGroupMemberInfoResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp, nil
}

// GetChatroomList 获取群聊列表
//
// 获取当前账号的所有群聊列表
//
// 返回:
//   - []types.ChatroomItem: 群聊列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	chatrooms, err := client.Room.GetChatroomList()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个群聊\n", len(chatrooms))
//	for _, room := range chatrooms {
//	    fmt.Printf("- %s (%s)\n", room.NickName, room.Username)
//	    if room.Remark != "" {
//	        fmt.Printf("  备注: %s\n", room.Remark)
//	    }
//	}
//
// 注意:
//   - 返回的列表包含所有群聊的基本信息
//   - 每个群聊包含昵称、群ID、头像URL等信息
//   - Username 字段是群聊ID，格式为 "数字@chatroom"
func (a *RoomAPI) GetChatroomList() ([]types.ChatroomItem, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_chatroom_list", nil)
	if err != nil {
		return nil, fmt.Errorf("get chatroom list: %w", err)
	}

	var resp types.GetChatroomListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// ModChatroomTopic 修改群名称
//
// 修改指定群聊的名称
//
// 参数:
//   - wxid: 群聊ID，格式为 "数字@chatroom"，例如 "45220347292@chatroom"
//   - topic: 新的群名称
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	err := client.Room.ModChatroomTopic(
//	    "45220347292@chatroom",
//	    "新的群名称",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群名称修改成功")
//
// 注意:
//   - 需要有修改群名称的权限（群主或管理员）
//   - 群名称长度有限制
//   - 修改后所有群成员都会看到新名称
func (a *RoomAPI) ModChatroomTopic(wxid, topic string) error {
	req := types.ModChatroomTopicRequest{
		Wxid:  wxid,
		Topic: topic,
	}

	respBody, err := a.client.DoRequest("POST", "/api/mod_chatroom_topic", req)
	if err != nil {
		return fmt.Errorf("mod chatroom topic: %w", err)
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

// BatchGetRoomContact 获取所有群的资料
//
// 从网络获取所有群聊的详细资料并保存到本地缓存
// 注意：这是一个长耗时接口，会批量获取所有群的资料
//
// 返回:
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	err := client.Room.BatchGetRoomContact()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("所有群资料已更新")
//
// 注意:
//   - 这是一个长耗时接口，可能需要较长时间完成
//   - 会从服务器获取所有群聊的最新资料
//   - 获取的资料会保存到本地缓存
//   - 建议在后台异步执行，避免阻塞主流程
//   - 适合在启动时或定期更新群资料时使用
func (a *RoomAPI) BatchGetRoomContact() error {
	respBody, err := a.client.DoRequest("POST", "/api/batch_getroom_contact", nil)
	if err != nil {
		return fmt.Errorf("batch get room contact: %w", err)
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

// BatchGetRoomCache 获取所有群资料(缓存)
//
// 从本地缓存快速获取所有群聊的详细资料
// 注意：这是一个速度极快的接口，直接读取本地缓存
//
// 返回:
//   - interface{}: 群资料列表（原始数据）
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	data, err := client.Room.BatchGetRoomCache()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("获取群资料成功\n")
//
// 注意:
//   - 这是一个速度极快的接口，直接读取本地缓存
//   - 返回的数据可能不是最新的，需要先调用 BatchGetRoomContact 更新
//   - 包含群名称、群主、成员数量等详细信息
//   - 适合频繁查询群资料的场景
//   - 返回的数据结构复杂，建议根据实际需要解析
func (a *RoomAPI) BatchGetRoomCache() (interface{}, error) {
	respBody, err := a.client.DoRequest("POST", "/api/batch_getroom_cache", nil)
	if err != nil {
		return nil, fmt.Errorf("batch get room cache: %w", err)
	}

	var resp types.BatchGetRoomCacheResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d", resp.Code)
	}

	return resp.Data, nil
}

// InitRooms 初始化群聊
//
// 初始化所有群聊的成员资料缓存
// 注意：如果遇到群里成员发送消息但无成员资料，可以调用此接口进行缓存
//
// 返回:
//   - error: 初始化失败时返回错误信息
//
// 示例:
//
//	err := client.Room.InitRooms()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群聊初始化成功")
//
// 注意:
//   - 此接口用于初始化群聊成员资料缓存
//   - 适用于事件处理场景，当收到群消息但缺少成员资料时调用
//   - 会缓存所有群聊的成员信息
//   - 建议在启动时或发现成员资料缺失时调用
func (a *RoomAPI) InitRooms() error {
	respBody, err := a.client.DoRequest("POST", "/api/init_rooms", nil)
	if err != nil {
		return fmt.Errorf("init rooms: %w", err)
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

// GetMemberNick 获取群成员简要信息
//
// 获取指定群成员的昵称和基本信息
//
// 参数:
//   - wxid: 群成员微信ID
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//
// 返回:
//   - *types.MemberNickData: 群成员简要信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	memberInfo, err := client.Room.GetMemberNick(
//	    "wxid_8zggbw1yo5ib22",
//	    "18402658081@chatroom",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("成员昵称: %s\n", memberInfo.NickName)
//	fmt.Printf("邀请人: %s\n", memberInfo.InviterUserName)
//
// 注意:
//   - 用于快速获取群成员的昵称和头像信息
//   - 包含邀请人信息和加群场景
//   - 适合在处理群消息时获取发送者信息
func (a *RoomAPI) GetMemberNick(wxid, roomID string) (*types.MemberNickData, error) {
	req := types.GetMemberNickRequest{
		Wxid:   wxid,
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_member_nick", req)
	if err != nil {
		return nil, fmt.Errorf("get member nick: %w", err)
	}

	var resp types.GetMemberNickResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.ErrCode != 1 && resp.ErrCode != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return &resp.Data, nil
}

// TransferChatroomOwner 转让群主
//
// 将群主权限转让给指定的群成员
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"
//   - toWxid: 新群主的微信ID
//
// 返回:
//   - error: 转让失败时返回错误信息
//
// 示例:
//
//	err := client.Room.TransferChatroomOwner(
//	    "18402658081@chatroom",
//	    "wxid_8zggbw1yo5ib22",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群主转让成功")
//
// 注意:
//   - 只有当前群主才能执行此操作
//   - 新群主必须是群成员
//   - 转让后原群主将失去群主权限
//   - 此操作不可撤销
func (a *RoomAPI) TransferChatroomOwner(roomID, toWxid string) error {
	req := types.TransferChatroomOwnerRequest{
		RoomID: roomID,
		ToWxid: toWxid,
	}

	respBody, err := a.client.DoRequest("POST", "/api/transferchatroomowner", req)
	if err != nil {
		return fmt.Errorf("transfer chatroom owner: %w", err)
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

// GetGroupMemberBySQL 获取群成员数据(简要不包含头像)
//
// 通过 SQL 查询获取群成员的简要信息，不包含头像等详细资料。
// 此接口返回数据更轻量，适合只需要基本成员信息的场景。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "18402658081@chatroom"
//
// 返回:
//   - []types.SimpleMember: 简要成员信息列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	members, err := client.Room.GetGroupMemberBySQL("18402658081@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, member := range members {
//	    fmt.Printf("- %s (%s) 邀请人: %s\n",
//	        member.DisplayName, member.UserName, member.InviterUserName)
//	}
//
// 注意:
//   - 此接口不返回头像URL，数据更轻量
//   - 适合只需要成员列表和基本信息的场景
//   - 返回的 DisplayName 是群昵称，如果没有设置则为空
func (a *RoomAPI) GetGroupMemberBySQL(roomID string) ([]types.SimpleMember, error) {
	req := types.GetGroupMemberBySQLRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_groupmember_bysql", req)
	if err != nil {
		return nil, fmt.Errorf("get group member by sql: %w", err)
	}

	var resp types.GetGroupMemberBySQLResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return resp.Members, nil
}

// ModChatRoomSelfNickName 修改自己在群里的昵称
//
// 修改自己在指定群聊中显示的昵称（群昵称）。
// 群昵称只在该群内生效，不影响其他群或好友看到的昵称。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "49767299448@chatroom"
//   - nickName: 新的群昵称
//
// 返回:
//   - error: 修改失败时返回错误信息
//
// 示例:
//
//	err := client.Room.ModChatRoomSelfNickName(
//	    "49767299448@chatroom",
//	    "綦奕泽",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("群昵称修改成功")
//
// 注意:
//   - 群昵称只在该群内生效
//   - 不同群可以设置不同的群昵称
//   - 群昵称长度有限制
//   - 修改后群内所有成员都会看到新昵称
func (a *RoomAPI) ModChatRoomSelfNickName(roomID, nickName string) error {
	req := types.ModChatRoomSelfNickNameRequest{
		RoomID:   roomID,
		NickName: nickName,
	}

	respBody, err := a.client.DoRequest("POST", "/api/mod_chat_room_self_nick_name", req)
	if err != nil {
		return fmt.Errorf("mod chat room self nick name: %w", err)
	}

	var resp types.ModChatRoomSelfNickNameResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// GetRoomWxids 获取所有群wxid
//
// 从网络获取所有群聊的微信ID列表。
// 注意：这是一个长耗时接口，可能需要较长时间完成。
//
// 返回:
//   - []string: 群聊微信ID列表（格式为 "数字@chatroom"）
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	wxids, err := client.Room.GetRoomWxids()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个群聊\n", len(wxids))
//	for i, wxid := range wxids {
//	    fmt.Printf("%d. %s\n", i+1, wxid)
//	}
//
// 注意:
//   - 这是一个长耗时接口，可能需要较长时间完成
//   - 从网络获取最新的群聊列表
//   - 建议在后台异步执行，避免阻塞主流程
//   - 适合在启动时或定期更新群聊列表时使用
//   - 只返回群wxid，不包含其他详细信息
//   - 返回的wxid格式为 "数字@chatroom"
func (a *RoomAPI) GetRoomWxids() ([]string, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_room_wxids", nil)
	if err != nil {
		return nil, fmt.Errorf("get room wxids: %w", err)
	}

	var resp types.GetRoomWxidsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Wxids, nil
}

// GetRoomsInfo 获取群成员数量和群昵称
//
// 获取所有群聊的基本信息，包括群昵称和成员数量。
// 这是一个轻量级接口，快速获取群聊概览信息。
//
// 返回:
//   - []types.RoomInfo: 群信息列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	rooms, err := client.Room.GetRoomsInfo()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个群聊\n", len(rooms))
//	for _, room := range rooms {
//	    fmt.Printf("群名: %s\n", room.NickName)
//	    fmt.Printf("群ID: %s\n", room.RoomID)
//	    fmt.Printf("成员数: %d\n", room.MemberCount)
//	    fmt.Println("---")
//	}
//
// 注意:
//   - 这是一个轻量级接口，只返回基本信息
//   - 包含群昵称和成员数量
//   - 不包含成员详细列表
//   - 适合快速获取群聊概览的场景
func (a *RoomAPI) GetRoomsInfo() ([]types.RoomInfo, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_rooms_info", nil)
	if err != nil {
		return nil, fmt.Errorf("get rooms info: %w", err)
	}

	var rooms []types.RoomInfo
	if err := json.Unmarshal(respBody, &rooms); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return rooms, nil
}

// GetChatroomAnnouncement 获取群公告
//
// 获取指定群聊的公告内容。
//
// 参数:
//   - roomID: 群聊ID，格式为 "数字@chatroom"，例如 "38879414299@chatroom"
//
// 返回:
//   - *types.GetChatroomAnnouncementResponse: 群公告信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.Room.GetChatroomAnnouncement("38879414299@chatroom")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if resp.Announcement != "" {
//	    fmt.Printf("群公告: %s\n", resp.Announcement)
//	    if resp.Editor != "" {
//	        fmt.Printf("编辑者: %s\n", resp.Editor)
//	    }
//	    if resp.PublishTime > 0 {
//	        fmt.Printf("发布时间: %d\n", resp.PublishTime)
//	    }
//	} else {
//	    fmt.Println("该群暂无公告")
//	}
//
// 注意:
//   - 如果群没有设置公告，Announcement 字段为空
//   - PublishTime 是Unix时间戳
//   - Editor 是公告编辑者的微信ID
func (a *RoomAPI) GetChatroomAnnouncement(roomID string) (*types.GetChatroomAnnouncementResponse, error) {
	req := types.GetChatroomAnnouncementRequest{
		RoomID: roomID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_chatroom_announcement", req)
	if err != nil {
		return nil, fmt.Errorf("get chatroom announcement: %w", err)
	}

	var resp types.GetChatroomAnnouncementResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return &resp, nil
}

// GetAllRoomDetail 获取所有群聊的详细信息
//
// 获取所有群聊的完整信息，包括成员列表、群头像、昵称等数据。
// 这是一个综合性接口，返回所有群的详细资料。
//
// 返回:
//   - []types.RoomDetailInfo: 群详细信息列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	rooms, err := client.Room.GetAllRoomDetail()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个群聊\n", len(rooms))
//	for _, room := range rooms {
//	    fmt.Printf("\n群名: %s\n", room.NickName)
//	    fmt.Printf("群ID: %s\n", room.RoomID)
//	    fmt.Printf("群主: %s\n", room.Owner)
//	    fmt.Printf("成员数: %d\n", room.MemberCount)
//	    if room.Announcement != "" {
//	        fmt.Printf("群公告: %s\n", room.Announcement)
//	    }
//	    fmt.Printf("头像: %s\n", room.SmallHeadImgUrl)
//	    fmt.Println("成员列表:")
//	    for i, member := range room.Members {
//	        displayName := member.DisplayName
//	        if displayName == "" {
//	            displayName = member.NickName
//	        }
//	        fmt.Printf("  %d. %s (%s)\n", i+1, displayName, member.Wxid)
//	    }
//	    fmt.Println("---")
//	}
//
// 注意:
//   - 这是一个综合性接口，返回所有群的详细信息
//   - 包含群昵称、头像、成员列表、群主、公告等完整数据
//   - 成员列表包含每个成员的微信ID、昵称和群昵称
//   - 数据量较大，建议在需要完整群信息时使用
func (a *RoomAPI) GetAllRoomDetail() ([]types.RoomDetailInfo, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_all_room_detail", nil)
	if err != nil {
		return nil, fmt.Errorf("get all room detail: %w", err)
	}

	var resp types.GetAllRoomDetailResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Rooms, nil
}
