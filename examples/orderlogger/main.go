package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sinisterminister/currencytrader/types/order"

	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/simulated"
	"github.com/sirupsen/logrus"
)

func main() {
	// Start up the simulated provider
	provider := simulated.New(simulated.ProviderConfig{})

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Get the available markets
	markets := trader.MarketSvc().GetMarkets()

	// Setup a close channel
	killSwitch := make(chan bool)

	// Stream the tickers to output log
	for _, mkt := range markets {
		go placeOrders(killSwitch, mkt)
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

func placeOrders(stop <-chan bool, mkt types.Market) {
	// Get a ticker to get a price
	ticker, err := mkt.Ticker()

	// Bail on error
	if err != nil {
		logrus.WithError(err).Error("could not get ticker")
		return
	}

	logrus.Infof("ticker for market %s is %s", mkt.Name(), ticker.ToDTO())
	buy, err := mkt.AttemptOrder(order.Buy, ticker.Price(), decimal.NewFromFloat(10))
	// Bail on error
	if err != nil {
		logrus.WithError(err).Error("could not place order")
		return
	}
	t := time.NewTicker(1 * time.Second)

	// Watch and wait for the order to be fulfilled or canceled
	for {
		select {
		case <-t.C:
			logrus.Infof("order %s status %s", buy.ID(), buy.Status())
			if buy.Status() == order.Filled || buy.Status() == order.Canceled {
				logrus.Infof("order %s finished with a status of %s", buy.ID(), buy.Status())
				return
			}
		}
	}
}
