package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// PeerInfo stores details of each peer.
type PeerInfo struct {
	ID       string
	Addr     string
	LastSeen time.Time
}

// BootstrapServer manages peer registrations.
type BootstrapServer struct {
	peers    map[string]PeerInfo
	mutex    sync.RWMutex
	listener net.Listener
	done     chan struct{}
}

func NewBootstrapServer() *BootstrapServer {
	return &BootstrapServer{
		peers:  make(map[string]PeerInfo),
		done:   make(chan struct{}),
	}
}

func (bs *BootstrapServer) Start(port string) error {
	var err error
	bs.listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to start bootstrap server: %w", err)
	}

	go bs.handleSignals()
	go bs.cleanupInactivePeers()

	log.Printf("Bootstrap server running on port %s", port)

	for {
		select {
		case <-bs.done:
			return nil
		default:
			conn, err := bs.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				log.Printf("Accept error: %v", err)
				continue
			}
			go bs.handleConnection(conn)
		}
	}
}

func (bs *BootstrapServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var msg struct {
		Type string
		ID   string
		Addr string
	}

	if err := decoder.Decode(&msg); err != nil {
		log.Printf("Failed to decode message: %v", err)
		return
	}

	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	switch msg.Type {
	case "register":
		bs.peers[msg.ID] = PeerInfo{ID: msg.ID, Addr: msg.Addr, LastSeen: time.Now()}
		log.Printf("Registered peer: %s (%s)", msg.ID, msg.Addr)
		encoder.Encode(map[string]string{"status": "ok"})

	case "get_peers":
		var peers []PeerInfo
		for _, peer := range bs.peers {
			if peer.ID != msg.ID {
				peers = append(peers, peer)
			}
		}
		encoder.Encode(peers)

	case "heartbeat":
		if peer, exists := bs.peers[msg.ID]; exists {
			peer.LastSeen = time.Now()
			bs.peers[msg.ID] = peer
		}

	default:
		encoder.Encode(map[string]string{"status": "error", "error": "unknown request"})
	}
}

func (bs *BootstrapServer) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	close(bs.done)
	bs.listener.Close()
	log.Println("Bootstrap server shutting down")
}

func (bs *BootstrapServer) cleanupInactivePeers() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bs.mutex.Lock()
		now := time.Now()
		for id, peer := range bs.peers {
			if now.Sub(peer.LastSeen) > 10*time.Minute {
				delete(bs.peers, id)
				log.Printf("Removed inactive peer: %s", id)
			}
		}
		bs.mutex.Unlock()
	}
}

func main() {
	server := NewBootstrapServer()
	if err := server.Start("9999"); err != nil {
		log.Fatal(err)
	}
}