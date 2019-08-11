package trader

import (
	"github.com/sinisterminister/moneytrader/pkg/market"
	"github.com/sinisterminister/moneytrader/pkg/provider"
	"github.com/sinisterminister/moneytrader/pkg/wallet"
)

type Trader struct {
	provider provider.Provider
	markets  []market.Market
	wallets  []*wallet.Wallet
}

func New(provider provider.Provider) (t *Trader) {
	t = &Trader{
		provider: provider,
	}
	return t
}
