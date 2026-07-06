// LAN 广播检测
//
// 监听 Minecraft "对局域网开放" 的 UDP 组播广播，
// 自动获取集成服务器的端口号。
//
// MC 1.3.1+ 通过 UDP 组播 224.0.2.60:4445 广播:
//
//	[MOTD]玩家名 - 世界名[/MOTD][AD]端口号[/AD]
//
// 频率约 1 次/秒，UTF-8 编码。

package tunnel

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var lanPortRegex = regexp.MustCompile(`\[AD\](\d+)\[/AD\]`)

// DetectLANPort 通过 UDP 组播检测 MC LAN 端口
//
// 监听最多 timeout 时间，如果检测到广播包则返回端口号。
// 如果超时或组播不可用，回退到进程端口枚举。
func DetectLANPort(timeout time.Duration) (int, string, error) {
	port, motd := detectViaMulticast(timeout)
	if port > 0 {
		return port, motd, nil
	}

	// 回退：枚举 javaw.exe 的监听端口
	port, err := detectViaNetstat()
	if err != nil {
		return 0, "", fmt.Errorf("未检测到 MC LAN 端口（组播和进程扫描均失败）")
	}
	return port, "", nil
}

// detectViaMulticast 通过组播检测 LAN 端口
func detectViaMulticast(timeout time.Duration) (int, string) {
	addr, err := net.ResolveUDPAddr("udp", "224.0.2.60:4445")
	if err != nil {
		return 0, ""
	}

	// 监听所有网络接口的 UDP
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		// 尝试普通 UDP 监听
		conn2, err2 := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
		if err2 != nil {
			return 0, ""
		}
		conn = conn2
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, 512)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		data := string(buf[:n])
		if matches := lanPortRegex.FindStringSubmatch(data); len(matches) >= 2 {
			port, err := strconv.Atoi(matches[1])
			if err != nil || port < 1024 || port > 65535 {
				continue
			}

			// 提取 MOTD（可选）
			motd := ""
			if idx := strings.Index(data, "[MOTD]"); idx >= 0 {
				end := strings.Index(data, "[/MOTD]")
				if end > idx {
					motd = data[idx+6 : end]
				}
			}

			return port, motd
		}
	}

	return 0, ""
}

// detectViaNetstat 通过进程端口枚举回退检测
func detectViaNetstat() (int, error) {
	// 方法 1: PowerShell 查 java.exe 的监听端口
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-NetTCPConnection -OwningProcess (Get-Process java -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Id) -ErrorAction SilentlyContinue | Where-Object State -eq 'Listen' | Select-Object -ExpandProperty LocalPort | Sort-Object -Unique) -join ','`)
	out, err := cmd.Output()
	if err == nil {
		ports := strings.TrimSpace(string(out))
		if ports != "" {
			for _, p := range strings.Split(ports, ",") {
				port, err := strconv.Atoi(strings.TrimSpace(p))
				if err == nil && port > 1024 && port < 65535 && port != 25565 {
					return port, nil
				}
			}
		}
	}

	// 方法 2: netstat 找 java.exe 的非标准端口
	cmd2 := exec.Command("cmd", "/c", "netstat -ano | findstr LISTENING | findstr java")
	out2, err2 := cmd2.Output()
	if err2 == nil {
		lines := strings.Split(string(out2), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				addr := parts[1]
				if idx := strings.LastIndex(addr, ":"); idx >= 0 {
					portStr := addr[idx+1:]
					port, err := strconv.Atoi(portStr)
					if err == nil && port > 1024 && port < 65535 && port != 25565 {
						return port, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("未找到 java 监听端口")
}
