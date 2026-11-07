package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Save(ctx context.Context, budget domain.Budget) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO budgets (category, limit_amount)
		VALUES ($1, $2)
		ON CONFLICT (category) DO UPDATE SET limit_amount = EXCLUDED.limit_amount
	`, budget.Category, budget.Limit)
	return err
}

func (r *BudgetRepository) GetByCategory(ctx context.Context, category string) (domain.Budget, error) {
	var budget domain.Budget
	err := r.db.QueryRowContext(ctx, `
		SELECT category, limit_amount FROM budgets WHERE category = $1
	`, category).Scan(&budget.Category, &budget.Limit)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Budget{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Budget{}, err
	}
	return budget, nil
}

func (r *BudgetRepository) List(ctx context.Context) ([]domain.Budget, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT category, limit_amount FROM budgets ORDER BY category
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Budget, 0)
	for rows.Next() {
		var budget domain.Budget
		if err := rows.Scan(&budget.Category, &budget.Limit); err != nil {
			return nil, err
		}
		items = append(items, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
