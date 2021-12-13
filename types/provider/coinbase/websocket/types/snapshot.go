package types

import "github.com/shopspring/decimal"

type Snapshot struct {
	Message
	ProductID string              `json:"product_id"`
	Bids      [][]decimal.Decimal `json:"bids"`
	Asks      [][]decimal.Decimal `json:"asks"`
}
