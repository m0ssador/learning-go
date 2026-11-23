package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/app"
	"github.com/m0ssador/learning-go/functions/ledger/service"
)

var svc service.Service

func TestMain(m *testing.M) {
	var closeFn func() error
	if os.Getenv("DATABASE_URL") != "" {
		var err error
		svc, closeFn, err = app.New(context.Background())
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
			os.Exit(1)
		}
	}
	code := m.Run()
	if closeFn != nil {
		_ = closeFn()
	}
	os.Exit(code)
}

func TestBudgetsAPI(t *testing.T) {
	if svc == nil {
		t.Skip("DATABASE_URL is not set")
	}
	srv := newAPIServer(t)

	t.Run("ok", func(t *testing.T) {
		app.Reset(context.Background())
		t.Cleanup(func() { app.Reset(context.Background()) })

		resp := postJSON(t, srv.URL+"/api/budgets", CreateBudgetRequest{
			Category: "еда",
			Limit:    5000,
		})
		assertStatus(t, resp, http.StatusCreated)
		assertContentType(t, resp)
		var created BudgetResponse
		decode(t, resp, &created)
		if created.Category != "еда" || created.Limit != 5000 {
			t.Fatalf("created budget: %+v", created)
		}

		resp = getJSON(t, srv.URL+"/api/budgets")
		assertStatus(t, resp, http.StatusOK)
		assertContentType(t, resp)
		var list []BudgetResponse
		decode(t, resp, &list)
		if len(list) != 1 || list[0].Category != "еда" || list[0].Limit != 5000 {
			t.Fatalf("budgets: %+v", list)
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		app.Reset(context.Background())
		t.Cleanup(func() { app.Reset(context.Background()) })

		resp := postRaw(t, srv.URL+"/api/budgets", []byte("{"))
		assertStatus(t, resp, http.StatusBadRequest)
		assertContentType(t, resp)
		if msg := readError(t, resp); msg != "invalid json" {
			t.Fatalf("error %q", msg)
		}
	})

	t.Run("bad_limit", func(t *testing.T) {
		app.Reset(context.Background())
		t.Cleanup(func() { app.Reset(context.Background()) })

		resp := postJSON(t, srv.URL+"/api/budgets", CreateBudgetRequest{
			Category: "еда",
			Limit:    0,
		})
		assertStatus(t, resp, http.StatusBadRequest)
		assertContentType(t, resp)
		if msg := readError(t, resp); msg != "budget limit must be positive" {
			t.Fatalf("error %q", msg)
		}
	})
}

func TestTransactionsAPI(t *testing.T) {
	if svc == nil {
		t.Skip("DATABASE_URL is not set")
	}
	srv := newAPIServer(t)

	t.Run("ok", func(t *testing.T) {
		app.Reset(context.Background())
		t.Cleanup(func() { app.Reset(context.Background()) })

		resp := postJSON(t, srv.URL+"/api/budgets", CreateBudgetRequest{
			Category: "еда",
			Limit:    5000,
		})
		assertStatus(t, resp, http.StatusCreated)
		resp.Body.Close()

		resp = postJSON(t, srv.URL+"/api/transactions", CreateTransactionRequest{
			Amount:      450,
			Category:    "еда",
			Description: "ланч",
			Date:        "2025-09-10",
		})
		assertStatus(t, resp, http.StatusCreated)
		assertContentType(t, resp)
		var created TransactionResponse
		decode(t, resp, &created)
		if created.ID == 0 || created.Amount != 450 || created.Category != "еда" || created.Description != "ланч" || created.Date != "2025-09-10" {
			t.Fatalf("created transaction: %+v", created)
		}

		resp = getJSON(t, srv.URL+"/api/transactions")
		assertStatus(t, resp, http.StatusOK)
		assertContentType(t, resp)
		var list []TransactionResponse
		decode(t, resp, &list)
		if len(list) != 1 || list[0].ID != created.ID || list[0].Amount != 450 || list[0].Date != "2025-09-10" {
			t.Fatalf("transactions: %+v", list)
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		app.Reset(context.Background())
		t.Cleanup(func() { app.Reset(context.Background()) })

		resp := postJSON(t, srv.URL+"/api/budgets", CreateBudgetRequest{
			Category: "еда",
			Limit:    5000,
		})
		assertStatus(t, resp, http.StatusCreated)
		resp.Body.Close()

		resp = postJSON(t, srv.URL+"/api/transactions", CreateTransactionRequest{
			Amount:      4500,
			Category:    "еда",
			Description: "закупка",
			Date:        "2025-09-10",
		})
		assertStatus(t, resp, http.StatusCreated)
		resp.Body.Close()

		resp = postJSON(t, srv.URL+"/api/transactions", CreateTransactionRequest{
			Amount:      600,
			Category:    "еда",
			Description: "сверх лимита",
			Date:        "2025-09-11",
		})
		assertStatus(t, resp, http.StatusConflict)
		assertContentType(t, resp)
		if msg := readError(t, resp); msg != "budget exceeded" {
			t.Fatalf("error %q", msg)
		}

		resp = getJSON(t, srv.URL+"/api/transactions")
		assertStatus(t, resp, http.StatusOK)
		var list []TransactionResponse
		decode(t, resp, &list)
		if len(list) != 1 || list[0].Amount != 4500 {
			t.Fatalf("transactions after reject: %+v", list)
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		app.Reset(context.Background())
		t.Cleanup(func() { app.Reset(context.Background()) })

		resp := postRaw(t, srv.URL+"/api/transactions", []byte("{"))
		assertStatus(t, resp, http.StatusBadRequest)
		assertContentType(t, resp)
		if msg := readError(t, resp); msg != "invalid json" {
			t.Fatalf("error %q", msg)
		}
	})
}

func TestBulkTransactions(t *testing.T) {
	stub := &stubLedger{summary: service.ImportSummary{
		Accepted: 1,
		Rejected: 1,
		Errors: []service.ImportFailure{{
			Index:   1,
			Message: "budget exceeded",
		}},
	}}
	h := New(stub)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions/bulk", h.BulkTransactions)

	t.Run("ok", func(t *testing.T) {
		rr := postBulk(t, mux, "/api/transactions/bulk?workers=3", `[{"amount":10,"category":"еда","description":"хлеб","date":"2025-09-10"},{"amount":6000,"category":"еда","description":"сверх","date":"2025-09-11"}]`)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Fatalf("Content-Type %q", got)
		}
		if stub.workers != 3 {
			t.Fatalf("workers %d", stub.workers)
		}
		var body BulkImportResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Accepted != 1 || body.Rejected != 1 || len(body.Errors) != 1 || body.Errors[0].Index != 1 || body.Errors[0].Error != "budget exceeded" {
			t.Fatalf("body %+v", body)
		}
	})

	t.Run("default_workers", func(t *testing.T) {
		stub.workers = 0
		stub.summary = service.ImportSummary{}
		rr := postBulk(t, mux, "/api/transactions/bulk", `[]`)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
		if stub.workers != service.DefaultImportWorkers {
			t.Fatalf("workers %d", stub.workers)
		}
		var body BulkImportResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Errors == nil {
			t.Fatal("errors is null")
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		rr := postBulk(t, mux, "/api/transactions/bulk", `{`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status %d", rr.Code)
		}
	})

	t.Run("bad_workers", func(t *testing.T) {
		rr := postBulk(t, mux, "/api/transactions/bulk?workers=0", `[]`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("timeout", func(t *testing.T) {
		slow := &stubLedger{wait: true}
		slowMux := http.NewServeMux()
		slowMux.HandleFunc("/api/transactions/bulk", New(slow).BulkTransactions)
		handler := RequestTimeout(20 * time.Millisecond)(slowMux)
		rr := postBulk(t, handler, "/api/transactions/bulk", `[{"amount":10,"category":"еда","description":"хлеб","date":"2025-09-10"}]`)
		if rr.Code != http.StatusGatewayTimeout {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
		var body errorBody
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Error != "request timeout" {
			t.Fatalf("error %q", body.Error)
		}
	})
}

func TestBulkTransactionsDB(t *testing.T) {
	if svc == nil {
		t.Skip("DATABASE_URL is not set")
	}
	srv := newAPIServer(t)
	app.Reset(context.Background())
	t.Cleanup(func() { app.Reset(context.Background()) })

	resp := postJSON(t, srv.URL+"/api/budgets", CreateBudgetRequest{Category: "еда", Limit: 5000})
	assertStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	resp = postRaw(t, srv.URL+"/api/transactions/bulk?workers=2", []byte(`[
		{"amount":100,"category":"еда","description":"хлеб","date":"2025-09-10"},
		{"amount":0,"category":"еда","description":"пусто","date":"2025-09-10"},
		{"amount":6000,"category":"еда","description":"сверх","date":"2025-09-11"}
	]`))
	assertStatus(t, resp, http.StatusOK)
	assertContentType(t, resp)
	var body BulkImportResponse
	decode(t, resp, &body)
	if body.Accepted != 1 || body.Rejected != 2 {
		t.Fatalf("summary: %+v", body)
	}
	if len(body.Errors) != 2 || body.Errors[0].Index != 1 || body.Errors[1].Index != 2 || body.Errors[1].Error != "budget exceeded" {
		t.Fatalf("errors: %+v", body.Errors)
	}

	resp = getJSON(t, srv.URL+"/api/transactions")
	assertStatus(t, resp, http.StatusOK)
	var list []TransactionResponse
	decode(t, resp, &list)
	if len(list) != 1 || list[0].Amount != 100 {
		t.Fatalf("saved: %+v", list)
	}
}

func postBulk(t *testing.T, handler http.Handler, url, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

type stubLedger struct {
	summary service.ImportSummary
	workers int
	wait    bool
}

func (s *stubLedger) SetBudget(context.Context, service.Budget) error { return nil }

func (s *stubLedger) ListBudgets(context.Context) ([]service.Budget, error) { return nil, nil }

func (s *stubLedger) AddTransaction(context.Context, service.Transaction) (service.Transaction, error) {
	return service.Transaction{}, nil
}

func (s *stubLedger) ListTransactions(context.Context) ([]service.Transaction, error) {
	return nil, nil
}

func (s *stubLedger) ReportSummary(context.Context, time.Time, time.Time) (map[string]float64, error) {
	return nil, nil
}

func (s *stubLedger) ImportTransactions(ctx context.Context, _ []service.Transaction, workers int) (service.ImportSummary, error) {
	s.workers = workers
	if s.wait {
		<-ctx.Done()
		return service.ImportSummary{}, ctx.Err()
	}
	return s.summary, nil
}

func newAPIServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := New(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions", h.Transactions)
	mux.HandleFunc("/api/transactions/bulk", h.BulkTransactions)
	mux.HandleFunc("/api/budgets", h.Budgets)
	srv := httptest.NewServer(LogRequests(mux))
	t.Cleanup(srv.Close)
	return srv
}

func getJSON(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func postJSON(t *testing.T, url string, payload any) *http.Response {
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
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, status int) {
	t.Helper()
	if resp.StatusCode != status {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("status %d, want %d, body %s", resp.StatusCode, status, body)
	}
}

func assertContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	const want = "application/json; charset=utf-8"
	if got := resp.Header.Get("Content-Type"); got != want {
		t.Fatalf("Content-Type %q, want %q", got, want)
	}
}

func readError(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	var body errorBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body.Error
}

func decode(t *testing.T, resp *http.Response, dest any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		t.Fatal(err)
	}
}
