package ledger

import "errors"

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

	mu.Lock()
	defer mu.Unlock()

	if b, ok := budgets[tx.Category]; ok {
		var spent float64
		for _, existing := range transactions {
			if existing.Category == tx.Category {
				spent += existing.Amount
			}
		}
		if spent+tx.Amount > b.Limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	tx.ID = len(transactions) + 1
	transactions = append(transactions, tx)
	return tx, nil
}

func ListTransactions() []Transaction {
	mu.Lock()
	defer mu.Unlock()

	result := make([]Transaction, len(transactions))
	copy(result, transactions)
	return result
}
