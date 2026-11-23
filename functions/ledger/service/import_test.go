package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
	"github.com/m0ssador/learning-go/functions/ledger/service"
)

func TestImportTransactionsMixed(t *testing.T) {
	budgets := &memBudgets{items: map[string]domain.Budget{
		"еда": {Category: "еда", Limit: 5000},
	}}
	expenses := &memExpenses{}
	svc := service.New(budgets, expenses)

	txs := []service.Transaction{
		{Amount: 100, Category: "еда", Description: "хлеб", Date: "2025-09-10"},
		{Amount: 0, Category: "еда", Description: "пусто", Date: "2025-09-10"},
		{Amount: 6000, Category: "еда", Description: "сверх", Date: "2025-09-11"},
		{Amount: 50, Category: "", Description: "без категории", Date: "2025-09-10"},
		{Amount: 80, Category: "транспорт", Description: "метро", Date: "2025-09-12"},
	}

	summary, err := svc.ImportTransactions(context.Background(), txs, 3)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Accepted != 2 || summary.Rejected != 3 {
		t.Fatalf("summary: %+v", summary)
	}
	if len(summary.Errors) != 3 {
		t.Fatalf("errors: %+v", summary.Errors)
	}
	if summary.Errors[0].Index != 1 || summary.Errors[0].Message != "transaction amount must not be zero" {
		t.Fatalf("error 0: %+v", summary.Errors[0])
	}
	if summary.Errors[1].Index != 2 || summary.Errors[1].Message != "budget exceeded" {
		t.Fatalf("error 1: %+v", summary.Errors[1])
	}
	if summary.Errors[2].Index != 3 || summary.Errors[2].Message != "transaction category must not be empty" {
		t.Fatalf("error 2: %+v", summary.Errors[2])
	}

	saved, err := svc.ListTransactions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 {
		t.Fatalf("saved: %+v", saved)
	}
}

func TestImportTransactionsCancel(t *testing.T) {
	svc := service.New(&memBudgets{items: map[string]domain.Budget{}}, &memExpenses{})
	txs := []service.Transaction{
		{Amount: 10, Category: "еда", Date: "2025-09-10"},
		{Amount: 20, Category: "еда", Date: "2025-09-10"},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	var summary service.ImportSummary
	var err error
	go func() {
		defer close(done)
		summary, err = svc.ImportTransactions(ctx, txs, 2)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("import hung")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v", err)
	}
	if summary.Accepted+summary.Rejected > len(txs) {
		t.Fatalf("summary: %+v", summary)
	}
}

type memBudgets struct {
	mu    sync.Mutex
	items map[string]domain.Budget
}

func (m *memBudgets) Save(ctx context.Context, budget domain.Budget) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[budget.Category] = budget
	return nil
}

func (m *memBudgets) GetByCategory(ctx context.Context, category string) (domain.Budget, error) {
	if err := ctx.Err(); err != nil {
		return domain.Budget{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	budget, ok := m.items[category]
	if !ok {
		return domain.Budget{}, domain.ErrNotFound
	}
	return budget, nil
}

func (m *memBudgets) List(ctx context.Context) ([]domain.Budget, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Budget, 0, len(m.items))
	for _, budget := range m.items {
		out = append(out, budget)
	}
	return out, nil
}

type memExpenses struct {
	mu    sync.Mutex
	items []domain.Transaction
	next  int
}

func (m *memExpenses) Add(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	if err := ctx.Err(); err != nil {
		return domain.Transaction{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	tx.ID = m.next
	m.items = append(m.items, tx)
	return tx, nil
}

func (m *memExpenses) List(ctx context.Context) ([]domain.Transaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Transaction, len(m.items))
	copy(out, m.items)
	return out, nil
}

func (m *memExpenses) SumByCategory(ctx context.Context, category string) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var sum float64
	for _, tx := range m.items {
		if tx.Category == category {
			sum += tx.Amount
		}
	}
	return sum, nil
}

func (m *memExpenses) Categories(context.Context, time.Time, time.Time) ([]string, error) {
	return nil, nil
}

func (m *memExpenses) SumBetween(context.Context, string, time.Time, time.Time) (float64, error) {
	return 0, nil
}

func (m *memExpenses) Summary(context.Context, time.Time, time.Time) ([]domain.SummaryItem, error) {
	return nil, nil
}
