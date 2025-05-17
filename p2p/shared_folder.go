package p2p

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// SharedFolder represents the folder where shared files are stored.
type SharedFolder struct {
	FolderPath string
}

// NewSharedFolder creates a new instance of SharedFolder and ensures the folder exists.
func NewSharedFolder(folderPath string) *SharedFolder {
	// Ensure the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create shared folder: %v", err)
		}
		log.Printf("Created shared folder at: %s", folderPath)
	}

	return &SharedFolder{FolderPath: folderPath}
}

// ListFiles returns the list of filenames in the shared folder.
func (s *SharedFolder) ListFiles() ([]string, error) {
	files, err := ioutil.ReadDir(s.FolderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// AddFile adds a new file to the shared folder.
func (s *SharedFolder) AddFile(filename string, data []byte) error {
	filePath := filepath.Join(s.FolderPath, filename)
	err := ioutil.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filename, err)
	}
	log.Printf("File added: %s", filename)
	return nil
}

// RemoveFile removes a file from the shared folder.
func (s *SharedFolder) RemoveFile(filename string) error {
	filePath := filepath.Join(s.FolderPath, filename)
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove file %s: %v", filename, err)
	}
	log.Printf("File removed: %s", filename)
	return nil
}
