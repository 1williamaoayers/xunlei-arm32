package thunder

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Task 下载任务
type Task struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	StatusName  string `json:"status_name"`
	FileID      string `json:"file_id"`
	FileSize    string `json:"file_size"`
	Progress    int    `json:"progress"`
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	Message     string `json:"message"`
}

// TaskList 任务列表响应
type TaskList struct {
	Tasks         []Task `json:"tasks"`
	NextPageToken string `json:"next_page_token"`
}

const (
	TASK_API_URL = API_URL + "/tasks"
)

// ListTasks 列出下载任务
func (c *Client) ListTasks() ([]Task, error) {
	var tasks []Task
	var pageToken string

	for {
		var resp struct {
			ErrResp
			TaskList
		}

		_, err := c.AuthRequest(TASK_API_URL, http.MethodGet, func(r *resty.Request) {
			r.SetQueryParams(map[string]string{
				"page_token": pageToken,
				"limit":      "100",
				"type":       "offline",
				"filters":    `{"phase":{"in":"PHASE_TYPE_RUNNING,PHASE_TYPE_ERROR,PHASE_TYPE_PENDING,PHASE_TYPE_COMPLETE"}}`,
			})
		}, &resp)

		if err != nil {
			return nil, err
		}

		if resp.ErrResp.IsError() {
			return nil, &resp.ErrResp
		}

		tasks = append(tasks, resp.Tasks...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return tasks, nil
}

// CreateOfflineTask 创建离线下载任务
func (c *Client) CreateOfflineTask(url, parentID, name string) (*Task, error) {
	if parentID == "" {
		parentID = "" // 根目录
	}

	var resp struct {
		ErrResp
		Task Task `json:"task"`
	}

	_, err := c.AuthRequest(TASK_API_URL, http.MethodPost, func(r *resty.Request) {
		r.SetBody(map[string]interface{}{
			"type":       "offline",
			"url":        url,
			"parent_id":  parentID,
			"file_name":  name,
			"folder_type": "",
		})
	}, &resp)

	if err != nil {
		return nil, err
	}

	if resp.ErrResp.IsError() {
		return nil, &resp.ErrResp
	}

	return &resp.Task, nil
}

// GetTask 获取任务详情
func (c *Client) GetTask(taskID string) (*Task, error) {
	var resp struct {
		ErrResp
		Task
	}

	_, err := c.AuthRequest(TASK_API_URL+"/"+taskID, http.MethodGet, nil, &resp)
	if err != nil {
		return nil, err
	}

	if resp.ErrResp.IsError() {
		return nil, &resp.ErrResp
	}

	return &resp.Task, nil
}

// DeleteTask 删除任务
func (c *Client) DeleteTask(taskID string) error {
	var resp ErrResp

	_, err := c.AuthRequest(TASK_API_URL+"/"+taskID, http.MethodDelete, nil, &resp)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return &resp
	}

	return nil
}
