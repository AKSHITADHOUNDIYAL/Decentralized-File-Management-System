package p2p

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

// Message struct for handling different request types.
type Message struct {
	Type     string `json:"type"`
	Filename string `json:"filename,omitempty"`
	Content  []byte `json:"content,omitempty"`
}

// StartTCPServerWithListener starts the TCP server and listens for file requests.
func StartTCPServerWithListener(localPeer *Peer, msgChan chan<- string, listener net.Listener, quit <-chan struct{}) {
	log.Printf("[TCP] Listening on %s", listener.Addr().String())

	for {
		select {
		case <-quit:
			log.Println("[TCP] Server shutting down...")
			listener.Close()
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Println("[TCP] Accept error:", err)
				continue
			}
			go handleConnection(conn, msgChan)
		}
	}
}

// handleConnection processes incoming TCP messages safely.
func handleConnection(conn net.Conn, msgChan chan<- string) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var request Message

	if err := decoder.Decode(&request); err != nil {
		log.Printf("[ERROR] Failed to parse incoming request: %v", err)
		return
	}

	log.Printf("[DEBUG] Received request type: %s, filename: %s", request.Type, request.Filename)

	// Handle file request
	if request.Type == "request_file" {
		log.Printf("[DEBUG] File request received, calling sendFile() for %s", request.Filename)
		sendFile(conn, request.Filename)
		log.Printf("[DEBUG] Finished executing sendFile() for %s", request.Filename)
	}
}

// sendFile streams the requested file in chunks, ensuring integrity.
func sendFile(conn net.Conn, filename string) {
	filePath := filepath.Join("shared_folder", filename)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] File not found: %s", filename)
		response := Message{
			Type:     "error",
			Filename: filename,
			Content:  []byte("File not found"),
		}
		json.NewEncoder(conn).Encode(response)
		return
	}
	defer file.Close()

	buffer := make([]byte, 4096) // 4KB chunks
	encoder := json.NewEncoder(conn)

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[ERROR] Error reading file: %s", filename)
			return
		}

		log.Printf("[DEBUG] Read file chunk (%d bytes)", n)

		response := Message{
			Type:    "send_file_chunk",
			Content: buffer[:n],
		}

		if err := encoder.Encode(response); err != nil {
			log.Printf("[ERROR] Failed to send file chunk: %s", filename)
			return
		}

		log.Printf("[DEBUG] Sent file chunk (%d bytes)", n)
	}

	// Send explicit "end_of_file" signal
	endMessage := Message{
		Type:     "end_of_file",
		Filename: filename,
	}
	if err := encoder.Encode(endMessage); err != nil {
		log.Printf("[ERROR] Failed to send end-of-file signal: %s", filename)
		return
	}

	log.Printf("[DEBUG] File %s sent successfully.", filename)
}
