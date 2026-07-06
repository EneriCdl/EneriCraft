//go:build windows

package auth

import (
	"log"
	"os"
	"strings"
)

// SetupHosts 将 Mojang 认证域名重定向到本地
// 需要管理员权限
func SetupHosts() {
	hostsPath := os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		log.Printf("[Auth] 无法读取 hosts 文件: %v", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	toAdd := []string{
		"127.0.0.1 sessionserver.mojang.com  # EneriCraft",
		"127.0.0.1 authserver.mojang.com     # EneriCraft",
		"127.0.0.1 api.mojang.com            # EneriCraft",
	}

	changed := false
	for _, entry := range toAdd {
		domain := strings.Fields(entry)[1]
		found := false
		for _, line := range lines {
			if strings.Contains(line, domain) && strings.Contains(line, "EneriCraft") {
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, entry)
			changed = true
			log.Printf("[Auth] 添加 hosts: %s → 127.0.0.1", domain)
		}
	}

	if changed {
		newData := strings.Join(lines, "\n")
		if err := os.WriteFile(hostsPath, []byte(newData), 0644); err != nil {
			log.Printf("[Auth] 写入 hosts 失败: %v (需要管理员权限)", err)
		} else {
			log.Println("[Auth] Hosts 文件已更新，Mojang 验证已绕过")
		}
	}
}
