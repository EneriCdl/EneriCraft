// STUN 客户端
//
// 实现 RFC 5389 STUN Binding Request，
// 用于获取公网 IP:Port 和判断 NAT 类型。

package tunnel

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const (
	stunMagicCookie = 0x2112A442
	stunTimeout     = 3 * time.Second

	// STUN 方法
	methodBinding = 0x0001

	// STUN 属性类型
	attrMappedAddress  = 0x0001
	attrXorMappedAddr  = 0x0020
	attrChangeRequest  = 0x0003
	attrResponseOrigin = 0x802b
)

// 公共 STUN 服务器
var PublicSTUNServers = []string{
	"stun.l.google.com:19302",
	"stun1.l.google.com:19302",
}

// StunResponse STUN 响应
type StunResponse struct {
	PublicIP   string
	PublicPort int
}

// SendBindingRequest 发送 STUN Binding Request
//
// 返回从 STUN 服务器视角看到的公网地址。
func SendBindingRequest(stunServer string) (*StunResponse, error) {
	conn, err := net.DialTimeout("udp", stunServer, stunTimeout)
	if err != nil {
		return nil, fmt.Errorf("连接 STUN 服务器失败: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(stunTimeout))

	// 构造 STUN Binding Request (RFC 5389)
	request := buildBindingRequest()

	if _, err := conn.Write(request); err != nil {
		return nil, fmt.Errorf("发送 STUN 请求失败: %w", err)
	}

	// 读取响应
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("读取 STUN 响应失败: %w", err)
	}

	return parseBindingResponse(buf[:n])
}

// buildBindingRequest 构造 20 字节 STUN Binding Request
func buildBindingRequest() []byte {
	// STUN 消息头: 20 字节
	//  0-1:  消息类型 = 0x0001 (Binding Request)
	//  2-3:  消息长度 = 0 (无属性)
	//  4-7:  Magic Cookie = 0x2112A442
	//  8-19: Transaction ID (12 字节随机数)

	msg := make([]byte, 20)
	binary.BigEndian.PutUint16(msg[0:2], 0x0001) // Binding Request
	binary.BigEndian.PutUint16(msg[2:4], 0)      // Length = 0
	binary.BigEndian.PutUint32(msg[4:8], stunMagicCookie)

	// 生成简单 Transaction ID（生产环境应用 crypto/rand）
	txID := uint64(time.Now().UnixNano())
	binary.BigEndian.PutUint64(msg[8:16], txID)

	return msg
}

// parseBindingResponse 解析 STUN Binding Response
func parseBindingResponse(data []byte) (*StunResponse, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("STUN 响应太短: %d 字节", len(data))
	}

	// 验证消息类型（Binding Success Response = 0x0101）
	msgType := binary.BigEndian.Uint16(data[0:2])
	if msgType != 0x0101 {
		return nil, fmt.Errorf("非 Binding Success Response: 0x%04x", msgType)
	}

	// 验证 Magic Cookie
	cookie := binary.BigEndian.Uint32(data[4:8])
	if cookie != stunMagicCookie {
		return nil, fmt.Errorf("STUN Magic Cookie 不匹配")
	}

	// 消息长度
	msgLen := int(binary.BigEndian.Uint16(data[2:4]))
	attrs := data[20 : 20+msgLen]

	// 解析属性，查找 XOR-MAPPED-ADDRESS 或 MAPPED-ADDRESS
	for i := 0; i+4 <= len(attrs); {
		attrType := binary.BigEndian.Uint16(attrs[i : i+2])
		attrLen := int(binary.BigEndian.Uint16(attrs[i+2 : i+4]))
		paddedLen := (attrLen + 3) & ^3 // 4字节对齐

		if i+4+attrLen > len(attrs) {
			break
		}

		switch attrType {
		case attrXorMappedAddr:
			return parseXorMappedAddress(attrs[i+4:i+4+attrLen], data[4:20])
		case attrMappedAddress:
			return parseMappedAddress(attrs[i+4 : i+4+attrLen])
		}

		i += 4 + paddedLen
	}

	return nil, fmt.Errorf("未找到 MAPPED-ADDRESS 属性")
}

// parseXorMappedAddress 解析 XOR-MAPPED-ADDRESS (RFC 5389)
func parseXorMappedAddress(data []byte, magicAndID []byte) (*StunResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("XOR-MAPPED-ADDRESS 太短")
	}

	// 第一个字节忽略，第二个字节是地址族 (0x01 = IPv4)
	family := data[1]
	if family != 0x01 {
		return nil, fmt.Errorf("不支持的地址族: %d", family)
	}

	// 端口: XOR 前2字节的 Magic Cookie
	port := binary.BigEndian.Uint16(data[2:4])
	port ^= uint16(stunMagicCookie >> 16)

	// IP: XOR Magic Cookie
	ip := binary.BigEndian.Uint32(data[4:8])
	ip ^= binary.BigEndian.Uint32(magicAndID[0:4])

	return &StunResponse{
		PublicIP:   fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)),
		PublicPort: int(port),
	}, nil
}

// parseMappedAddress 解析 MAPPED-ADDRESS (传统格式)
func parseMappedAddress(data []byte) (*StunResponse, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("MAPPED-ADDRESS 太短")
	}
	port := binary.BigEndian.Uint16(data[2:4])
	ip := net.IPv4(data[4], data[5], data[6], data[7])

	return &StunResponse{
		PublicIP:   ip.String(),
		PublicPort: int(port),
	}, nil
}
