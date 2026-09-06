package gvxsdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/olaria01/gVxSdk/sdk/go/api"
)

// Client 微信 SDK 客户端
type Client struct {
	baseURL    string
	authToken  string // 可选的鉴权 Token，为空则不添加鉴权头
	httpClient *http.Client

	// API 模块
	Emotion  *api.EmotionAPI
	Video    *api.VideoAPI
	Message  *api.MessageAPI
	CDN      *api.CDNAPI
	Voice    *api.VoiceAPI
	SNS      *api.SNSAPI
	System   *api.SystemAPI
	Room     *api.RoomAPI
	Favorite *api.FavoriteAPI
	Contact  *api.ContactAPI
	Social   *api.SocialAPI
	WxWork   *api.WxWorkAPI
	Payment  *api.PaymentAPI
}

// NewClient 创建新的客户端实例
// baseURL: 微信中间件的 HTTP 服务地址，例如 "http://127.0.0.1:19088"
// authToken: 可选的鉴权 Token，例如 "your-local-auth-token"，传空字符串则不使用鉴权
func NewClient(baseURL string, authToken ...string) *Client {
	token := ""
	if len(authToken) > 0 {
		token = authToken[0]
	}

	c := &Client{
		baseURL:   baseURL,
		authToken: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 初始化 API 模块（注入 client 依赖）
	c.Emotion = api.NewEmotionAPI(c)
	c.Video = api.NewVideoAPI(c)
	c.Message = api.NewMessageAPI(c)
	c.CDN = api.NewCDNAPI(c)
	c.Voice = api.NewVoiceAPI(c)
	c.SNS = api.NewSNSAPI(c)
	c.System = api.NewSystemAPI(c)
	c.Room = api.NewRoomAPI(c)
	c.WxWork = api.NewWxWorkAPI(c)
	c.Favorite = api.NewFavoriteAPI(c)
	c.Contact = api.NewContactAPI(c)
	c.Social = api.NewSocialAPI(c)
	c.Payment = api.NewPaymentAPI(c)

	return c
}

// DoRequest 执行 HTTP 请求
func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加鉴权头
	if c.authToken != "" {
		req.Header.Set("X-Auth-Token", c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
