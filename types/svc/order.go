package svc

import (
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type order struct {
	trader internal.Trader
}

func NewOrder(trader internal.Trader) types.OrderSvc {
	svc := &order{
		trader: trader,
	}
	return svc
}

func (svc *order) GetOrder(id string) (order types.Order, err error) {
	dto, err := svc.trader.Provider().GetOrder(id)
	order = ord.NewOrder(svc.trader, dto)
	return
}

func (svc *order) AttemptOrder(mkt types.Market, req types.OrderRequest) (order types.Order, err error) {
	dto, err := svc.trader.Provider().AttemptOrder(mkt.ToDTO(), req.ToDTO())
	order = ord.NewOrder(svc.trader, dto)
	return
}

func (svc *order) CancelOrder(order types.Order) error {
	return svc.trader.Provider().CancelOrder(order.ToDTO())
}
