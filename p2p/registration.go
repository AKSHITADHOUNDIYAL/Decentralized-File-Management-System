package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// BootstrapPeerInfo represents peer information received from the bootstrap server.
type BootstrapPeerInfo struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

// RegisterWithBootstrap registers the local peer with the bootstrap server.
// It sends a JSON message with type "register", the peer's ID and address,
// then reads the server's response.
func RegisterWithBootstrap(localPeer Peer, bootstrapAddr string) error {
	conn, err := net.DialTimeout("tcp", bootstrapAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to bootstrap server: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	msg := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Addr string `json:"addr"`
	}{
		Type: "register",
		ID:   localPeer.ID,
		Addr: localPeer.Address(),
	}

	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send register message: %w", err)
	}

	var resp map[string]string
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to decode bootstrap response: %w", err)
	}
	if resp["status"] != "ok" {
		return fmt.Errorf("bootstrap registration failed: %s", resp["error"])
	}

	return nil
}

// GetPeersFromBootstrap queries the bootstrap server for active peers.
// It sends a message with type "get_peers" and decodes the returned peer list.
func GetPeersFromBootstrap(localPeer Peer, bootstrapAddr string) ([]BootstrapPeerInfo, error) {
	conn, err := net.DialTimeout("tcp", bootstrapAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bootstrap server: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	msg := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "get_peers",
		ID:   localPeer.ID,
	}

	if err := encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send get_peers request: %w", err)
	}

	var peers []BootstrapPeerInfo
	if err := decoder.Decode(&peers); err != nil {
		return nil, fmt.Errorf("failed to decode peers list: %w", err)
	}

	return peers, nil
}

// SendHeartbeatToBootstrap notifies the bootstrap server that the peer is still active.
// It sends a JSON message with type "heartbeat" and the peer's ID.
func SendHeartbeatToBootstrap(localPeer Peer, bootstrapAddr string) error {
	conn, err := net.DialTimeout("tcp", bootstrapAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to bootstrap server: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	msg := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "heartbeat",
		ID:   localPeer.ID,
	}

	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send heartbeat message: %w", err)
	}

	return nil
}
