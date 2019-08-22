package svc

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type order struct {
	trader internal.Trader

	mutex   sync.RWMutex
	running sync.Once
	stop    chan bool
}

func NewOrder(trader internal.Trader) types.OrderSvc {
	svc := &order{
		trader: trader,
	}
	return svc
}

func (svc *order) Order(id string) (order types.Order, err error) {
	dto, err := svc.trader.Provider().Order(id)
	if err != nil {
		return
	}
	order = svc.buildOrder(dto)
	return
}

func (svc *order) AttemptOrder(mkt types.Market, req types.OrderRequest) (order types.Order, err error) {
	dto, err := svc.trader.Provider().AttemptOrder(mkt.ToDTO(), req.ToDTO())
	if err != nil {
		return
	}
	order = svc.buildOrder(dto)
	return
}

func (svc *order) CancelOrder(order types.Order) error {
	return svc.trader.Provider().CancelOrder(order.ToDTO())
}

func (svc *order) buildOrder(dto types.OrderDTO) types.Order {
	ord := ord.NewOrder(svc.trader, dto)
	go svc.handleOrderStream(ord)
	return ord
}

func (svc *order) handleOrderStream(o internal.Order) {
	// Bail if the order is already closed
	if o.Status() == ord.Filled || o.Status() == ord.Canceled {
		return
	}

	stream, err := svc.trader.Provider().OrderStream(svc.stop, o.ToDTO())
	if err != nil {
		logrus.WithError(err).Errorf("could not get order stream for order %s", o.ID())
	}

	for {
		select {
		case <-svc.stop:
			return
		case data := <-stream:
			select {
			case <-svc.stop:
				return
			default:
				o.Update(data)
				if data.Status == ord.Filled || data.Status == ord.Canceled {
					return
				}
			}
		}
	}
}

func (svc *order) Start() {
	svc.running.Do(func() {
		svc.mutex.Lock()
		defer svc.mutex.Unlock()
		svc.stop = make(chan bool)
	})
}

func (svc *order) Stop() {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()
	close(svc.stop)
	svc.running = sync.Once{}
}
