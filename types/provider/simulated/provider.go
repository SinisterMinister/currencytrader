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

func (p *provider) GetMarkets() (markets []types.MarketDTO, err error) {
	markets = getMarkets()
	return
}

func (p *provider) GetCurrencies() (currencies []types.CurrencyDTO, err error) {
	currencies = getCurrencies()
	return
}

func (p *provider) GetTicker(market types.MarketDTO) (ticker types.TickerDTO, err error) {
	ticker = getTicker(market)
	return
}

func (p *provider) GetTickerStream(stop <-chan bool, market types.MarketDTO) (dataChan <-chan types.TickerDTO, err error) {
	dataChan = getTickerStream(stop, market)
	return
}

func (p *provider) GetWallets() (wallets []types.WalletDTO, err error) {
	return
}
