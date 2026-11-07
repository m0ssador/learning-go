package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/m0ssador/learning-go/functions/gateway/internal/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	})
	mux.HandleFunc("/api/transactions", api.Transactions)
	mux.HandleFunc("/api/budgets", api.Budgets)

	addr := ":8080"
	log.Printf("Gateway listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, api.LogRequests(mux)))
}
