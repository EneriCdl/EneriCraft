package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"net"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/mc-connector/internal/connectcode"
	"github.com/mc-connector/internal/server"
	"github.com/mc-connector/internal/tunnel"
)

type RoomStatus struct {
	Connected      bool     `json:"connected"`
	ConnectCode    string   `json:"connect_code,omitempty"`
	ConnectionType string   `json:"connection_type"`
	NATType        string   `json:"nat_type,omitempty"`
	MCVersion      string   `json:"mc_version"`
	ServerRunning  bool     `json:"server_running"`
	Players        []PlayerEntry `json:"players"`
	Step           string   `json:"step"`
	Mode           string   `json:"mode"`
	NeedOpenLAN    bool     `json:"need_open_lan"`
}

func (a *APIHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeError(w, 405, "仅支持 POST"); return }
	var req struct {
		GameMode string `json:"game_mode"`; RoomName string `json:"room_name"`
		Version  string `json:"mc_version"`; Mode     string `json:"mode"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Mode == "" { req.Mode = "lan" }

	mcVersion, username := req.Version, ""
	if mcVersion == "" || !isCleanMCVersion(mcVersion) {
		v, u := detectRunningMC(); if v != "" { mcVersion, username = v, u }
	}
	if mcVersion == "" || !isCleanMCVersion(mcVersion) {
		writeError(w, 400, "未检测到运行中的 Minecraft，请先打开游戏"); return
	}
	log.Printf("[创建房间] MC版本: %s, 用户: %s, 模式: %s", mcVersion, username, req.Mode)

	if req.Mode == "lan" { a.createRoomLAN(w, mcVersion, username, req.GameMode); return }
	a.createRoomPaper(w, req, mcVersion)
}

func (a *APIHandler) createRoomLAN(w http.ResponseWriter, mcVersion, username, gameMode string) {
	log.Printf("[LAN] 检测端口...")
	lanPort, _, err := tunnel.DetectLANPort(5 * time.Second)
	if err != nil || lanPort == 0 {
		writeJSON(w, RoomStatus{Connected: false, ConnectionType: "none", MCVersion: mcVersion, Mode: "lan", NeedOpenLAN: true, Step: "need_open_lan"}); return
	}
	log.Printf("[LAN] 端口: %d", lanPort)

	udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	quicPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	tlsCfg, pubKey, _ := tunnel.GenerateTLSConfig()
	quicListener, err := quic.Listen(udpConn, tlsCfg, tunnel.DefaultQuicConfig())
	if err != nil { writeError(w, 500, "QUIC 监听失败: "+err.Error()); return }

	tmgr.SetListener(quicListener, lanPort)
	tunnel.AddFirewallRule(quicPort)
	log.Printf("[LAN] QUIC 0.0.0.0:%d → LAN:%d", quicPort, lanPort)

	// UPnP
	upnpPort := quicPort
	if p, e := tunnel.AddUPnPMapping(quicPort); e == nil { upnpPort = p; log.Printf("[LAN] UPnP 外网:%d→本机:%d", upnpPort, quicPort) }

	// 端点
	stun, _ := tunnel.GetPublicAddr()
	pubIP := ""; if stun != nil { pubIP = stun.PublicIP }
	eps := []connectcode.Endpoint{}
	if addrs, _ := tunnel.GetLocalAddrs(); addrs != nil {
		for _, a := range addrs { eps = append(eps, connectcode.Endpoint{IP: a, Port: quicPort}) }
	}
	if pubIP != "" { eps = append(eps, connectcode.Endpoint{IP: pubIP, Port: upnpPort}) }
	if len(eps) == 0 { eps = append(eps, connectcode.Endpoint{IP: "127.0.0.1", Port: quicPort}) }

	code, _ := connectcode.Generate(eps, pubKey, mcVersion, "")
	// 主机侧连中继
	relayAddr := a.cfg.RelayServer
	if relayAddr == "" { relayAddr = tunnel.DefaultRelay }
	if relayAddr != "" {
		go func() {
			if rc, err := tunnel.ConnectRelay(relayAddr, code); err == nil {
				tmgr.HostRelay(rc, lanPort)
				log.Printf("[LAN] 中继通道已建立")
			}
		}()
	}

	SetRoom(code, "p2p", mcVersion)
	log.Printf("[LAN] 完成! LAN=%d QUIC=%d 端点=%d", lanPort, quicPort, len(eps))
	writeJSON(w, RoomStatus{Connected: false, ConnectCode: code, ConnectionType: "p2p", NATType: "p2p", MCVersion: mcVersion, ServerRunning: true, Players: []PlayerEntry{{Name: username, Online: true}, {Name: "等待朋友加入...", Online: false}}, Step: "waiting", Mode: "lan"})
}

func (a *APIHandler) createRoomPaper(w http.ResponseWriter, req struct {
	GameMode string `json:"game_mode"`; RoomName string `json:"room_name"`
	Version  string `json:"mc_version"`; Mode     string `json:"mode"`
}, mcVersion string) {
	java := server.FindBestJava(mcVersion)
	if java == nil { writeError(w, 400, fmt.Sprintf("未找到 Java %d+", server.RequiredJava(mcVersion))); return }
	home, _ := os.UserHomeDir()
	jarPath, err := server.DownloadPaper(mcVersion, filepath.Join(home, ".enericraft", "servers", mcVersion))
	if err != nil { writeError(w, 500, "下载 Paper 失败: "+err.Error()); return }

	cfg := &server.Config{MCVersion: mcVersion, JavaPath: java.Path, JarPath: jarPath, ServerDir: filepath.Dir(jarPath), MemoryMB: 2048, GameMode: "survival", Difficulty: "normal", MaxPlayers: 8, Port: 25565}
	if req.GameMode != "" { cfg.GameMode = req.GameMode }
	_, err = StartServer(cfg)
	if err != nil { writeError(w, 500, "启动 Paper 失败: "+err.Error()); return }

	udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	quicPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	tlsCfg, pubKey, _ := tunnel.GenerateTLSConfig()
	quicListener, err := quic.Listen(udpConn, tlsCfg, tunnel.DefaultQuicConfig())
	if err != nil { writeError(w, 500, "QUIC 失败: "+err.Error()); return }
	tmgr.SetListener(quicListener, 25565); tunnel.AddFirewallRule(quicPort)

	stun, _ := tunnel.GetPublicAddr(); pubIP := ""; if stun != nil { pubIP = stun.PublicIP }
	eps := []connectcode.Endpoint{}
	if addrs, _ := tunnel.GetLocalAddrs(); addrs != nil {
		for _, a := range addrs { eps = append(eps, connectcode.Endpoint{IP: a, Port: quicPort}) }
	}
	if pubIP != "" { eps = append(eps, connectcode.Endpoint{IP: pubIP, Port: quicPort}) }
	if len(eps) == 0 { eps = append(eps, connectcode.Endpoint{IP: "127.0.0.1", Port: quicPort}) }

	code, _ := connectcode.Generate(eps, pubKey, mcVersion, "")
	SetRoom(code, "p2p", mcVersion)
	writeJSON(w, RoomStatus{Connected: false, ConnectCode: code, ConnectionType: "p2p", NATType: "p2p", MCVersion: mcVersion, ServerRunning: true, Players: []PlayerEntry{{Name: "等待朋友加入...", Online: false}}, Step: "waiting", Mode: "paper"})
}

func (a *APIHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeError(w, 405, "仅支持 POST"); return }
	var req struct{ Code string }; json.NewDecoder(r.Body).Decode(&req)
	parsed, err := connectcode.Parse(req.Code)
	if err != nil { writeError(w, 400, "连接码无效: "+err.Error()); return }
	log.Printf("[加入] 版本=%s %d端点", parsed.MV, len(parsed.EP))

	clientTLS, _, _ := tunnel.GenerateTLSConfig(); clientTLS.ServerName = "localhost"
	connected := false
	for _, ep := range parsed.EP {
		addr := fmt.Sprintf("%s:%d", ep.IP, ep.Port)
		log.Printf("[加入] 尝试: %s", addr)
		if err := tmgr.Connect(addr, clientTLS, 25566); err != nil {
			log.Printf("[加入] %s 失败: %v", addr, err); continue
		}
		connected = true; log.Printf("[加入] 已连接 %s", addr); break
	}

	if !connected {
		// P2P 失败 → 走中继
		relayAddr := a.cfg.RelayServer
		if relayAddr == "" { relayAddr = tunnel.DefaultRelay }
		if relayAddr != "" {
			log.Printf("[加入] P2P失败, 走中继: %s", relayAddr)
			relayConn, err := tunnel.ConnectRelay(relayAddr, req.Code)
			if err == nil {
				tmgr.ConnectRelay(relayConn)
				connected = true
				log.Printf("[加入] 中继连接成功")
			} else {
				log.Printf("[加入] 中继失败: %v", err)
			}
		}
	}

	if !connected {
		writeError(w, 500, "P2P 直连失败，中继也连不上。请检查网络后重试")
		return
	}

	SetRoom(req.Code, "p2p", parsed.MV); AddPlayer("朋友")
	writeJSON(w, RoomStatus{Connected: true, ConnectCode: req.Code, ConnectionType: "p2p", MCVersion: parsed.MV, ServerRunning: true, Players: []PlayerEntry{{Name: "你", Online: true}, {Name: "房主", Online: true}}, Step: "connected"})
}

func (a *APIHandler) PunchRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeError(w, 405, "仅支持 POST"); return }
	var req struct{ PunchCode string `json:"punch_code"` }; json.NewDecoder(r.Body).Decode(&req)
	parsed, err := connectcode.Parse(req.PunchCode)
	if err != nil { writeError(w, 400, "回执码无效: "+err.Error()); return }
	if len(parsed.EP) == 0 { writeError(w, 400, "回执码无地址"); return }

	peerAddr := fmt.Sprintf("%s:%d", parsed.EP[0].IP, parsed.EP[0].Port)
	log.Printf("[打洞] 房主→客户端: %s", peerAddr)
	tlsCfg, _, _ := tunnel.GenerateTLSConfig(); tlsCfg.ServerName = "localhost"

	go func() {
		for i := 0; i < 60; i++ {
			conn, err := tunnel.DialQUIC(peerAddr, tlsCfg)
			if err == nil {
				log.Printf("[打洞] 成功! (第%d次)", i+1)
				tmgr.HostRelay(conn, 25565)
				return
			}
			if i%10 == 0 { log.Printf("[打洞] 第%d次...", i+1) }
			time.Sleep(2 * time.Second)
		}
		log.Printf("[打洞] 放弃")
	}()
	writeJSON(w, map[string]string{"status": "punching", "message": "打洞中..."})
}

func (a *APIHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	ClearRoom(); log.Println("[离开] 已清理"); writeJSON(w, map[string]string{"status": "ok"})
}

func (a *APIHandler) RoomStatus(w http.ResponseWriter, r *http.Request) {
	s := GetState()
	writeJSON(w, RoomStatus{Connected: s.Connected, ConnectCode: s.ConnectCode, ConnectionType: s.ConnectionType, NATType: s.NATType, MCVersion: s.MCVersion, ServerRunning: IsServerRunning(), Players: s.Players})
}
