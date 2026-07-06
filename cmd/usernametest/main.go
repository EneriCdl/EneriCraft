package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process -Filter "name='java.exe'" | ForEach-Object { $_.CommandLine }`)
	data, err := cmd.Output()
	if err != nil {
		fmt.Printf("PowerShell failed: %v\n", err)
		return
	}

	output := string(data)
	fmt.Printf("Output length: %d chars\n", len(output))
	fmt.Printf("First 300 chars: %s\n\n", output[:min(300, len(output))])

	for i, line := range strings.Split(output, "\n") {
		fmt.Printf("Line %d (len=%d)\n", i, len(line))

		if strings.Contains(line, "paper-") || strings.Contains(line, "--nogui") {
			fmt.Println("  SKIP (paper/nogui)")
			continue
		}

		if idx := strings.Index(line, "--username "); idx >= 0 {
			rest := line[idx+len("--username "):]
			fmt.Printf("  Found --username at %d, rest=%q\n", idx, rest[:min(20, len(rest))])
			if end := strings.IndexAny(rest, " \n\r"); end > 0 {
				u := strings.TrimSpace(rest[:end])
				fmt.Printf("  Username = %q\n", u)
			}
		}

		if idx := strings.Index(line, "--version "); idx >= 0 {
			rest := line[idx+len("--version "):]
			fmt.Printf("  Found --version at %d, rest=%q\n", idx, rest[:min(20, len(rest))])
			if end := strings.IndexAny(rest, " \n\r"); end > 0 {
				v := strings.TrimSpace(rest[:end])
				fmt.Printf("  Version = %q\n", v)
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
