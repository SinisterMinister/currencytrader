package ticker

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

// Ticker TODO
type Ticker struct {
	ask       decimal.Decimal
	bid       decimal.Decimal
	price     decimal.Decimal
	quantity  decimal.Decimal
	timestamp time.Time
	volume    decimal.Decimal
}

func NewTicker(ask decimal.Decimal, bid decimal.Decimal, price decimal.Decimal,
	quantity decimal.Decimal, timestamp time.Time, volume decimal.Decimal) types.Ticker {
	return &Ticker{
		ask:       ask,
		bid:       bid,
		price:     price,
		quantity:  quantity,
		timestamp: timestamp,
		volume:    volume,
	}
}

func (t *Ticker) Ask() decimal.Decimal {
	return t.ask
}

func (t *Ticker) Bid() decimal.Decimal {
	return t.bid
}

func (t *Ticker) Price() decimal.Decimal {
	return t.price
}

func (t *Ticker) Quantity() decimal.Decimal {
	return t.quantity
}

func (t *Ticker) Timestamp() time.Time {
	return t.timestamp
}

func (t *Ticker) Volume() decimal.Decimal {
	return t.volume
}
