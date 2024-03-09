package Indexer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// constants settings
const (
	ZincSearchURL  = "http://localhost:4080/api/_bulkv2"
	ZincSearchUser = "admin"
	ZincSearchPass = "Complexpass#123"
)

// func sendToZincSearch
// based on email Pkg provided sends a bulkv2 petition to ZincSearch
// sendToZincSearch based on email Pkg provided sends a bulkv2 petition to ZincSearch
func SendToZincSearch(emails []Email) {
	fmt.Println("Sending Package to ZincSearch... - Size: ", len(emails))

	go func() {
		reqBody, err := json.Marshal(map[string]interface{}{
			"index":   "emails",
			"records": emails,
		})

		if err != nil {
			fmt.Println("Error marshalling email pkg: ", err)
			return
		}

		req, err := http.NewRequest("POST", ZincSearchURL, strings.NewReader(string(reqBody)))
		if err != nil {
			fmt.Println("Error creating request: ", err)
			return
		}

		req.SetBasicAuth(ZincSearchUser, ZincSearchPass)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error sending request: ", err)
			return
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error in response. Status code: ", resp.StatusCode)
			return
		}
	}()
}
