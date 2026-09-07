package types

// GetWxBasePathResponse 获取微信缓存目录响应
type GetWxBasePathResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"` // 微信缓存目录路径
}

// DecryptDBRequest 解密数据库请求
type DecryptDBRequest struct {
	DpPath  string `json:"dp_path"`  // 加密数据库路径
	OutPath string `json:"out_path"` // 解密后输出路径
	Key     string `json:"key"`      // 解密密钥
}

// DecryptDBResponse 解密数据库响应
type DecryptDBResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// CheckLoginResponse 获取登录状态响应
type CheckLoginResponse struct {
	AccountWxid string `json:"account_wxid"`
	ErrCode     int    `json:"errCode"`
	ErrMsg      string `json:"errMsg"`
	Data        struct {
		Status bool `json:"status"` // true-已登录 false-未登录
	} `json:"data"`
}

// QRScanRequest 二维码识别请求
type QRScanRequest struct {
	Path string `json:"path"` // 二维码图片路径
}

// QRScanResponse 二维码识别响应
type QRScanResponse struct {
	AccountWxid string     `json:"account_wxid"`
	Data        QRScanData `json:"data"`
	ErrCode     int        `json:"errCode"`
	ErrMsg      string     `json:"errMsg"`
}

// QRScanData 二维码识别数据
type QRScanData struct {
	ScanRes string `json:"scan_res"` // 识别结果
}

// RefreshQRCodeResponse 获取登录二维码响应
type RefreshQRCodeResponse struct {
	BaseResponse              BaseResponse              `json:"baseResponse"`
	QRCode                    QRCodeBuffer              `json:"qrcode"`
	Uuid                      string                    `json:"uuid"`
	CheckTime                 int                       `json:"checkTime"`
	NotifyKey                 QRCodeBuffer              `json:"notifyKey"`
	ExpiredTime               int                       `json:"expiredTime"`
	BlueToothBroadCastContent BlueToothBroadCastContent `json:"blueToothBroadCastContent"`
}

// QRCodeBuffer 二维码缓冲区
type QRCodeBuffer struct {
	ILen   int    `json:"iLen"`
	Buffer string `json:"buffer,omitempty"`
}

// BlueToothBroadCastContent 蓝牙广播内容
type BlueToothBroadCastContent struct {
	ILen int `json:"iLen"`
}

// GetA8KeyRequest 获取A8key请求
type GetA8KeyRequest struct {
	URL     string `json:"url"`     // URL地址
	URLType string `json:"urlType"` // URL类型，例如 "0"
	Scene   string `json:"scene"`   // 场景，例如 "0"
}

// GetA8KeyResponse 获取A8key响应
type GetA8KeyResponse struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	// 根据实际返回添加其他字段
}

// JsLoginRequest 获取小程序code请求
type JsLoginRequest struct {
	WaID string `json:"waId"` // 小程序AppID
}

// JsLoginResponse 获取小程序code响应
type JsLoginResponse struct {
	BaseResponse      BaseResponse      `json:"baseResponse"`
	JsapiBaseresponse JsapiBaseresponse `json:"jsapiBaseresponse"`
	Code              string            `json:"code"`
	State             string            `json:"state"`
}

// JsapiBaseresponse JSAPI基础响应
type JsapiBaseresponse struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	ErrorNumber int    `json:"errorNumber"`
}

// GetDBHandleResponse 获取数据库句柄响应
type GetDBHandleResponse struct {
	Data []DBHandleInfo `json:"data"`
}

// DBHandleInfo 数据库句柄信息
type DBHandleInfo struct {
	Handle int64  `json:"handle"` // 数据库句柄
	Name   string `json:"name"`   // 数据库名称
}

// Sqlite3ExecRequest 执行数据库查询请求
type Sqlite3ExecRequest struct {
	DBName string `json:"db_name"` // 数据库名称，例如 "contact.db"
	SqlFmt string `json:"sql_fmt"` // SQL查询语句
}

// Sqlite3ExecResponse 执行数据库查询响应
// 返回的是一个数组，每个元素是一个map，key是列名，value是列值
type Sqlite3ExecResponse []map[string]interface{}

// GetConfigPathResponse 获取配置文件保存目录响应
type GetConfigPathResponse struct {
	Code       int         `json:"code"`
	ConfigPath string      `json:"configPath"`
	Data       interface{} `json:"data"`
}

// AntiRevokeRequest 防撤回请求
type AntiRevokeRequest struct {
	Switch string `json:"swtich"` // 开关状态，"true" 或 "false"（注意：字段名拼写为 swtich）
}

// AntiRevokeResponse 防撤回响应
type AntiRevokeResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// WechatInitResponse 微信初始化响应
type WechatInitResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// BackupDatabaseRequest 数据库备份请求
type BackupDatabaseRequest struct {
	OutputDir string `json:"outputDir"` // 输出目录
	Name      string `json:"name"`      // 数据库文件名
}

// BackupDatabaseResponse 数据库备份响应
type BackupDatabaseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
