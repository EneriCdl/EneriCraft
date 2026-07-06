package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mc-connector/internal/connectcode"
	"github.com/mc-connector/internal/tunnel"
)

// GenerateConnectCode 生成连接码
func (a *APIHandler) GenerateConnectCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "仅支持 POST")
		return
	}

	var req struct {
		MCVersion string `json:"mc_version"`
		ModHash   string `json:"mod_hash"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.MCVersion == "" {
		req.MCVersion = "1.21"
	}

	// 获取公网地址
	stunResp, err := tunnel.GetPublicAddr()
	if err != nil {
		writeError(w, 500, "无法获取公网地址: "+err.Error())
		return
	}

	// 获取内网地址
	localAddrs, _ := tunnel.GetLocalAddrs()

	// 生成密钥
	key, err := connectcode.GenerateKey()
	if err != nil {
		writeError(w, 500, "生成密钥失败")
		return
	}

	// 构建端点
	endpoints := []connectcode.Endpoint{
		{IP: stunResp.PublicIP, Port: stunResp.PublicPort},
	}
	for _, addr := range localAddrs {
		endpoints = append(endpoints, connectcode.Endpoint{IP: addr, Port: stunResp.PublicPort})
	}

	// 生成连接码
	code, err := connectcode.Generate(endpoints, key, req.MCVersion, req.ModHash)
	if err != nil {
		writeError(w, 500, "生成连接码失败: "+err.Error())
		return
	}

	log.Printf("生成连接码: %d 字符, %d 个端点", len(code), len(endpoints))
	writeJSON(w, map[string]string{
		"code":         code,
		"nat_type":     "检测中...",
		"mc_version":   req.MCVersion,
		"public_ip":    stunResp.PublicIP,
		"public_port":  formatInt(stunResp.PublicPort),
	})
}

// ParseConnectCode 解析连接码
func (a *APIHandler) ParseConnectCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "仅支持 POST")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "无效请求")
		return
	}

	parsed, err := connectcode.Parse(req.Code)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}

	log.Printf("解析连接码: 版本=%s, %d 个端点", parsed.MV, len(parsed.EP))

	var endpoints []map[string]interface{}
	for _, ep := range parsed.EP {
		endpoints = append(endpoints, map[string]interface{}{
			"ip":   ep.IP,
			"port": ep.Port,
		})
	}

	writeJSON(w, map[string]interface{}{
		"version":    parsed.MV,
		"endpoints":  endpoints,
		"mod_hash":   parsed.MH,
		"mc_version": parsed.MV,
	})
}

func formatInt(n int) string {
	return json.Number(json.Number(fmt.Sprintf("%d", n))).String()
}
