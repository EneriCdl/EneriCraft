package main

import (
	"fmt"
	"os"
	"path/filepath"
)

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

func getLatestFromFiles() string {
	searchPaths := []string{"E:/PCL/.minecraft", "D:/.minecraft", "E:/.minecraft"}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		searchPaths = append(searchPaths, filepath.Join(appdata, ".minecraft"))
	}
	var best string
	for _, mcDir := range searchPaths {
		entries, err := os.ReadDir(filepath.Join(mcDir, "versions"))
		if err != nil { continue }
		for _, entry := range entries {
			if !entry.IsDir() { continue }
			vid := entry.Name()
			if !isCleanMCVersion(vid) { continue }
			if best == "" || compareVerStr(vid, best) > 0 { best = vid }
		}
	}
	return best
}

func main() {
	fmt.Println("=== getLatestFromFiles 测试 ===")
	result := getLatestFromFiles()
	fmt.Printf("最新版本: '%s'\n", result)
	fmt.Printf("是否干净: %v\n", isCleanMCVersion(result))

	fmt.Println("\n=== 所有版本检查 ===")
	searchPaths := []string{"E:/PCL/.minecraft", "D:/.minecraft", "E:/.minecraft"}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		searchPaths = append(searchPaths, filepath.Join(appdata, ".minecraft"))
	}
	for _, mcDir := range searchPaths {
		entries, err := os.ReadDir(filepath.Join(mcDir, "versions"))
		if err != nil { continue }
		for _, entry := range entries {
			if !entry.IsDir() { continue }
			vid := entry.Name()
			clean := isCleanMCVersion(vid)
			mark := "❌"
			if clean { mark = "✅" }
			if clean {
				fmt.Printf("  %s %s (纯数字版本)\n", mark, vid)
			}
		}
	}
}
