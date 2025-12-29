package thunder

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// FileList 文件列表响应
type FileList struct {
	Kind          string  `json:"kind"`
	NextPageToken string  `json:"next_page_token"`
	Files         []File  `json:"files"`
}

// File 文件信息
type File struct {
	Kind           string    `json:"kind"`
	ID             string    `json:"id"`
	ParentID       string    `json:"parent_id"`
	Name           string    `json:"name"`
	Size           string    `json:"size"`
	WebContentLink string    `json:"web_content_link"`
	CreatedTime    time.Time `json:"created_time"`
	ModifiedTime   time.Time `json:"modified_time"`
	IconLink       string    `json:"icon_link"`
	ThumbnailLink  string    `json:"thumbnail_link"`
	Hash           string    `json:"hash"`
	Trashed        bool      `json:"trashed"`
	OriginalURL    string    `json:"original_url"`
	Medias         []struct {
		Link struct {
			URL    string    `json:"url"`
			Token  string    `json:"token"`
			Expire time.Time `json:"expire"`
		} `json:"link"`
	} `json:"medias"`
}

// IsFolder 判断是否是文件夹
func (f *File) IsFolder() bool {
	return f.Kind == FOLDER
}

// ListFiles 列出文件
func (c *Client) ListFiles(parentID string) ([]File, error) {
	var files []File
	var pageToken string

	for {
		var resp struct {
			ErrResp
			FileList
		}

		_, err := c.AuthRequest(FILE_API_URL, http.MethodGet, func(r *resty.Request) {
			r.SetQueryParams(map[string]string{
				"parent_id":  parentID,
				"page_token": pageToken,
				"__type":     "drive",
				"refresh":    "true",
				"__sync":     "true",
				"with_audit": "true",
				"limit":      "100",
				"filters":    `{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}}`,
			})
		}, &resp)

		if err != nil {
			return nil, err
		}

		if resp.ErrResp.IsError() {
			return nil, &resp.ErrResp
		}

		files = append(files, resp.Files...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return files, nil
}

// GetFile 获取文件详情
func (c *Client) GetFile(fileID string) (*File, error) {
	var resp struct {
		ErrResp
		File
	}

	_, err := c.AuthRequest(FILE_API_URL+"/"+fileID, http.MethodGet, nil, &resp)
	if err != nil {
		return nil, err
	}

	if resp.ErrResp.IsError() {
		return nil, &resp.ErrResp
	}

	return &resp.File, nil
}

// GetDownloadURL 获取下载链接
func (c *Client) GetDownloadURL(fileID string) (string, error) {
	file, err := c.GetFile(fileID)
	if err != nil {
		return "", err
	}

	// 优先使用 media link
	for _, media := range file.Medias {
		if media.Link.URL != "" {
			return media.Link.URL, nil
		}
	}

	// 使用 web content link
	if file.WebContentLink != "" {
		return file.WebContentLink, nil
	}

	return "", nil
}

// CreateFolder 创建文件夹
func (c *Client) CreateFolder(parentID, name string) (*File, error) {
	var resp struct {
		ErrResp
		File
	}

	_, err := c.AuthRequest(FILE_API_URL, http.MethodPost, func(r *resty.Request) {
		r.SetBody(map[string]interface{}{
			"kind":      FOLDER,
			"name":      name,
			"parent_id": parentID,
		})
	}, &resp)

	if err != nil {
		return nil, err
	}

	if resp.ErrResp.IsError() {
		return nil, &resp.ErrResp
	}

	return &resp.File, nil
}

// DeleteFile 删除文件（移到回收站）
func (c *Client) DeleteFile(fileID string) error {
	var resp ErrResp

	_, err := c.AuthRequest(FILE_API_URL+"/"+fileID+"/trash", http.MethodPatch, func(r *resty.Request) {
		r.SetBody("{}")
	}, &resp)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return &resp
	}

	return nil
}
