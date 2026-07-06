package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mc-connector/internal/server"
)

func main() {
	mcVersion := "1.21.11"
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".enericraft", "servers", mcVersion)

	fmt.Printf("=== 测试 Paper 服务端启动 ===\n")
	fmt.Printf("MC 版本: %s\n", mcVersion)
	fmt.Printf("缓存目录: %s\n\n", cacheDir)

	// 1. Java 检测
	fmt.Println("1. Java 检测...")
	javas := server.DetectJava()
	for _, j := range javas {
		fmt.Printf("   找到 Java %d: %s (%s)\n", j.MajorVer, j.Path, j.Version)
	}

	required := server.RequiredJava(mcVersion)
	fmt.Printf("   需要的 Java 版本: %d+\n", required)

	java := server.FindBestJava(mcVersion)
	if java == nil {
		fmt.Printf("   ❌ 未找到合适的 Java!\n")
		return
	}
	fmt.Printf("   ✅ 选中 Java %d: %s\n\n", java.MajorVer, java.Path)

	// 2. 下载 Paper
	fmt.Println("2. 下载 Paper...")
	jarPath, err := server.DownloadPaper(mcVersion, cacheDir)
	if err != nil {
		fmt.Printf("   ❌ 下载失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ JAR: %s\n\n", jarPath)

	// 3. 启动服务端
	fmt.Println("3. 启动 MC 服务端...")
	cfg := &server.Config{
		MCVersion:  mcVersion,
		JavaPath:   java.Path,
		JarPath:    jarPath,
		ServerDir:  cacheDir,
		MemoryMB:   2048,
		GameMode:   "survival",
		Difficulty: "normal",
		MaxPlayers: 8,
		Port:       25565,
	}

	proc, err := server.Start(cfg)
	if err != nil {
		fmt.Printf("   ❌ 启动失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 进程已启动 (PID %d)\n", proc)
	fmt.Printf("   Running: %v\n", proc.Running())

	// 等待一下看看进程是否还活着
	fmt.Println("\n等待 5 秒... (按 Ctrl+C 退出)")
	select {}
}
