package main

import (
	"os"
	"os/signal"

	coinbase "github.com/preichenberger/go-coinbasepro"
	"github.com/sinisterminister/currencytrader"
	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/provider/coinbasepro"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup a coinbase client
	client := coinbase.NewClient()
	client.UpdateConfig(&coinbase.ClientConfig{
		Key:        "f561da92e7e431717e01b81339a92240",
		Passphrase: "throwback",
		Secret:     "YY7CvMVlA1/Ld9joXidr1brEc2xn9MOIacGijym7md3yv6heK9Z52IDFD7rhY3fwQvNaamZX8KcVHvAjnTpMng==",
	})
	// Start up a coinbase provider
	provider := coinbasepro.New(client)

	// Get an instance of the trader
	trader := currencytrader.New(provider)
	trader.Start()

	// Get the available markets
	markets := trader.MarketSvc().Markets()

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
	<-interrupt

	// Let the user know what happened
	logrus.Warn("Received an interrupt signal! Shutting down!")

	// Kill the streams
	close(killSwitch)

	// Shutdown the trader
	trader.Stop()
}

func streamTicker(stop <-chan bool, market types.Market) {
	// Get the ticker stream from the market
	stream := market.TickerStream(stop)

	// Watch the stream and log any data sent over it
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

		// Data received
		case data := <-stream:
			if data != nil {
				logrus.WithField("data", data.ToDTO()).Infof("stream data recieved for %s market", market.Name())
			}
		}
	}
}
