// MC 联机器 — 信令服务器（可选增强）
//
// 提供房间发现、NAT 协调、好友状态同步等功能。
// 仅在用户配置了信令服务器地址时使用。
// 零服务器模式下不需要此组件。

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("MC 联机器 - 信令服务器 v0.1.0")
	log.Println("监听地址: :8080")

	// TODO: 启动 WebSocket 服务
	// TODO: 启动 STUN 服务

	// 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("\n正在关闭...")
}
