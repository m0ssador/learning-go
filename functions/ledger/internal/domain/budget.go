package domain

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

func (b Budget) Validate() error {
	if b.Limit <= 0 {
		return invalid("budget limit must be positive")
	}
	if b.Category == "" {
		return invalid("budget category must not be empty")
	}
	return nil
}
