package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mcserver "github.com/mc-connector/internal/server"
)

type ServerStatus struct {
	Running       bool    `json:"running"`
	Version       string  `json:"version"`
	TPS           float64 `json:"tps"`
	MemoryUsage   float64 `json:"memory_usage"`
	UptimeSecs    uint64  `json:"uptime_secs"`
	PlayersOnline int     `json:"players_online"`
	MaxPlayers    int     `json:"max_players"`
}

func (a *APIHandler) StartServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "仅支持 POST")
		return
	}
	log.Println("API: 启动 MC 服务端")
	writeJSON(w, ServerStatus{Running: true, Version: "1.21", TPS: 20.0, MaxPlayers: 8})
}

func (a *APIHandler) StopServer(w http.ResponseWriter, r *http.Request) {
	log.Println("API: 停止 MC 服务端")
	writeJSON(w, map[string]string{"status": "ok"})
}

func (a *APIHandler) ServerStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, ServerStatus{
		Running: IsServerRunning(),
		Version: GetState().MCVersion,
	})
}

// ServerCommand 向服务端发送控制台命令
func (a *APIHandler) ServerCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "仅支持 POST")
		return
	}
	var req struct{ Command string }
	json.NewDecoder(r.Body).Decode(&req)
	if req.Command == "" {
		writeError(w, 400, "请输入命令")
		return
	}

	state.mu.RLock()
	proc := state.ServerProc
	state.mu.RUnlock()

	if proc == nil || !proc.Running() {
		writeError(w, 400, "服务端未运行")
		return
	}

	if err := proc.SendCommand(req.Command); err != nil {
		log.Printf("[命令] 发送失败: %v", err)
		writeError(w, 500, "命令发送失败: "+err.Error())
		return
	}

	log.Printf("[命令] %s", req.Command)
	writeJSON(w, map[string]string{"status": "ok", "command": req.Command})
}

func (a *APIHandler) DetectVersions(w http.ResponseWriter, r *http.Request) {
	// 只从运行进程检测，不做文件回退
	runningVer, runningDir := detectProcessOnly()
	fromProcess := runningVer != ""

	// 也列出所有已安装版本
	var versions []map[string]interface{}
	searchPaths := mcserver.GetAllMinecraftDirs()
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		searchPaths = append(searchPaths, filepath.Join(appdata, ".minecraft"))
	}

	bestDir := runningDir
	bestCount := -1

	for _, mcDir := range searchPaths {
		if _, err := os.Stat(mcDir); os.IsNotExist(err) {
			continue
		}
		versionsDir := filepath.Join(mcDir, "versions")
		entries, err := os.ReadDir(versionsDir)
		if err != nil {
			continue
		}
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			count++
			vid := entry.Name()
			vtype := "unknown"
			if data, err := os.ReadFile(filepath.Join(versionsDir, vid, vid+".json")); err == nil {
				var info struct {
					Type         string `json:"type"`
					InheritsFrom string `json:"inheritsFrom"`
				}
				if json.Unmarshal(data, &info) == nil {
					if info.Type != "" {
						vtype = info.Type
					} else if info.InheritsFrom != "" {
						vtype = "modded"
					}
				}
			}
			versions = append(versions, map[string]interface{}{
				"id": vid, "type": vtype, "dir": mcDir,
			})
		}
		if count > bestCount {
			bestCount = count
			if bestDir == "" {
				bestDir = mcDir
			}
		}
	}

	if bestDir == "" {
		bestDir = filepath.Join(os.Getenv("APPDATA"), ".minecraft")
	}

	// 如果没有检测到运行进程，用文件扫描的最新版本
	latest := runningVer
	if latest == "" {
		latest = getLatestFromFiles()
	}
	writeJSON(w, map[string]interface{}{
		"versions":       versions,
		"latest":         latest,
		"running":        runningVer,
		"from_process":   fromProcess,
		"minecraft_dir":  bestDir,
	})
}

// detectRunningMC 检测正在运行的 Minecraft 进程版本和用户名
func detectRunningMC() (version string, username string) {
	// 方法1: PowerShell 进程检测
	if v, u := detectViaPowerShell(); v != "" {
		return v, u
	}
	// 方法2: 找最近使用过的版本
	if v, _ := detectRecentVersion(); v != "" {
		return v, ""
	}
	return "", ""
}

