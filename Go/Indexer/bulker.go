package Indexer

import (
	"encoding/json"
	"fmt"
	"io"
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
	// fmt.Println("Sending Package to ZincSearch...")

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
	// req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request: ", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error in response. Status code: ", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err)
		return
	}

	// fmt.Println("Package sent successfully!")
	fmt.Println("Response body: ", string(body))
}
