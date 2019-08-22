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

func (p *provider) Markets() (markets []types.MarketDTO, err error) {
	markets = getMarkets()
	return
}

func (p *provider) Currencies() (currencies []types.CurrencyDTO, err error) {
	currencies = getCurrencies()
	return
}

func (p *provider) Ticker(market types.MarketDTO) (ticker types.TickerDTO, err error) {
	ticker = getTicker(market)
	return
}

func (p *provider) TickerStream(stop <-chan bool, market types.MarketDTO) (dataChan <-chan types.TickerDTO, err error) {
	dataChan = getTickerStream(stop, market)
	return
}

func (p *provider) Wallets() (wallets []types.WalletDTO, err error) {
	wallets = getWallets()
	return
}

func (p *provider) Wallet(currency types.CurrencyDTO) (wallet types.WalletDTO, err error) {
	wallet = getWallet(currency)
	return
}

func (p *provider) WalletStream(stop <-chan bool, currency types.CurrencyDTO) (stream <-chan types.WalletDTO, err error) {
	stream = getWalletStream(stop, currency)
	return
}

func (p *provider) AttemptOrder(mkt types.MarketDTO, ord types.OrderRequestDTO) (types.OrderDTO, error) {
	return attemptOrder(mkt, ord)
}

func (p *provider) CancelOrder(order types.OrderDTO) error {
	return cancelOrder(order)
}

func (p *provider) Order(id string) (types.OrderDTO, error) {
	return getOrder(id)
}

func (p *provider) OrderStream(stop <-chan bool, order types.OrderDTO) (ch <-chan types.OrderDTO, err error) {
	return getOrderStream(stop, order)
}
