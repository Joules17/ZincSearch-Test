package Indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const MaxThreads = 900 // max number of concurrent go routines

// read directory function
// looks for files given a path as arg (recursively)
func ReadDirectory(directoryPath string, wg *sync.WaitGroup, bulkSize int, emailChannel chan Email) {
	semaphore := make(chan struct{}, MaxThreads)

	files, err := os.ReadDir(directoryPath)

	if err != nil {
		fmt.Printf("Error reading directory %s: %v", directoryPath, err)
		return
	}

	for _, file := range files {

		filePath := filepath.Join(directoryPath, file.Name())
		if file.IsDir() {
			ReadDirectory(filePath, wg, bulkSize, emailChannel)
		} else {
			wg.Add(1)
			semaphore <- struct{}{} // add 1 to the semaphore
			go func(filePath string, file os.DirEntry) {
				defer func() {
					<-semaphore // release the semaphore
					wg.Done()
				}()
				ProcessFile(filePath, file, wg, bulkSize, emailChannel)

			}(filePath, file)
		}
	}
}

// process file function
// process file content and add it to channel
func ProcessFile(filePath string, file os.DirEntry, wg *sync.WaitGroup, bulkSize int, emailChannel chan Email) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", filePath)
		return // ignore file
	}

	email, err := ScanFile(file.Name(), string(fileContent))

	if err != nil {
		fmt.Println("Error scanning file: ", filePath)
		return // ignore file
	}

	emailChannel <- email
}
