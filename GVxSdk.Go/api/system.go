package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVXkernel/GVxSdk.Go/types"
)

// SystemAPI 系统相关接口
type SystemAPI struct {
	client Client
}

// NewSystemAPI 创建系统API实例
func NewSystemAPI(client Client) *SystemAPI {
	return &SystemAPI{client: client}
}

// GetWxBasePath 获取微信缓存目录
//
// 获取当前微信的缓存目录路径
//
// 返回:
//   - string: 微信缓存目录路径
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	path, err := client.System.GetWxBasePath()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("微信缓存目录: %s\n", path)
func (a *SystemAPI) GetWxBasePath() (string, error) {
	respBody, err := a.client.DoRequest("POST", "/api/getwxbasepath", nil)
	if err != nil {
		return "", fmt.Errorf("get wx base path: %w", err)
	}

	var resp types.GetWxBasePathResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return "", fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}

// AutoLogin 自动登录
//
// 触发微信自动登录
//
// 返回:
//   - error: 登录失败时返回错误信息
//
// 示例:
//
//	err := client.System.AutoLogin()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("自动登录成功")
//
// 注意:
//   - 需要微信中间件支持自动登录功能
//   - 通常用于重启后自动恢复登录状态
func (a *SystemAPI) AutoLogin() error {
	respBody, err := a.client.DoRequest("POST", "/api/auto_login", nil)
	if err != nil {
		return fmt.Errorf("auto login: %w", err)
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

// DecryptDB 解密数据库
//
// 解密微信加密的数据库文件
//
// 参数:
//   - dpPath: 加密数据库文件路径
//   - outPath: 解密后输出文件路径
//   - key: 解密密钥（32位十六进制字符串）
//
// 返回:
//   - error: 解密失败时返回错误信息
//
// 示例:
//
//	err := client.System.DecryptDB(
//	    "E:\\xwechat_files\\wxid_xxx\\db_storage\\contact\\contact.db",
//	    "E:\\xwechat_files\\wxid_xxx\\db_storage\\contact\\contact2.db",
//	    "910e9301f30a4250aefaf9c51fb8e1646103c228ae1b4cc7899a8456762cdb16",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("数据库解密成功")
//
// 注意:
//   - 密钥必须是正确的32位十六进制字符串
//   - 输出路径的目录必须存在
//   - 如果输出文件已存在，将被覆盖
func (a *SystemAPI) DecryptDB(dpPath, outPath, key string) error {
	req := types.DecryptDBRequest{
		DpPath:  dpPath,
		OutPath: outPath,
		Key:     key,
	}

	respBody, err := a.client.DoRequest("POST", "/api/decrypt_db", req)
	if err != nil {
		return fmt.Errorf("decrypt db: %w", err)
	}

	var resp types.DecryptDBResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// RefreshQRCode 获取登录二维码
//
// 获取微信登录二维码，用于扫码登录。
// 返回的二维码数据可以保存为图片或直接展示给用户扫描。
//
// 返回:
//   - *types.RefreshQRCodeResponse: 二维码数据和相关信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.System.RefreshQRCode()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("UUID: %s\n", resp.Uuid)
//	fmt.Printf("二维码长度: %d\n", resp.QRCode.ILen)
//	fmt.Printf("过期时间: %d秒\n", resp.ExpiredTime)
//	fmt.Printf("检查时间: %d秒\n", resp.CheckTime)
//	// resp.QRCode.Buffer 包含二维码图片数据
//	// 可以保存为图片文件或直接展示
//
// 注意:
//   - 二维码有过期时间（ExpiredTime），过期后需要重新获取
//   - CheckTime 表示建议的检查间隔时间
//   - UUID 用于标识此次登录会话
//   - NotifyKey 用于接收登录通知
func (a *SystemAPI) RefreshQRCode() (*types.RefreshQRCodeResponse, error) {
	respBody, err := a.client.DoRequest("POST", "/api/reflash_qrcode", nil)
	if err != nil {
		return nil, fmt.Errorf("refresh qrcode: %w", err)
	}

	var resp types.RefreshQRCodeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// GetA8Key 获取A8key
//
// 获取A8key，用于各种场景，如群聊邀请等。
// A8key是微信内部使用的一种授权密钥，用于访问特定资源。
//
// 参数:
//   - url: URL地址，例如群聊邀请链接
//   - urlType: URL类型，通常为 "0"
//   - scene: 场景类型，通常为 "0"
//
// 返回:
//   - *types.GetA8KeyResponse: A8key响应信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	url := "https://support.weixin.qq.com/cgi-bin/mmsupport-bin/addchatroombyinvite?ticket=AwfZ4kSJ9P2FbmFK6LPrpg%3D%3D"
//	resp, err := client.System.GetA8Key(url, "0", "0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("A8key获取成功")
//
// 注意:
//   - A8key有多种使用场景，示例展示的是群聊邀请场景
//   - urlType 和 scene 参数根据具体场景可能有不同的值
//   - 不同场景的 URL 格式可能不同
func (a *SystemAPI) GetA8Key(url, urlType, scene string) (*types.GetA8KeyResponse, error) {
	req := types.GetA8KeyRequest{
		URL:     url,
		URLType: urlType,
		Scene:   scene,
	}

	respBody, err := a.client.DoRequest("POST", "/api/get_a8key", req)
	if err != nil {
		return nil, fmt.Errorf("get a8key: %w", err)
	}

	var resp types.GetA8KeyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	return &resp, nil
}

// JsLogin 获取小程序code
//
// 获取指定小程序的登录code，用于小程序授权登录。
// 返回的code可以用于调用小程序的后端接口进行用户身份验证。
//
// 参数:
//   - waID: 小程序的AppID，例如 "wxfec93dd30abcc9ad"
//
// 返回:
//   - *types.JsLoginResponse: 包含小程序code的响应信息
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.System.JsLogin("wxfec93dd30abcc9ad")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("小程序code: %s\n", resp.Code)
//	fmt.Printf("状态: %s\n", resp.State)
//	if resp.JsapiBaseresponse.Errcode == 0 {
//	    fmt.Println("获取成功")
//	}
//
// 注意:
//   - code有时效性，需要及时使用
//   - 需要检查 JsapiBaseresponse.Errcode 是否为 0 来判断是否成功
//   - 不同小程序的AppID不同，需要使用正确的AppID
func (a *SystemAPI) JsLogin(waID string) (*types.JsLoginResponse, error) {
	req := types.JsLoginRequest{
		WaID: waID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/js_login", req)
	if err != nil {
		return nil, fmt.Errorf("js login: %w", err)
	}

	var resp types.JsLoginResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.BaseResponse.Ret != 0 {
		return nil, fmt.Errorf("api error: ret=%d, msg=%v",
			resp.BaseResponse.Ret, resp.BaseResponse.ErrMsg)
	}

	if resp.JsapiBaseresponse.Errcode != 0 {
		return nil, fmt.Errorf("jsapi error: code=%d, msg=%s",
			resp.JsapiBaseresponse.Errcode, resp.JsapiBaseresponse.Errmsg)
	}

	return &resp, nil
}

// GetDBHandle 获取数据库句柄
//
// 获取微信所有数据库的句柄信息，包括数据库名称和对应的句柄值。
// 这些句柄可以用于后续的数据库操作。
//
// 返回:
//   - *types.GetDBHandleResponse: 包含所有数据库句柄信息的列表
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	resp, err := client.System.GetDBHandle()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("共有 %d 个数据库\n", len(resp.Data))
//	for _, db := range resp.Data {
//	    fmt.Printf("数据库: %s, 句柄: %d\n", db.Name, db.Handle)
//	}
//
// 注意:
//   - 返回的数据库包括：message_0.db（消息）、contact.db（联系人）、sns.db（朋友圈）等
//   - 句柄值是一个大整数，用于标识特定的数据库
//   - 这些句柄可以用于执行SQL查询等数据库操作
func (a *SystemAPI) GetDBHandle() (*types.GetDBHandleResponse, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_db_handle", nil)
	if err != nil {
		return nil, fmt.Errorf("get db handle: %w", err)
	}

	var resp types.GetDBHandleResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// Sqlite3Exec 执行数据库查询
//
// 对指定的微信数据库执行SQL查询语句。
// 可以查询联系人、消息、群聊等各种数据。
//
// 参数:
//   - dbName: 数据库名称，例如 "contact.db"、"message_0.db"、"sns.db" 等
//   - sqlFmt: SQL查询语句，支持标准的SQLite语法
//
// 返回:
//   - types.Sqlite3ExecResponse: 查询结果，是一个数组，每个元素是一行数据（map格式）
//   - error: 查询失败时返回错误信息
//
// 示例:
//
//	// 查询表结构
//	resp, err := client.System.Sqlite3Exec("contact.db", "select * from sqlite_master")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 查询群聊信息
//	sql := `SELECT
//	    cr.username AS room_wxid,
//	    cr.owner AS manager_wxid,
//	    c_room.nick_name AS nickname,
//	    COUNT(cm.member_id) AS total_member
//	FROM chat_room cr
//	LEFT JOIN contact c_room ON c_room.UserName = cr.username
//	LEFT JOIN chatroom_member cm ON cm.room_id = cr.id
//	GROUP BY cr.id`
//	resp, err = client.System.Sqlite3Exec("contact.db", sql)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, row := range resp {
//	    fmt.Printf("群ID: %v, 群主: %v, 昵称: %v, 成员数: %v\n",
//	        row["room_wxid"], row["manager_wxid"], row["nickname"], row["total_member"])
//	}
//
// 注意:
//   - 查询前建议先查询表结构：select * from sqlite_master
//   - 带二进制数据的字段可能会被忽略
//   - 支持复杂的SQL查询，包括JOIN、GROUP BY等
//   - 常用数据库：contact.db（联系人）、message_0.db（消息）、sns.db（朋友圈）
func (a *SystemAPI) Sqlite3Exec(dbName, sqlFmt string) (types.Sqlite3ExecResponse, error) {
	req := types.Sqlite3ExecRequest{
		DBName: dbName,
		SqlFmt: sqlFmt,
	}

	respBody, err := a.client.DoRequest("POST", "/api/sqlite3_exec", req)
	if err != nil {
		return nil, fmt.Errorf("sqlite3 exec: %w", err)
	}

	var resp types.Sqlite3ExecResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return resp, nil
}

// GetConfigPath 获取配置文件保存目录
//
// 获取微信工具的配置文件保存目录路径。
//
// 返回:
//   - string: 配置文件保存目录路径
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	configPath, err := client.System.GetConfigPath()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("配置文件目录: %s\n", configPath)
//
// 注意:
//   - 返回的路径通常类似 "C:\Users\Admin\AppData\Roaming\WechatTools\config.ini"
//   - 路径格式取决于操作系统
//   - 可用于读取或修改配置文件
func (a *SystemAPI) GetConfigPath() (string, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_config_path", nil)
	if err != nil {
		return "", fmt.Errorf("get config path: %w", err)
	}

	var resp types.GetConfigPathResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return "", fmt.Errorf("api error: code=%d", resp.Code)
	}

	return resp.ConfigPath, nil
}

// AntiRevoke 防撤回
//
// 开启或关闭消息防撤回功能。
// 开启后，对方撤回的消息仍然可以看到。
//
// 参数:
//   - enable: true 表示开启防撤回，false 表示关闭防撤回
//
// 返回:
//   - error: 设置失败时返回错误信息
//
// 示例:
//
//	// 开启防撤回
//	err := client.System.AntiRevoke(true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("防撤回已开启")
//
//	// 关闭防撤回
//	err = client.System.AntiRevoke(false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("防撤回已关闭")
//
// 注意:
//   - 开启后，对方撤回的消息仍然可以看到
//   - 防撤回功能可能影响消息同步
//   - 建议根据实际需求开启或关闭
func (a *SystemAPI) AntiRevoke(enable bool) error {
	switchValue := "false"
	if enable {
		switchValue = "true"
	}

	req := types.AntiRevokeRequest{
		Switch: switchValue,
	}

	respBody, err := a.client.DoRequest("POST", "/api/anti_revoke", req)
	if err != nil {
		return fmt.Errorf("anti revoke: %w", err)
	}

	var resp types.AntiRevokeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// WechatInit 微信初始化好友列表和群列表
//
// 初始化微信的好友列表和群列表数据。
// 建议在登录后首次使用时调用，用于加载和缓存联系人数据。
//
// 返回:
//   - error: 初始化失败时返回错误信息
//
// 示例:
//
//	err := client.System.WechatInit()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("微信初始化成功")
//
// 注意:
//   - 建议在登录后首次使用时调用
//   - 初始化会加载所有好友和群聊数据
//   - 这是一个长耗时操作，具体时间取决于好友和群聊数量
//   - 初始化后使用 GetContactFast 等接口查询速度会更快
//   - 与 UpdateAllFriend 接口配合使用效果更佳
func (a *SystemAPI) WechatInit() error {
	respBody, err := a.client.DoRequest("POST", "/api/wechat_init", nil)
	if err != nil {
		return fmt.Errorf("wechat init: %w", err)
	}

	var resp types.WechatInitResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// BackupDatabase 数据库备份
//
// 备份微信数据库文件到指定目录。
//
// 参数:
//   - outputDir: 输出目录，例如 "C:\\Users\\Admin\\AppData\\Roaming\\WechatTools\\wxid_8543785438012_databasebackup"
//   - name: 数据库文件名，例如 "contact.db"
//
// 返回:
//   - error: 备份失败时返回错误信息
//
// 示例:
//
//	err := client.System.BackupDatabase(
//	    "C:\\Users\\Admin\\AppData\\Roaming\\WechatTools\\wxid_8543785438012_databasebackup",
//	    "contact.db",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("数据库备份成功")
//
// 注意:
//   - outputDir 必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 输出目录必须存在，否则备份会失败
//   - name 参数是数据库文件名，常见的有：
//   - contact.db (联系人数据库)
//   - msg.db (消息数据库)
//   - emotion.db (表情数据库)
//   - 备份的是加密数据库，需要使用 DecryptDB 解密后才能查看
func (a *SystemAPI) BackupDatabase(outputDir, name string) error {
	req := types.BackupDatabaseRequest{
		OutputDir: outputDir,
		Name:      name,
	}

	respBody, err := a.client.DoRequest("POST", "/api/backup_database", req)
	if err != nil {
		return fmt.Errorf("backup database: %w", err)
	}

	var resp types.BackupDatabaseResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// CheckLogin 获取登录状态
//
// 检查当前微信是否已登录
//
// 返回:
//   - bool: true-已登录 false-未登录
//   - error: 检查失败时返回错误信息
//
// 示例:
//
//	isLoggedIn, err := client.System.CheckLogin()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if isLoggedIn {
//	    fmt.Println("微信已登录")
//	} else {
//	    fmt.Println("微信未登录")
//	}
//
// 注意:
//   - 此接口用于检查微信客户端的登录状态
//   - 可用于判断是否需要重新登录
//   - 建议在执行其他操作前先检查登录状态
func (a *SystemAPI) CheckLogin() (bool, error) {
	respBody, err := a.client.DoRequest("POST", "/api/check_login", nil)
	if err != nil {
		return false, fmt.Errorf("check login: %w", err)
	}

	var resp types.CheckLoginResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return false, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.ErrCode != 1 && resp.ErrCode != 0 {
		return false, fmt.Errorf("api error: code=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return resp.Data.Status, nil
}

// QRScan 二维码识别
//
// 识别图片中的二维码内容
//
// 参数:
//   - path: 二维码图片路径，例如 "d:\\qr2.png"
//
// 返回:
//   - string: 识别出的二维码内容
//   - error: 识别失败时返回错误信息
//
// 示例:
//
//	result, err := client.System.QRScan("d:\\qr2.png")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("二维码内容: %s\n", result)
//
// 注意:
//   - path 必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 支持常见的图片格式（PNG、JPG等）
//   - 图片中必须包含清晰可识别的二维码
func (a *SystemAPI) QRScan(path string) (string, error) {
	req := types.QRScanRequest{
		Path: path,
	}

	respBody, err := a.client.DoRequest("POST", "/api/qrscan", req)
	if err != nil {
		return "", fmt.Errorf("qr scan: %w", err)
	}

	var resp types.QRScanResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.ErrCode != 1 && resp.ErrCode != 0 {
		return "", fmt.Errorf("api error: code=%d, msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return resp.Data.ScanRes, nil
}

// Logout 退出登录
//
// 退出当前微信登录状态
//
// 返回:
//   - error: 退出失败时返回错误信息
//
// 示例:
//
//	err := client.System.Logout()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已退出登录")
//
// 注意:
//   - 此接口会退出当前微信的登录状态
//   - 退出后需要重新扫码登录
//   - 建议在需要切换账号时使用
func (a *SystemAPI) Logout() error {
	respBody, err := a.client.DoRequest("POST", "/api/logout", nil)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
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
