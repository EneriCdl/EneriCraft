package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

var (
	pending   = map[string]chan *quic.Conn{}
	pendingMu sync.Mutex
)

func main() {
	port := "9000"
	if len(os.Args) > 1 { port = os.Args[1] }
	tlsCfg := generateTLS()
	udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: atoi(port)})
	listener, _ := quic.Listen(udpConn, tlsCfg, &quic.Config{
		KeepAlivePeriod: 15 * time.Second, MaxIdleTimeout: 120 * time.Second,
	})
	log.Printf("Relay v0.8 :%s", port)

	// 清理由 select timeout 处理，不需要额外清理

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil { continue }
		go handle(conn)
	}
}

func handle(conn *quic.Conn) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil { return }
	buf := make([]byte, 512)
	n, err := stream.Read(buf)
	if err != nil { return }
	code := strings.TrimSpace(string(buf[:n]))
	hash := hex.EncodeToString([]byte(code)[:min(8, len(code))])
	log.Printf("[JOIN] hash=%s len=%d from=%s", hash, len(code), conn.RemoteAddr())

	pendingMu.Lock()
	// Debug: list all pending
	keys := []string{}
	for k := range pending { keys = append(keys, k) }
	log.Printf("[DEBUG] pending=%d keys=%v", len(pending), func() []string {
		var hs []string
		for _, k := range keys { hs = append(hs, hex.EncodeToString([]byte(k)[:min(8, len(k))])) }
		return hs
	}())

	ch, exists := pending[code]
	if exists {
		log.Printf("[PAIR] Match! key=%s", hash)
		delete(pending, code)
		pendingMu.Unlock()
		ch <- conn
		return
	}

	ch = make(chan *quic.Conn, 1)
	pending[code] = ch
	pendingMu.Unlock()
	log.Printf("[WAIT] key=%s", hash)

	select {
	case peer := <-ch:
		log.Printf("[PAIRED] key=%s", hash)
		bridge(conn, stream, peer)
	case <-time.After(5 * time.Minute):
		log.Printf("[TIMEOUT] key=%s", hash)
		pendingMu.Lock(); delete(pending, code); pendingMu.Unlock()
	}
}

func bridge(a *quic.Conn, aFirst *quic.Stream, b *quic.Conn) {
	go func() {
		for {
			s, err := a.AcceptStream(a.Context())
			if err != nil { return }
			go func(s *quic.Stream) {
				peer, err := b.OpenStreamSync(context.Background())
				if err != nil { return }
				go func() { io.Copy(peer, s); peer.Close() }()
				io.Copy(s, peer); s.Close()
			}(s)
		}
	}()
	for {
		s, err := b.AcceptStream(b.Context())
		if err != nil { return }
		go func(s *quic.Stream) {
			peer, err := a.OpenStreamSync(context.Background())
			if err != nil { return }
			go func() { io.Copy(peer, s); peer.Close() }()
			io.Copy(s, peer); s.Close()
		}(s)
	}
}

func generateTLS() *tls.Config {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	t := x509.Certificate{SerialNumber: big.NewInt(1), DNSNames: []string{"relay"}}
	d, _ := x509.CreateCertificate(rand.Reader, &t, &t, &key.PublicKey, key)
	return &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{d}, PrivateKey: key}}, NextProtos: []string{"mc-connector-relay"}, InsecureSkipVerify: true}
}

func min(a, b int) int { if a < b { return a }; return b }
func atoi(s string) int { var n int; for _, c := range s { n = n*10 + int(c-'0') }; return n }
