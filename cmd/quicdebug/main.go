package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

func main() {
	// 生成自签名证书
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{"localhost"},
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	tlsCert := tls.Certificate{Certificate: [][]byte{certDER}, PrivateKey: key}

	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"mc-connector"},
	}
	clientTLS := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"mc-connector"},
		ServerName:         "localhost",
	}

	udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	fmt.Printf("Listen 127.0.0.1:%d\n", port)

	quicCfg := &quic.Config{KeepAlivePeriod: 10 * time.Second, MaxIdleTimeout: 60 * time.Second}
	listener, err := quic.Listen(udpConn, serverTLS, quicCfg)
	if err != nil {
		fmt.Printf("FAIL Listen: %v\n", err)
		return
	}
	defer listener.Close()

	errCh := make(chan error, 1)
	go func() {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			errCh <- err
			return
		}
		fmt.Printf("ACCEPTED: %s\n", conn.RemoteAddr())
		conn.CloseWithError(0, "test done")
		errCh <- nil
	}()

	time.Sleep(300 * time.Millisecond)

	fmt.Printf("Dial 127.0.0.1:%d...\n", port)
	conn, err := quic.DialAddr(context.Background(), fmt.Sprintf("127.0.0.1:%d", port), clientTLS, quicCfg)
	if err != nil {
		fmt.Printf("FAIL Dial: %v\n", err)
	} else {
		fmt.Printf("SUCCESS! %s\n", conn.RemoteAddr())
		conn.CloseWithError(0, "ok")
	}

	select {
	case e := <-errCh:
		if e != nil {
			fmt.Printf("Accept error: %v\n", e)
		}
	case <-time.After(2 * time.Second):
		fmt.Println("Timeout")
	}
}
