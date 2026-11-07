package ledger

import "sync"

var (
	mu           sync.Mutex
	transactions []Transaction
	budgets      map[string]Budget
)

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	transactions = nil
	budgets = nil
}
