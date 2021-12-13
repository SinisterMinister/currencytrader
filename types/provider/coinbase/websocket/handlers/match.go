package handlers

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/sinisterminister/currencytrader/types/provider/coinbase/websocket/types"
	"github.com/spf13/viper"
)

type match struct {
	log      log.Entry
	stop     <-chan bool
	incoming chan types.DataPackage
	outgoing chan types.Match
}

func Match(stop <-chan bool) (msgHandler types.MatchMessageHandler, err error) {
	var handler match = match{
		incoming: make(chan types.DataPackage, viper.GetInt("coinbase.websocket.handlers.match.incomingBufferSize")),
		outgoing: make(chan types.Match, viper.GetInt("coinbase.websocket.handlers.match.outgoingBufferSize")),
		stop:     stop,
		log:      log.WithField("source", "coinbase.websocket.handlers.match"),
	}

	// Start handler
	go handler.start()

	// Return the handler
	return &handler, err
}

func (handler *match) Name() string {
	return MATCH_HANDLER_NAME
}

func (handler *match) Input() chan<- types.DataPackage {
	return handler.incoming
}

func (handler *match) Output() <-chan types.Match {
	return handler.outgoing
}

func (handler *match) start() {
	for {
		select {
		// Kill switch flipped
		case <-handler.stop:
			return

		// Handle incoming data
		case pkg := <-handler.incoming:
			// Capture data
			var data types.Match
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
