package decimal

import (
	"github.com/shopspring/decimal"
)

func Add(a, b string) string {
	return decimal.RequireFromString(a).Add(decimal.RequireFromString(b)).String()
}

func Sub(a, b string) string {
	return decimal.RequireFromString(a).Sub(decimal.RequireFromString(b)).String()
}

func LessThan(a, b string) bool {
	return decimal.RequireFromString(a).LessThan(decimal.RequireFromString(b))
}

func GreaterThan(a, b string) bool {
	return decimal.RequireFromString(a).GreaterThan(decimal.RequireFromString(b))
}

func Equal(a, b string) bool {
	return decimal.RequireFromString(a).Equal(decimal.RequireFromString(b))
}
