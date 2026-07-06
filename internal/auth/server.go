// 本地 Yggdrasil 认证服务器
//
// 模拟 Mojang 的验证接口，MC 客户端连接前向本服务器"认证"时，
// 直接返回有效的用户信息，从而绕过"无效的会话"错误。
//
// 配合 authlib-injector 或 hosts 文件使用。

package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type PlayerProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AuthServer 本地认证服务器
type AuthServer struct {
	mu       sync.RWMutex
	players  map[string]PlayerProfile // username → profile
	serverID string
}

var Default = &AuthServer{
	players: map[string]PlayerProfile{},
}

// Start 启动认证服务器
func (s *AuthServer) Start(port string) {
	mux := http.NewServeMux()

	// Yggdrasil API 端点
	mux.HandleFunc("/authserver/authenticate", s.handleAuthenticate)
	mux.HandleFunc("/authserver/refresh", s.handleRefresh)
	mux.HandleFunc("/authserver/validate", s.handleValidate)
	mux.HandleFunc("/sessionserver/session/minecraft/join", s.handleJoin)
	mux.HandleFunc("/sessionserver/session/minecraft/hasJoined", s.handleHasJoined)
	mux.HandleFunc("/api/yggdrasil/", s.handleYggdrasil)
	mux.HandleFunc("/", s.handleDefault)

	go func() {
		log.Printf("[Auth] 本地认证服务器 127.0.0.1:%s", port)
		if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
			log.Printf("[Auth] 启动失败: %v", err)
		}
	}()
}

func (s *AuthServer) handleDefault(w http.ResponseWriter, r *http.Request) {
	// 对任何请求返回合法的认证响应
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "EneriCraftPlayer"
	}
	profile := s.getOrCreateProfile(username)
	json.NewEncoder(w).Encode(profile)
}

func (s *AuthServer) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" {
		req.Username = "EneriCraftPlayer"
	}

	profile := s.getOrCreateProfile(req.Username)

	// 返回标准 Yggdrasil 认证响应
	resp := map[string]interface{}{
		"accessToken":       "enericraft-zero-token",
		"clientToken":       "enericraft-client",
		"availableProfiles": []PlayerProfile{profile},
		"selectedProfile":   profile,
		"user": map[string]interface{}{
			"id":         profile.ID,
			"properties": []interface{}{},
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *AuthServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccessToken string `json:"accessToken"`
		ClientToken string `json:"clientToken"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	// 无论如何都返回成功
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accessToken": "enericraft-zero-token",
		"clientToken": req.ClientToken,
	})
}

func (s *AuthServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	// 总是返回 204 No Content（验证通过）
	w.WriteHeader(204)
}

func (s *AuthServer) handleJoin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccessToken     string `json:"accessToken"`
		SelectedProfile string `json:"selectedProfile"`
		ServerID        string `json:"serverId"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	// 接受任何 join 请求
	w.WriteHeader(204)
}

func (s *AuthServer) handleHasJoined(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "EneriCraftPlayer"
	}
	profile := s.getOrCreateProfile(username)

	// 返回标准 hasJoined 响应
	resp := map[string]interface{}{
		"id":         profile.ID,
		"name":       profile.Name,
		"properties": []interface{}{},
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *AuthServer) handleYggdrasil(w http.ResponseWriter, r *http.Request) {
	// 通用 Yggdrasil API 处理
	path := strings.TrimPrefix(r.URL.Path, "/api/yggdrasil/")
	switch {
	case strings.Contains(path, "authenticate"):
		s.handleAuthenticate(w, r)
	case strings.Contains(path, "refresh"):
		s.handleRefresh(w, r)
	case strings.Contains(path, "validate"):
		s.handleValidate(w, r)
	case strings.Contains(path, "join"):
		s.handleJoin(w, r)
	case strings.Contains(path, "hasJoined"):
		s.handleHasJoined(w, r)
	default:
		w.WriteHeader(204)
	}
}

func (s *AuthServer) getOrCreateProfile(username string) PlayerProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.players[username]; ok {
		return p
	}
	// 生成固定 UUID（offline 模式）
	uuid := "00000000-0000-0000-0000-" + usernameToUUID(username)
	p := PlayerProfile{ID: uuid, Name: username}
	s.players[username] = p
	return p
}

func usernameToUUID(name string) string {
	// 生成 12 位 hex（补全为 UUID 后缀）
	h := 0
	for _, c := range name {
		h = h*31 + int(c)
	}
	if h < 0 { h = -h }
	return strings.Repeat("0", 12-len(itoa(h))) + itoa(h)
}

func itoa(n int) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if len(s) < 12 {
		s = strings.Repeat("0", 12-len(s)) + s
	}
	return s
}
