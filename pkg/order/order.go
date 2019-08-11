package order

import (
	"time"

	"github.com/sinisterminister/moneytrader/pkg"

	"github.com/shopspring/decimal"
)

// Order represents an order that has been accepted by the provider
type Order struct {
	CreationTime time.Time
	Filled       decimal.Decimal
	ID           string
	Request      pkg.OrderRequest
	Status       pkg.Status
}
