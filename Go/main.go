package main

import (
	"ZincSearchTest/Go/Indexer"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

const (
	bulkSize    = 30000
	num_workers = 10
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file to index next time.")
		os.Exit(1)
	}

	directoryPath := os.Args[1]

	// Enable CPU profiling
	cpuProfileFile, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Println("Error creating CPU profile:", err)
		os.Exit(1)
	}
	defer cpuProfileFile.Close()

	pprof.StartCPUProfile(cpuProfileFile)

	// check index if exist
	Indexer.CheckIndex()

	// Initialize indexing
	StartIndexing(directoryPath)

	// Stop CPU profiling
	pprof.StopCPUProfile()

	// Initialize Server
	// Server.Initialize_Server()
}

// Start_Index:
// function that initialize indexing process
func StartIndexing(directoryPath string) {
	// WaitGroups:
	var readDirWg sync.WaitGroup
	var mainWg sync.WaitGroup
	var workerWg sync.WaitGroup

	// ----------------------------------------------------------------------------
	fmt.Println("Iniciando indexador...")
	mainWg.Add(1)

	// email channel
	emailChannel := make(chan []Indexer.Email, bulkSize)

	for i := 0; i < num_workers; i++ {
		workerWg.Add(1)
		go Indexer.Worker(&workerWg, emailChannel)
	}

	// reading dir
	readDirWg.Add(1)
	go Indexer.ReadDirectory(directoryPath, &readDirWg, bulkSize, emailChannel)

	// wait for all goroutines readers to finish
	go func() {
		// reader
		readDirWg.Wait()
		if len(Indexer.EmailsList) >= 0 {
			emailChannel <- Indexer.EmailsList
		}
		close(emailChannel)

		// workers
		workerWg.Wait()
		mainWg.Done()
	}()

	// counting time of execution:
	startTime := time.Now()

	// wait for all goroutines to finish
	mainWg.Wait()

	// counting time of execution
	elapsedTime := time.Since(startTime)
	fmt.Println("Indexing took: ", elapsedTime)
}
