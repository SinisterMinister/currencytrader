package simulated

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/sinisterminister/currencytrader/types/order"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types"
)

var (
	mutex       sync.Mutex
	orders      map[string]types.OrderDTO
	streams     map[string][]chan types.OrderDTO
	cancelChans map[string]chan bool
	wallets     map[string]types.WalletDTO
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
			if !contains(markets, base.Symbol+quote.Symbol) && !contains(markets, quote.Symbol+base.Symbol) && base.Symbol != quote.Symbol {
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

	go func(ch chan types.TickerDTO) {
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

func getWallets() []types.WalletDTO {
	if wallets == nil {
		wallets = make(map[string]types.WalletDTO)
		currencies := getCurrencies()
		for _, cur := range currencies {
			id, _ := uuid.NewUUID()
			wallets[id.String()] = types.WalletDTO{
				ID:       id.String(),
				Currency: cur,
				Free:     decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(cur.Precision)),
				Locked:   decimal.NewFromFloat((rand.Float64() / 2) * float64(rand.Intn(100))).Round(int32(cur.Precision)),
			}
		}
	}

	wals := []types.WalletDTO{}
	for _, wal := range wallets {
		wals = append(wals, wal)
	}
	return wals
}

func getWallet(id string) types.WalletDTO {
	return wallets[id]
}

func getWalletStream(stop <-chan bool, wal types.WalletDTO) <-chan types.WalletDTO {
	ch := make(chan types.WalletDTO)
	go func(stop <-chan bool, wal types.WalletDTO, ch chan types.WalletDTO) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-stop:
				return
			default:
			}

			select {
			case <-stop:
				return
			case <-ticker.C:
				ch <- getWallet(wal.ID)
			}
		}
	}(stop, wal, ch)
	return ch
}

func attemptOrder(mk types.MarketDTO, req types.OrderRequestDTO) (types.OrderDTO, error) {
	order := types.OrderDTO{
		Market:       mk,
		CreationTime: time.Now(),
		Filled:       decimal.Zero,
		ID:           uuid.New().String(),
		Request:      req,
		Status:       order.Pending,
	}

	registerOrder(order)

	return order, nil
}

func registerOrder(o types.OrderDTO) {
	mutex.Lock()
	if orders == nil {
		orders = make(map[string]types.OrderDTO)
	}
	if streams == nil {
		streams = make(map[string][]chan types.OrderDTO)
	}
	if cancelChans == nil {
		cancelChans = make(map[string]chan bool)
	}

	orders[o.ID] = o
	streams[o.ID] = []chan types.OrderDTO{}
	cancelChans[o.ID] = processOrder(o)
	mutex.Unlock()
}

func updateOrder(o types.OrderDTO) {
	mutex.Lock()
	orders[o.ID] = o

	chs, ok := streams[o.ID]
	mutex.Unlock()
	if !ok {
		return
	}

	if len(chs) > 0 {
		for _, ch := range chs {
			select {
			case ch <- o:
			default:
				log.Warn("skipping blocked order update channel")
			}
		}
	}
}

func cleanupOrder(o types.OrderDTO) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(cancelChans, o.ID)
	delete(streams, o.ID)
}

func processOrder(o types.OrderDTO) chan bool {
	stop := make(chan bool)
	go func(stop chan bool) {
		ticker := time.NewTicker(time.Duration(rand.Intn(5000)) * time.Millisecond)
		defer cleanupOrder(o)
		defer ticker.Stop()

		select {
		case <-stop:
			o.Status = order.Canceled
		case <-ticker.C:
			o.Status = order.Partial
		}
		updateOrder(o)

		if o.Status == order.Canceled {
			return
		}

		select {
		case <-stop:
			o.Status = order.Canceled
		case <-ticker.C:
			o.Status = order.Filled
		}
		updateOrder(o)
	}(stop)

	return stop
}

func getOrder(mkt types.MarketDTO, id string) (types.OrderDTO, error) {
	order, ok := orders[id]

	if !ok {
		return order, fmt.Errorf("could not find order for ID %s", id)
	}

	return order, nil
}

func getOrderStream(stop <-chan bool, o types.OrderDTO) (<-chan types.OrderDTO, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := streams[o.ID]; !ok {
		return nil, fmt.Errorf("cannot get update stream for order %s", o.ID)
	}
	ch := make(chan types.OrderDTO)
	streams[o.ID] = append(streams[o.ID], ch)
	return ch, nil
}

func cancelOrder(o types.OrderDTO) error {
	stop, ok := cancelChans[o.ID]
	if !ok {
		return fmt.Errorf("could not cancel order %s", o.ID)
	}
	close(stop)
	return nil
}

func getCandles(mkt types.MarketDTO, interval types.CandleInterval, start time.Time, end time.Time) []types.CandleDTO {
	candles := []types.CandleDTO{}

	for index := 0; index < int(end.Sub(start).Minutes()); index++ {
		candles = append(candles, types.CandleDTO{})
	}

	return candles
}

func randDecimal(min float64, max float64) decimal.Decimal {
	space := max - min
	init := rand.Float64() * space
	return decimal.NewFromFloat(init + min)
}
