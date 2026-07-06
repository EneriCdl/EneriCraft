//go:build windows

package tunnel

import (
	"fmt"
	"os/exec"
)

// AddFirewallRule 添加 Windows 防火墙入站规则
func AddFirewallRule(port int) error {
	ruleName := fmt.Sprintf("EneriCraft QUIC UDP %d", port)
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name="+ruleName,
		"dir=in",
		"action=allow",
		"protocol=UDP",
		fmt.Sprintf("localport=%d", port),
		"profile=any",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("添加防火墙规则失败: %w", err)
	}
	return nil
}

// RemoveFirewallRule 删除防火墙规则
func RemoveFirewallRule(port int) {
	ruleName := fmt.Sprintf("EneriCraft QUIC UDP %d", port)
	exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
		"name="+ruleName).Run()
}
