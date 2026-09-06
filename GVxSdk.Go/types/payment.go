package types

// TenPayTransferConfirmRequest 确认收款请求
type TenPayTransferConfirmRequest struct {
	InvalidTime int64  `json:"invalid_time"` // 过期时间，来自消息里的 invalid_time
	TransferID  string `json:"transferid"`   // 传输ID，来自消息里的 transferid
}

// TenPayTransferConfirmResponse 确认收款响应
type TenPayTransferConfirmResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// UnTenPayTransferConfirmRequest 拒绝收款请求
type UnTenPayTransferConfirmRequest struct {
	InvalidTime int64  `json:"invalid_time"` // 过期时间，来自消息里的 invalid_time
	TransferID  string `json:"transferid"`   // 传输ID，来自消息里的 transferid
}

// UnTenPayTransferConfirmResponse 拒绝收款响应
type UnTenPayTransferConfirmResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
