package binance

import (
	"strconv"
	"time"

	"github.com/imdario/mergo"

	"github.com/joeshaw/envdecode"

	"github.com/sinisterminister/currencytrader/types/candle"

	"github.com/go-playground/log"
	"github.com/shopspring/decimal"
	"github.com/sinisterminister/coinfactory"
	api "github.com/sinisterminister/coinfactory/pkg/binance"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/order"
	validator "gopkg.in/go-playground/validator.v9"
)

var (
	validate *validator.Validate

	currencies  map[string]types.CurrencyDTO
	updateTimer time.Timer
)

type provider struct {
	stopChan <-chan bool
	config   Config
}

func New(stopChan <-chan bool, config Config) types.Provider {
	return &provider{
		config:   parseConfig(config),
		stopChan: stopChan,
	}

}

func init() {
	currencies = make(map[string]types.CurrencyDTO)
	validate = validator.New()
	coinfactory.Start()
}

func parseConfig(config Config) Config {
	defaults := Config{}
	// Load in env vars
	err := envdecode.Decode(&defaults)
	if err != nil {
		log.WithError(err).Fatal("could not load config from environment")
	}

	// Merge into provided config
	mergo.Merge(&config, defaults)

	// Validate config; exit if fails
	err = validate.Struct(config)
	if err != nil {
		log.WithError(err).Fatal("invalid provider configuration!")
	}

	return config
}

func (p *provider) Wallet(cur types.CurrencyDTO) (types.WalletDTO, error) {
	return types.WalletDTO{
		Currency: cur,
		Free:     coinfactory.GetBalanceManager().GetAvailableBalance(cur.Name),
		Locked:   coinfactory.GetBalanceManager().GetFrozenBalance(cur.Name),
	}, nil
}

func (p *provider) WalletStream(stop <-chan bool, cur types.CurrencyDTO) (<-chan types.WalletDTO, error) {
	stream := make(chan types.WalletDTO)

	go func(stop <-chan bool, stream chan types.WalletDTO) {
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
				dto, _ := p.Wallet(cur)
				stream <- dto
			}
		}
	}(stop, stream)
	return stream, nil
}

func (p *provider) Ticker(market types.MarketDTO) (types.TickerDTO, error) {
	ticker := coinfactory.GetSymbolService().GetSymbol(market.Name).GetTicker()
	return types.TickerDTO{
		Ask:       ticker.AskPrice,
		Bid:       ticker.BidPrice,
		Price:     ticker.CurrentClosePrice,
		Quantity:  ticker.CloseTradeQuantity,
		Volume:    ticker.BaseVolume,
		Timestamp: time.Now(),
	}, nil
}

func (p *provider) TickerStream(stop <-chan bool, market types.MarketDTO) (<-chan types.TickerDTO, error) {
	tickerStream := coinfactory.GetSymbolService().GetSymbol(market.Name).GetTickerStream(stop)
	stream := make(chan types.TickerDTO)

	go func(stop <-chan bool, tickerStream <-chan api.SymbolTickerData, stream chan types.TickerDTO) {
		for {
			select {
			case <-stop:
				return
			default:
			}

			select {
			case <-stop:
				return
			case ticker := <-tickerStream:
				stream <- types.TickerDTO{
					Ask:       ticker.AskPrice,
					Bid:       ticker.BidPrice,
					Price:     ticker.CurrentClosePrice,
					Quantity:  ticker.CloseTradeQuantity,
					Volume:    ticker.BaseVolume,
					Timestamp: time.Now(),
				}
			}
		}
	}(stop, tickerStream, stream)
	return stream, nil
}

func (p *provider) OrderStream(stop <-chan bool, dto types.OrderDTO) (<-chan types.OrderDTO, error) {
	order, err := coinfactory.GetOrderService().GetOrder(dto.Market.Name, convertIDToInt(dto.ID))
	if err != nil {
		return nil, err
	}
	stream := make(chan types.OrderDTO)

	go func(stop <-chan bool, stream chan types.OrderDTO) {
		updateChan := order.GetUpdateChan()

		for {
			select {
			case <-stop:
				return
			default:
			}

			select {
			case <-stop:
				return
			case next := <-updateChan:
				if next {
					stream <- orderToDTO(dto.Market, order)
				}
			}
		}

	}(stop, stream)

	return stream, nil
}

func convertIDToInt(i string) (d int) {
	d, _ = strconv.Atoi(i)
	return
}

func convertIDToString(i int) (d string) {
	d = strconv.Itoa(i)
	return
}

func (p *provider) AttemptOrder(req types.OrderRequestDTO) (types.OrderDTO, error) {
	// Create the order request
	request := coinfactory.OrderRequest{
		Symbol:   req.Market.Name,
		Side:     string(req.Side),
		Price:    req.Price,
		Quantity: req.Quantity,
		Type:     string(req.Type),
	}

	// Place the order
	rawOrder, err := coinfactory.GetOrderService().AttemptOrder(request)
	if err != nil {
		return types.OrderDTO{}, err
	}

	// Create the dto
	dto := orderToDTO(req.Market, rawOrder)

	return dto, nil
}

func convertType(t string) types.OrderType {
	if t == "LIMIT" {
		return order.Limit
	}
	if t == "MARKET" {
		return order.Market
	}
	return order.Limit
}

