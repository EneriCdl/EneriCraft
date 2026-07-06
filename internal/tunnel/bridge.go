// TCP ↔ QUIC 桥接
//
// 核心功能：
//   Host 模式: 本地监听 TCP → 数据转发到 QUIC Stream
//   Client 模式: QUIC Stream 数据 → 转发到本地 TCP (MC服务端)
//
// 这是实现"透明联机"的关键 ——
// 玩家无需任何配置，打开 MC 客户端连接本地端口即可。

package tunnel

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

// BridgeMode 桥接模式
type BridgeMode int

const (
	ModeHost   BridgeMode = iota // 主机: TCP → QUIC
	ModeClient                   // 客户端: QUIC → TCP
)

// Bridge TCP-QUIC 桥接器
type Bridge struct {
	mode        BridgeMode
	localPort   int
	listener    net.Listener
	conn        *quic.Conn
	streams     map[int64]*quic.Stream
	mu          sync.Mutex
	stopCh      chan struct{}
	running     bool
	OnActivity  func() // 首次数据流动时回调
	fired       bool
}

// NewHostBridge 创建主机端桥接
//
// 监听 localPort，将每个接入的 TCP 连接转发到 QUIC Stream。
// MC 客户端 → 连接 localhost:localPort → TCP → QUIC → 远端
func NewHostBridge(localPort int, conn *quic.Conn) (*Bridge, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return nil, fmt.Errorf("TCP 监听失败 :%d: %w", localPort, err)
	}

	return &Bridge{
		mode:      ModeHost,
		localPort: localPort,
		listener:  listener,
		conn:      conn,
		streams:   make(map[int64]*quic.Stream),
		stopCh:    make(chan struct{}),
	}, nil
}

// NewClientBridge 创建客户端桥接
//
// 从 QUIC Stream 读取数据，转发到本地 TCP (MC 客户端)。
// QUIC Stream → TCP → localhost:25565 → MC 服务端
func NewClientBridge(conn *quic.Conn) *Bridge {
	return &Bridge{
		mode:    ModeClient,
		conn:    conn,
		streams: make(map[int64]*quic.Stream),
		stopCh:  make(chan struct{}),
	}
}

// LocalPort 返回本地监听端口
func (b *Bridge) LocalPort() int {
	return b.localPort
}

func (b *Bridge) SetLocalPort(port int) {
	b.localPort = port
}

// Start 启动桥接
func (b *Bridge) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("桥接已在运行")
	}
	b.running = true
	b.mu.Unlock()

	switch b.mode {
	case ModeHost:
		return b.runHost()
	case ModeClient:
		return b.runClient()
	default:
		return fmt.Errorf("未知桥接模式")
	}
}

// runHost 运行主机模式 (TCP接受 → QUIC转发)
func (b *Bridge) runHost() error {
	log.Printf("[Bridge:Host] 监听 127.0.0.1:%d，等待 MC 客户端连接...", b.localPort)

	for {
		select {
		case <-b.stopCh:
			return nil
		default:
		}

		tcpConn, err := b.listener.Accept()
		if err != nil {
			select {
			case <-b.stopCh:
				return nil
			default:
				return fmt.Errorf("接受 TCP 连接失败: %w", err)
			}
		}

		// 为每个 TCP 连接打开 QUIC Stream
		go b.handleHostConnection(tcpConn)
	}
}

func (b *Bridge) fireActivity() {
	if !b.fired && b.OnActivity != nil {
		b.fired = true
		b.OnActivity()
	}
}

// handleHostConnection 处理单个主机端连接
func (b *Bridge) handleHostConnection(tcpConn net.Conn) {
	defer tcpConn.Close()
	b.fireActivity()

	stream, err := OpenStream(b.conn)
	if err != nil {
		log.Printf("[Bridge:Host] 打开 QUIC Stream 失败: %v", err)
		return
	}
	defer stream.Close()

	// 双向拷贝: TCP ← → QUIC Stream
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(tcpConn, stream)
		tcpConn.SetReadDeadline(time.Now())
	}()

	go func() {
		defer wg.Done()
		io.Copy(stream, tcpConn)
	}()

	wg.Wait()
}

// runClient 运行客户端模式 (QUIC Stream接收 → TCP转发)
func (b *Bridge) runClient() error {
	log.Printf("[Bridge:Client] 等待 QUIC Stream...")

	for {
		select {
		case <-b.stopCh:
			return nil
		default:
		}

		stream, err := b.conn.AcceptStream(b.conn.Context())
		if err != nil {
			select {
			case <-b.stopCh:
				return nil
			default:
				return fmt.Errorf("接受 QUIC Stream 失败: %w", err)
			}
		}

		// 为每个 Stream 连接到本地 MC 服务端
		go b.handleClientStream(stream)
	}
}

// handleClientStream 处理单个客户端 Stream
func (b *Bridge) handleClientStream(stream *quic.Stream) {
	defer stream.Close()
	b.fireActivity()

	// 连接到本地 MC 服务端
	tcpConn, err := net.DialTimeout("tcp",
		fmt.Sprintf("127.0.0.1:%d", b.localPort), 5*time.Second)
	if err != nil {
		log.Printf("[Bridge:Client] 连接本地 MC 失败: %v", err)
		return
	}
	defer tcpConn.Close()

	// 双向拷贝
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(tcpConn, stream)
		tcpConn.SetReadDeadline(time.Now())
	}()

	go func() {
		defer wg.Done()
		io.Copy(stream, tcpConn)
	}()

	wg.Wait()
}

// Stop 停止桥接
func (b *Bridge) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return
	}
	close(b.stopCh)
	if b.listener != nil {
		b.listener.Close()
	}
	b.running = false
	log.Printf("[Bridge] 桥接已停止")
}
