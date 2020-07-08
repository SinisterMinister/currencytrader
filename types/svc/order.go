package svc

import (
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/market"

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

func (svc *order) AttemptOrder(m types.Market, t types.OrderType, s types.OrderSide, price decimal.Decimal, quantity decimal.Decimal) (order types.Order, err error) {
	dto, err := svc.trader.Provider().AttemptOrder(types.OrderRequestDTO{
		Market:   m.ToDTO(),
		Type:     t,
		Side:     s,
		Price:    price,
		Quantity: quantity,
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
		log.Info("bailing on order stream")
		return
	}

	log.Debugf("starting the order stream for order %s", o.ID())
	stream, err := svc.trader.Provider().OrderStream(svc.stop, o.ToDTO())
	if err != nil {
		log.WithError(err).Errorf("could not get order stream for order %s", o.ID())
	}

	// Watch for updates
	timer := time.NewTimer(1 * time.Second)
	for {
		select {
		case <-timer.C:
			log.Debugf("fetching latest order status for order %s to preload stream", o.ID())
			dto, err := svc.trader.Provider().Order(o.Market().ToDTO(), o.ID())
			if err != nil {
				log.WithError(err).Errorf("could not fetch order status for order %s", o.ID())
			}

			// No need to watch if it's already done
			if dto.Status == ord.Filled || dto.Status == ord.Canceled {
				go o.Update(dto)
				return
			}
		case <-svc.stop:
			return
		case data := <-stream:
			select {
			case <-svc.stop:
				return
			default:
				go o.Update(data)
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

// Request represents an order to be placed by the provider
type request struct {
	trader internal.Trader
	dto    types.OrderRequestDTO
}

func NewRequestFromDTO(trader internal.Trader, dto types.OrderRequestDTO) types.OrderRequest {
	return &request{trader, dto}
}

func NewRequest(trader internal.Trader, oType types.OrderType, side types.OrderSide, quantity decimal.Decimal, price decimal.Decimal) types.OrderRequest {
	return &request{trader, types.OrderRequestDTO{
		Type:     oType,
		Side:     side,
		Price:    price,
		Quantity: quantity,
	}}
}

func (r *request) ToDTO() types.OrderRequestDTO {
	return r.dto
}

func (r *request) Side() types.OrderSide { return r.dto.Side }

func (r *request) Quantity() decimal.Decimal { return r.dto.Quantity }

func (r *request) Price() decimal.Decimal { return r.dto.Price }

func (r *request) Type() types.OrderType { return r.dto.Type }

func (r *request) Market() types.Market { return market.New(r.trader, r.dto.Market) }
