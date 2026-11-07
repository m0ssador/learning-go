package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

func LoadBudgets(ctx context.Context, svc Service, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read budgets: %w", err)
	}

	var items []domain.Budget
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parse budgets: %w", err)
	}

	for i, budget := range items {
		if err := budget.Validate(); err != nil {
			return fmt.Errorf("budget %d: %w", i+1, err)
		}
	}
	for _, budget := range items {
		if err := svc.SetBudget(ctx, budget); err != nil {
			return err
		}
	}
	return nil
}
