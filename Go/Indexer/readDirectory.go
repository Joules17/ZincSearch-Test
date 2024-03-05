package Indexer

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

// read directory function
// looks for files given a path as arg (recursively)
func ReadDirectory(directoryPath string, wg *sync.WaitGroup, pkgChan chan<- EmailPkg, bulkSize int) {
	// fmt.Println("Iniciando lectura: ", directoryPath)
	defer wg.Done()

	files, err := os.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	var emails []Email

	for _, file := range files {
		filePath := filepath.Join(directoryPath, file.Name())

		if file.IsDir() {
			wg.Add(1)
			go ReadDirectory(filePath, wg, pkgChan, bulkSize)
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

			emails = append(emails, email)

			if len(emails) >= bulkSize {
				pkgChan <- EmailPkg{Emails: emails[:bulkSize]}
				emails = emails[bulkSize:]
			}
		}
	}

	if len(emails) > 0 {
		pkgChan <- EmailPkg{Emails: emails}
	}
}
