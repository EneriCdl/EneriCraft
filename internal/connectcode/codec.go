package connectcode

import (
	"bytes"
	"compress/zlib"
	"log"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"
)

const (
	CodePrefix = "EC1-" // 短前缀: EneriCraft v1
	MaxAge     = 24 * time.Hour
)

type Code struct {
	V  int        `json:"v"`
	EP []Endpoint `json:"ep"`
	MV string     `json:"mv"`
	MH string     `json:"mh"`
	TS int64      `json:"ts"`
}

type Endpoint struct {
	IP   string `json:"ip"`
	Port int    `json:"p"`
}

// Generate 生成连接码
func Generate(endpoints []Endpoint, _ []byte, mcVersion, modHash string) (string, error) {
	// 生成简短随机 ID（不用完整公钥，太长容易被微信截断）
	sessionID, _ := rand.Int(rand.Reader, big.NewInt(99999999))
	_ = sessionID

	code := Code{
		V:  1,
		EP: endpoints,
		MV: mcVersion,
		MH: modHash,
		TS: time.Now().Unix(),
	}

	log.Println("*** USING NEW CODEC: EC1- prefix ***"); jsonData, err := json.Marshal(code)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %w", err)
	}

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(jsonData); err != nil {
		return "", fmt.Errorf("压缩失败: %w", err)
	}
	w.Close()

	// 用标准 Base64（不用 RawURLEncoding，兼容性更好）
	encoded := base64.StdEncoding.EncodeToString(compressed.Bytes())

	return CodePrefix + encoded, nil
}

func Parse(raw string) (*Code, error) {
	// 支持 EC1-（连接码）和 EP1-（回执码）
	if len(raw) < 4 { return nil, fmt.Errorf("无效的码，太短") }
	var stripped string
	switch {
	case strings.HasPrefix(raw, "EC1-"):
		stripped = raw[4:]
	case strings.HasPrefix(raw, "EP1-"):
		stripped = raw[4:]
	default:
		return nil, fmt.Errorf("无效的码，应以 EC1- 或 EP1- 开头")
	}

	// 清理：去掉可能的空白字符（微信/QQ可能加换行）
	stripped = cleanBase64(stripped)

	compressed, err := base64.StdEncoding.DecodeString(stripped)
	if err != nil {
		return nil, fmt.Errorf("连接码格式错误，请确认完整复制（未截断）")
	}

	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}
	defer reader.Close()

	jsonData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("连接码不完整，请重新复制（不要漏掉末尾字符）")
	}

	var code Code
	if err := json.Unmarshal(jsonData, &code); err != nil {
		return nil, fmt.Errorf("连接码已损坏: %w", err)
	}

	age := time.Since(time.Unix(code.TS, 0))
	if age > MaxAge {
		return nil, fmt.Errorf("连接码已过期（%.0f小时前）", age.Hours())
	}

	return &code, nil
}

func cleanBase64(s string) string {
	var result []byte
	for _, c := range []byte(s) {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' {
			result = append(result, c)
		}
	}
	return string(result)
}

// GeneratePunchCode 生成回执码（客户端 → 房主）
// 包含客户端的公网地址，让房主也能向客户端打洞
func GeneratePunchCode(ip string, port int) string {
	code := Code{
		V: 2,
		EP: []Endpoint{{IP: ip, Port: port}},
		TS: time.Now().Unix(),
	}
	jsonData, _ := json.Marshal(code)
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(jsonData)
	w.Close()
	return "EP1-" + base64.StdEncoding.EncodeToString(compressed.Bytes())
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
