// EneriCraft — 主入口
//
// Go 后端 + 内嵌 React 前端的轻量架构。
// 编译为单个 exe 文件，无外部依赖。
// 启动后自动打开浏览器显示 UI。

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/mc-connector/internal/api"
	"github.com/mc-connector/internal/auth"
	"github.com/mc-connector/internal/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("EneriCraft v0.9.0 启动中...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败，使用默认配置: %v", err)
		cfg = config.Default()
	}

	// 启动本地认证服务器 + 绕过 Mojang 验证
	auth.Default.Start("25568")
	auth.SetupHosts()
	log.Println("本地认证: 127.0.0.1:25568")

	// 创建 API 路由
	handler := api.NewHandler(cfg)

	// 找到可用端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("无法绑定端口: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// HTTP 服务器
	server := &http.Server{Handler: handler}
	go func() {
		log.Printf("本地服务: http://127.0.0.1:%d", port)
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Printf("HTTP 服务错误: %v", err)
		}
	}()

	// 自动打开浏览器
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	go openBrowser(url)

	log.Printf("请在浏览器中打开: %s", url)
	log.Println("按 Ctrl+C 退出")

	// 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("正在关闭...")
	server.Shutdown(context.Background())
}

// openBrowser 打开默认浏览器
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("无法打开浏览器: %v", err)
	}
}
