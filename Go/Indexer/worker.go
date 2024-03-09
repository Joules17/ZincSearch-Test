package Indexer

import (
	"fmt"
	"sync"
)

func Worker(wg *sync.WaitGroup, emailChannel <-chan Email, bulkSize int) {
	defer wg.Done()

	var EmailList []Email

	for emails := range emailChannel {
		EmailList = append(EmailList, emails)
		if len(EmailList) >= bulkSize {
			fmt.Println("Pkg created!!!!! ")
			go SendToZincSearch(EmailList)
			EmailList = nil
		}
	}

	if len(EmailList) > 0 {
		fmt.Println("Sending remaining emails")
		go SendToZincSearch(EmailList)
	}
}
