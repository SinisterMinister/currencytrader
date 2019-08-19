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
	wallets = getWallets()
	return
}

func (p *provider) GetWallet(currency types.CurrencyDTO) (wallet types.WalletDTO, err error) {
	wallet = getWallet(currency)
	return
}

func (p *provider) GetWalletStream(stop <-chan bool, currency types.CurrencyDTO) (stream <-chan types.WalletDTO, err error) {
	stream = getWalletStream(stop, currency)
	return
}

func (p *provider) AttemptOrder(mkt types.MarketDTO, ord types.OrderRequestDTO) (types.OrderDTO, error) {
	return attemptOrder(mkt, ord)
}

func (p *provider) CancelOrder(order types.OrderDTO) error {
	return cancelOrder(order)
}

func (p *provider) GetOrder(id string) (types.OrderDTO, error) {
	return getOrder(id)
}

func (p *provider) GetOrderStream(stop <-chan bool, order types.OrderDTO) (ch <-chan types.OrderDTO, err error) {
	return getOrderStream(stop, order)
}
