package ledger

import (
	"errors"
	"os"
	"testing"
)

func TestTransactionValidate(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{
			name: "valid",
			tx: Transaction{
				Amount:      1500.50,
				Category:    "еда",
				Description: "продукты",
				Date:        "2025-09-10",
			},
		},
		{
			name:    "zero amount",
			tx:      Transaction{Amount: 0, Category: "еда", Date: "2025-09-10"},
			wantErr: true,
		},
		{
			name:    "negative amount",
			tx:      Transaction{Amount: -20, Category: "еда", Date: "2025-09-10"},
			wantErr: true,
		},
		{
			name:    "empty category",
			tx:      Transaction{Amount: 100, Category: "", Date: "2025-09-10"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.tx.Validate()
			if tt.wantErr {
				assertError(t, err)
				return
			}
			assertNoError(t, err)
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
	}{
		{
			name:   "positive limit",
			budget: Budget{Category: "еда", Limit: 5000, Period: "month"},
		},
		{
			name:    "zero limit",
			budget:  Budget{Category: "еда", Limit: 0},
			wantErr: true,
		},
		{
			name:    "negative limit",
			budget:  Budget{Category: "еда", Limit: -1},
			wantErr: true,
		},
		{
			name:    "empty category",
			budget:  Budget{Category: "", Limit: 5000},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.budget.Validate()
			if tt.wantErr {
				assertError(t, err)
				return
			}
			assertNoError(t, err)
		})
	}
}

func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") != "" {
		if err := Connect(); err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

func TestAddTransactionBudgetLimit(t *testing.T) {
	if !Connected() {
		t.Skip("DATABASE_URL is not set")
	}
	Reset()
	t.Cleanup(Reset)

	assertNoError(t, SetBudget(Budget{Category: "еда", Limit: 5000}))

	_, err := AddTransaction(Transaction{
		Amount:      1500.50,
		Category:    "еда",
		Description: "продукты",
		Date:        "2025-09-10",
	})
	assertNoError(t, err)

	before, err := ListTransactions()
	assertNoError(t, err)
	_, err = AddTransaction(Transaction{
		Amount:      4000,
		Category:    "еда",
		Description: "сверх лимита",
		Date:        "2025-09-11",
	})
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("got %v, want budget exceeded", err)
	}
	after, err := ListTransactions()
	assertNoError(t, err)
	if len(after) != len(before) {
		t.Fatalf("transactions after reject: %d, want %d", len(after), len(before))
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("got nil, want error")
	}
}
