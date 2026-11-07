package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/m0ssador/learning-go/functions/gateway/internal/api"
	"github.com/m0ssador/learning-go/functions/ledger/app"
)

func main() {
	svc, closeFn, err := app.New(context.Background())
	if err != nil {
		log.Fatalf("не удалось подключиться к базе: %v", err)
	}
	defer closeFn()

	h := api.New(svc)
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
	mux.HandleFunc("/api/transactions", h.Transactions)
	mux.HandleFunc("/api/budgets", h.Budgets)
	mux.HandleFunc("/api/reports/summary", h.Summary)

	addr := ":8080"
	log.Printf("Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, api.LogRequests(mux)); err != nil {
		log.Printf("gateway: %v", err)
	}
}
