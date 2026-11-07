package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

func main() {
	transactions = make([]Transaction, 0)
	budgets = make(map[string]Budget)
	fmt.Println("Ledger service started")

	checks := []struct {
		name string
		v    Validatable
	}{
		{"Transaction", Transaction{Amount: 100, Category: "еда", Date: "2026-10-01"}},
		{"Transaction", Transaction{Amount: 0, Category: "еда", Date: "2026-10-01"}},
		{"Transaction", Transaction{Amount: -5, Category: "еда", Date: "2026-10-01"}},
		{"Transaction", Transaction{Amount: 10, Category: "", Date: "2026-10-01"}},
		{"Budget", Budget{Category: "еда", Limit: 5000, Period: "month"}},
		{"Budget", Budget{Category: "", Limit: 5000}},
		{"Budget", Budget{Category: "еда", Limit: 0}},
		{"Budget", Budget{Category: "еда", Limit: -100}},
	}
	fmt.Println("CheckValid:")
	for _, check := range checks {
		if err := CheckValid(check.v); err != nil {
			fmt.Printf("отказ: %s: %v\n", check.name, err)
			continue
		}
		fmt.Printf("ok: %s\n", check.name)
	}

	f, err := os.Open("budgets.json")
	if err != nil {
		fmt.Printf("не удалось открыть budgets.json: %v\n", err)
	} else {
		defer f.Close()
		if err := LoadBudgets(bufio.NewReader(f)); err != nil {
			fmt.Printf("не удалось загрузить бюджеты: %v\n", err)
		}
	}

	if err := SetBudget(Budget{
		Category: "жильё",
		Limit:    10000,
		Period:   "month",
	}); err != nil {
		fmt.Printf("отказ: бюджет не установлен: %v\n", err)
	}

	invalidBudgets := []Budget{
		{Category: "", Limit: 1000},
		{Category: "еда", Limit: 0},
		{Category: "еда", Limit: -1},
	}
	for _, b := range invalidBudgets {
		if err := SetBudget(b); err != nil {
			fmt.Printf("отказ: бюджет не установлен (категория %q, лимит %.2f): %v\n",
				b.Category, b.Limit, err)
		}
	}

	fmt.Println("Budgets:")
	printBudgets()

	samples := []Transaction{
		{
			Amount:      1500.50,
			Category:    "еда",
			Description: "продукты на неделю",
			Date:        "2026-09-26",
		},
		{
			Amount:      3499.50,
			Category:    "еда",
			Description: "ровно до лимита",
			Date:        "2026-09-26",
		},
		{
			Amount:      100.00,
			Category:    "еда",
			Description: "сверх лимита на еду",
			Date:        "2026-09-26",
		},
		{
			Amount:      89.00,
			Category:    "транспорт",
			Description: "проездной",
			Date:        "2026-09-27",
		},
		{
			Amount:      150.00,
			Category:    "транспорт",
			Description: "сверх лимита на транспорт",
			Date:        "2026-09-27",
		},
		{
			Amount:      3200.00,
			Category:    "жильё",
			Description: "коммунальные платежи",
			Date:        "2026-09-28",
		},
		{
			Amount:      450.00,
			Category:    "развлечения",
			Description: "кино, бюджета на категорию нет",
			Date:        "2026-09-29",
		},
		{
			Amount:      0,
			Category:    "еда",
			Description: "нулевая сумма",
			Date:        "2026-09-30",
		},
		{
			Amount:      -20,
			Category:    "еда",
			Description: "отрицательная сумма",
			Date:        "2026-09-30",
		},
		{
			Amount:      50,
			Category:    "",
			Description: "пустая категория",
			Date:        "2026-09-30",
		},
	}

	for _, tx := range samples {
		if err := AddTransaction(tx); err != nil {
			fmt.Printf("отказ: транзакция не добавлена (категория %s, сумма %.2f): %v\n",
				tx.Category, tx.Amount, err)
			continue
		}
		fmt.Printf("добавлено: категория %s, сумма %.2f, %s\n",
			tx.Category, tx.Amount, tx.Description)
	}

	fmt.Println("Transactions:")
	for _, tx := range ListTransactions() {
		fmt.Printf("ID=%d Amount=%.2f Category=%s Description=%s Date=%s\n",
			tx.ID, tx.Amount, tx.Category, tx.Description, tx.Date)
	}
}

func printBudgets() {
	categories := make([]string, 0, len(budgets))
	for category := range budgets {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	for _, category := range categories {
		b := budgets[category]
		fmt.Printf("Category=%s Limit=%.2f Period=%s\n", b.Category, b.Limit, b.Period)
	}
}
