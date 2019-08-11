package simulated

import "github.com/sinisterminister/currencytrader/types"

type provider struct {
}

type ProviderConfig struct {
}

func New(config ProviderConfig) types.Provider {
	p := &provider{}
	return p
}

func (p *provider) GetMarkets() (markets []types.Market, err error) {
	return
}

func (p *provider) GetCurrencies() (currencies []types.Currency, err error) {
	return
}

func (p *provider) GetTicker(market types.Market) (ticker types.Ticker, err error) {
	return
}

func (p *provider) GetTickerStream(stop <-chan bool, market types.Market) (dataChan <-chan types.Ticker, err error) {
	return
}

func (p *provider) GetWallets() (wallets []types.Wallet, err error) {
	return
}
