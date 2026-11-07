package cached

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

const (
	budgetsKey = "budgets:all"
	budgetsTTL = 30 * time.Second
)

type Budgets struct {
	next domain.BudgetRepository
}

func NewBudgets(next domain.BudgetRepository) *Budgets {
	return &Budgets{next: next}
}

func (r *Budgets) Save(ctx context.Context, budget domain.Budget) error {
	if err := r.next.Save(ctx, budget); err != nil {
		return err
	}
	return cache.Del(ctx, budgetsKey)
}

func (r *Budgets) GetByCategory(ctx context.Context, category string) (domain.Budget, error) {
	return r.next.GetByCategory(ctx, category)
}

func (r *Budgets) List(ctx context.Context) ([]domain.Budget, error) {
	if raw, ok, err := cache.Get(ctx, budgetsKey); err != nil {
		return nil, err
	} else if ok {
		var items []domain.Budget
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, err
		}
		log.Printf("cache hit %s", budgetsKey)
		return items, nil
	}

	log.Printf("cache miss %s", budgetsKey)
	items, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	if err := cache.Set(ctx, budgetsKey, string(raw), budgetsTTL); err != nil {
		return nil, err
	}
	return items, nil
}
