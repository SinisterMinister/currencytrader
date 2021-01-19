package order

import (
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/market"
	"github.com/spf13/viper"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

type order struct {
	log     log.Entry
	trader  internal.Trader
	mutex   sync.RWMutex
	dto     types.OrderDTO
	streams []chan types.OrderStatus
	done    chan bool
}

func NewOrder(trader internal.Trader, dto types.OrderDTO) internal.Order {
	ord := &order{
		trader: trader,
		dto:    dto,
		log:    log.WithField("source", "currencytrader.order"),
		done:   make(chan bool),
	}

	// Close the done channel if the order is already closed
	switch dto.Status {
	case Filled:
		fallthrough
	case Canceled:
		fallthrough
	case Expired:
		fallthrough
	case Rejected:
		close(ord.done)
	}

	return ord
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

func (o *order) Refresh() (err error) {
	dto, err := o.trader.Provider().RefreshOrder(o.ToDTO())
	if err != nil {
		return
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.dto = dto
	return
}

func (o *order) Request() types.OrderRequest {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return NewRequestFromDTO(market.New(o.trader, o.dto.Market), o.dto.Request)
}

func (o *order) Fees() (types.OrderSide, decimal.Decimal) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.FeesSide, o.dto.Fees
}

func (o *order) Filled() decimal.Decimal {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.Filled
}

func (o *order) Paid() decimal.Decimal {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.Paid
}

func (o *order) Status() types.OrderStatus {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto.Status
}

func (o *order) Market() types.Market {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return market.New(o.trader, o.dto.Market)
}

func (o *order) StatusStream(stop <-chan bool) <-chan types.OrderStatus {
	stream := make(chan types.OrderStatus, viper.GetInt("currencytrader.order.streamBufferSize"))
	o.registerStream(stop, stream)
	go o.broadcastToStreams(o.Status())
	return stream
}

func (o *order) IsDone() bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Refresh done first
	o.refreshDone()

	select {
	case <-o.done:
		return true
	default:
		return false
	}
}

func (o *order) Done() <-chan bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	// Refresh done first
	o.refreshDone()
	return o.done
}

func (o *order) refreshDone() {
	switch o.dto.Status {
	case Canceled:
		fallthrough
	case Expired:
		fallthrough
	case Rejected:
		fallthrough
	case Filled:
		// close the done channel if needed
		select {
		case <-o.done:
		default:
			close(o.done)
		}
	default:
	}
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
	switch dto.Status {
	case Filled:
		fallthrough
	case Canceled:
		fallthrough
	case Expired:
		fallthrough
	case Rejected:
		// Close the done channel as the
		select {
		case <-o.done:
		default:
			close(o.done)
		}
		fallthrough
	case Partial:
		o.dto.Filled = o.dto.Filled.Add(dto.Filled)
		fallthrough
	default:
		o.dto = dto
	}
	go o.broadcastToStreams(dto.Status)

	// TODO: close streams
}

func (o *order) broadcastToStreams(status types.OrderStatus) {
	o.mutex.RLock()
	o.log.Debugf("broadcasting status %s to streams for order %s", status, o.dto.ID)
	streams := o.streams[:0]
	for _, stream := range o.streams {
		select {
		case stream <- status:
			if status == Filled || status == Canceled {
				o.log.Debugf("closing status streams for order %s", o.dto.ID)
				close(stream)
				continue
			}
		default:
			// skip blocked channels
			log.Warnf("skipping blocked order status channel for order %s", o.dto.ID)
		}
		streams = append(streams, stream)
	}
	o.mutex.RUnlock()
	o.mutex.Lock()
	o.streams = streams
	o.mutex.Unlock()
}

func (o *order) ToDTO() types.OrderDTO {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.dto
}
