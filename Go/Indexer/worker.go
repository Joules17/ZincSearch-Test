package Indexer

import (
	"fmt"
	"sync"
)

func Worker(wg *sync.WaitGroup, emailChannel <-chan []Email) {
	defer wg.Done()

	for emails := range emailChannel {
		fmt.Println("Pkg created and sent it to ZincSearch...")
		SendToZincSearch(emails)
	}
}
