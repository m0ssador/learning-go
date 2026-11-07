package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/service"
)

type Handler struct {
	svc service.Service
}

func New(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTransaction(w, r)
	case http.MethodGet:
		h.listTransactions(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) Budgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createBudget(w, r)
	case http.MethodGet:
		h.listBudgets(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	created, err := h.svc.AddTransaction(r.Context(), req.ToTransaction())
	if errors.Is(err, service.ErrBudgetExceeded) {
		writeError(w, http.StatusConflict, "budget exceeded")
		return
	}
	var invalid *service.ValidationError
	if errors.As(err, &invalid) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Printf("add transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, TransactionResponseFrom(created))
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListTransactions(r.Context())
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

func (h *Handler) createBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	budget := req.ToBudget()
	if err := h.svc.SetBudget(r.Context(), budget); err != nil {
		var invalid *service.ValidationError
		if errors.As(err, &invalid) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("set budget: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, BudgetResponseFrom(budget))
}

func (h *Handler) listBudgets(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListBudgets(r.Context())
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

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.svc.ReportSummary(r.Context(), from, to)
	if err != nil {
		log.Printf("report summary: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
