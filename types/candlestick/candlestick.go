package candlestick

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

type candlestick struct {
	dto types.CandlestickDTO
}

func New(dto types.CandlestickDTO) types.Candlestick {
	return &candlestick{dto}
}

func (c *candlestick) Close() decimal.Decimal { return c.dto.Close }

func (c *candlestick) High() decimal.Decimal { return c.dto.High }

func (c *candlestick) Low() decimal.Decimal { return c.dto.Low }

func (c *candlestick) Open() decimal.Decimal { return c.dto.Open }

func (c *candlestick) Timestamp() time.Time { return c.dto.Timestamp }

func (c *candlestick) Volume() decimal.Decimal { return c.dto.Volume }

func (c *candlestick) ToDTO() types.CandlestickDTO { return c.dto }
