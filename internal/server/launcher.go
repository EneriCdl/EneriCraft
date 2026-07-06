// MC 服务端进程管理器
//
// 负责 Paper 服务端的启动、停止、崩溃重启。

package server

import (
	"fmt"
	"log"
	"os"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config 服务端配置
type Config struct {
	MCVersion  string `json:"mc_version"`
	JavaPath   string `json:"java_path"`
	JarPath    string `json:"jar_path"`
	ServerDir  string `json:"server_dir"`
	MemoryMB   int    `json:"memory_mb"`
	GameMode   string `json:"game_mode"`
	Difficulty string `json:"difficulty"`
	MaxPlayers int    `json:"max_players"`
	Port       int    `json:"port"`
}

// DefaultConfig 返回默认服务端配置
func DefaultConfig() *Config {
	return &Config{
		MCVersion:  "1.21",
		JavaPath:   "java",
		MemoryMB:   2048,
		GameMode:   "survival",
		Difficulty: "normal",
		MaxPlayers: 8,
		Port:       25565,
	}
}

// Process 服务端进程
type Process struct {
	cmd    *exec.Cmd
	config *Config
	logCh  chan string
	done   chan struct{}
	stdin  io.WriteCloser
}

// SendCommand 向服务端控制台发送命令（如 op、gamemode 等）
func (p *Process) SendCommand(cmd string) error {
	if p.stdin == nil {
		return fmt.Errorf("服务端 stdin 不可用")
	}
	_, err := io.WriteString(p.stdin, cmd+"\n")
	return err
}

// Start 启动 MC 服务端
func Start(config *Config) (*Process, error) {
	if config.ServerDir == "" {
		home, _ := os.UserHomeDir()
		config.ServerDir = filepath.Join(home, ".mc-connector", "servers", config.MCVersion)
	}
	os.MkdirAll(config.ServerDir, 0755)

	// 写 EULA
	eulaPath := filepath.Join(config.ServerDir, "eula.txt")
	if _, err := os.Stat(eulaPath); os.IsNotExist(err) {
		os.WriteFile(eulaPath, []byte("eula=true\n"), 0644)
	}

	// 写 server.properties
	propsPath := filepath.Join(config.ServerDir, "server.properties")
	props := generateProperties(config)
	os.WriteFile(propsPath, []byte(props), 0644)

	// JVM 参数
	jvmArgs := []string{
		fmt.Sprintf("-Xms%dM", config.MemoryMB/2),
		fmt.Sprintf("-Xmx%dM", config.MemoryMB),
		"-XX:+UseG1GC",
		"-XX:+ParallelRefProcEnabled",
		"-XX:MaxGCPauseMillis=200",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+DisableExplicitGC",
		"-jar", config.JarPath,
		"nogui",
	}

	cmd := exec.Command(config.JavaPath, jvmArgs...)
	cmd.Dir = config.ServerDir
	cmd.Stdout = &logWriter{prefix: "[MC] "}
	cmd.Stderr = &logWriter{prefix: "[MC:ERR] "}

	// 创建 stdin 管道，用于发送命令
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdin 管道失败: %w", err)
	}

	// Windows: 分离进程，防止父进程退出时子进程被一起杀死
	setDetached(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 MC 服务端失败: %w", err)
	}

	p := &Process{
		cmd:    cmd,
		config: config,
		logCh:  make(chan string, 100),
		done:   make(chan struct{}),
		stdin:  stdin,
	}

	go func() {
		cmd.Wait()
		close(p.done)
	}()

	log.Printf("MC 服务端已启动: %s (PID %d)", config.MCVersion, cmd.Process.Pid)
	return p, nil
}

// Stop 停止服务端（发送 /stop 命令后优雅关闭）
func (p *Process) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	// 发送 SIGTERM 优雅关闭
	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		return p.cmd.Process.Kill()
	}
	p.cmd.Wait()
	log.Println("MC 服务端已停止")
	return nil
}

// Running 检查是否在运行
func (p *Process) Running() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	select {
	case <-p.done:
		return false
	default:
		return true
	}
}

// generateProperties 生成 server.properties 内容
func generateProperties(c *Config) string {
	lines := []string{
		fmt.Sprintf("server-port=%d", c.Port),
		fmt.Sprintf("gamemode=%s", c.GameMode),
		fmt.Sprintf("difficulty=%s", c.Difficulty),
		fmt.Sprintf("max-players=%d", c.MaxPlayers),
		"server-ip=127.0.0.1", // 仅本地访问（安全）
		"online-mode=false",
		"allow-nether=true",
		"allow-flight=false",
		"spawn-animals=true",
		"spawn-monsters=true",
		"spawn-npcs=true",
		"view-distance=12",
		"simulation-distance=8",
		"enable-query=false",
		"enable-rcon=false",
		"enable-jmx-monitoring=false",
	}
	return strings.Join(lines, "\n")
}

// logWriter 日志写入器
type logWriter struct {
	prefix string
}

func (w *logWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\r\n")
	if msg != "" {
		log.Printf("%s%s", w.prefix, msg)
	}
	return len(p), nil
}
