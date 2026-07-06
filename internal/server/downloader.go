package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const paperAPI = "https://fill.papermc.io/v3/projects/paper"

func DownloadPaper(mcVersion, cacheDir string) (string, error) {
	os.MkdirAll(cacheDir, 0755)

	// 尝试多个 Paper 版本名: 完全匹配 → 二级版本 → 一级版本
	paperVersions := []string{mcVersion}
	parts := strings.Split(mcVersion, ".")
	if len(parts) >= 3 {
		paperVersions = append(paperVersions, strings.Join(parts[:2], "."))
	}
	if len(parts) >= 2 {
		paperVersions = append(paperVersions, strings.Join(parts[:2], "."))
	}

	home, _ := os.UserHomeDir()

	// 先检查本地缓存
	for _, pv := range paperVersions {
		jarName := fmt.Sprintf("paper-%s.jar", pv)
		for _, d := range []string{
			cacheDir,
			filepath.Join(home, ".enericraft", "servers", mcVersion),
			filepath.Join(home, ".enericraft", "servers", pv),
		} {
			jp := filepath.Join(d, jarName)
			if info, err := os.Stat(jp); err == nil && info.Size() > 1000000 {
				log.Printf("Paper 已缓存: %s (%.1fMB)", jp, float64(info.Size())/(1024*1024))
				return jp, nil
			}
		}
	}

	// 从 Paper Fill v3 API 下载
	var lastErr error
	for _, pv := range paperVersions {
		jarPath, err := downloadFromFill(pv, cacheDir)
		if err == nil {
			return jarPath, nil
		}
		lastErr = err
		log.Printf("Paper %s 下载失败: %v, 尝试下一个版本...", pv, err)
	}

	return "", fmt.Errorf("Paper 下载失败 (已尝试 %v): %w\n请手动下载 Paper JAR 放到:\n%s",
		paperVersions, lastErr, filepath.Join(home, ".enericraft", "servers", mcVersion))
}

func downloadFromFill(mcVersion, cacheDir string) (string, error) {
	// 步骤1: 获取版本信息（build 列表）
	verURL := fmt.Sprintf("%s/versions/%s", paperAPI, mcVersion)
	resp, err := http.Get(verURL)
	if err != nil {
		return "", fmt.Errorf("API 不可达: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("Paper 没有 %s 版本", mcVersion)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API 返回 HTTP %d", resp.StatusCode)
	}

	var verInfo struct {
		Builds []int `json:"builds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&verInfo); err != nil {
		return "", fmt.Errorf("解析版本信息失败: %w", err)
	}
	if len(verInfo.Builds) == 0 {
		return "", fmt.Errorf("MC %s 没有可用的 Paper Build", mcVersion)
	}

	// 取最新 build
	latestBuild := verInfo.Builds[len(verInfo.Builds)-1]

	// 步骤2: 获取 build 详情（包含下载 URL）
	buildURL := fmt.Sprintf("%s/versions/%s/builds/%d", paperAPI, mcVersion, latestBuild)
	resp2, err := http.Get(buildURL)
	if err != nil {
		return "", fmt.Errorf("获取 build 信息失败: %w", err)
	}
	defer resp2.Body.Close()

	var buildInfo struct {
		Downloads map[string]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			Size int64  `json:"size"`
		} `json:"downloads"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&buildInfo); err != nil {
		return "", fmt.Errorf("解析 build 信息失败: %w", err)
	}

	// 获取 server JAR 下载链接
	dl, ok := buildInfo.Downloads["server:default"]
	if !ok {
		return "", fmt.Errorf("build %d 没有 server JAR", latestBuild)
	}

	log.Printf("下载 Paper %s build #%d (%.1fMB): %s",
		mcVersion, latestBuild, float64(dl.Size)/(1024*1024), dl.URL)

	// 步骤3: 下载
	resp3, err := http.Get(dl.URL)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != 200 {
		return "", fmt.Errorf("下载失败 HTTP %d", resp3.StatusCode)
	}

	jarName := fmt.Sprintf("paper-%s.jar", mcVersion)
	jarPath := filepath.Join(cacheDir, jarName)
	tmpPath := jarPath + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(f, resp3.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", err
	}
	f.Close()

	if err := os.Rename(tmpPath, jarPath); err != nil {
		return "", err
	}

	log.Printf("Paper 下载完成: %s (%.1fMB)", jarPath, float64(dl.Size)/(1024*1024))
	return jarPath, nil
}
