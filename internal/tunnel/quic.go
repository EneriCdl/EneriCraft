// QUIC 隧道管理器
//
// 基于 quic-go v0.60+ 实现 P2P 加密隧道。

package tunnel

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// GenerateTLSConfig 生成自签名 TLS 配置 (仅用于 QUIC 加密层)
func GenerateTLSConfig() (*tls.Config, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("生成密钥失败: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("生成证书失败: %w", err)
	}

	pubKeyDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("导出公钥失败: %w", err)
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		NextProtos:         []string{"mc-connector"},
		InsecureSkipVerify: true,
	}

	return tlsConfig, pubKeyDER, nil
}

// DefaultQuicConfig 返回默认 QUIC 配置
func DefaultQuicConfig() *quic.Config {
	return &quic.Config{
		KeepAlivePeriod: 10 * time.Second,
		MaxIdleTimeout: 60 * time.Second,
	}
}

// ListenQUIC 作为主机监听 QUIC 连接
func ListenQUIC(addr string, tlsConfig *tls.Config) (*quic.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("解析地址失败: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("UDP 监听失败: %w", err)
	}

	listener, err := quic.Listen(udpConn, tlsConfig, DefaultQuicConfig())
	if err != nil {
		return nil, fmt.Errorf("QUIC 监听失败: %w", err)
	}

	return listener, nil
}

// DialQUIC 作为客户端主动连接主机
func DialQUIC(addr string, tlsConfig *tls.Config) (*quic.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, addr, tlsConfig, DefaultQuicConfig())
	if err != nil {
		return nil, fmt.Errorf("QUIC 连接失败: %w", err)
	}

	return conn, nil
}

// OpenStream 在连接上打开新的双向 Stream
func OpenStream(conn *quic.Conn) (*quic.Stream, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("打开 Stream 失败: %w", err)
	}
	return stream, nil
}

// PemEncode 将 DER 编码的公钥转为 PEM 格式
func PemEncode(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	})
}
