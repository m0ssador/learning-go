package ledger

import (
	"database/sql"
	"errors"
	"time"

	"github.com/m0ssador/learning-go/functions/ledger/internal/db"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

func (t Transaction) Validate() error {
	if t.Amount == 0 {
		return errors.New("transaction amount must not be zero")
	}
	if t.Amount < 0 {
		return errors.New("transaction amount must not be negative")
	}
	if t.Category == "" {
		return errors.New("transaction category must not be empty")
	}
	return nil
}

func AddTransaction(tx Transaction) (Transaction, error) {
	if err := tx.Validate(); err != nil {
		return Transaction{}, err
	}
	spentOn, err := time.Parse(time.DateOnly, tx.Date)
	if err != nil {
		return Transaction{}, errors.New("transaction date must be YYYY-MM-DD")
	}
	sqlDB := db.SQL()
	if sqlDB == nil {
		return Transaction{}, errNoDatabase
	}

	var limit float64
	err = sqlDB.QueryRow(`SELECT limit_amount FROM budgets WHERE category = $1`, tx.Category).Scan(&limit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, err
	}
	if err == nil {
		var spent float64
		if err := sqlDB.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1`, tx.Category).Scan(&spent); err != nil {
			return Transaction{}, err
		}
		if spent+tx.Amount > limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	if err := sqlDB.QueryRow(`
		INSERT INTO expenses (amount, category, description, date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, tx.Amount, tx.Category, tx.Description, spentOn).Scan(&tx.ID); err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func ListTransactions() ([]Transaction, error) {
	sqlDB := db.SQL()
	if sqlDB == nil {
		return nil, errNoDatabase
	}

	rows, err := sqlDB.Query(`
		SELECT id, amount, category, COALESCE(description, ''), date
		FROM expenses
		ORDER BY date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Transaction, 0)
	for rows.Next() {
		var tx Transaction
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
