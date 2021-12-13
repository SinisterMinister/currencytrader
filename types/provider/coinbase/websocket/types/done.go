package types

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type Done struct {
	Message
	ProductID     string          `json:"product_id"`
	Time          time.Time       `json:"time"`
	Sequence      int             `json:"sequence"`
	OrderID       string          `json:"order_id"`
	RemainingSize decimal.Decimal `json:"remaining_size"`
	Price         decimal.Decimal `json:"price"`
	Side          string          `json:"side"`
	Reason        string          `json:"reason"`
}

func (d *Done) ToDTO(order types.OrderDTO) types.OrderDTO {
	var status types.OrderStatus
	switch d.Reason {
	case "filled":
		status = ord.Filled
	case "canceled":
		status = ord.Canceled
	}
	return types.OrderDTO{
		Market:       order.Market,
		CreationTime: order.CreationTime,
		Filled:       order.Request.Quantity.Sub(d.RemainingSize),
		ID:           order.ID,
		Request:      order.Request,
		Status:       status,
		Fees:         order.Fees,
		FeesSide:     order.FeesSide,
		Paid:         order.Paid,
	}
}