func detectViaPowerShell() (version string, username string) {
	var data []byte

	// 方法1: Get-CimInstance（需要 WinRM/PowerShell 3.0+）
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process -Filter "name='java.exe'" | ForEach-Object { $_.CommandLine }`)
	data, err := cmd.Output()
	if err != nil || len(data) < 50 {
		// 方法2: Get-WmiObject（更兼容，不需要 WinRM）
		cmd2 := exec.Command("powershell", "-NoProfile", "-Command",
			`Get-WmiObject Win32_Process -Filter "name='java.exe'" | Select-Object -ExpandProperty CommandLine`)
		data2, err2 := cmd2.Output()
		if err2 != nil || len(data2) < 50 {
			// 方法3: wmic（传统方式）
			cmd3 := exec.Command("wmic", "process", "where", "name='java.exe'", "get", "commandline", "/format:list")
			data3, err3 := cmd3.Output()
			if err3 != nil {
				return "", ""
			}
			data = data3
		} else {
			data = data2
		}
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "paper-") || strings.Contains(line, "--nogui") {
			continue
		}
		// 提取 --version
		if version == "" {
			if idx := strings.Index(line, "--version "); idx >= 0 {
				rest := line[idx+len("--version "):]
				if end := strings.IndexAny(rest, " \n\r"); end > 0 {
					v := strings.TrimSpace(rest[:end])
					if isCleanMCVersion(v) {
						version = v
					}
				}
			}
		}
		// 提取 --username
		if username == "" {
			if idx := strings.Index(line, "--username "); idx >= 0 {
				rest := line[idx+len("--username "):]
				if end := strings.IndexAny(rest, " \n\r"); end > 0 {
					username = strings.TrimSpace(rest[:end])
				}
			}
		}
		if version != "" && username != "" {
			return
		}
	}
	return version, username
}

// isCleanMCVersion 检查版本号是否为纯 MC release（如 1.21.11），排除 Fabric/Forge/OptiFine 等
func isCleanMCVersion(vid string) bool {
	if len(vid) == 0 || vid[0] < '0' || vid[0] > '9' {
		return false
	}
	for _, c := range vid {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}
	return true
}

// detectRecentVersion 找最近修改过的纯 MC release 版本目录
func detectRecentVersion() (version string, minecraftDir string) {
	searchPaths := mcserver.GetAllMinecraftDirs()
	var bestVer string
	var bestDir string
	var bestTime int64

	for _, mcDir := range searchPaths {
		entries, err := os.ReadDir(filepath.Join(mcDir, "versions"))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			vid := entry.Name()
			// 只接受纯数字+点的版本号，排除 Fabric/Forge/整合包
			if !isCleanMCVersion(vid) {
				continue
			}

			// 检查版本 JAR 的修改时间
			jarPath := filepath.Join(mcDir, "versions", vid, vid+".jar")
			info, err := os.Stat(jarPath)
			if err != nil {
				info = nil
			}

			if info != nil && info.ModTime().Unix() > bestTime {
				bestTime = info.ModTime().Unix()
				bestVer = vid
				bestDir = mcDir
			}
		}
	}

	return bestVer, bestDir
}

// detectProcessOnly 仅从进程检测，不回落文件
func detectProcessOnly() (version string, minecraftDir string) {
	v, _ := detectViaPowerShell()
	return v, ""
}

// getLatestFromFiles 从文件系统获取最新纯数字 release 版本号
func getLatestFromFiles() string {
	searchPaths := mcserver.GetAllMinecraftDirs()
	var best string
	for _, mcDir := range searchPaths {
		entries, err := os.ReadDir(filepath.Join(mcDir, "versions"))
		if err != nil { continue }
		for _, entry := range entries {
			if !entry.IsDir() { continue }
			vid := entry.Name()
			// 只接受纯数字+点的版本号，排除 Fabric/Forge/整合包
			if !isCleanMCVersion(vid) { continue }
			if best == "" || compareVerStr(vid, best) > 0 { best = vid }
		}
	}
	return best
}

func compareVerStr(a, b string) int {
	pa := parseVer(a)
	pb := parseVer(b)
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] < pb[i] { return -1 }
		if pa[i] > pb[i] { return 1 }
	}
	return len(pa) - len(pb)
}

func parseVer(v string) []int {
	var p []int
	n := 0
	for _, c := range v {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else if c == '.' {
			p = append(p, n)
			n = 0
		} else {
			break
		}
	}
	p = append(p, n)
	return p
}

// execCmd 执行命令并返回输出
func execCmd(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}
