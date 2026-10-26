package domain

type CustomerLevel string

const (
	CustomerNew     CustomerLevel = "new"
	CustomerRegular CustomerLevel = "regular"
	CustomerVIP     CustomerLevel = "vip"
)

type Customer struct {
	ID    int64
	Level CustomerLevel
}

func (c Customer) OrderLimit() float64 {
	if c.Level == CustomerVIP {
		return 1_000_000
	}
	return 100_000
}
