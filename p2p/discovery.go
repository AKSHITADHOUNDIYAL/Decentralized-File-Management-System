package p2p

import (
	"log"
	"net"
	"strings"
	"time"
)

const BroadcastInterval = 5 * time.Second

func DiscoverPeers(localPeer *Peer, udpPort string, peerChan chan<- *Peer) {
	listenAddr, _ := net.ResolveUDPAddr("udp4", ":"+udpPort)
	recvConn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		log.Fatalf("[ERROR] UDP listen error: %v", err)
	}
	defer recvConn.Close()

	sendAddr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:"+udpPort)
	sendConn, err := net.DialUDP("udp4", nil, sendAddr)
	if err != nil {
		log.Fatalf("[ERROR] UDP dial error: %v", err)
	}
	defer sendConn.Close()

	// Broadcast presence
	go func() {
		for {
			msg := "PING:" + localPeer.Address()
			_, err := sendConn.Write([]byte(msg))
			if err != nil {
				log.Printf("[WARN] UDP send error: %v", err)
			}
			time.Sleep(BroadcastInterval)
		}
	}()

	// Listen for pings
	buf := make([]byte, 1024)
	for {
		n, addr, err := recvConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("[ERROR] UDP read error: %v", err)
			continue
		}

		msg := string(buf[:n])
		if strings.HasPrefix(msg, "PING:") {
			peerAddr := msg[5:]
			if peerAddr != localPeer.Address() {
				ip, port, err := net.SplitHostPort(peerAddr)
				if err != nil {
					log.Printf("[ERROR] Failed to split peer address: %v", err)
					continue
				}
				peerChan <- NewPeer(addr.String(), ip, port) // Assign correct peer ID dynamically
				log.Printf("[INFO] Discovered peer: %s @ %s", ip, port)
			}
		}
	}
}