package p2p

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Define the fixed size for each chunk.
const chunkSize = 4096

// MetaData holds metadata that maps a filename to a list of chunk hashes.
type MetaData struct {
	Filename    string   `json:"filename"`
	ChunkHashes []string `json:"chunk_hashes"`
}

// StoreFile breaks the file data into chunks, computes a SHA-256 hash for each chunk,
// stores each chunk on disk, and saves corresponding metadata.
func StoreFile(filename string, data []byte) error {
	chunks, hashes := chunkData(data, chunkSize)

	// Ensure that the "chunks" directory exists.
	if err := os.MkdirAll("chunks", 0755); err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}

	// Write each chunk to a separate file, named by its hash.
	for i, chunk := range chunks {
		chunkFilename := filepath.Join("chunks", hashes[i])

		// Skip writing if the chunk already exists.
		if _, err := os.Stat(chunkFilename); os.IsNotExist(err) {
			if err := os.WriteFile(chunkFilename, chunk, 0644); err != nil {
				return fmt.Errorf("failed to write chunk %s: %w", hashes[i], err)
			}
		}
		fmt.Printf("Stored chunk %s for file %s (%d bytes)\n", hashes[i], filename, len(chunk))
	}

	// Create metadata for the file and save it.
	meta := MetaData{
		Filename:    filename,
		ChunkHashes: hashes,
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metaFilename := filepath.Join("chunks", filename+".meta")
	if err := os.WriteFile(metaFilename, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	fmt.Printf("Stored metadata for file %s\n", filename)
	return nil
}

// RetrieveFile reconstructs a file from its chunks using stored metadata.
func RetrieveFile(filename string) ([]byte, error) {
	metaFilename := filepath.Join("chunks", filename+".meta")
	metaData, err := os.ReadFile(metaFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta MetaData
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	var fileData []byte
	for _, hash := range meta.ChunkHashes {
		chunkFilename := filepath.Join("chunks", hash)
		chunk, err := os.ReadFile(chunkFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk file %s: %w", hash, err)
		}
		fileData = append(fileData, chunk...)
	}

	fmt.Printf("Reconstructed file %s from %d chunks (%d bytes)\n", filename, len(meta.ChunkHashes), len(fileData))
	return fileData, nil
}

// chunkData splits data into fixed-size chunks and computes SHA-256 hashes for each.
func chunkData(data []byte, size int) ([][]byte, []string) {
	var chunks [][]byte
	var hashes []string

	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		hashVal := sha256.Sum256(chunk)
		hashStr := fmt.Sprintf("%x", hashVal)
		chunks = append(chunks, chunk)
		hashes = append(hashes, hashStr)
	}

	return chunks, hashes
}
