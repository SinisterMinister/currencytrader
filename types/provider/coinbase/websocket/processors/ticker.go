package processors

import (
	"github.com/go-playground/log/v7"
	traderTypes "github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Ticker interface {
	types.OrderProcessor
	Input() chan<- types.Ticker
}

func NewTicker(stop <-chan bool) (Ticker, error) {
	instance := ticker{
		log:    log.WithField("source", "coinbase.websocket.processors.ticker"),
		input:  make(chan types.Ticker, viper.GetInt("coinbase.websocket.processors.ticker.inputBufferSize")),
		output: make(chan traderTypes.OrderDTO, viper.GetInt("coinbase.websocket.processors.ticker.outputBufferSize")),
	}

	return &instance, nil
}

type ticker struct {
	log    log.Entry
	input  chan types.Ticker
	output chan traderTypes.OrderDTO
}

func (ch *ticker) Input() (input chan<- types.Ticker) {
	return
}

func (ch *ticker) Output() (output <-chan traderTypes.OrderDTO) {
	return
}
