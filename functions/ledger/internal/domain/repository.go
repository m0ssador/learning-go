package domain

import (
	"context"
	"time"
)

type SummaryItem struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

type BudgetRepository interface {
	Save(ctx context.Context, budget Budget) error
	GetByCategory(ctx context.Context, category string) (Budget, error)
	List(ctx context.Context) ([]Budget, error)
}

type ExpenseRepository interface {
	Add(ctx context.Context, tx Transaction) (Transaction, error)
	List(ctx context.Context) ([]Transaction, error)
	SumByCategory(ctx context.Context, category string) (float64, error)
	Summary(ctx context.Context, from, to time.Time) ([]SummaryItem, error)
}
