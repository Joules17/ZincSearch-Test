package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Query struct {
	Query string `json:"input"`
}

func main() {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Options("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})

	r.Post("/search", search)

	http.ListenAndServe(":3000", r)
}

func search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	var body Query
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Construir la consulta a Zinc Search
	query := `{
		"query":{
			"bool":{
				"must":[
					{
						"query_string":{
							"query": "` + body.Query + `"
						}
					}
				]
			}
		},
		"sort":[
			"-@timestamp"
		],
		"from":0,
		"size":100,
		"aggs":{
			"histogram":{
				"auto_date_histogram":{
					"field":"@timestamp",
					"buckets":100
				}
			}
		}
	}`

	// Enviar la consulta a Zinc Search
	req, err := http.NewRequest("POST", "http://localhost:4080/es/emails/_search", strings.NewReader(query))
	if err != nil {
		http.Error(w, "Failed to create request to Zinc Search", http.StatusInternalServerError)
		return
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request to Zinc Search", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Leer y escribir la respuesta al cliente
	results, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from Zinc Search", http.StatusInternalServerError)
		return
	}

	w.Write(results)
}
