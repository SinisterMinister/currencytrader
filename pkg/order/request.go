package order

import "github.com/shopspring/decimal"

// Side represents which side the order will be placed
type Side int

const (
	// BuySide represents a buy sided order
	BuySide Side = iota

	// SellSide represents a sell sided order
	SellSide
)

// Request represents an order to be placed by the provider
type Request struct {
	Side     Side            `json:"side"`
	Quantity decimal.Decimal `json:"qty"`
	Price    decimal.Decimal `json:"price"`
}
