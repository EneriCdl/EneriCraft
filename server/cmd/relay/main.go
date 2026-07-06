// MC 联机器 — 中继服务器（可选增强）
//
// 当 P2P 直连失败时，通过中继服务器转发游戏数据包。
// 社区成员可以自行部署中继服务器。
// 一个 8MB 的单文件二进制，零配置即可运行。
//
// 启动方式:
//   ./relay-server --port 3478

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := flag.Int("port", 3478, "中继服务器监听端口")
	maxRooms := flag.Int("max-rooms", 100, "最大房间数")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("MC 联机器 - 中继服务器 v0.1.0")
	log.Printf("监听端口: %d | 最大房间数: %d", *port, *maxRooms)
	log.Println("零配置 · 纯转发 · 不存储任何数据")

	// TODO: 启动 QUIC 中继服务
	// TODO: 会话管理
	// TODO: 速率限制

	// 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\n正在关闭...")
}
