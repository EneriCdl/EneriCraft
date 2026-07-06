package web

import (
	"embed"
	"io/fs"
	"log"
	"os"
)

//go:embed dist/*
var embeddedFrontend embed.FS

// Frontend 返回前端文件系统
// 开发模式：从 frontend/dist 读取
// 生产模式：从内嵌文件读取
func Frontend() fs.FS {
	f, err := fs.Sub(embeddedFrontend, "dist")
	if err != nil {
		log.Printf("⚠ 前端未内嵌，从文件系统读取")
		return os.DirFS("frontend/dist")
	}
	return f
}
