package types

import "github.com/shopspring/decimal"

type Activate struct {
	Message
	ProductID    string          `json:"product_id"`
	Timestamp    string          `json:"timestamp"`
	UserID       string          `json:"user_id"`
	ProfileID    string          `json:"profile_id"`
	OrderID      string          `json:"order_id"`
	StopType     string          `json:"stop_type"`
	StopPrice    decimal.Decimal `json:"stop_price"`
	Side         string          `json:"side"`
	Size         decimal.Decimal `json:"size"`
	Funds        decimal.Decimal `json:"funds"`
	TakerFeeRate decimal.Decimal `json:"taker_fee_rate"`
	Private      bool            `json:"private"`
}
