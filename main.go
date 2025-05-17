package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"FDS/p2p"
)

func main() {
	peerID := flag.String("id", "", "Unique peer ID")
	bootstrapAddr := flag.String("bootstrap", "", "Bootstrap server address (host:port)")
	fileRequest := flag.String("file", "", "Filename to request from peers")
	targetPeer := flag.String("target", "", "Target peer ID to request file from")
	flag.Parse()

	if *peerID == "" {
		log.Fatalln("Please provide a peer ID using -id")
	}
	if *bootstrapAddr == "" {
		log.Fatalln("Please provide a bootstrap server address using -bootstrap")
	}

	ip := p2p.GetLocalIP()
	listener, portStr, err := p2p.CreateTCPListener(ip, "0") // Auto-assign port
	if err != nil {
		log.Fatalf("[ERROR] Failed to create TCP listener: %v", err)
	}

	localPeer := p2p.NewPeer(*peerID, ip, portStr) // localPeer is of type *p2p.Peer

	// NOTICE: Dereference localPeer so that we're passing a value rather than a pointer.
	if err := p2p.RegisterWithBootstrap(*localPeer, *bootstrapAddr); err != nil {
		log.Fatalf("[ERROR] Registration with bootstrap failed: %v", err)
	}
	log.Printf("[INFO] Peer %s registered with bootstrap at %s", *peerID, *bootstrapAddr)

	// Start sending heartbeat every 10 seconds
	stopHeartbeat := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := p2p.SendHeartbeatToBootstrap(*localPeer, *bootstrapAddr); err != nil {
					log.Printf("[WARN] Heartbeat error: %v", err)
				} else {
					log.Printf("[DEBUG] Heartbeat sent")
				}
			case <-stopHeartbeat:
				return
			}
		}
	}()

	// Start TCP server for incoming connections
	msgChan := make(chan string)
	quit := make(chan struct{})
	go p2p.StartTCPServerWithListener(localPeer, msgChan, listener, quit)
	// Periodically fetch peer list
	peerList := make(map[string]p2p.BootstrapPeerInfo)
	var listMutex sync.Mutex
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			peers, err := p2p.GetPeersFromBootstrap(*localPeer, *bootstrapAddr)
			if err != nil {
				log.Printf("[WARN] Failed to get peers from bootstrap: %v", err)
				continue
			}
			listMutex.Lock()
			peerList = make(map[string]p2p.BootstrapPeerInfo)
			for _, p := range peers {
				peerList[p.ID] = p
			}
			listMutex.Unlock()
			log.Printf("[INFO] Discovered %d peers", len(peers))
		}
	}()
	
	// Wait for a few seconds to allow the peer list to update
	log.Printf("[INFO] Waiting for peers to register...")
	time.Sleep(10 * time.Second)

	// Handle file request via bootstrap peer list
	if *fileRequest != "" && *targetPeer != "" {
		log.Printf("[INFO] Requesting file '%s' from peer %s via bootstrap...", *fileRequest, *targetPeer)

		listMutex.Lock()
		target, exists := peerList[*targetPeer]
		listMutex.Unlock()

		if !exists {
			log.Fatalf("[ERROR] Peer %s not found in the bootstrap peer list", *targetPeer)
		}

		// Extract IP and port from target.Addr
		peerIP, peerPort := parsePeerAddress(target.Addr)
		if peerIP == "" || peerPort == "" {
			log.Fatalf("[ERROR] Invalid peer address format: %s", target.Addr)
		}

		// Here RequestFile is called with a value of type p2p.Peer, constructed using the target's data.
		fileData, err := p2p.RequestFile(p2p.Peer{
			ID:   target.ID,
			IP:   peerIP,
			Port: peerPort,
		}, *fileRequest)
		if err != nil {
			log.Fatalf("[ERROR] File request failed: %v", err)
		}

		// Save received file
		if err := os.WriteFile("received_"+*fileRequest, fileData, 0644); err != nil {
			log.Fatalf("[ERROR] Failed to save received file: %v", err)
		}
		log.Printf("[INFO] File '%s' received and saved as 'received_%s'", *fileRequest, *fileRequest)
	}

	// Graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("[INFO] Shutting down peer...")

	close(quit)
	close(stopHeartbeat)
}

// parsePeerAddress extracts IP and Port from the peer's Addr field
func parsePeerAddress(addr string) (string, string) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		log.Printf("[WARN] Invalid peer address format: %s", addr)
		return "", ""
	}
	return parts[0], parts[1]
}
