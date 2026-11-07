package main

import "errors"

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

var transactions []Transaction

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

func AddTransaction(tx Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	if b, ok := budgets[tx.Category]; ok {
		var spent float64
		for _, existing := range transactions {
			if existing.Category == tx.Category {
				spent += existing.Amount
			}
		}
		if spent+tx.Amount > b.Limit {
			return errors.New("budget exceeded")
		}
	}
	tx.ID = len(transactions) + 1
	transactions = append(transactions, tx)
	return nil
}

func ListTransactions() []Transaction {
	result := make([]Transaction, len(transactions))
	copy(result, transactions)
	return result
}
