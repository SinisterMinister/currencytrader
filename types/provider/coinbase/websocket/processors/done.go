package processors

import (
	"github.com/go-playground/log/v7"
	traderTypes "github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Done interface {
	types.OrderProcessor
	Input() chan<- types.Done
}

func NewDone(stop <-chan bool) (Done, error) {
	instance := done{
		log:    log.WithField("source", "coinbase.websocket.processors.done"),
		input:  make(chan types.Done, viper.GetInt("coinbase.websocket.processors.done.inputBufferSize")),
		output: make(chan traderTypes.OrderDTO, viper.GetInt("coinbase.websocket.processors.done.outputBufferSize")),
	}

	return &instance, nil
}

type done struct {
	log    log.Entry
	input  chan types.Done
	output chan traderTypes.OrderDTO
}

func (ch *done) Input() (input chan<- types.Done) {
	return
}

func (ch *done) Output() (output <-chan traderTypes.OrderDTO) {
	return
}
