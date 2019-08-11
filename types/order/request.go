package order

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/moneytrader/pkg"
)

// Request represents an order to be placed by the provider
type Request struct {
	Side     types.Side
	Quantity decimal.Decimal
	Price    decimal.Decimal
}
