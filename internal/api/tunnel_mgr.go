package api

import (
	"context"
	"crypto/tls"
	"log"
	"sync"

	"github.com/mc-connector/internal/tunnel"
	"github.com/quic-go/quic-go"
)

type TunnelManager struct {
	mu         sync.Mutex
	listener   *quic.Listener
	conn       *quic.Conn
	hostBridge *tunnel.Bridge
	cliBridge  *tunnel.Bridge
	connected  bool
	targetPort int // 主机桥接目标端口（MC 服务端/LAN 端口）
}

var tmgr = &TunnelManager{targetPort: 25565}

// SetListener 设置 QUIC listener 并指定桥接目标端口
func (m *TunnelManager) SetListener(l *quic.Listener, targetPort int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listener != nil {
		m.listener.Close()
	}
	m.targetPort = targetPort
	m.listener = l
	go m.acceptLoop()
}

func (m *TunnelManager) StartHost(addr string, tlsCfg *tls.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.listener != nil {
		m.listener.Close()
	}

	listener, err := tunnel.ListenQUIC(addr, tlsCfg)
	if err != nil {
		return err
	}
	m.listener = listener
	go m.acceptLoop()
	log.Printf("[Tunnel] 主机监听 %s", addr)
	return nil
}

func (m *TunnelManager) acceptLoop() {
	m.mu.Lock()
	port := m.targetPort
	m.mu.Unlock()

	for {
		conn, err := m.listener.Accept(context.Background())
		if err != nil {
			log.Printf("[Tunnel] 监听停止: %v", err)
			return
		}
		log.Printf("[Tunnel] 客户端已连入: %s", conn.RemoteAddr())

		m.mu.Lock()
		m.conn = conn
		bridge := tunnel.NewClientBridge(conn)
		bridge.SetLocalPort(port)
		bridge.OnActivity = func() {
			UpdatePlayerStatus("朋友", true)
			log.Printf("[Tunnel] 检测到活动: 朋友已连接")
		}
		m.hostBridge = bridge
		m.connected = true
		m.mu.Unlock()

		go bridge.Start()
		log.Printf("[Tunnel] 桥接: QUIC ↔ 127.0.0.1:%d", port)
	}
}

// Connect 作为客户端连接主机
// localPort: 本地监听的 TCP 端口（MC 客户端连接此端口）
func (m *TunnelManager) Connect(addr string, tlsCfg *tls.Config, localPort int) error {
	conn, err := tunnel.DialQUIC(addr, tlsCfg)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.conn = conn
	bridge, err := tunnel.NewHostBridge(localPort, conn)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	bridge.OnActivity = func() {
		UpdatePlayerStatus("你", true)
		log.Printf("[Tunnel] 连接活跃")
	}
	m.cliBridge = bridge
	m.connected = true
	m.mu.Unlock()

	go bridge.Start()
	log.Printf("[Tunnel] 桥接: 127.0.0.1:%d ↔ QUIC ↔ 主机", localPort)
	return nil
}

// ConnectRelay 客户端通过中继连接
func (m *TunnelManager) ConnectRelay(conn *quic.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conn = conn
	bridge, err := tunnel.NewHostBridge(25566, conn)
	if err != nil {
		log.Printf("[Tunnel] 中继桥接失败: %v", err)
		return
	}
	bridge.OnActivity = func() {
		UpdatePlayerStatus("你", true)
		log.Printf("[Tunnel] 连接活跃")
	}
	m.cliBridge = bridge
	m.connected = true
	go bridge.Start()
	log.Printf("[Tunnel] 客户端中继: 127.0.0.1:25566 ↔ 中继 ↔ 主机")
}

// HostRelay 主机通过中继暴露（将中继连接桥接到 MC LAN 端口）
func (m *TunnelManager) HostRelay(conn *quic.Conn, lanPort int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已有 P2P 桥接则不覆盖
	if m.hostBridge != nil {
		log.Printf("[Tunnel] 中继已作为备用通道连接")
		return
	}

	bridge := tunnel.NewClientBridge(conn)
	bridge.SetLocalPort(lanPort)
		bridge.OnActivity = func() {
			UpdatePlayerStatus("朋友", true)
			log.Printf("[Tunnel] 主机检测到朋友连接")
		}
	m.hostBridge = bridge
	go bridge.Start()
	log.Printf("[Tunnel] 主机中继: 中继 ↔ 127.0.0.1:%d", lanPort)
}

func (m *TunnelManager) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *TunnelManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cliBridge != nil {
		m.cliBridge.Stop()
	}
	if m.hostBridge != nil {
		m.hostBridge.Stop()
	}
	if m.listener != nil {
		m.listener.Close()
	}
	m.conn = nil
	m.listener = nil
	m.connected = false
}
