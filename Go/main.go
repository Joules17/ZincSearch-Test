package main

import (
	"ZincSearchTest/Go/Indexer"
	"ZincSearchTest/Go/Server"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

const (
	bulkSize    = 40000
	num_workers = 4
)

func main() {
	// ask user for choice
	choice, directoryPath := GetUserChoice()
	switch choice {
	case 1:
		// Start Indexing
		// Enable CPU profiling
		cpuProfileFile, err := os.Create("cpu.prof")
		if err != nil {
			fmt.Println("Error creating CPU profile:", err)
			os.Exit(1)
		}
		defer cpuProfileFile.Close()

		runtime.LockOSThread()
		pprof.StartCPUProfile(cpuProfileFile)

		// check index if exist
		Indexer.CheckIndex()

		// Initialize indexing
		StartIndexing(directoryPath)

		// Stop CPU profiling
		time.Sleep(5 * time.Second)
		pprof.StopCPUProfile()
		runtime.UnlockOSThread()
	case 2:
		// Initialize Server
		Server.InitializeServer()
	default:
		fmt.Println("Choose a valid option next time")
		os.Exit(1)
	}
}

// Start_Index:
// function that initialize indexing process
func StartIndexing(directoryPath string) {
	// WaitGroups:
	var mainWg sync.WaitGroup
	var readFileWg sync.WaitGroup
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
		fmt.Println("Se ha lanzado un Worker")
		go Indexer.Worker(&workersWg, emailChannel, bulkSize)
	}

	// reading dir
	Indexer.ReadDirectory(directoryPath, &readFileWg, bulkSize, emailChannel)

	// wait for all goroutines readers to finish
	go func() {
		// reader
		readFileWg.Wait()
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

// GetUserChoice:
// 1: Start Indexing, 2: Start Server
func GetUserChoice() (int, string) {
	fmt.Println("Select an option: ")
	fmt.Println("1: Start Indexing")
	fmt.Println("2: Start Server")

	var choice int
	var directoryPath string

	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Error reading input: ", err)
		os.Exit(1)
	}

	if choice == 1 {
		fmt.Println("Enter the directory path: ")

		_, err := fmt.Scan(&directoryPath)
		if err != nil {
			fmt.Println("Error reading input (directory path): ", err)
			os.Exit(1)
		}
	}

	return choice, directoryPath
}
