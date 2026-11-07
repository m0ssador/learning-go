package domain

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

func (t Transaction) Validate() error {
	if t.Amount == 0 {
		return invalid("transaction amount must not be zero")
	}
	if t.Amount < 0 {
		return invalid("transaction amount must not be negative")
	}
	if t.Category == "" {
		return invalid("transaction category must not be empty")
	}
	return nil
}
