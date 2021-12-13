package processors

import (
	"github.com/go-playground/log/v7"
	traderTypes "github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Open interface {
	types.OrderProcessor
	Input() chan<- types.Open
}

func NewOpen(stop <-chan bool) (Open, error) {
	instance := open{
		log:    log.WithField("source", "coinbase.websocket.processors.open"),
		input:  make(chan types.Open, viper.GetInt("coinbase.websocket.processors.open.inputBufferSize")),
		output: make(chan traderTypes.OrderDTO, viper.GetInt("coinbase.websocket.processors.open.outputBufferSize")),
	}

	return &instance, nil
}

type open struct {
	log    log.Entry
	input  chan types.Open
	output chan traderTypes.OrderDTO
}

func (ch *open) Input() (input chan<- types.Open) {
	return
}

func (ch *open) Output() (output <-chan traderTypes.OrderDTO) {
	return
}
