package thunder

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// TokenResp Token 响应
type TokenResp struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Sub          string `json:"sub"`
	UserID       string `json:"user_id"`
}

func (t *TokenResp) Token() string {
	return fmt.Sprintf("%s %s", t.TokenType, t.AccessToken)
}

// CaptchaTokenRequest 验证码 Token 请求
type CaptchaTokenRequest struct {
	Action       string            `json:"action"`
	CaptchaToken string            `json:"captcha_token"`
	ClientID     string            `json:"client_id"`
	DeviceID     string            `json:"device_id"`
	Meta         map[string]string `json:"meta"`
	RedirectUri  string            `json:"redirect_uri"`
}

// CaptchaTokenResponse 验证码 Token 响应
type CaptchaTokenResponse struct {
	CaptchaToken string `json:"captcha_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Url          string `json:"url"`
}

// Client 迅雷客户端
type Client struct {
	*Common
	*TokenResp

	Username string
	Password string
}

// NewClient 创建迅雷客户端
func NewClient(username, password string) *Client {
	c := &Client{
		Common:   NewCommon(),
		Username: username,
		Password: password,
	}
	c.SetDeviceID(username + "_xunlei_arm32")
	return c
}

// Login 登录
func (c *Client) Login() error {
	// 获取验证码 Token
	if err := c.refreshCaptchaToken("POST:/v1/auth/signin"); err != nil {
		return fmt.Errorf("获取验证码 Token 失败: %w", err)
	}

	// 登录
	url := XLUSER_API_URL + "/auth/signin"
	
	// 构建请求体
	body := map[string]interface{}{
		"client_id":     c.ClientID,
		"client_secret": c.ClientSecret,
		"username":      c.Username,
		"password":      c.Password,
	}

	req := c.client.R().
		SetHeader("User-Agent", c.UserAgent).
		SetHeader("Accept", "application/json;charset=UTF-8").
		SetHeader("X-Device-ID", c.DeviceID).
		SetHeader("X-Client-ID", c.ClientID).
		SetHeader("X-Client-Version", c.ClientVersion).
		SetHeader("X-Captcha-Token", c.GetCaptchaToken()).
		SetBody(body)

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}

	// 打印原始响应用于调试
	fmt.Printf("[DEBUG] Login Response Status: %d\n", resp.StatusCode())
	fmt.Printf("[DEBUG] Login Response Body: %s\n", string(resp.Body()))

	// 解析响应
	var tokenResp struct {
		ErrResp
		TokenResp
	}
	
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return fmt.Errorf("解析登录响应失败: %w, body: %s", err, string(resp.Body()))
	}

	if tokenResp.ErrResp.IsError() {
		return fmt.Errorf("登录失败: %s", tokenResp.ErrResp.Error())
	}

	c.TokenResp = &tokenResp.TokenResp
	return nil
}

// RefreshToken 刷新 Token
func (c *Client) RefreshToken() error {
	if c.TokenResp == nil || c.TokenResp.RefreshToken == "" {
		return fmt.Errorf("没有可用的 RefreshToken")
	}

	url := XLUSER_API_URL + "/auth/token"
	var resp struct {
		ErrResp
		TokenResp
	}

	_, err := c.Request(url, http.MethodPost, func(r *resty.Request) {
		r.SetBody(map[string]interface{}{
			"client_id":     c.ClientID,
			"client_secret": c.ClientSecret,
			"grant_type":    "refresh_token",
			"refresh_token": c.TokenResp.RefreshToken,
		})
	}, &resp)

	if err != nil {
		return err
	}

	if resp.ErrResp.IsError() {
		return fmt.Errorf("刷新 Token 失败: %s", resp.ErrResp.Error())
	}

	c.TokenResp = &resp.TokenResp
	return nil
}

// refreshCaptchaToken 刷新验证码 Token
func (c *Client) refreshCaptchaToken(action string) error {
	url := XLUSER_API_URL + "/shield/captcha/init"
	var resp struct {
		ErrResp
		CaptchaTokenResponse
	}

	_, err := c.Request(url, http.MethodPost, func(r *resty.Request) {
		r.SetBody(&CaptchaTokenRequest{
			Action:      action,
			ClientID:    c.ClientID,
			DeviceID:    c.DeviceID,
			Meta:        map[string]string{"email": c.Username},
			RedirectUri: "xlaccsdk01://xunlei.com/callback?state=harbor",
		})
	}, &resp)

	if err != nil {
		return err
	}

	if resp.CaptchaToken != "" {
		c.SetCaptchaToken(resp.CaptchaToken)
	}

	// 如果需要验证，返回验证 URL
	if resp.Url != "" {
		return fmt.Errorf("需要验证，请访问: %s", resp.Url)
	}

	return nil
}

// IsLogin 检查是否已登录
func (c *Client) IsLogin() bool {
	return c.TokenResp != nil && c.TokenResp.AccessToken != ""
}

// AuthRequest 发送带认证的请求
func (c *Client) AuthRequest(url string, method string, callback func(*resty.Request), resp interface{}) ([]byte, error) {
	if !c.IsLogin() {
		return nil, fmt.Errorf("未登录")
	}

	return c.Request(url, method, func(r *resty.Request) {
		r.SetHeaders(map[string]string{
			"Authorization":   c.TokenResp.Token(),
			"X-Captcha-Token": c.GetCaptchaToken(),
		})
		if callback != nil {
			callback(r)
		}
	}, resp)
}
