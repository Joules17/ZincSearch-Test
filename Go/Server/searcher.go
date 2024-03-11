package Server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Query struct {
	Term     string `json:"input"`
	From     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// funct search
// makes a search petition to ZincSearch for every fields of index
func Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	var body Query

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Query:", body.Term, body.From, body.PageSize)

	// query for zincsearch
	query := map[string]interface{}{
		"search_type": "match",
		"query": map[string]interface{}{
			"term": body.Term,
		},
		"from":        body.From,
		"max_results": body.PageSize,
		"_source":     []string{},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ZincSearch Query:", string(queryJSON))

	// make request to zincsearch
	zincURL := "http://localhost:4080/api/emails/_search"
	req, err := http.NewRequest("POST", zincURL, strings.NewReader(string(queryJSON)))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	log.Println("ZincSearch Response Status:", resp.StatusCode)

	// results
	results, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(results)
}
