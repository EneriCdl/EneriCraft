// 中继客户端
//
// 当 P2P 直连失败时，通过中继服务器转发。

package tunnel

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/quic-go/quic-go"
)

// RelayConfig 中继配置
type RelayConfig struct {
	ServerAddr string // 中继服务器地址 (host:port)
}

// DefaultRelay 社区中继服务器
const DefaultRelay = "120.77.255.112:9000"

// ConnectRelay 通过中继服务器建立连接
// roomCode: 连接码（用于配对）
func ConnectRelay(serverAddr, roomCode string) (*quic.Conn, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"mc-connector-relay"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, serverAddr, tlsCfg, &quic.Config{
		KeepAlivePeriod: 10 * time.Second,
		MaxIdleTimeout:  120 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("连接中继失败: %w", err)
	}

	// 发送 JOIN 消息（不关闭 stream，让中继继续使用）
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		conn.CloseWithError(0, "stream open failed")
		return nil, fmt.Errorf("打开中继流失败: %w", err)
	}

	joinMsg := "JOIN:" + roomCode + "\n"
	if _, err := stream.Write([]byte(joinMsg)); err != nil {
		conn.CloseWithError(0, "write failed")
		return nil, fmt.Errorf("发送加入消息失败: %w", err)
	}
	// 不关闭 stream — 中继需要读取配对后的数据

	log.Printf("[中继] 已连接到 %s, 房间: %s...", serverAddr, roomCode[:min(20, len(roomCode))])
	return conn, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
