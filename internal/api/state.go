package api

import (
	"log"
	"sync"

	"github.com/mc-connector/internal/server"
)

type PlayerEntry struct {
	Name   string `json:"name"`
	Online bool   `json:"online"`
}

type RoomState struct {
	mu             sync.RWMutex
	Connected      bool          `json:"connected"`
	ConnectCode    string        `json:"connect_code"`
	NATType        string        `json:"nat_type"`
	MCVersion      string        `json:"mc_version"`
	Players        []PlayerEntry `json:"players"`
	ConnectionType string        `json:"connection_type"`
	ServerProc     *server.Process
}

var state = &RoomState{MCVersion: "1.21"}

func GetState() RoomState {
	state.mu.RLock()
	defer state.mu.RUnlock()
	players := make([]PlayerEntry, len(state.Players))
	copy(players, state.Players)
	return RoomState{
		Connected: state.Connected, ConnectCode: state.ConnectCode,
		NATType: state.NATType, MCVersion: state.MCVersion,
		Players: players, ConnectionType: state.ConnectionType,
	}
}

func IsServerRunning() bool {
	state.mu.RLock(); defer state.mu.RUnlock()
	if state.ServerProc == nil { return false }
	return state.ServerProc.Running()
}

func SetRoom(connectCode, connType, mcVersion string) {
	state.mu.Lock(); defer state.mu.Unlock()
	state.Connected = true; state.ConnectCode = connectCode
	state.ConnectionType = connType; state.MCVersion = mcVersion
	state.Players = []PlayerEntry{{Name: "房主", Online: true}}
}

func SetServerProc(proc *server.Process) {
	state.mu.Lock(); defer state.mu.Unlock()
	state.ServerProc = proc
}

func AddPlayer(name string) {
	state.mu.Lock(); defer state.mu.Unlock()
	state.Players = append(state.Players, PlayerEntry{Name: name, Online: false})
}

func UpdatePlayerStatus(name string, online bool) {
	state.mu.Lock(); defer state.mu.Unlock()
	for i := range state.Players {
		if state.Players[i].Name == name {
			state.Players[i].Online = online
			log.Printf("[状态] 玩家 %s 在线=%v", name, online)
			return
		}
	}
	state.Players = append(state.Players, PlayerEntry{Name: name, Online: online})
	log.Printf("[状态] 新玩家 %s 在线=%v", name, online)
}

func ClearRoom() {
	state.mu.Lock(); defer state.mu.Unlock()
	tmgr.Stop()
	if state.ServerProc != nil {
		if state.ServerProc.Running() { state.ServerProc.Stop() }
		state.ServerProc = nil
	}
	state.Connected = false; state.ConnectCode = ""
	state.NATType = ""; state.Players = nil
	state.ConnectionType = "none"
}

func StartServer(cfg *server.Config) (*server.Process, error) {
	proc, err := server.Start(cfg)
	if err != nil { return nil, err }
	SetServerProc(proc)
	return proc, nil
}
