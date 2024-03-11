package Indexer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var semaphore = make(chan struct{}, 900)

// read directory function
// looks for files given a path as arg (recursively)
func ReadDirectory(directoryPath string, wg *sync.WaitGroup, bulkSize int, emailChannel chan Email) {
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
	fileHandle, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", filePath, err)
		return
	}
	defer fileHandle.Close()

	fileContent, err := io.ReadAll(fileHandle)
	if err != nil {
		fmt.Println("Error reading file:", filePath, err)
		return // ignore file
	}

	email, err := ScanFile(file.Name(), string(fileContent))
	if err != nil {
		fmt.Println("Error scanning file: ", filePath)
		return // ignore file
	}

	// Close the file explicitly after reading
	if err := fileHandle.Close(); err != nil {
		fmt.Println("Error closing file:", filePath, err)
	}

	emailChannel <- email
}
