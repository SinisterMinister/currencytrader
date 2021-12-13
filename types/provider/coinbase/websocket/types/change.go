package types

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type Change struct {
	Message
	ProductID string          `json:"product_id"`
	Time      time.Time       `json:"time"`
	Sequence  int             `json:"sequence"`
	OrderID   string          `json:"order_id"`
	Price     decimal.Decimal `json:"price"`
	Side      string          `json:"side"`
	OldFunds  decimal.Decimal `json:"old_funds"`
	NewFunds  decimal.Decimal `json:"new_funds"`
	OldSize   decimal.Decimal `json:"old_size"`
	NewSize   decimal.Decimal `json:"new_size"`
}

func (c *Change) ToDTO(order types.OrderDTO) types.OrderDTO {
	request := order.Request
	request.Price = c.Price
	request.Quantity = c.NewSize
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       order.Filled,
		ID:           order.ID,
		Request:      request,
		Status:       ord.Updated,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}
