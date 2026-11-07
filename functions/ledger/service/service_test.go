package service_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/m0ssador/learning-go/functions/ledger/app"
	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

func TestAddTransactionBudgetLimit(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx := context.Background()
	svc, closeFn, err := app.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeFn() })
	app.Reset(ctx)
	t.Cleanup(func() { app.Reset(ctx) })

	if err := svc.SetBudget(ctx, domain.Budget{Category: "еда", Limit: 5000}); err != nil {
		t.Fatal(err)
	}

	_, err = svc.AddTransaction(ctx, domain.Transaction{
		Amount:      1500.50,
		Category:    "еда",
		Description: "продукты",
		Date:        "2025-09-10",
	})
	if err != nil {
		t.Fatal(err)
	}

	before, err := svc.ListTransactions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.AddTransaction(ctx, domain.Transaction{
		Amount:      4000,
		Category:    "еда",
		Description: "сверх лимита",
		Date:        "2025-09-11",
	})
	if !errors.Is(err, domain.ErrBudgetExceeded) {
		t.Fatalf("got %v, want budget exceeded", err)
	}
	after, err := svc.ListTransactions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("transactions after reject: %d, want %d", len(after), len(before))
	}
}
