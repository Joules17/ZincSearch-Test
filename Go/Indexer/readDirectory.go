package Indexer

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

var EmailsList []Email
var mu sync.Mutex

// read directory function
// looks for files given a path as arg (recursively)
func ReadDirectory(directoryPath string, wg *sync.WaitGroup, bulkSize int, emailChannel chan []Email) {
	defer wg.Done()

	files, err := os.ReadDir(directoryPath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", directoryPath, err)
		return
	}

	for _, file := range files {
		filePath := filepath.Join(directoryPath, file.Name())
		if file.IsDir() {
			wg.Add(1)
			go ReadDirectory(filePath, wg, bulkSize, emailChannel)
		} else {
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading file %s: %v", filePath, err)
				continue
			}

			email, err := ScanFile(file.Name(), string(fileContent))

			if err != nil {
				log.Printf("Error scanning file %s: %v", filePath, err)
				continue
			}

			mu.Lock()
			EmailsList = append(EmailsList, email)
			if len(EmailsList) >= bulkSize {
				emailChannel <- EmailsList[:bulkSize]
				EmailsList = EmailsList[bulkSize:]
			}
			mu.Unlock()
		}
	}
}
