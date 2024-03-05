package main

import (
	"ZincSearchTest/Go/Indexer"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	bulkSize    = 1000
	num_workers = 6
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file to index next time. ")
		os.Exit(1)
	}

	directoryPath := os.Args[1]

	// check index if exist
	Indexer.CheckIndex()

	// initialize indexing
	StartIndexing(directoryPath)
}

// Start_Index:
// function that initialize indexing process
func StartIndexing(directoryPath string) {
	// WaitGroups:
	var workerWg sync.WaitGroup
	var readDirWg sync.WaitGroup
	var mainWg sync.WaitGroup

	// ----------------------------------------------------------------------------
	fmt.Println("Iniciando indexador...")
	mainWg.Add(1)

	fmt.Println("Main goroutine started")
	defer fmt.Println("Main goroutine completed")

	// setting channel
	pkgChan := make(chan Indexer.EmailPkg, bulkSize)

	for i := 0; i < num_workers; i++ {
		workerWg.Add(1)
		go Indexer.Worker(pkgChan, &workerWg, bulkSize)
	}

	// reading dirs
	readDirWg.Add(1)
	go Indexer.ReadDirectory(directoryPath, &readDirWg, pkgChan, bulkSize)

	// wait for all go routines readers to finish
	go func() {
		// reader
		readDirWg.Wait()
		close(pkgChan)

		// workers
		workerWg.Wait()
		mainWg.Done()
	}()

	// counting time of execution:
	startTime := time.Now()

	// wait for all go routines to finish
	mainWg.Wait()

	// counting time of execution
	elapsedTime := time.Since(startTime)
	fmt.Println("Indexing took: ", elapsedTime)
}
