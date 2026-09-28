package main

import "fmt"

func main() {
	transactions = make([]Transaction, 0)
	fmt.Println("Ledger service started")

	samples := []Transaction{
		{
			Amount:      1500.50,
			Category:    "еда",
			Description: "продукты на неделю",
			Date:        "2026-09-26",
		},
		{
			Amount:      89.00,
			Category:    "транспорт",
			Description: "проездной",
			Date:        "2026-09-27",
		},
		{
			Amount:      3200.00,
			Category:    "жильё",
			Description: "коммунальные платежи",
			Date:        "2026-09-28",
		},
	}

	for _, tx := range samples {
		if err := AddTransaction(tx); err != nil {
			fmt.Printf("не удалось добавить транзакцию: %v\n", err)
			continue
		}
	}

	fmt.Println("Transactions:")
	for _, tx := range ListTransactions() {
		fmt.Printf("ID=%d Amount=%.2f Category=%s Description=%s Date=%s\n",
			tx.ID, tx.Amount, tx.Category, tx.Description, tx.Date)
	}
}
