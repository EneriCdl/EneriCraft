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

func main() {
	searchPaths := []string{"E:/PCL/.minecraft", "D:/.minecraft", "E:/.minecraft"}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		searchPaths = append(searchPaths, filepath.Join(appdata, ".minecraft"))
	}

	for _, mcDir := range searchPaths {
		entries, err := os.ReadDir(filepath.Join(mcDir, "versions"))
		if err != nil {
			continue
		}
		fmt.Printf("📁 %s:\n", mcDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			vid := entry.Name()
			clean := isCleanMCVersion(vid)
			mark := "❌"
			if clean {
				mark = "✅"
			}
			fmt.Printf("  %s %s\n", mark, vid)
		}
	}
}
