package api

import (
	"encoding/json"
	"fmt"

	"github.com/hk5jj7grw4-debug/CVXkernel/GVxSdk.Go/types"
)

// CDNAPI CDN相关接口
type CDNAPI struct {
	client Client
}

// NewCDNAPI 创建CDN API实例
func NewCDNAPI(client Client) *CDNAPI {
	return &CDNAPI{client: client}
}

// SendCDNImgMsg CDN发送图片（用于转发消息）
//
// 通过CDN发送图片，通常用于转发已存在于CDN上的图片消息
//
// 参数:
//   - toWxid: 接收人微信ID
//   - totalLen: 图片总长度（字符串格式）
//   - fileID: CDN文件ID
//   - aesKey: AES加密密钥
//   - cdnMidImgSize: CDN中等图片大小（字符串格式）
//   - cdnThumbImgSize: CDN缩略图大小（字符串格式）
//   - encryVer: 加密版本（字符串格式）
//
// 返回:
//   - error: 发送失败时返回错误信息
//
// 示例:
//
//	err := client.CDN.SendCDNImgMsg(
//	    "filehelper",
//	    "4876183",
//	    "307002010204693067020104020445fb609102030f4fed...",
//	    "cea539047a38291fb96776572f464625",
//	    "4876183",
//	    "4876183",
//	    "1",
//	)
//
// 注意:
//   - 此接口主要用于转发已存在的图片消息
//   - 需要从原始消息中获取相关的CDN参数
func (a *CDNAPI) SendCDNImgMsg(toWxid, totalLen, fileID, aesKey, cdnMidImgSize, cdnThumbImgSize, encryVer string) error {
	req := types.SendCDNImgMsgRequest{
		ToWxid:          toWxid,
		TotalLen:        totalLen,
		FileID:          fileID,
		AESKey:          aesKey,
		CDNMidImgSize:   cdnMidImgSize,
		CDNThumbImgSize: cdnThumbImgSize,
		EncryVer:        encryVer,
	}

	respBody, err := a.client.DoRequest("POST", "/api/send_cdn_img_msg", req)
	if err != nil {
		return fmt.Errorf("send cdn img msg: %w", err)
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

// DownloadImg 下载图片
//
// 下载消息中的图片到本地。所有参数都来自于图片消息。
//
// 参数:
//   - toUser: 自己的 wxid（消息接收方）
//   - fromUser: 发送图片的人的 wxid
//   - startPos: 起始位置，通常为 0
//   - totalLen: 图片总长度，来自消息的 cdnmidimgurl_size
//   - dataLen: 数据长度，来自消息的 cdnmidimgurl_size
//   - compressType: 压缩类型，默认为 0
//   - msgID: 消息 ID
//   - path: 保存路径，例如 "d:\\7878787878.jpg"
//
// 返回:
//   - *types.DownloadImgResponse: 下载结果信息
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	resp, err := client.CDN.DownloadImg(
//	    "wxid_ozyqateb85un22",
//	    "wxid_8543785438012",
//	    0,
//	    44041,
//	    44041,
//	    0,
//	    1213935352,
//	    "d:\\7878787878.jpg",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if resp.Status == "success" {
//	    fmt.Printf("图片下载成功: %s\n", resp.Path)
//	}
//
// 注意:
//   - 所有参数都应该从图片消息中获取
//   - 保存路径必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 下载成功后 Status 字段为 "success"
func (a *CDNAPI) DownloadImg(toUser, fromUser string, startPos, totalLen, dataLen, compressType int, msgID int64, path string) (*types.DownloadImgResponse, error) {
	req := types.DownloadImgRequest{
		ToUser:       toUser,
		FromUser:     fromUser,
		StartPos:     startPos,
		TotalLen:     totalLen,
		DataLen:      dataLen,
		CompressType: compressType,
		MsgID:        msgID,
		Path:         path,
	}

	respBody, err := a.client.DoRequest("POST", "/api/download_img", req)
	if err != nil {
		return nil, fmt.Errorf("download img: %w", err)
	}

	var resp types.DownloadImgResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("download failed: status=%s", resp.Status)
	}

	return &resp, nil
}

// DownloadFile 下载文件
//
// 下载消息中的文件到本地。所有参数都来自于文件消息。
//
// 参数:
//   - fromUser: 发送文件的人的 wxid
//   - totalLen: 文件总长度（字符串格式）
//   - msgID: 消息 ID（数字类型）
//   - path: 保存路径，例如 "d:\\罗泽南8月考勤表.xlsx"
//   - attachID: 附件 ID，来自消息中的 attachid 字段
//   - fileType: 文件类型，例如 "6"
//
// 返回:
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	err := client.CDN.DownloadFile(
//	    "wxid_xxx",
//	    "31538",
//	    2754393265637994500,
//	    "d:\\罗泽南8月考勤表.xlsx",
//	    "@cdn_3057020100044b3049020100020403e0b2d502032df85f...",
//	    "6",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("文件下载成功")
//
// 注意:
//   - 所有参数都应该从文件消息中获取
//   - 保存路径必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - attachID 是一个很长的字符串，必须完整传递
func (a *CDNAPI) DownloadFile(fromUser, totalLen string, msgID float64, path, attachID, fileType string) error {
	req := types.DownloadFileRequest{
		FromUser: fromUser,
		TotalLen: totalLen,
		MsgID:    msgID,
		Path:     path,
		AttachID: attachID,
		Type:     fileType,
	}

	respBody, err := a.client.DoRequest("POST", "/api/download_file", req)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}

	var resp types.DownloadFileResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// CDNDownload CDN下载
//
// 通用的CDN下载接口，支持下载高清图片、普通图片、缩略图、视频和文件
//
// 参数:
//   - fileID: 文件ID，从消息中获取
//   - aesKey: AES密钥，从消息中获取
//   - imgType: 文件类型
//     1 - 高清图片
//     2 - 普通图片
//     3 - 缩略图
//     4 - 视频
//     5 - 文件
//   - out: 输出路径，例如 "D:\\test1.jpg"
//
// 返回:
//   - error: 下载失败时返回错误信息
//
// 示例:
//
//	// 下载普通图片
//	err := client.CDN.CDNDownload(
//	    "3057020100044b3049020100020403e0b2d502032e1e4102046e8ae473...",
//	    "04a64148a121f65a94ad23396e565a73",
//	    2,
//	    "D:\\test1.jpg",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("文件下载成功")
//
// 注意:
//   - fileID 和 aesKey 需要从消息中获取
//   - 输出路径必须是微信中间件所在机器上的本地路径
//   - Windows 路径需要使用双反斜杠或单正斜杠
//   - 根据不同的 imgType 下载不同质量的资源
func (a *CDNAPI) CDNDownload(fileID, aesKey string, imgType int, out string) error {
	req := types.CDNDownloadRequest{
		FileID:  fileID,
		AESKey:  aesKey,
		ImgType: imgType,
		Out:     out,
	}

	respBody, err := a.client.DoRequest("POST", "/api/cdn_download", req)
	if err != nil {
		return fmt.Errorf("cdn download: %w", err)
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

// GetCDNInfo 获取CDN信息
//
// 获取当前账号的CDN配置信息
//
// 返回:
//   - interface{}: CDN信息数据
//   - error: 获取失败时返回错误信息
//
// 示例:
//
//	cdnInfo, err := client.CDN.GetCDNInfo()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("CDN信息: %+v\n", cdnInfo)
//
// 注意:
//   - 返回的数据包含CDN配置相关信息
//   - 具体数据结构根据实际返回内容解析
func (a *CDNAPI) GetCDNInfo() (interface{}, error) {
	respBody, err := a.client.DoRequest("POST", "/api/get_cdn_info", nil)
	if err != nil {
		return nil, fmt.Errorf("get cdn info: %w", err)
	}

	var resp types.GetCDNInfoResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Code != 1 && resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return resp.Data, nil
}
