package types

import "time"

type Heartbeat struct {
	Message
	Sequence    int       `json:"sequence"`
	LastTradeID int       `json:"last_trade_id"`
	ProductID   string    `json:"product_id"`
	Time        time.Time `json:"time"`
}
