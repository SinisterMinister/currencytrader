package processors

import (
	"github.com/go-playground/log/v7"
	traderTypes "github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Match interface {
	types.OrderProcessor
	Input() chan<- types.Match
}

func NewMatch(stop <-chan bool) (Match, error) {
	instance := match{
		log:    log.WithField("source", "coinbase.websocket.processors.match"),
		input:  make(chan types.Match, viper.GetInt("coinbase.websocket.processors.match.inputBufferSize")),
		output: make(chan traderTypes.OrderDTO, viper.GetInt("coinbase.websocket.processors.match.outputBufferSize")),
	}

	return &instance, nil
}

type match struct {
	log    log.Entry
	input  chan types.Match
	output chan traderTypes.OrderDTO
}

func (ch *match) Input() (input chan<- types.Match) {
	return
}

func (ch *match) Output() (output <-chan traderTypes.OrderDTO) {
	return
}
