package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type ticker struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Ticker
}

func Ticker(stop <-chan bool) (msgHandler types.TickerMessageHandler, err error) {
	var handler ticker = ticker{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.ticker.incomingBufferSize")),
		outgoing: make(chan types.Ticker, viper.GetInt("coinbase.websocket.handlers.ticker.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.ticker"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *ticker) Name() string {
	return TICKER_HANDLER_NAME
}

func (handler *ticker) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *ticker) Output() <-chan types.Ticker {
	return handler.outgoing
}

func (handler *ticker) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Ticker
			handler.log.Debugf("handling %s payload", handler.Name())
			e := json.Unmarshal(pkg.Data, &data)
			if e != nil {
				handler.log.Errorf("could not parse data for %s", handler.Name())
				continue
			}
			handler.outgoing <- data
		}
	}
}
