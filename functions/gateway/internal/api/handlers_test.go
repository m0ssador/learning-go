package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransactionsAndBudgets(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions", Transactions)
	mux.HandleFunc("/api/budgets", Budgets)
	srv := httptest.NewServer(LogRequests(mux))
	t.Cleanup(srv.Close)

	resp := get(t, srv.URL+"/api/transactions")
	if got := readBody(t, resp); strings.TrimSpace(got) != "[]" {
		t.Fatalf("empty transactions: %s", got)
	}

	resp = get(t, srv.URL+"/api/budgets")
	if got := readBody(t, resp); strings.TrimSpace(got) != "[]" {
		t.Fatalf("empty budgets: %s", got)
	}

	resp = post(t, srv.URL+"/api/transactions", CreateTransactionRequest{Amount: 0, Category: "еда"})
	expectStatus(t, resp, http.StatusBadRequest)
	if msg := readError(t, resp); msg != "transaction amount must not be zero" {
		t.Fatalf("zero amount error: %s", msg)
	}

	resp = post(t, srv.URL+"/api/transactions", CreateTransactionRequest{Amount: -5, Category: "еда"})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	resp = post(t, srv.URL+"/api/transactions", CreateTransactionRequest{Amount: 10, Category: ""})
	expectStatus(t, resp, http.StatusBadRequest)
	if msg := readError(t, resp); msg != "transaction category must not be empty" {
		t.Fatalf("empty category error: %s", msg)
	}

	resp = postRaw(t, srv.URL+"/api/transactions", []byte("{"))
	expectStatus(t, resp, http.StatusBadRequest)
	if msg := readError(t, resp); msg != "invalid json" {
		t.Fatalf("invalid json error: %s", msg)
	}

	resp = post(t, srv.URL+"/api/transactions", CreateTransactionRequest{
		Amount: 450, Category: "еда", Description: "ланч", Date: "2025-09-10",
	})
	expectStatus(t, resp, http.StatusCreated)
	var created TransactionResponse
	decode(t, resp, &created)
	if created.ID != 1 || created.Amount != 450 || created.Category != "еда" || created.Description != "ланч" || created.Date != "2025-09-10" {
		t.Fatalf("created transaction: %+v", created)
	}

	resp = post(t, srv.URL+"/api/budgets", CreateBudgetRequest{Category: "еда", Limit: 0})
	expectStatus(t, resp, http.StatusBadRequest)
	if msg := readError(t, resp); msg != "budget limit must be positive" {
		t.Fatalf("budget limit error: %s", msg)
	}

	resp = post(t, srv.URL+"/api/budgets", CreateBudgetRequest{Category: "", Limit: 1000})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	resp = post(t, srv.URL+"/api/budgets", CreateBudgetRequest{Category: "еда", Limit: 500})
	expectStatus(t, resp, http.StatusCreated)
	var budget BudgetResponse
	decode(t, resp, &budget)
	if budget.Category != "еда" || budget.Limit != 500 {
		t.Fatalf("created budget: %+v", budget)
	}

	resp = post(t, srv.URL+"/api/transactions", CreateTransactionRequest{
		Amount: 100, Category: "еда", Description: "сверх лимита", Date: "2025-09-10",
	})
	expectStatus(t, resp, http.StatusConflict)
	if msg := readError(t, resp); msg != "budget exceeded" {
		t.Fatalf("conflict error: %s", msg)
	}

	resp = post(t, srv.URL+"/api/transactions", CreateTransactionRequest{
		Amount: 50, Category: "еда", Description: "ровно до лимита", Date: "2025-09-11",
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	resp = get(t, srv.URL+"/api/transactions")
	expectStatus(t, resp, http.StatusOK)
	var txs []TransactionResponse
	decode(t, resp, &txs)
	if len(txs) != 2 || txs[0].Amount != 450 || txs[1].Amount != 50 {
		t.Fatalf("transactions after rejects: %+v", txs)
	}

	resp = get(t, srv.URL+"/api/budgets")
	expectStatus(t, resp, http.StatusOK)
	var budgets []BudgetResponse
	decode(t, resp, &budgets)
	if len(budgets) != 1 || budgets[0].Category != "еда" || budgets[0].Limit != 500 {
		t.Fatalf("budgets: %+v", budgets)
	}

	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/transactions", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, resp, http.StatusMethodNotAllowed)
	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type %q", ct)
	}
	resp.Body.Close()
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	expectJSON(t, resp)
	return resp
}

func post(t *testing.T, url string, payload any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return postRaw(t, url, body)
}

func postRaw(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	expectJSON(t, resp)
	return resp
}

func expectJSON(t *testing.T, resp *http.Response) {
	t.Helper()
	if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type %q", ct)
	}
}

func expectStatus(t *testing.T, resp *http.Response, status int) {
	t.Helper()
	if resp.StatusCode != status {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d, want %d, body %s", resp.StatusCode, status, body)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func readError(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	var body errorBody
	decode(t, resp, &body)
	return body.Error
}

func decode(t *testing.T, resp *http.Response, dest any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		t.Fatal(err)
	}
}
