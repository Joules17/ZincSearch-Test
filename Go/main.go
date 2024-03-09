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
	bulkSize    = 40000
	num_workers = 4
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
	var mainWg sync.WaitGroup
	var readDirWg sync.WaitGroup
	var workersWg sync.WaitGroup
	// ----------------------------------------------------------------------------
	fmt.Println("Iniciando indexador...")

	// counting time of execution:
	startTime := time.Now()

	mainWg.Add(1)
	// email channel
	emailChannel := make(chan Indexer.Email, bulkSize)

	// workers
	for i := 0; i < num_workers; i++ {
		workersWg.Add(1)
		go Indexer.Worker(&workersWg, emailChannel, bulkSize)
	}

	// reading dir
	Indexer.ReadDirectory(directoryPath, &readDirWg, bulkSize, emailChannel)

	// wait for all goroutines readers to finish
	go func() {
		// reader
		readDirWg.Wait()
		// close email channel
		close(emailChannel)

		// workers
		workersWg.Wait()
		mainWg.Done()
	}()

	mainWg.Wait()

	// counting time of execution
	elapsedTime := time.Since(startTime)
	fmt.Println("Indexing took: ", elapsedTime)
}
