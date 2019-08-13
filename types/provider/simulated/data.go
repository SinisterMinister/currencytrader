package simulated

import (
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

func getCurrencies() []types.CurrencyDTO {
	return append([]types.CurrencyDTO{},
		types.CurrencyDTO{
			Name:      "US Dollar",
			Symbol:    "USD",
			Precision: 2,
		},
		types.CurrencyDTO{
			Name:      "Bitcoin",
			Symbol:    "BTC",
			Precision: 8,
		},
		types.CurrencyDTO{
			Name:      "Etherium",
			Symbol:    "ETH",
			Precision: 8,
		},
		types.CurrencyDTO{
			Name:      "Ripple",
			Symbol:    "XRP",
			Precision: 2,
		},
	)
}

func getMarkets() []types.MarketDTO {
	currencies := getCurrencies()
	markets := []types.MarketDTO{}

	contains := func(markets []types.MarketDTO, symbol string) bool {
		for _, m := range markets {
			if m.Name == symbol {
				return true
			}
		}
		return false
	}

	for _, base := range currencies {
		for _, quote := range currencies {
			if !contains(markets, base.Symbol+quote.Symbol) && !contains(markets, quote.Symbol+base.Symbol) {
				markets = append(markets, types.MarketDTO{
					Name:          base.Symbol + quote.Symbol,
					BaseCurrency:  base,
					QuoteCurrency: quote,
				})
			}
		}
	}
	return markets
}

func getTicker(mkt types.MarketDTO) types.TickerDTO {
	return types.TickerDTO{
		Ask:       decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Bid:       decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Price:     decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Quantity:  decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(mkt.QuoteCurrency.Precision)),
		Timestamp: time.Now(),
		Volume:    decimal.NewFromFloat(rand.Float64() * float64(rand.Intn(10000))).Round(int32(mkt.QuoteCurrency.Precision)),
	}
}

func getTickerStream(stop <-chan bool, mkt types.MarketDTO) <-chan types.TickerDTO {
	ch := make(chan types.TickerDTO)

	func(ch chan types.TickerDTO) {
		ticker := time.NewTicker(1 * time.Second)

		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			default:
			}

			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				ch <- getTicker(mkt)
			}
		}

	}(ch)

	return ch
}
