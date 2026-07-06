// TCP Bridge 集成测试
//
// 模拟完整 P2P 联机链路:
//   Host:   QUIC Listener → ClientBridge → TCP 15565 (模拟 MC 服务端)
//   Client: HostBridge (TCP 15566) → QUIC Dial → Host
//
// 测试: Client TCP → HostBridge → QUIC → ClientBridge → MC服务端(15565)
//        MC服务端(15565) → ClientBridge → QUIC → HostBridge → Client TCP

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mc-connector/internal/tunnel"
	"github.com/quic-go/quic-go"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	os.Exit(run())
}

func run() int {
	// ========== 1. 启动模拟 MC 服务端 (TCP 15565) ==========
	mockServer, err := net.Listen("tcp", "127.0.0.1:15565")
	if err != nil {
		log.Printf("❌ 模拟 MC 服务端启动失败: %v", err)
		return 1
	}
	defer mockServer.Close()
	log.Println("✅ 模拟 MC 服务端: 127.0.0.1:15565")

	// MC 服务端接受连接
	serverDone := make(chan string, 1)
	go func() {
		conn, err := mockServer.Accept()
		if err != nil {
			log.Printf("❌ MC 服务端 Accept 失败: %v", err)
			serverDone <- ""
			return
		}
		defer conn.Close()
		log.Printf("✅ MC 服务端: 收到连接 %s", conn.RemoteAddr())

		// 读客户端发来的数据
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("❌ MC 服务端读失败: %v", err)
			serverDone <- ""
			return
		}
		msg := string(buf[:n])
		log.Printf("✅ MC 服务端收到: %q", msg)

		// 回复
		reply := fmt.Sprintf("ECHO: %s", msg)
		conn.Write([]byte(reply))
		log.Printf("✅ MC 服务端回复: %q", reply)
		serverDone <- msg
	}()

	// ========== 2. Host: 启动 QUIC 监听 ==========
	hostTLS, _, err := tunnel.GenerateTLSConfig()
	if err != nil {
		log.Printf("❌ 生成 Host TLS 失败: %v", err)
		return 1
	}

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		log.Printf("❌ UDP 监听失败: %v", err)
		return 1
	}
	hostPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	log.Printf("✅ Host QUIC 监听: 127.0.0.1:%d", hostPort)

	quicListener, err := quic.Listen(udpConn, hostTLS, tunnel.DefaultQuicConfig())
	if err != nil {
		log.Printf("❌ QUIC 监听失败: %v", err)
		return 1
	}

	// Host accept goroutine
	hostReady := make(chan *quic.Conn, 1)
	go func() {
		conn, err := quicListener.Accept(context.Background())
		if err != nil {
			log.Printf("❌ Host Accept 失败: %v", err)
			return
		}
		log.Printf("✅ Host: 接受 QUIC 连接 %s", conn.RemoteAddr())

		// 创建 ClientBridge: QUIC Stream → TCP 15565
		bridge := tunnel.NewClientBridge(conn)
		bridge.SetLocalPort(15565)
		go bridge.Start()

		hostReady <- conn
	}()

	// ========== 3. Client: QUIC Dial ==========
	time.Sleep(200 * time.Millisecond)

	clientTLS, _, _ := tunnel.GenerateTLSConfig()
	clientTLS.ServerName = "localhost" // 关键: 匹配证书 DNSNames

	addr := fmt.Sprintf("127.0.0.1:%d", hostPort)
	log.Printf("Client 正在连接: %s", addr)

	conn, err := tunnel.DialQUIC(addr, clientTLS)
	if err != nil {
		log.Printf("❌ Client Dial 失败: %v", err)
		return 1
	}
	log.Printf("✅ Client: QUIC 已连接 %s", conn.RemoteAddr())

	// ========== 4. Client: 创建 HostBridge ==========
	bridge, err := tunnel.NewHostBridge(15566, conn)
	if err != nil {
		log.Printf("❌ Client HostBridge 创建失败: %v", err)
		return 1
	}
	go bridge.Start()
	log.Println("✅ Client HostBridge: 监听 127.0.0.1:15566")

	// ========== 5. 模拟 MC 客户端连接 HostBridge ==========
	time.Sleep(300 * time.Millisecond)

	log.Println("--- 模拟 MC 客户端连接 127.0.0.1:15566 ---")
	mcClient, err := net.DialTimeout("tcp", "127.0.0.1:15566", 5*time.Second)
	if err != nil {
		log.Printf("❌ MC 客户端连接失败: %v", err)
		return 1
	}
	defer mcClient.Close()
	log.Println("✅ MC 客户端: 已连接")

	// 发送数据
	testMsg := "HELLO_MINECRAFT_PROTOCOL"
	mcClient.Write([]byte(testMsg))
	log.Printf("✅ MC 客户端发送: %q", testMsg)

	// 读取回复
	mcClient.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)
	n, err := mcClient.Read(buf)
	if err != nil {
		log.Printf("❌ MC 客户端读失败: %v", err)
		return 1
	}
	reply := string(buf[:n])
	log.Printf("✅ MC 客户端收到: %q", reply)

	// 验证
	select {
	case msg := <-serverDone:
		if msg == testMsg {
			log.Println("\n🎉🎉🎉 P2P Bridge 测试通过！双向数据正常！")
			return 0
		}
		log.Printf("❌ 数据不匹配: got %q, expected %q", msg, testMsg)
		return 1
	case <-time.After(10 * time.Second):
		log.Println("❌ 超时")
		return 1
	}
}
