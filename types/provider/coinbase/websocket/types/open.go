package types

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type Open struct {
	Message
	ProductID     string          `json:"product_id"`
	Time          time.Time       `json:"time"`
	Sequence      int             `json:"sequence"`
	OrderID       string          `json:"order_id"`
	RemainingSize decimal.Decimal `json:"remaining_size"`
	Price         decimal.Decimal `json:"price"`
	Side          string          `json:"side"`
}

func (o *Open) ToDTO(order types.OrderDTO) types.OrderDTO {
	var status types.OrderStatus
	if order.Request.Quantity.Equal(o.RemainingSize) {
		status = ord.Pending
	} else {
		status = ord.Partial
	}
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       decimal.Zero, // We fill zero here since a match event will cover the actual amount(s)
		ID:           order.ID,
		Request:      order.Request,
		Status:       status,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}
