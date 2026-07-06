package main

import (
	"fmt"
	"time"

	"github.com/mc-connector/internal/tunnel"
)

func main() {
	code := "EC1-TEST-RELAY-CODE"

	// 主机侧连接中继
	go func() {
		conn, err := tunnel.ConnectRelay("127.0.0.1:19000", code)
		if err != nil {
			fmt.Printf("HOST FAIL: %v\n", err)
			return
		}
		fmt.Println("HOST connected to relay")
		time.Sleep(3 * time.Second)
		conn.CloseWithError(0, "done")
	}()

	time.Sleep(1 * time.Second)

	// 客户端侧连接中继
	conn, err := tunnel.ConnectRelay("127.0.0.1:19000", code)
	if err != nil {
		fmt.Printf("CLIENT FAIL: %v\n", err)
		return
	}
	fmt.Println("CLIENT connected to relay")
	time.Sleep(2 * time.Second)
	conn.CloseWithError(0, "done")

	fmt.Println("SUCCESS: relay pairing works!")
}
