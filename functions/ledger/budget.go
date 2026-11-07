package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
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

	mu.Lock()
	defer mu.Unlock()

	if budgets == nil {
		budgets = make(map[string]Budget)
	}
	budgets[b.Category] = b
	return nil
}

func ListBudgets() []Budget {
	mu.Lock()
	defer mu.Unlock()

	result := make([]Budget, 0, len(budgets))
	for _, b := range budgets {
		result = append(result, b)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Category < result[j].Category
	})
	return result
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
