package main

import "errors"

// Transaction — финансовая транзакция.
type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

// transactions — хранилище транзакций в памяти.
var transactions []Transaction

// AddTransaction добавляет транзакцию в хранилище.
// ID заполняется автоинкрементом. Сумма 0 считается ошибкой.
func AddTransaction(tx Transaction) error {
	if tx.Amount == 0 {
		return errors.New("transaction amount must not be zero")
	}
	tx.ID = len(transactions) + 1
	transactions = append(transactions, tx)
	return nil
}

// ListTransactions возвращает копию всех сохранённых транзакций.
func ListTransactions() []Transaction {
	result := make([]Transaction, len(transactions))
	copy(result, transactions)
	return result
}
