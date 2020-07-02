package currency

import (
	"github.com/go-playground/log/v7"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

// Currency TODO
type currency struct {
	trader internal.Trader
	dto    types.CurrencyDTO
}

func New(trader internal.Trader, cur types.CurrencyDTO) types.Currency {
	return &currency{trader, cur}
}

func (c *currency) Increment() decimal.Decimal { return c.dto.Increment }

func (c *currency) Name() string { return c.dto.Name }

func (c *currency) Symbol() string { return c.dto.Symbol }

func (c *currency) Precision() int { return c.dto.Precision }

func (c *currency) ToDTO() types.CurrencyDTO { return c.dto }

func (c *currency) Wallet() types.Wallet {
	wallet, err := c.trader.AccountSvc().Wallet(c)
	if err != nil {
		log.WithField("currency", c.dto).WithError(err).Fatal("was unable to retrieve wallet from service")
	}
	return wallet
}
