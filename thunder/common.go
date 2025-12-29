package thunder

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	API_URL        = "https://api-pan.xunlei.com/drive/v1"
	FILE_API_URL   = API_URL + "/files"
	XLUSER_API_URL = "https://xluser-ssl.xunlei.com/v1"

	FOLDER = "drive#folder"
	FILE   = "drive#file"
)

// Common 通用配置
type Common struct {
	client *resty.Client

	captchaToken    string
	captchaTokenExp time.Time

	Algorithms        []string
	DeviceID          string
	ClientID          string
	ClientSecret      string
	ClientVersion     string
	PackageName       string
	UserAgent         string
	DownloadUserAgent string

	// 签名
	Timestamp    string
	CaptchaSign  string
	CreditKey    string
}

// NewCommon 创建通用配置
func NewCommon() *Common {
	return &Common{
		client: resty.New().SetTimeout(30 * time.Second),
		Algorithms: []string{
			"9uJNVj/wLmdwKrJaVj/omlQ",
			"Oz64Lp0GigmChHMf/6TNfxx7O9PyopcczMsnf",
			"Eb+L7Ce+Ej48u",
			"jKY0",
			"ASr0zCl6v8W4aidjPK5KHd1Lq3t+vBFf41dqv5+fnOd",
			"wQlozdg6r1qxh0eRmt3QgNXOvSZO6q/GXK",
			"gmirk+ciAvIgA/cxUUCema47jr/YToixTT+Q6O",
			"5IiCoM9B1/788ntB",
			"P07JH0h6qoM6TSUAK2aL9T5s2QBVeY9JWvalf",
			"+oK0AN",
		},
		ClientID:          "Xp6vsxz_7IYVw2BB",
		ClientSecret:      "Xp6vsy4tN9toTVdMSpomVdXpRmES",
		ClientVersion:     "8.31.0.9726",
		PackageName:       "com.xunlei.downloadprovider",
		UserAgent:         "ANDROID-com.xunlei.downloadprovider/8.31.0.9726 netWorkType/WIFI appid/40 deviceName/Xiaomi_M2004j7ac deviceModel/M2004J7AC OSVersion/12 protocolVersion/301 platformVersion/10 sdkVersion/512000 Oauth2Client/0.9 (Linux 4_14_186-perf) (JAVA 0)",
		DownloadUserAgent: "Dalvik/2.1.0 (Linux; U; Android 12; M2004J7AC Build/SP1A.210812.016)",
	}
}

// SetDeviceID 设置设备 ID（32位 MD5）
func (c *Common) SetDeviceID(id string) {
	if len(id) != 32 {
		hash := md5.Sum([]byte(id))
		c.DeviceID = hex.EncodeToString(hash[:])
	} else {
		c.DeviceID = id
	}
}

// SetCaptchaToken 设置验证码 Token
func (c *Common) SetCaptchaToken(token string) {
	c.captchaToken = token
	c.captchaTokenExp = time.Now().Add(2 * time.Hour)
}

// GetCaptchaToken 获取验证码 Token
func (c *Common) GetCaptchaToken() string {
	if time.Now().After(c.captchaTokenExp) {
		return ""
	}
	return c.captchaToken
}

// SetCreditKey 设置信任密钥
func (c *Common) SetCreditKey(key string) {
	c.CreditKey = key
}

// Request 发送请求
func (c *Common) Request(url string, method string, callback func(*resty.Request), resp interface{}) ([]byte, error) {
	req := c.client.R()
	req.SetHeaders(map[string]string{
		"User-Agent":      c.UserAgent,
		"Accept":          "application/json;charset=UTF-8",
		"X-Device-ID":     c.DeviceID,
		"X-Client-ID":     c.ClientID,
		"X-Client-Version": c.ClientVersion,
	})

	if callback != nil {
		callback(req)
	}

	if resp != nil {
		req.SetResult(resp)
	}

	var (
		data *resty.Response
		err  error
	)

	switch strings.ToUpper(method) {
	case http.MethodGet:
		data, err = req.Get(url)
	case http.MethodPost:
		data, err = req.Post(url)
	case http.MethodPatch:
		data, err = req.Patch(url)
	case http.MethodDelete:
		data, err = req.Delete(url)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	// 检查错误响应
	if errResp, ok := resp.(*ErrResp); ok && errResp.IsError() {
		return data.Body(), errResp
	}

	return data.Body(), nil
}

// ErrResp 错误响应
type ErrResp struct {
	ErrorCode        int64  `json:"error_code"`
	ErrorMsg         string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *ErrResp) IsError() bool {
	if e.ErrorMsg == "success" {
		return false
	}
	return e.ErrorCode != 0 || e.ErrorMsg != "" || e.ErrorDescription != ""
}

func (e *ErrResp) Error() string {
	return fmt.Sprintf("ErrorCode: %d, Error: %s, ErrorDescription: %s", e.ErrorCode, e.ErrorMsg, e.ErrorDescription)
}
