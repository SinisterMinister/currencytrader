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

type TickerConfig struct {
	types.TickerDTO
}

func New(config TickerConfig) types.Ticker {
	return &Ticker{
		ask:       config.Ask,
		bid:       config.Bid,
		price:     config.Price,
		quantity:  config.Quantity,
		timestamp: config.Timestamp,
		volume:    config.Volume,
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
