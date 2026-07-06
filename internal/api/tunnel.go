package api

import (
	"encoding/json"
	"net/http"
)

// TunnelStatus 隧道状态
type TunnelStatus struct {
	Connected      bool   `json:"connected"`
	ConnectionType string `json:"connection_type"`
	LatencyMs      int    `json:"latency_ms"`
	BytesSent      uint64 `json:"bytes_sent"`
	BytesReceived  uint64 `json:"bytes_received"`
}

// TunnelStatus 获取隧道状态
func (a *APIHandler) TunnelStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, TunnelStatus{
		Connected:      false,
		ConnectionType: "none",
	})
}

// HandleConfig 配置管理
func (a *APIHandler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, a.cfg)
	case http.MethodPost:
		var body map[string]interface{}; json.NewDecoder(r.Body).Decode(&body); if v, ok := body["nickname"]; ok { a.cfg.Nickname = v.(string) }; if v, ok := body["relay_server"]; ok { a.cfg.RelayServer = v.(string) }
		a.cfg.Save()
		writeJSON(w, a.cfg)
	default:
		writeError(w, 405, "仅支持 GET/POST")
	}
}
