package Indexer

import (
	"fmt"
	"net/http"
)

// constants
const (
	IndexName          = "emails"
	ZincSearchURLCheck = "http://localhost:4080/api/index/"
)

// check index function
// sends a check index petition to ZincSearch
func CheckIndex() {
	fmt.Println("Checking if index exists...")

	req, err := http.NewRequest("HEAD", ZincSearchURLCheck+IndexName, nil)
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return
	}

	req.SetBasicAuth(ZincSearchUser, ZincSearchPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error sending request ", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("Index already exists")
		DeleteIndex()
		return
	}
}

// delete index function
// sends a delete index petition to ZincSearch
func DeleteIndex() {
	fmt.Println("Deleting index...")

	req, err := http.NewRequest("DELETE", ZincSearchURLCheck+IndexName, nil)
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return
	}

	req.SetBasicAuth(ZincSearchUser, ZincSearchPass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error sending request")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error in response. Status code: ", resp.StatusCode)
		return
	}
}
