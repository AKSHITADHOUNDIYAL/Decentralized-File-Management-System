package p2p

import (
	"log"
	"net"
	"os"
	"strings"
)

// Peer defines a network peer with its ID, IP, and Port.
type Peer struct {
	ID   string
	IP   string
	Port string
}

// NewPeer creates and returns a new Peer instance.
func NewPeer(id, ip, port string) *Peer {
	return &Peer{
		ID:   id,
		IP:   ip,
		Port: port,
	}
}

// Address returns the peer's network address in "IP:Port" format.
func (p *Peer) Address() string {
	return net.JoinHostPort(p.IP, p.Port)
}

// GetLocalIP retrieves a non-loopback local IP of the host.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println("Error getting local IP:", err)
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() != nil &&
			!ipnet.IP.IsLinkLocalUnicast() {
			return ipnet.IP.String()
		}
	}
	return ""
}

// GetHostname returns the system's hostname.
func GetHostname() string {
	name, err := os.Hostname()
	if err != nil {
		log.Printf("Error retrieving hostname: %v", err)
		return "unknown"
	}
	return name
}
func parsePeerAddress(addr string) (string, string) {
    parts := strings.Split(addr, ":")
    if len(parts) != 2 {
        log.Fatalf("[ERROR] Invalid peer address format: %s", addr)
    }
    return parts[0], parts[1]
}
