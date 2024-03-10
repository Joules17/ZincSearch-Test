package Server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitializeServer() {
	fmt.Println("Initializing server...")
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Use(CorsHandler)

	r.Post("/search", Search)

	serverAddr := ":3000"
	fmt.Printf("Server is now running and listening on %s\n", serverAddr)

	err := http.ListenAndServe(serverAddr, r)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func CorsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
