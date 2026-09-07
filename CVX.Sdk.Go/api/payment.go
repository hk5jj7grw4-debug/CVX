package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVX/CVX.Sdk.Go/types"
)

// PaymentAPI 支付相关接口
type PaymentAPI struct {
	client Client
}

// NewPaymentAPI 创建支付API实例
func NewPaymentAPI(client Client) *PaymentAPI {
	return &PaymentAPI{client: client}
}

// TenPayTransferConfirm 确认收款
//
// 确认接收微信转账。当收到转账消息时，可以调用此接口确认收款。
//
// 参数:
//   - invalidTime: 过期时间，来自转账消息里的 invalid_time 字段
//   - transferID: 传输ID，来自转账消息里的 transferid 字段
//
// 返回:
//   - error: 确认失败时返回错误信息
//
// 示例:
//
//	// 从转账消息中获取参数
//	invalidTime := int64(1765008493)
//	transferID := "1000050001202512051421286682351"
//
//	err := client.Payment.TenPayTransferConfirm(invalidTime, transferID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("收款确认成功")
//
// 注意:
//   - invalidTime 和 transferID 必须从转账消息中获取
//   - 转账有时效性，过期后无法确认收款
//   - 确认收款后，金额会立即到账
//   - 如果不确认，转账会在24小时后自动退回给发送方
func (a *PaymentAPI) TenPayTransferConfirm(invalidTime int64, transferID string) error {
	req := types.TenPayTransferConfirmRequest{
		InvalidTime: invalidTime,
		TransferID:  transferID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/ten_pay_trans_fer_confirm", req)
	if err != nil {
		return fmt.Errorf("ten pay transfer confirm: %w", err)
	}

	var resp types.TenPayTransferConfirmResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// UnTenPayTransferConfirm 拒绝收款
//
// 拒绝接收微信转账。当收到转账消息时，可以调用此接口拒绝收款，转账将退回给发送方。
//
// 参数:
//   - invalidTime: 过期时间，来自转账消息里的 invalid_time 字段
//   - transferID: 传输ID，来自转账消息里的 transferid 字段
//
// 返回:
//   - error: 拒绝失败时返回错误信息
//
// 示例:
//
//	// 从转账消息中获取参数
//	invalidTime := int64(1765008493)
//	transferID := "1000050001202512051421286682351"
//
//	err := client.Payment.UnTenPayTransferConfirm(invalidTime, transferID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("已拒绝收款，转账将退回给发送方")
//
// 注意:
//   - invalidTime 和 transferID 必须从转账消息中获取
//   - 拒绝收款后，转账会立即退回给发送方
//   - 拒绝操作不可撤销
//   - 如果不主动确认或拒绝，转账会在24小时后自动退回
func (a *PaymentAPI) UnTenPayTransferConfirm(invalidTime int64, transferID string) error {
	req := types.UnTenPayTransferConfirmRequest{
		InvalidTime: invalidTime,
		TransferID:  transferID,
	}

	respBody, err := a.client.DoRequest("POST", "/api/un_ten_pay_trans_fer_confirm", req)
	if err != nil {
		return fmt.Errorf("un ten pay transfer confirm: %w", err)
	}

	var resp types.UnTenPayTransferConfirmResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}
