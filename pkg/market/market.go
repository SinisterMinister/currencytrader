package market

import (
	"github.com/shopspring/decimal"
	cur "github.com/sinisterminister/moneytrader/pkg/currency"
)

// Market is where you can trade one currency for another.
type Market struct {
	Name             string
	BaseCurrency     cur.Currency
	QuoteCurrency    cur.Currency
	MinPrice         decimal.Decimal
	MaxPrice         decimal.Decimal
	PriceIncrement   decimal.Decimal
	MinQuantity      decimal.Decimal
	MaxQuantity      decimal.Decimal
	QuantityStepSize decimal.Decimal
}
