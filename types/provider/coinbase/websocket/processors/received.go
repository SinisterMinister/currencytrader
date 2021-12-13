package processors

import (
	"github.com/go-playground/log/v7"
	traderTypes "github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type Received interface {
	types.OrderProcessor
	Input() chan<- types.Received
}

func NewReceived(stop <-chan bool) (Received, error) {
	instance := received{
		log:    log.WithField("source", "coinbase.websocket.processors.log"),
		input:  make(chan types.Received, viper.GetInt("coinbase.websocket.processors.received.inputBufferSize")),
		output: make(chan traderTypes.OrderDTO, viper.GetInt("coinbase.websocket.processors.received.outputBufferSize")),
	}

	instance.start(stop)

	return &instance, nil
}

type received struct {
	log    log.Entry
	input  chan types.Received
	output chan traderTypes.OrderDTO
}

func (proc *received) Input() (input chan<- types.Received) {
	return proc.input
}

func (proc *received) Output() (output <-chan traderTypes.OrderDTO) {
	return proc.output
}

func (proc *received) start(stop <-chan bool) {

	for {
		select {
		// Kill switch flipped
		case <-stop:
			return

		// Handle input data
		case orderData := <-proc.input:
			// Bail out if there's no ClientOrderID as it isn't an order we care about
			clientId := orderData.ClientOrderID
			if clientId == "" {
				continue
			}

		}
	}
}
