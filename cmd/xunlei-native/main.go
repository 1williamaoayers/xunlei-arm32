package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cnk3x/xunlei/thunder"
)

//go:embed web
var webFS embed.FS

var (
	version = "dev"
	client  *thunder.Client
)

type Config struct {
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DataDir  string `json:"data_dir"`
}

func main() {
	log.Printf("迅雷 ARM32 原生客户端 v%s\n", version)

	// 解析配置
	cfg := Config{
		Port:    2345,
		DataDir: "./data",
	}

	// 从环境变量读取
	if port := os.Getenv("XL_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}
	if username := os.Getenv("XL_USERNAME"); username != "" {
		cfg.Username = username
	}
	if password := os.Getenv("XL_PASSWORD"); password != "" {
		cfg.Password = password
	}
	if dataDir := os.Getenv("XL_DATA_DIR"); dataDir != "" {
		cfg.DataDir = dataDir
	}

	// 创建数据目录
	os.MkdirAll(cfg.DataDir, 0755)

	// 创建 HTTP 服务
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/files", handleFiles)
	mux.HandleFunc("/api/tasks", handleTasks)
	mux.HandleFunc("/api/task/create", handleCreateTask)

	// 静态文件
	webContent, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	// 启动服务
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("服务启动在 http://0.0.0.0:%d\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v\n", err)
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

// API 响应
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func jsonResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("收到登录请求")
	
	if r.Method != http.MethodPost {
		log.Println("登录请求方法错误:", r.Method)
		jsonResponse(w, 405, "Method Not Allowed", nil)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("登录请求解析错误:", err)
		jsonResponse(w, 400, "请求格式错误", nil)
		return
	}

	log.Printf("尝试登录用户: %s\n", req.Username)
	
	client = thunder.NewClient(req.Username, req.Password)
	if err := client.Login(); err != nil {
		log.Printf("登录失败: %v\n", err)
		jsonResponse(w, 500, fmt.Sprintf("登录失败: %v", err), nil)
		return
	}

	log.Printf("登录成功, UserID: %s, Sub: %s, AccessToken长度: %d\n", 
		client.TokenResp.UserID, client.TokenResp.Sub, len(client.TokenResp.AccessToken))
	log.Printf("IsLogin: %v\n", client.IsLogin())
	
	userID := client.TokenResp.UserID
	if userID == "" {
		userID = client.TokenResp.Sub
	}
	
	jsonResponse(w, 0, "登录成功", map[string]interface{}{
		"user_id": userID,
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if client == nil || !client.IsLogin() {
		jsonResponse(w, 401, "未登录", nil)
		return
	}

	jsonResponse(w, 0, "已登录", map[string]interface{}{
		"user_id":   client.TokenResp.UserID,
		"logged_in": true,
	})
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	if client == nil || !client.IsLogin() {
		jsonResponse(w, 401, "未登录", nil)
		return
	}

	parentID := r.URL.Query().Get("parent_id")
	files, err := client.ListFiles(parentID)
	if err != nil {
		jsonResponse(w, 500, fmt.Sprintf("获取文件列表失败: %v", err), nil)
		return
	}

	jsonResponse(w, 0, "success", files)
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if client == nil || !client.IsLogin() {
		jsonResponse(w, 401, "未登录", nil)
		return
	}

	tasks, err := client.ListTasks()
	if err != nil {
		jsonResponse(w, 500, fmt.Sprintf("获取任务列表失败: %v", err), nil)
		return
	}

	jsonResponse(w, 0, "success", tasks)
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, 405, "Method Not Allowed", nil)
		return
	}

	if client == nil || !client.IsLogin() {
		jsonResponse(w, 401, "未登录", nil)
		return
	}

	var req struct {
		URL      string `json:"url"`
		ParentID string `json:"parent_id"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, 400, "请求格式错误", nil)
		return
	}

	task, err := client.CreateOfflineTask(req.URL, req.ParentID, req.Name)
	if err != nil {
		jsonResponse(w, 500, fmt.Sprintf("创建任务失败: %v", err), nil)
		return
	}

	jsonResponse(w, 0, "任务创建成功", task)
}
