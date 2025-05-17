package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// CreateTCPListener sets up a TCP listener with proper error handling.
func CreateTCPListener(ip, port string) (net.Listener, string, error) {
	addr := net.JoinHostPort(ip, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", fmt.Errorf("tcp listen failed on %s: %v", addr, err)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	return listener, fmt.Sprintf("%d", actualPort), nil
}

// SendMessage allows a peer to send a message via TCP.
func SendMessage(peer Peer, message string) error {
	conn, err := net.DialTimeout("tcp", peer.Address(), 10*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to peer %s: %w", peer.Address(), err)
	}
	defer conn.Close()

	msg := Message{
		Type:    "message",
		Content: []byte(message),
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// RequestFile sends a file request and receives the file in chunks.
func RequestFile(peer Peer, filename string) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", peer.Address(), 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to connect to peer %s: %w", peer.Address(), err)
	}
	defer conn.Close()

	// Send file request as JSON
	request := Message{
		Type:     "request_file",
		Filename: filename,
	}
	encoder := json.NewEncoder(conn)

	log.Printf("[DEBUG] Sending file request: %s to %s", filename, peer.Address())

	if err := encoder.Encode(request); err != nil {
		return nil, fmt.Errorf("[ERROR] Failed to send file request: %w", err)
	}

	// Read file chunks
	var receivedData []byte
	decoder := json.NewDecoder(conn)

	for {
		var response Message
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("[ERROR] Error receiving file chunk: %w", err)
		}

		log.Printf("[DEBUG] Received message type: %s", response.Type)

		if response.Type == "error" {
			return nil, fmt.Errorf("[ERROR] Peer responded with error: %s", string(response.Content))
		}

		if response.Type == "send_file_chunk" {
			log.Printf("[DEBUG] Received file chunk: %d bytes", len(response.Content))
			receivedData = append(receivedData, response.Content...)
		}

		if response.Type == "end_of_file" {
			log.Println("[DEBUG] File transfer complete!")
			break
		}
	}

	// Confirm final data length before saving
	log.Printf("[DEBUG] Total received file size: %d bytes", len(receivedData))

	return receivedData, nil
}

// SaveFile writes received data to disk.
func SaveFile(filename string, data []byte) error {
	log.Printf("[DEBUG] Writing %d bytes to file: %s", len(data), filename)

	file, err := os.Create("received_" + filename)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to write file: %w", err)
	}

	writer.Flush()
	log.Printf("[DEBUG] Successfully saved file: %s", filename)
	return nil
}
