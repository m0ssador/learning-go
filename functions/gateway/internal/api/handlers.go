package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger"
)

func Transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createTransaction(w, r)
	case http.MethodGet:
		listTransactions(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func Budgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createBudget(w, r)
	case http.MethodGet:
		listBudgets(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	tx := req.ToTransaction()
	if err := tx.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := ledger.AddTransaction(tx)
	if errors.Is(err, ledger.ErrBudgetExceeded) {
		writeError(w, http.StatusConflict, "budget exceeded")
		return
	}
	if err != nil {
		log.Printf("add transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, TransactionResponseFrom(created))
}

func listTransactions(w http.ResponseWriter, _ *http.Request) {
	items, err := ledger.ListTransactions()
	if err != nil {
		log.Printf("list transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	out := make([]TransactionResponse, 0, len(items))
	for _, tx := range items {
		out = append(out, TransactionResponseFrom(tx))
	}
	writeJSON(w, http.StatusOK, out)
}

func createBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	budget := req.ToBudget()
	if err := budget.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := ledger.SetBudget(budget); err != nil {
		log.Printf("set budget: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, BudgetResponseFrom(budget))
}

func listBudgets(w http.ResponseWriter, _ *http.Request) {
	items, err := ledger.ListBudgets()
	if err != nil {
		log.Printf("list budgets: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	out := make([]BudgetResponse, 0, len(items))
	for _, b := range items {
		out = append(out, BudgetResponseFrom(b))
	}
	writeJSON(w, http.StatusOK, out)
}

func Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	from, err := time.Parse(time.DateOnly, r.URL.Query().Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date")
		return
	}
	to, err := time.Parse(time.DateOnly, r.URL.Query().Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date")
		return
	}
	if from.After(to) {
		writeError(w, http.StatusBadRequest, "invalid date")
		return
	}

	items, err := ledger.GetReportSummary(r.Context(), from, to)
	if err != nil {
		log.Printf("report summary: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
