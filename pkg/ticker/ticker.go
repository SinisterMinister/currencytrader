package ticker

import (
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/moneytrader/pkg/market"
)

// Ticker TODO
type Ticker struct {
	Market market.Market
	Price  decimal.Decimal
}
