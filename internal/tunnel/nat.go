// NAT 类型探测
//
// 基于 RFC 5780 算法，通过向 STUN 服务器发送多个请求判断 NAT 类型。

package tunnel

import (
	"fmt"
	"net"
)

// NatType NAT 类型
type NatType int

const (
	NatUnknown           NatType = iota // 未知
	NatOpen                             // 公网直连 (无 NAT)
	NatFullCone                         // 完全锥形
	NatRestrictedCone                   // 受限锥形
	NatPortRestrictedCone               // 端口受限锥形
	NatSymmetric                        // 对称 NAT (最难打洞)
)

func (n NatType) String() string {
	switch n {
	case NatOpen:
		return "公网直连"
	case NatFullCone:
		return "完全锥形 NAT"
	case NatRestrictedCone:
		return "受限锥形 NAT"
	case NatPortRestrictedCone:
		return "端口受限锥形 NAT"
	case NatSymmetric:
		return "对称 NAT"
	default:
		return "未知"
	}
}

// CanP2P 判断该 NAT 类型是否支持 P2P 打洞
func (n NatType) CanP2P() bool {
	return n == NatOpen || n == NatFullCone || n == NatRestrictedCone || n == NatPortRestrictedCone
}

// DetectNat 探测本机 NAT 类型
//
// 简化算法 (使用单个 STUN 服务器):
//  1. 发送 Binding Request → 获取公网 IP:Port (A:P1)
//  2. 从另一个本地端口发送 → 获取公网 IP:Port (A:P2)
//  3. 比较:
//     - 如果两个公网地址相同但端口不同 → 端口受限锥形或对称
//     - 精确判断需要两个 STUN 服务器配合 CHANGE-REQUEST
//
// 完整算法需要两个不同 IP 的 STUN 服务器，这里简化实现。
func DetectNat() (NatType, error) {
	// 使用第一个 STUN 服务器探测两次
	resp1, err := SendBindingRequest(PublicSTUNServers[0])
	if err != nil {
		return NatUnknown, fmt.Errorf("第一次探测失败: %w", err)
	}

	// 第二次探测 (从不同源端口)
	resp2, err := SendBindingRequest(PublicSTUNServers[0])
	if err != nil {
		return NatUnknown, fmt.Errorf("第二次探测失败: %w", err)
	}

	// 判断逻辑
	// 注意: 这是简化实现。完整的 NAT 类型判断需要两个不同 IP 的 STUN 服务器。
	if resp1.PublicIP == resp2.PublicIP && resp1.PublicPort == resp2.PublicPort {
		// 两次映射相同 → 完全锥形或受限锥形
		return NatFullCone, nil
	}

	// 端口不同但 IP 相同 → 对称 NAT 或端口受限锥形
	// 无法仅凭一个 STUN 服务器区分，保守判断为对称
	if resp1.PublicIP == resp2.PublicIP && resp1.PublicPort != resp2.PublicPort {
		return NatSymmetric, nil
	}

	return NatUnknown, nil
}

// GetPublicAddr 获取本机公网地址
func GetPublicAddr() (*StunResponse, error) {
	for _, server := range PublicSTUNServers {
		resp, err := SendBindingRequest(server)
		if err != nil {
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("所有 STUN 服务器均不可达")
}

// GetLocalAddrs 获取本机内网地址
func GetLocalAddrs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var local []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			ip := ipnet.IP
			if ip.IsPrivate() && !ip.IsLoopback() && ip.To4() != nil {
				local = append(local, ip.String())
			}
		}
	}
	return local, nil
}
