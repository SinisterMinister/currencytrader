package trader

import (
	"github.com/sinisterminister/moneytrader/pkg"
	"github.com/sinisterminister/moneytrader/pkg/market"
	"github.com/sinisterminister/moneytrader/pkg/wallet"
)

type Trader struct {
	provider pkg.Provider
	markets  []market.Market
	wallets  []*wallet.Wallet
}

func New(provider pkg.Provider) (t *Trader) {
	t = &Trader{
		provider: provider,
	}
	return t
}
