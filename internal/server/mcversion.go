package server

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type MCVersion struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Dir  string `json:"dir"`
}

// GetAllMinecraftDirs 获取所有可能的 .minecraft 目录
func GetAllMinecraftDirs() []string {
	var dirs []string
	seen := map[string]bool{}

	add := func(p string) {
		clean := filepath.Clean(p)
		if seen[clean] {
			return
		}
		seen[clean] = true
		if info, err := os.Stat(clean); err == nil && info.IsDir() {
			dirs = append(dirs, clean)
		}
	}

	// 优先：盘符根目录（PCL2 最常用的位置）
	for _, drive := range []string{"D:\\", "E:\\", "F:\\", "C:\\"} {
		add(filepath.Join(drive, ".minecraft"))
	}

	// PCL/PCL2/HMCL 启动器子目录
	for _, launcher := range []string{"PCL", "PCL2", "HMCL", "BakaXL"} {
		for _, drive := range []string{"D:\\", "E:\\", "F:\\", "C:\\"} {
			add(filepath.Join(drive, launcher, ".minecraft"))
		}
	}

	// 最后才查 APPDATA（通常版本少）
	appdata := os.Getenv("APPDATA")
	if appdata != "" {
		add(filepath.Join(appdata, ".minecraft"))
	}

	return dirs
}

// DetectVersions 扫描所有 .minecraft 目录的版本
func DetectVersions() []MCVersion {
	var all []MCVersion
	seenIDs := map[string]bool{}

	for _, mcDir := range GetAllMinecraftDirs() {
		versionsDir := filepath.Join(mcDir, "versions")
		entries, err := os.ReadDir(versionsDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || seenIDs[entry.Name()] {
				continue
			}

			vid := entry.Name()
			versionDir := filepath.Join(versionsDir, vid)
			jsonPath := filepath.Join(versionDir, vid+".json")
			jarPath := filepath.Join(versionDir, vid+".jar")

			if _, err := os.Stat(jarPath); os.IsNotExist(err) {
				parent := findParentJar(jsonPath, versionsDir)
				if parent != "" {
					jarPath = parent
				} else {
					continue
				}
			}

			v := MCVersion{ID: vid, Dir: mcDir}

			if data, err := os.ReadFile(jsonPath); err == nil {
				var info struct {
					Type         string `json:"type"`
					InheritsFrom string `json:"inheritsFrom"`
				}
				if json.Unmarshal(data, &info) == nil {
					if info.Type != "" {
						v.Type = info.Type
					} else if info.InheritsFrom != "" {
						v.Type = "modded"
					}
				}
			}
			if v.Type == "" {
				v.Type = "unknown"
			}

			_ = jarPath
			all = append(all, v)
			seenIDs[vid] = true
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return compareVersions(all[i].ID, all[j].ID) > 0
	})

	return all
}

func GetLatestVersion() string {
	for _, v := range DetectVersions() {
		if v.Type == "release" || v.Type == "modded" {
			return v.ID
		}
	}
	return ""
}

func GetRunningVersion() string {
	out, err := exec.Command("jps", "-l").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "net.minecraft") {
				_ = line
			}
		}
	}
	return GetLatestVersion()
}

func findParentJar(jsonPath, versionsDir string) string {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return ""
	}
	var info struct {
		InheritsFrom string `json:"inheritsFrom"`
	}
	if json.Unmarshal(data, &info) != nil || info.InheritsFrom == "" {
		return ""
	}
	parentJar := filepath.Join(versionsDir, info.InheritsFrom, info.InheritsFrom+".jar")
	if _, err := os.Stat(parentJar); err == nil {
		return parentJar
	}
	return ""
}

func FindMinecraftDir() string {
	best := ""
	bestCount := -1
	for _, dir := range GetAllMinecraftDirs() {
		entries, err := os.ReadDir(filepath.Join(dir, "versions"))
		if err != nil {
			continue
		}
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				count++
			}
		}
		if count > bestCount {
			bestCount = count
			best = dir
		}
	}
	if best != "" {
		return best
	}
	appdata := os.Getenv("APPDATA")
	if appdata != "" {
		return filepath.Join(appdata, ".minecraft")
	}
	return ""
}

func compareVersions(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	n := len(pa)
	if len(pb) < n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) []int {
	var parts []int
	for _, s := range strings.Split(v, ".") {
		n := 0
		for _, c := range s {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			} else {
				break
			}
		}
		parts = append(parts, n)
	}
	return parts
}
