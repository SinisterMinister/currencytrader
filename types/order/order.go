package order

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

type order struct {
	creationTime time.Time
	id           string
	request      types.OrderRequest

	mutex  sync.RWMutex
	filled decimal.Decimal
	status types.OrderStatus
}

func NewOrder(trader internal.Trader, dto types.OrderDTO) internal.Order {
	return &order{
		creationTime: dto.CreationTime,
		id:           dto.ID,
		request:      NewRequestFromDTO(dto.Request),
		filled:       dto.Filled,
		status:       dto.Status,
	}
}

func (o *order) CreationTime() time.Time {
	return o.creationTime
}

func (o *order) ID() string {
	return o.id
}

func (o *order) Request() types.OrderRequest {
	return o.request
}

func (o *order) Filled() decimal.Decimal {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.filled
}

func (o *order) Status() types.OrderStatus {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.status
}

func (o *order) Update(dto types.OrderDTO) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.status = dto.Status
	o.filled = dto.Filled
}

func (o *order) ToDTO() types.OrderDTO {
	return types.OrderDTO{
		CreationTime: o.creationTime,
		Filled:       o.filled,
		ID:           o.id,
		Request:      o.request.ToDTO(),
		Status:       o.status,
	}
}
