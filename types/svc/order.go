package svc

import (
	"sync"
	"time"

	"github.com/go-playground/log/v7"

	"github.com/shopspring/decimal"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
	ord "github.com/sinisterminister/currencytrader/types/order"
)

type order struct {
	trader internal.Trader

	mutex   sync.RWMutex
	running sync.Once
	stop    chan bool
	working []*types.Order
}

func NewOrder(trader internal.Trader) types.OrderSvc {
	svc := &order{
		trader: trader,
	}
	return svc
}

func (svc *order) Order(mkt types.Market, id string) (order types.Order, err error) {
	dto, err := svc.trader.Provider().Order(mkt.ToDTO(), id)
	if err != nil {
		return
	}
	order = svc.buildOrder(dto)
	return
}

func (svc *order) AttemptOrder(m types.Market, t types.OrderType, s types.OrderSide, price decimal.Decimal, quantity decimal.Decimal, forceMaker bool) (order types.Order, err error) {
	dto, err := svc.trader.Provider().AttemptOrder(types.OrderRequestDTO{
		Market:     m.ToDTO(),
		Type:       t,
		Side:       s,
		Price:      price,
		Quantity:   quantity,
		ForceMaker: forceMaker,
	})
	if err != nil {
		return
	}
	order = svc.buildOrder(dto)
	return
}

func (svc *order) CancelOrder(order types.Order) error {
	return svc.trader.Provider().CancelOrder(order.ToDTO())
}

func (svc *order) OrderFromDTO(dto types.OrderDTO) types.Order {
	return svc.buildOrder(dto)
}

func (svc *order) buildOrder(dto types.OrderDTO) types.Order {
	ord := ord.NewOrder(svc.trader, dto)
	go svc.handleOrderStream(ord)
	return ord
}

func (svc *order) handleOrderStream(o internal.Order) {
	// Bail if the order is already closed
	switch o.Status() {
	case ord.Filled:
		fallthrough
	case ord.Canceled:
		fallthrough
	case ord.Expired:
		fallthrough
	case ord.Rejected:
		log.Debugf("status is %s: bailing on order stream", o.Status())
		return
	}

	log.Debugf("starting the order stream for order %s", o.ID())
	stop := make(chan bool)
	stream, err := svc.trader.Provider().OrderStream(stop, o.ToDTO())
	if err != nil {
		log.WithError(err).Errorf("could not get order stream for order %s", o.ID())
		return
	}

	// Watch for updates
	timer := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-o.Done():
			close(stop)
			return

		case <-timer.C:
			// Refresh the order
			o.Refresh()

			// No need to watch if it's already done
			switch o.Status() {
			case ord.Unknown:
				timer.Reset(1 * time.Second)
			case ord.Filled:
				fallthrough
			case ord.Canceled:
				fallthrough
			case ord.Rejected:
				close(stop)
				return
			}

			// Reset the timer as a backup to the streams
			timer.Reset(5 * time.Second)
		case <-svc.stop:
			close(stop)
			return
		case data := <-stream:
			select {
			case <-svc.stop:
				close(stop)
				return
			default:
				go o.Update(data)
				if data.Status == ord.Filled || data.Status == ord.Canceled {
					close(stop)
					return
				}
				// Reset the timer as a backup to the streams
				timer.Reset(5 * time.Second)
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