func convertStatus(status string) types.OrderStatus {
	switch status {
	case "NEW":
		return order.Pending
	case "PARTIALLY_FILLED":
		return order.Partial
	case "FILLED":
		return order.Filled
	case "CANCELED":
		return order.Canceled
	case "PENDING_CANCEL":
		return order.Canceled
	case "REJECTED":
		return order.Rejected
	case "EXPIRED":
		return order.Expired
	}
	return order.Unknown
}

func (p *provider) CancelOrder(order types.OrderDTO) error {
	_, err := api.CancelOrder(api.OrderCancellationRequest{
		Symbol:  order.Market.Name,
		OrderID: convertIDToInt(order.ID),
	})
	return err
}

func (p *provider) Candles(mkt types.MarketDTO, interval types.CandleInterval, start time.Time, end time.Time) ([]types.CandleDTO, error) {
	candles := []types.CandleDTO{}

	// Get the Klines from binance
	rawCandles, err := coinfactory.GetSymbolService().GetSymbol(mkt.Name).GetKLines(toInterval(interval), start, end, 1000)
	if err != nil {
		return candles, err
	}

	for _, c := range rawCandles {
		candles = append(candles, types.CandleDTO{
			Close:     c.ClosePrice,
			High:      c.HighPrice,
			Low:       c.LowPrice,
			Open:      c.OpenPrice,
			Timestamp: c.OpenTime,
			Volume:    c.BaseVolume,
		})
	}
	return candles, nil
}

func (p *provider) Markets() (markets []types.MarketDTO, err error) {
	markets = []types.MarketDTO{}
	symbols := api.GetExchangeInfo().Symbols

	for _, symbol := range symbols {
		// Skip non trading markets
		if symbol.Status != "TRADING" {
			continue
		}
		baseCur := types.CurrencyDTO{
			Name:      symbol.BaseAsset,
			Symbol:    symbol.BaseAsset,
			Precision: symbol.BaseAssetPrecision,
		}
		quoteCur := types.CurrencyDTO{
			Name:      symbol.QuoteAsset,
			Symbol:    symbol.QuoteAsset,
			Precision: symbol.QuotePrecision,
		}

		m := types.MarketDTO{
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

	return markets, err
}

func (p *provider) Order(market types.MarketDTO, id string) (types.OrderDTO, error) {
	intId, _ := strconv.Atoi(id)
	rawOrder, err := coinfactory.GetOrderService().GetOrder(market.Name, intId)
	if err != nil {
		return types.OrderDTO{}, err
	}
	return orderToDTO(market, rawOrder), nil
}

func orderToDTO(market types.MarketDTO, rawOrder *coinfactory.Order) types.OrderDTO {
	return types.OrderDTO{
		Market:       market,
		CreationTime: rawOrder.GetCreationTime(),
		Filled:       rawOrder.GetStatus().ExecutedQuantity,
		ID:           convertIDToString(rawOrder.GetStatus().OrderID),
		Status:       convertStatus(rawOrder.GetStatus().Status),
		Request: types.OrderRequestDTO{
			Type:     convertType(rawOrder.Type),
			Side:     convertSides(rawOrder.Side),
			Price:    rawOrder.Price,
			Quantity: rawOrder.Quantity,
			Market:   market,
		},
	}
}

func convertSides(side string) (orderSide types.OrderSide) {
	if side == "BUY" {
		orderSide = order.Buy
	}
	if side == "SELL" {
		orderSide = order.Sell
	}
	return
}

func (p *provider) Currencies() ([]types.CurrencyDTO, error) {
	select {
	default:
		// We're not ready, bail
		cur := []types.CurrencyDTO{}
		for _, c := range currencies {
			cur = append(cur, c)
		}
		return cur, nil
	case <-updateTimer.C:
		// Update and return
	}

	updateCurrencies()

	cur := []types.CurrencyDTO{}
	for _, c := range currencies {
		cur = append(cur, c)
	}
	return cur, nil
}

func (p *provider) Wallets() (wallets []types.WalletDTO, err error) {
	data, err := api.GetUserData()
	if err != nil {
		return
	}

	wallets = []types.WalletDTO{}
	for _, bal := range data.Balances {
		cur, _ := currencies[bal.Asset]
		wallets = append(wallets, types.WalletDTO{cur, bal.Free, bal.Locked, decimal.Zero})
	}

	return wallets, err
}

func updateCurrencies() {
	symbols := api.GetExchangeInfo().Symbols
	for _, symbol := range symbols {
		if _, ok := currencies[symbol.BaseAsset]; !ok {
			currencies[symbol.BaseAsset] = types.CurrencyDTO{
				Symbol:    symbol.BaseAsset,
				Name:      symbol.BaseAsset,
				Precision: symbol.BaseAssetPrecision,
			}
		}

		if _, ok := currencies[symbol.QuoteAsset]; !ok {
			currencies[symbol.QuoteAsset] = types.CurrencyDTO{
				Symbol:    symbol.QuoteAsset,
				Name:      symbol.QuoteAsset,
				Precision: symbol.QuotePrecision,
			}
		}
	}
}

func toInterval(candleInt types.CandleInterval) (interval string) {
	switch candleInt {
	case candle.OneMinute:
		interval = "1m"
	case candle.FiveMinutes:
		interval = "5m"
	case candle.FifteenMinutes:
		interval = "15m"
	case candle.OneHour:
		interval = "1h"
	case candle.TwelveHours:
		interval = "12h"
	case candle.OneDay:
		interval = "1d"
	}
	return
}
