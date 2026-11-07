package cached

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

const reportTTL = 30 * time.Second

type Expenses struct {
	next domain.ExpenseRepository
}

func NewExpenses(next domain.ExpenseRepository) *Expenses {
	return &Expenses{next: next}
}

func (r *Expenses) Add(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	return r.next.Add(ctx, tx)
}

func (r *Expenses) List(ctx context.Context) ([]domain.Transaction, error) {
	return r.next.List(ctx)
}

func (r *Expenses) SumByCategory(ctx context.Context, category string) (float64, error) {
	return r.next.SumByCategory(ctx, category)
}

func (r *Expenses) Summary(ctx context.Context, from, to time.Time) ([]domain.SummaryItem, error) {
	key := fmt.Sprintf("report:summary:%s:%s", from.Format(time.DateOnly), to.Format(time.DateOnly))
	if raw, ok, err := cache.Get(ctx, key); err != nil {
		return nil, err
	} else if ok {
		var items []domain.SummaryItem
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, err
		}
		log.Printf("cache hit %s", key)
		return items, nil
	}

	log.Printf("cache miss %s", key)
	items, err := r.next.Summary(ctx, from, to)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	if err := cache.Set(ctx, key, string(raw), reportTTL); err != nil {
		return nil, err
	}
	return items, nil
}

func Clear(ctx context.Context) error {
	if err := cache.Del(ctx, budgetsKey); err != nil {
		return err
	}
	return cache.DelByPrefix(ctx, "report:summary:")
}
