package pg

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/domain"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Add(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	spentOn, err := time.Parse(time.DateOnly, tx.Date)
	if err != nil {
		return domain.Transaction{}, errors.New("transaction date must be YYYY-MM-DD")
	}

	err = r.db.QueryRowContext(ctx, `
		INSERT INTO expenses (amount, category, description, date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, tx.Amount, tx.Category, tx.Description, spentOn).Scan(&tx.ID)
	if err != nil {
		return domain.Transaction{}, err
	}
	return tx, nil
}

func (r *ExpenseRepository) List(ctx context.Context) ([]domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, amount, category, COALESCE(description, ''), date
		FROM expenses
		ORDER BY date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Transaction, 0)
	for rows.Next() {
		var tx domain.Transaction
		var spentOn time.Time
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &spentOn); err != nil {
			return nil, err
		}
		tx.Date = spentOn.Format(time.DateOnly)
		items = append(items, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ExpenseRepository) SumByCategory(ctx context.Context, category string) (float64, error) {
	var spent float64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1
	`, category).Scan(&spent)
	if err != nil {
		return 0, err
	}
	return spent, nil
}

func (r *ExpenseRepository) Summary(ctx context.Context, from, to time.Time) ([]domain.SummaryItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT category, COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE date >= $1 AND date <= $2
		GROUP BY category
		ORDER BY category
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SummaryItem, 0)
	for rows.Next() {
		var item domain.SummaryItem
		if err := rows.Scan(&item.Category, &item.Total); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
