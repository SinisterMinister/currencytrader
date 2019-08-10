package binance

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sinisterminister/moneytrader/pkg/wallet"

	"github.com/google/go-cmp/cmp"

	"github.com/sinisterminister/moneytrader/pkg/currency"
	"github.com/sinisterminister/moneytrader/pkg/market"
	"github.com/sinisterminister/moneytrader/pkg/provider/binance/api"
)

type Provider struct {
	currencies        map[string]currency.Currency
	currenciesRefresh *time.Timer
	markets           []market.Market
	marketsRefresh    *time.Timer

	stopChan <-chan bool
}

func newProvider(stopChan <-chan bool) *Provider {
	p := &Provider{
		stopChan: stopChan,
	}
	p.updateCurrencies()
	p.updateMarkets()

	return p
}

func (p *Provider) GetMarkets() (markets []market.Market, err error) {
	// Update markets if necessary
	p.updateMarkets()
	markets = p.markets

	return markets, err
}

func (p *Provider) GetMarket(currency0 currency.Currency, currency1 currency.Currency) (mkt market.Market, err error) {
	markets, _ := p.GetMarkets()

	for _, m := range markets {
		if cmp.Equal(mkt.BaseCurrency, currency0) && cmp.Equal(mkt.QuoteCurrency, currency1) {
			mkt = m
			return
		}
		if cmp.Equal(mkt.BaseCurrency, currency1) && cmp.Equal(mkt.QuoteCurrency, currency0) {
			mkt = m
			return
		}
	}

	return mkt, fmt.Errorf("no market for currencies %s, %s", currency0.Name, currency1.Name)
}

func (p *Provider) GetCurrency(symbol string) (cur currency.Currency, err error) {
	// First, update currencies if necessary
	p.updateCurrencies()

	// Get the currency from the cache
	cur, ok := p.currencies[symbol]
	if !ok {
		return cur, fmt.Errorf("currency %s not found", symbol)
	}
	return cur, err
}

func (p *Provider) GetCurrencies() (currencies []currency.Currency, err error) {
	// First, update currencies if necessary
	p.updateCurrencies()
	currencies = make([]currency.Currency, len(p.currencies))

	for _, cur := range p.currencies {
		currencies = append(currencies, cur)
	}

	return
}

func (p *Provider) GetWallets() (wallets []*wallet.Wallet, err error) {
	data, err := api.GetUserData()
	if err != nil {
		return
	}

	wallets = []*wallet.Wallet{}
	for _, bal := range data.Balances {
		cur, _ := p.GetCurrency(bal.Asset)
		w := wallet.NewWallet(cur, bal.Free.Add(bal.Locked), bal.Free, bal.Locked, decimal.Zero)

		wallets = append(wallets, w)
	}

	return wallets, err
}

func (p *Provider) updateMarkets() {
	if p.marketsRefresh != nil {
		select {
		default:
			// Nothing to do
			return
		case <-p.marketsRefresh.C:
			// Time to update, continue below
		}
	}

	markets := []market.Market{}
	symbols := api.GetExchangeInfo().Symbols

	for _, symbol := range symbols {
		baseCur, _ := p.GetCurrency(symbol.BaseAsset)
		quoteCur, _ := p.GetCurrency(symbol.QuoteAsset)

		m := market.Market{
			Name:             symbol.Symbol,
			BaseCurrency:     baseCur,
			QuoteCurrency:    quoteCur,
			MinPrice:         symbol.Filters.Price.MinPrice,
			MaxPrice:         symbol.Filters.Price.MaxPrice,
			PriceIncrement:   symbol.Filters.Price.TickSize,
			MinQuantity:      symbol.Filters.LotSize.MinQuantity,
			MaxQuantity:      symbol.Filters.LotSize.MaxQuantity,
			QuantityStepSize: symbol.Filters.LotSize.StepSize,
		}

		markets = append(markets, m)
	}

	p.markets = markets
	p.marketsRefresh = time.NewTimer(1 * time.Minute)
}

func (p *Provider) updateCurrencies() {
	if p.currenciesRefresh != nil {
		select {
		default:
			// Nothing to do
			return
		case <-p.currenciesRefresh.C:
			// Time to update, continue below
		}
	}

	symbols := api.GetExchangeInfo().Symbols
	currencies := make(map[string]currency.Currency)

	for _, symbol := range symbols {
		if _, ok := currencies[symbol.BaseAsset]; !ok {
			currencies[symbol.BaseAsset] = currency.Currency{
				Symbol:    symbol.BaseAsset,
				Name:      symbol.BaseAsset,
				Precision: symbol.BasePrecision,
			}
		}

		if _, ok := currencies[symbol.QuoteAsset]; !ok {
			currencies[symbol.QuoteAsset] = currency.Currency{
				Symbol:    symbol.QuoteAsset,
				Name:      symbol.QuoteAsset,
				Precision: symbol.QuotePrecision,
			}
		}
	}

	p.currencies = currencies
	p.currenciesRefresh = time.NewTimer(1 * time.Minute)
}
