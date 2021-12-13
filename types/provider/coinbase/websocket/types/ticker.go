package types

import (
	"time"

	"github.com/shopspring/decimal"
)

type Ticker struct {
	Message
	TradeID   int             `json:"trade_id"`
	Sequence  int             `json:"sequence"`
	Time      time.Time       `json:"time"`
	ProductID string          `json:"product_id"`
	Price     decimal.Decimal `json:"price"`
	Side      string          `json:"side"`
	LastSize  decimal.Decimal `json:"last_size"`
	BestBid   decimal.Decimal `json:"best_bid"`
	BestAsk   decimal.Decimal `json:"best_ask"`
}
