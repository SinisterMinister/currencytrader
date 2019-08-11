package order

import (
	"time"

	"github.com/sinisterminister/currencytrader/pkg"

	"github.com/shopspring/decimal"
)

// Order represents an order that has been accepted by the provider
type Order struct {
	CreationTime time.Time
	Filled       decimal.Decimal
	ID           string
	Request      types.OrderRequest
	Status       types.Status
}
