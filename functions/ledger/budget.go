package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/db"
)

const (
	budgetsCacheKey = "budgets:all"
	budgetsCacheTTL = 30 * time.Second
)

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

func (b Budget) Validate() error {
	if b.Limit <= 0 {
		return errors.New("budget limit must be positive")
	}
	if b.Category == "" {
		return errors.New("budget category must not be empty")
	}
	return nil
}

func SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	sqlDB := db.SQL()
	if sqlDB == nil {
		return errNoDatabase
	}

	_, err := sqlDB.Exec(`
		INSERT INTO budgets (category, limit_amount)
		VALUES ($1, $2)
		ON CONFLICT (category) DO UPDATE SET limit_amount = EXCLUDED.limit_amount
	`, b.Category, b.Limit)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := cache.Del(ctx, budgetsCacheKey); err != nil {
		return err
	}
	return nil
}

func ListBudgets() ([]Budget, error) {
	sqlDB := db.SQL()
	if sqlDB == nil {
		return nil, errNoDatabase
	}

	ctx := context.Background()
	if raw, ok, err := cache.Get(ctx, budgetsCacheKey); err != nil {
		return nil, err
	} else if ok {
		var items []Budget
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, err
		}
		log.Printf("cache hit %s", budgetsCacheKey)
		return items, nil
	}

	log.Printf("cache miss %s", budgetsCacheKey)
	rows, err := sqlDB.Query(`SELECT category, limit_amount FROM budgets ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Budget, 0)
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	if err := cache.Set(ctx, budgetsCacheKey, string(raw), budgetsCacheTTL); err != nil {
		return nil, err
	}
	return items, nil
}

func LoadBudgets(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read budgets: %w", err)
	}

	var items []Budget
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parse budgets: %w", err)
	}

	for i, b := range items {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("budget %d: %w", i+1, err)
		}
	}
	for _, b := range items {
		if err := SetBudget(b); err != nil {
			return err
		}
	}
	return nil
}
