package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type JavaInfo struct {
	Path     string
	Version  string
	MajorVer int
}

func parseJava(javaPath string) *JavaInfo {
	cmd := exec.Command(javaPath, "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ❌ 无法运行 %s: %v\n", javaPath, err)
		return nil
	}
	output := string(out)
	version := ""
	if idx := strings.Index(output, "version \""); idx >= 0 {
		start := idx + len("version \"")
		if end := strings.Index(output[start:], "\""); end >= 0 {
			version = output[start : start+end]
		}
	}
	if version == "" {
		return nil
	}
	major := 0
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		if v := parts[0]; v == "1" && len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &major)
		} else {
			fmt.Sscanf(v, "%d", &major)
		}
	}
	return &JavaInfo{Path: javaPath, Version: version, MajorVer: major}
}

func main() {
	fmt.Println("=== Java 检测测试 ===")

	// 测试 MC runtime Java
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		mcRuntime := filepath.Join(appdata, ".minecraft", "runtime")
		fmt.Printf("扫描: %s\n", mcRuntime)
		if entries, err := os.ReadDir(mcRuntime); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					javaExe := filepath.Join(mcRuntime, entry.Name(), "bin", "java.exe")
					fmt.Printf("  检查: %s\n", javaExe)
					if _, err := os.Stat(javaExe); err == nil {
						fmt.Printf("  文件存在 ✅\n")
						if j := parseJava(javaExe); j != nil {
							fmt.Printf("  → Java %d (%s)\n", j.MajorVer, j.Version)
						}
					} else {
						fmt.Printf("  文件不存在 ❌ (%v)\n", err)
					}
				}
			}
		} else {
			fmt.Printf("  读取目录失败: %v\n", err)
		}
	}

	// 测试系统 Java
	fmt.Println("\n=== 系统 PATH Java ===")
	if j := parseJava("java"); j != nil {
		fmt.Printf("  → Java %d (%s) at %s\n", j.MajorVer, j.Version, j.Path)
	}

	fmt.Println("\n=== 常见安装路径 ===")
	for _, base := range []string{
		"C:\\Program Files\\Java",
		"C:\\Program Files\\Microsoft",
	} {
		if _, err := os.Stat(base); os.IsNotExist(err) {
			continue
		}
		filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if info.Name() == "java.exe" {
				if j := parseJava(path); j != nil {
					fmt.Printf("  → Java %d (%s) at %s\n", j.MajorVer, j.Version, path)
				}
			}
			return nil
		})
	}
}
