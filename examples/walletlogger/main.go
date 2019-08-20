package main

import (
	"os"
	"os/signal"
	"time"

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

	// Get the wallets
	wallets := trader.WalletSvc().GetWallets()

	// Setup a close channel
	killSwitch := make(chan bool)

	// Log the wallets
	for _, wallet := range wallets {
		go logWallet(killSwitch, wallet)
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

func logWallet(stop <-chan bool, wal types.Wallet) {
	ticker := time.NewTicker(1 * time.Second)
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
		case <-ticker.C:
			logrus.WithField("wallet", wal.ToDTO()).Infof("wallet data for for %s", wal.Currency().Name())
		}
	}
}
