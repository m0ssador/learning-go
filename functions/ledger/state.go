package ledger

import "sync"

var (
	mu           sync.Mutex
	transactions []Transaction
	budgets      map[string]Budget
)
