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

func AddTransaction(tx Transaction) error {
	if tx.Amount == 0 {
		return errors.New("transaction amount must not be null")
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
