package trader

import (
	"github.com/sinisterminister/moneytrader/types"
	"github.com/sinisterminister/moneytrader/types/svc"
)

type Trader struct {
	provider  types.Provider
	marketSvc types.MarketSvc
	tickerSvc types.TickerSvc
}

func New(provider types.Provider) (t types.Trader) {
	t = &Trader{
		provider:  provider,
		marketSvc: svc.NewMarket(provider),
		tickerSvc: svc.NewTicker(provider),
	}
	return t
}

func (t *Trader) Launch(stop <-chan bool) {

}

func (t *Trader) OrderSvc() types.OrderSvc {
	return nil
}

func (t *Trader) WalletSvc() types.WalletSvc {
	return nil
}

func (t *Trader) MarketSvc() types.MarketSvc {
	return t.marketSvc
}

func (t *Trader) TickerSvc() types.TickerSvc {
	return t.tickerSvc
}
