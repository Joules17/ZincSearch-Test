package Indexer

import (
	"sync"
)

func Worker(pkgChan <-chan EmailPkg, wg *sync.WaitGroup, bulkSize int) {
	defer wg.Done()

	var emails []Email

	for pkg := range pkgChan {
		emails = append(emails, pkg.Emails...)

		if len(emails) >= bulkSize {
			SendToZincSearch(emails[:bulkSize])
			emails = emails[bulkSize:]
		}
	}

	if len(emails) > 0 {
		SendToZincSearch(emails)
	}

}
