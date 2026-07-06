package api

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/mc-connector/internal/config"
	"github.com/mc-connector/internal/web"
)

// NewHandler 创建 HTTP 路由处理器
func NewHandler(cfg *config.AppConfig) http.Handler {
	mux := http.NewServeMux()

	// API 路由
	api := &APIHandler{cfg: cfg}
	mux.HandleFunc("/api/room/create", api.CreateRoom)
	mux.HandleFunc("/api/room/join", api.JoinRoom)
	mux.HandleFunc("/api/room/leave", api.LeaveRoom)
	mux.HandleFunc("/api/room/punch", api.PunchRoom)
	mux.HandleFunc("/api/room/status", api.RoomStatus)
	mux.HandleFunc("/api/connect/generate", api.GenerateConnectCode)
	mux.HandleFunc("/api/connect/parse", api.ParseConnectCode)
	mux.HandleFunc("/api/server/start", api.StartServer)
	mux.HandleFunc("/api/server/stop", api.StopServer)
	mux.HandleFunc("/api/server/status", api.ServerStatus)
	mux.HandleFunc("/api/server/command", api.ServerCommand)
	mux.HandleFunc("/api/tunnel/status", api.TunnelStatus)
	mux.HandleFunc("/api/versions", api.DetectVersions)
	mux.HandleFunc("/api/config", api.HandleConfig)

	// 前端静态文件
	mux.Handle("/", serveSPA(web.Frontend()))

	return corsMiddleware(mux)
}

// APIHandler API 处理器
type APIHandler struct {
	cfg *config.AppConfig
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// corsMiddleware 跨域处理
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// serveSPA 单页应用静态文件服务
func serveSPA(fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 不处理 API 路由
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// 尝试读取请求的文件
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		f, err := fsys.Open(path)
		if err != nil {
			// 文件不存在 → 返回 index.html（SPA路由）
			path = "index.html"
			f, err = fsys.Open(path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()

		// 读取文件内容并直接返回
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, "读取文件失败", 500)
			return
		}

		// 设置 Content-Type
		contentType := "text/html; charset=utf-8"
		switch {
		case strings.HasSuffix(path, ".js"):
			contentType = "application/javascript"
		case strings.HasSuffix(path, ".css"):
			contentType = "text/css; charset=utf-8"
		case strings.HasSuffix(path, ".svg"):
			contentType = "image/svg+xml"
		case strings.HasSuffix(path, ".png"):
			contentType = "image/png"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(data)
	})
}

