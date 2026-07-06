// Java 运行时检测
//
// 扫描常见的 Java 安装路径，检测版本。

package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// JavaInfo Java 安装信息
type JavaInfo struct {
	Path        string `json:"path"`
	Version     string `json:"version"`
	MajorVer    int    `json:"major_version"`
	Is64Bit     bool   `json:"is_64bit"`
}

// DetectJava 检测系统已安装的 Java
func DetectJava() []JavaInfo {
	var results []JavaInfo
	seen := map[string]bool{}

	addJava := func(path string) {
		if seen[path] { return }
		j := parseJava(path)
		if j != nil {
			seen[path] = true
			results = append(results, *j)
		}
	}

	// 常见 Java 安装路径 (Windows)
	searchPaths := []string{
		"C:\\Program Files\\Java",
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Amazon Corretto",
		"C:\\Program Files\\Microsoft",
	}

	// 添加 MC 启动器自带的 Java runtime
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		mcRuntime := filepath.Join(appdata, ".minecraft", "runtime")
		if entries, err := os.ReadDir(mcRuntime); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					javaExe := filepath.Join(mcRuntime, entry.Name(), "bin", "java.exe")
					if _, err := os.Stat(javaExe); err == nil {
						addJava(javaExe)
					}
				}
			}
		}
		// PCL 的 .minecraft 在别的盘
		for _, drive := range []string{"D:", "E:"} {
			pclRuntime := filepath.Join(drive, "PCL", ".minecraft", "runtime")
			if entries, err := os.ReadDir(pclRuntime); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						javaExe := filepath.Join(pclRuntime, entry.Name(), "bin", "java.exe")
						if _, err := os.Stat(javaExe); err == nil {
							addJava(javaExe)
						}
					}
				}
			}
		}
	}

	for _, base := range searchPaths {
		if _, err := os.Stat(base); os.IsNotExist(err) { continue }
		filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil { return nil }
			if info.Name() == "java.exe" {
				addJava(path)
			}
			if strings.Count(path[len(base):], string(os.PathSeparator)) > 2 {
				return filepath.SkipDir
			}
			return nil
		})
	}

	// 尝试 PATH 中的 java
	addJava("java")

	return results
}

// FindBestJava 查找最适合目标 MC 版本的 Java
func FindBestJava(mcVersion string) *JavaInfo {
	javas := DetectJava()
	required := requiredJavaVersion(mcVersion)

	var best *JavaInfo
	for i := range javas {
		j := &javas[i]
		if j.MajorVer >= required {
			if best == nil || j.MajorVer > best.MajorVer {
				best = j
			}
		}
	}
	return best
}

// parseJava 解析 Java 版本
func parseJava(javaPath string) *JavaInfo {
	cmd := exec.Command(javaPath, "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	// 解析输出: java version "17.0.1" 2021-10-19 LTS
	// 或: openjdk version "21.0.1" 2023-10-17 LTS
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

	// 解析主版本号
	major := 0
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		if v := parts[0]; v == "1" && len(parts) > 1 {
			// Java 8 格式: 1.8.0_291
			fmt.Sscanf(parts[1], "%d", &major)
		} else {
			fmt.Sscanf(v, "%d", &major)
		}
	}

	return &JavaInfo{
		Path:     javaPath,
		Version:  version,
		MajorVer: major,
		Is64Bit:  strings.Contains(output, "64-Bit"),
	}
}

// RequiredJava 返回 MC 版本所需的最低 Java 主版本
func RequiredJava(mcVersion string) int {
	return requiredJavaVersion(mcVersion)
}

func requiredJavaVersion(mcVersion string) int {
	parts := strings.Split(mcVersion, ".")
	if len(parts) < 2 {
		return 21
	}

	major := 0
	minor := 0
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(parts[1], "%d", &minor)

	switch {
	case major == 1 && minor <= 16:
		return 8
	case major == 1 && minor == 17:
		return 16
	case major == 1 && minor <= 20:
		return 17
	default:
		// MC 1.21+ 需要 Java 21
		return 21
	}
}
