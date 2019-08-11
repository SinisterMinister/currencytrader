package main

import (
	"os"
	"os/signal"

	"github.com/sinisterminister/moneytrader"
	"github.com/sinisterminister/moneytrader/types"
	"github.com/sinisterminister/moneytrader/types/provider/simulated"
	"github.com/sirupsen/logrus"
)

func main() {
	// Start up the simulated provider
	provider := simulated.New(simulated.ProviderConfig{})

	// Get an instance of the trader
	trader := moneytrader.New(provider)
	trader.Start()

	// Get the available markets
	markets := trader.MarketSvc().GetMarkets()

	// Setup a close channel
	killSwitch := make(chan bool)

	// Stream the tickers to output log
	for _, mkt := range markets {
		go streamTicker(killSwitch, mkt)
	}

	// Intercept the interrupt signal and pass it along
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Wait for the interrupt
	select {
	case <-interrupt:
		// Let the user know what happened
		logrus.Warn("Received an interrupt signal! Shutting down!")

		// Kill the streams
		close(killSwitch)

		// Shutdown the
		trader.Stop()
	}
}

func streamTicker(stop <-chan bool, market types.Market) {
	stream := market.TickerStream(stop)
	for {
		// Bail out on stop
		select {
		case <-stop:
			return
		default:
		}

		select {
		//Backup bailout
		case <-stop:
			return
		case data := <-stream:
			logrus.WithField("data", data).Infof("stream data recieved for %s market", market.Name())
		}
	}
}
