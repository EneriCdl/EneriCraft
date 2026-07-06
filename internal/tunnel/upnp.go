// UPnP 端口映射
//
// 通过 UPnP 协议自动向路由器请求端口转发，
// 让外网能够访问本机的 QUIC 端口。

package tunnel

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// AddUPnPMapping 通过 UPnP 添加端口映射
// 返回外部端口号（可能与内部端口相同）
func AddUPnPMapping(internalPort int) (int, error) {
	// 步骤1: SSDP 发现 — 找到路由器 UPnP 服务
	gatewayURL, externalIP, err := discoverUPnP()
	if err != nil {
		return 0, fmt.Errorf("UPnP 发现失败: %w", err)
	}

	// 步骤2: 添加端口映射
	if err := addPortMapping(gatewayURL, internalPort, internalPort, "EneriCraft"); err != nil {
		// 端口可能被占用，尝试相邻端口
		for offset := 1; offset <= 10; offset++ {
			extPort := internalPort + offset
			if err2 := addPortMapping(gatewayURL, internalPort, extPort, "EneriCraft"); err2 == nil {
				log.Printf("[UPnP] 端口映射: %s:%d → 本机:%d", externalIP, extPort, internalPort)
				return extPort, nil
			}
		}
		return 0, fmt.Errorf("UPnP 端口映射失败: %w", err)
	}

	log.Printf("[UPnP] 端口映射: %s:%d → 本机:%d", externalIP, internalPort, internalPort)
	return internalPort, nil
}

// discoverUPnP 通过 SSDP 发现 UPnP 网关
func discoverUPnP() (gatewayURL string, externalIP string, err error) {
	// SSDP M-SEARCH 消息
	searchMsg := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 2\r\n" +
		"ST: urn:schemas-upnp-org:device:InternetGatewayDevice:1\r\n" +
		"\r\n"

	// 发送到 SSDP 组播地址
	addr, _ := net.ResolveUDPAddr("udp", "239.255.255.250:1900")
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	conn.Write([]byte(searchMsg))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return "", "", fmt.Errorf("SSDP 无响应 (路由器可能不支持 UPnP)")
	}

	response := string(buf[:n])

	// 提取 LOCATION (UPnP 描述文件 URL)
	locRe := regexp.MustCompile(`(?i)LOCATION:\s*(\S+)`)
	locMatch := locRe.FindStringSubmatch(response)
	if len(locMatch) < 2 {
		return "", "", fmt.Errorf("SSDP 响应格式错误")
	}
	locationURL := strings.TrimSpace(locMatch[1])

	// 获取 WANIPConnection 控制 URL
	controlURL, extIP, err := parseUPnPDescription(locationURL)
	if err != nil {
		return "", "", err
	}

	return controlURL, extIP, nil
}

// parseUPnPDescription 解析 UPnP 设备描述，获取控制 URL
func parseUPnPDescription(locationURL string) (controlURL string, externalIP string, err error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(locationURL)
	if err != nil {
		return "", "", fmt.Errorf("获取 UPnP 描述失败: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	xml := buf.String()

	// 找到 WANIPConnection 服务的控制 URL
	// 简单解析：找 <controlURL> 标签
	re := regexp.MustCompile(`(?s)<service>.*?<serviceType>urn:schemas-upnp-org:service:WANIPConnection:1</serviceType>.*?<controlURL>(.*?)</controlURL>.*?</service>`)
	match := re.FindStringSubmatch(xml)
	if len(match) < 2 {
		// 尝试 WANPPPConnection
		re2 := regexp.MustCompile(`(?s)<service>.*?<serviceType>urn:schemas-upnp-org:service:WANPPPConnection:1</serviceType>.*?<controlURL>(.*?)</controlURL>.*?</service>`)
		match = re2.FindStringSubmatch(xml)
	}
	if len(match) < 2 {
		return "", "", fmt.Errorf("未找到 WANIPConnection 服务")
	}

	controlPath := match[1]

	// 从 LOCATION URL 构造完整的控制 URL
	if strings.HasPrefix(controlPath, "/") {
		// 从 locationURL 提取 base URL
		baseRe := regexp.MustCompile(`^(https?://[^/]+)`)
		baseMatch := baseRe.FindStringSubmatch(locationURL)
		if len(baseMatch) >= 2 {
			controlURL = baseMatch[1] + controlPath
		}
	} else {
		controlURL = controlPath
	}

	// 获取外网 IP
	extIPRe := regexp.MustCompile(`(?i)<externalIPAddress>(.*?)</externalIPAddress>`)
	if extMatch := extIPRe.FindStringSubmatch(xml); len(extMatch) >= 2 {
		externalIP = extMatch[1]
	}

	return controlURL, externalIP, nil
}

// addPortMapping 发送 AddPortMapping SOAP 请求
func addPortMapping(controlURL string, internalPort, externalPort int, description string) error {
	soapBody := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:AddPortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:1">
<NewRemoteHost></NewRemoteHost>
<NewExternalPort>%d</NewExternalPort>
<NewProtocol>UDP</NewProtocol>
<NewInternalPort>%d</NewInternalPort>
<NewInternalClient>%s</NewInternalClient>
<NewEnabled>1</NewEnabled>
<NewPortMappingDescription>%s</NewPortMappingDescription>
<NewLeaseDuration>0</NewLeaseDuration>
</u:AddPortMapping>
</s:Body>
</s:Envelope>`, externalPort, internalPort, getLocalIP(), description)

	req, err := http.NewRequest("POST", controlURL, strings.NewReader(soapBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:WANIPConnection:1#AddPortMapping"`)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("UPnP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("UPnP 返回 HTTP %d", resp.StatusCode)
	}

	return nil
}

// getLocalIP 获取本机局域网 IP
func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			ip := ipnet.IP
			if ip.IsPrivate() && !ip.IsLoopback() && ip.To4() != nil {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

// RemoveUPnPMapping 删除 UPnP 端口映射
func RemoveUPnPMapping(externalPort int) {
	gatewayURL, _, err := discoverUPnP()
	if err != nil {
		return
	}

	soapBody := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:DeletePortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:1">
<NewRemoteHost></NewRemoteHost>
<NewExternalPort>%d</NewExternalPort>
<NewProtocol>UDP</NewProtocol>
</u:DeletePortMapping>
</s:Body>
</s:Envelope>`, externalPort)

	req, _ := http.NewRequest("POST", gatewayURL, strings.NewReader(soapBody))
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:WANIPConnection:1#DeletePortMapping"`)

	client := &http.Client{Timeout: 3 * time.Second}
	client.Do(req)
}
