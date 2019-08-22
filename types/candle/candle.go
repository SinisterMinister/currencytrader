package candle

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

type candle struct {
	dto types.CandleDTO
}

func New(dto types.CandleDTO) types.Candle {
	return &candle{dto}
}

func (c *candle) Close() decimal.Decimal { return c.dto.Close }

func (c *candle) High() decimal.Decimal { return c.dto.High }

func (c *candle) Low() decimal.Decimal { return c.dto.Low }

func (c *candle) Open() decimal.Decimal { return c.dto.Open }

func (c *candle) Timestamp() time.Time { return c.dto.Timestamp }

func (c *candle) Volume() decimal.Decimal { return c.dto.Volume }

func (c *candle) ToDTO() types.CandleDTO { return c.dto }
