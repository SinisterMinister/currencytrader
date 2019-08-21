package ticker

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

// Ticker TODO
type Ticker struct {
	dto types.TickerDTO
}

func New(dto types.TickerDTO) types.Ticker {
	return &Ticker{dto}
}

func (t *Ticker) ToDTO() types.TickerDTO { return t.dto }

func (t *Ticker) Ask() decimal.Decimal { return t.dto.Ask }

func (t *Ticker) Bid() decimal.Decimal { return t.dto.Bid }

func (t *Ticker) Price() decimal.Decimal { return t.dto.Price }

func (t *Ticker) Quantity() decimal.Decimal { return t.dto.Quantity }

func (t *Ticker) Timestamp() time.Time { return t.dto.Timestamp }

func (t *Ticker) Volume() decimal.Decimal { return t.dto.Volume }
