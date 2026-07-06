package main

import (
	"fmt"
	"time"

	"github.com/mc-connector/internal/tunnel"
)

func main() {
	fmt.Println("LAN port detection test...")
	fmt.Println("(Make sure MC is running and Open to LAN is enabled)")
	fmt.Println()

	port, motd, err := tunnel.DetectLANPort(10 * time.Second)
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
		fmt.Println()
		fmt.Println("Check:")
		fmt.Println("  1. MC is running")
		fmt.Println("  2. A world is loaded")
		fmt.Println("  3. Esc -> Open to LAN is clicked")
		return
	}

	fmt.Printf("SUCCESS: LAN port = %d\n", port)
	if motd != "" {
		fmt.Printf("  MOTD: %s\n", motd)
	}
}
