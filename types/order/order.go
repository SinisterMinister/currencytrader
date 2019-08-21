package order

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

type order struct {
	trader  internal.Trader
	mutex   sync.RWMutex
	dto     types.OrderDTO
	streams []chan types.OrderStatus
}

func NewOrder(trader internal.Trader, dto types.OrderDTO) internal.Order {
	return &order{
		trader: trader,
		dto:    dto,
	}
}

func (o *order) CreationTime() time.Time {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.CreationTime
}

func (o *order) ID() string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.ID
}

func (o *order) Request() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return NewRequestFromDTO(o.dto.Request)
}

func (o *order) Filled() decimal.Decimal {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.Filled
}

func (o *order) Status() types.OrderStatus {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.Status
}

func (o *order) GetStatusStream(stop <-chan bool) <-chan types.OrderStatus {
	stream := make(chan types.OrderStatus)
	o.registerStream(stop, stream)
	return stream
}

func (o *order) registerStream(stop <-chan bool, stream chan types.OrderStatus) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.streams = append(o.streams, stream)

	go func() {
		<-stop
		o.deregisterStream(stream)
	}()
}

func (o *order) deregisterStream(stream chan types.OrderStatus) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	filtered := o.streams[:0]
	for _, c := range o.streams {
		if c != stream {
			filtered = append(filtered, c)
		} else {
			// Close the channel gracefully
			select {
			case <-stream:
			default:
				close(stream)
			}
		}
	}

	// Clean up references
	for i := len(filtered); i < len(o.streams); i++ {
		o.streams[i] = nil
	}

	o.streams = filtered
}

func (o *order) Update(dto types.OrderDTO) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.dto.Status = dto.Status
	o.dto.Filled = dto.Filled
	for _, stream := range o.streams {
		select {
		case stream <- dto.Status:
		default:
			// skip blocked channels
			logrus.Warnf("skipping blocked order status channel for order %s", o.ID())
		}
	}
}

func (o *order) ToDTO() types.OrderDTO {
	return o.dto
}
