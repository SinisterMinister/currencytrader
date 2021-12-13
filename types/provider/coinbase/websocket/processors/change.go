package processors

import (
	"github.com/go-playground/log/v7"
	traderTypes "github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Change interface {
	types.OrderProcessor
	Input() chan<- types.Change
}

func NewChange(stop <-chan bool) (Change, error) {
	instance := change{
		log:    log.WithField("source", "coinbase.websocket.processors.change"),
		input:  make(chan types.Change, viper.GetInt("coinbase.websocket.processors.change.inputBufferSize")),
		output: make(chan traderTypes.OrderDTO, viper.GetInt("coinbase.websocket.processors.change.outputBufferSize")),
	}

	return &instance, nil
}

type change struct {
	log    log.Entry
	input  chan types.Change
	output chan traderTypes.OrderDTO
}

func (ch *change) Input() (input chan<- types.Change) {
	return
}

func (ch *change) Output() (output <-chan traderTypes.OrderDTO) {
	return
}
