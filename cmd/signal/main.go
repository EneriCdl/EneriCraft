// EneriCraft 信令服务器（极简版）
//
// 只交换双方公网地址，不中继游戏数据。
// 带宽消耗接近零——仅传递几百字节的地址信息。
// 一个最便宜的云服务器 ($3/月) 就能服务几千用户。
//
// 协议 (HTTP)：
//   POST /offer  — 主机上报地址 {code, stun_ip, stun_port}
//   GET  /answer?code=xxx — 客户端获取主机地址
//   POST /answer — 客户端上报自己地址 {code, stun_ip, stun_port}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Peer struct {
	Code      string `json:"code"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	CreatedAt time.Time
}

var (
	offers  = map[string]*Peer{} // 主机地址
	answers = map[string]*Peer{} // 客户端地址
	mu      sync.RWMutex
)

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/offer", handleOffer)
	http.HandleFunc("/answer", handleAnswer)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// 清理过期
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for k, v := range offers {
				if time.Since(v.CreatedAt) > 10*time.Minute {
					delete(offers, k)
				}
			}
			for k, v := range answers {
				if time.Since(v.CreatedAt) > 10*time.Minute {
					delete(answers, k)
				}
			}
			mu.Unlock()
		}
	}()

	log.Printf("EneriCraft Signal v0.8 :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleOffer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	var p Peer
	json.NewDecoder(r.Body).Decode(&p)
	p.CreatedAt = time.Now()

	mu.Lock()
	offers[p.Code] = &p
	// 检查是否客户端已经在等待
	if ans, ok := answers[p.Code]; ok {
		// 互相知道了——返回客户端地址给主机
		mu.Unlock()
		log.Printf("[配对] %s 主机:%s:%d <-> 客户端:%s:%d",
			p.Code[:20], p.IP, p.Port, ans.IP, ans.Port)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"peer": ans,
		})
		return
	}
	mu.Unlock()
	log.Printf("[主机] %s... %s:%d", p.Code[:20], p.IP, p.Port)
	json.NewEncoder(w).Encode(map[string]string{"status": "waiting"})
}

func handleAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 客户端查询主机地址
		code := r.URL.Query().Get("code")
		mu.RLock()
		offer, ok := offers[code]
		mu.RUnlock()
		if ok {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"peer": offer,
			})
			return
		}
		http.NotFound(w, r)
		return
	}

	if r.Method == "POST" {
		// 客户端上报自己地址
		var p Peer
		json.NewDecoder(r.Body).Decode(&p)
		p.CreatedAt = time.Now()

		mu.Lock()
		answers[p.Code] = &p
		// 返回主机地址
		offer, ok := offers[p.Code]
		mu.Unlock()

		if ok {
			log.Printf("[配对] %s 主机:%s:%d <-> 客户端:%s:%d",
				p.Code[:20], offer.IP, offer.Port, p.IP, p.Port)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"peer": offer,
			})
			return
		}
		log.Printf("[客户端] %s... %s:%d", p.Code[:20], p.IP, p.Port)
		json.NewEncoder(w).Encode(map[string]string{"status": "waiting"})
	}
}
