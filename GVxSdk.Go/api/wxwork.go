package api

import (
	"encoding/json"
	"fmt"

	"github.com/olaria01/gVxSdk/sdk/go/types"
)

// WxWorkAPI 企业微信相关接口
type WxWorkAPI struct {
	client Client
}

// NewWxWorkAPI 创建企业微信API实例
func NewWxWorkAPI(client Client) *WxWorkAPI {
	return &WxWorkAPI{client: client}
}

// DownloadWxWorkFile 下载企业文件/图片
//
// 下载企业微信中的文件或图片到本地。
// 需要提供文件URL、解密密钥、输出路径和认证密钥。
//
// 参数:
//   - url: 文件下载URL（从企业微信消息中获取）
//   - key: 解密密钥（用于解密文件内容）
//   - out: 输出文件路径，例如 "d:\揽收6.jpg"
//   - authkey: 认证密钥（用于验证下载权限）
//
// 返回:
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	err := client.WxWork.DownloadWxWorkFile(
//	    "https://wwfile.work.weixin.qq.com/cgi-bin/download?f=...",
//	    "0efef41ef3567c6e6be689f5b951bbac",
//	    "d:\\揽收6.jpg",
//	    "0A2B6F4E2D4D77415141414141415941646C674E577938334B6E34456C55367643703940696D2E7778776F726B10F392F231",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("文件下载成功")
//
// 注意:
//   - URL、key 和 authkey 通常从企业微信消息中获取
//   - 输出路径的目录必须存在
//   - 如果输出文件已存在，将被覆盖
//   - 支持下载图片、文档等各类文件
func (a *WxWorkAPI) DownloadWxWorkFile(url, key, out, authkey string) error {
	req := types.DownloadWxWorkFileRequest{
		URL:     url,
		Key:     key,
		Out:     out,
		AuthKey: authkey,
	}

	respBody, err := a.client.DoRequest("POST", "/api/download_wxwork_file", req)
	if err != nil {
		return fmt.Errorf("download wxwork file: %w", err)
	}

	var resp types.DownloadWxWorkFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}
