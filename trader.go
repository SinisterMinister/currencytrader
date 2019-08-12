package currencytrader

import (
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/trader"
)

func New(provider types.Provider) types.Trader {
	return trader.New(provider)
}
