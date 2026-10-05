package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

var budgets map[string]Budget

func SetBudget(b Budget) {
	if budgets == nil {
		budgets = make(map[string]Budget)
	}
	budgets[b.Category] = b
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
		if b.Category == "" {
			return fmt.Errorf("budget %d: empty category", i+1)
		}
	}
	for _, b := range items {
		SetBudget(b)
	}
	return nil
}
