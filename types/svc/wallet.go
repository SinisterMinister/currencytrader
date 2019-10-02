package svc

import (
	"errors"
	"sync"

	wal "github.com/sinisterminister/currencytrader/types/wallet"

	"github.com/go-playground/log"
	"github.com/sirupsen/logrus"

	"github.com/sinisterminister/currencytrader/types"
	"github.com/sinisterminister/currencytrader/types/internal"
)

type wallet struct {
	trader internal.Trader

	mutex   sync.RWMutex
	streams map[internal.Wallet]<-chan types.WalletDTO
	stop    chan bool
}

func NewWallet(trader internal.Trader) internal.WalletSvc {
	return &wallet{
		trader: trader,
	}
}

func (w *wallet) Start() {
	w.mutex.Lock()
	w.stop = make(chan bool)
	w.startWalletStreams()
	w.mutex.Unlock()
}
func (w *wallet) Stop() {
	w.mutex.Lock()
	close(w.stop)
	w.mutex.Unlock()
}

func (w *wallet) Wallet(currency types.Currency) (wal types.Wallet, err error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for wallet := range w.streams {
		if wallet.Currency() == currency {
			wal = wallet
			return
		}
	}
	err = errors.New("no wallet found for currency")
	return
}

func (w *wallet) Wallets() (wallets []types.Wallet) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	wallets = make([]types.Wallet, 0, len(w.streams))
	for w := range w.streams {
		wallets = append(wallets, w)
	}

	return
}

func (w *wallet) startWalletStreams() {

	wallets, err := w.trader.Provider().Wallets()
	if err != nil {
		log.WithError(err).Fatal("could not fetch wallets from provider!")
	}

	streams := make(map[internal.Wallet]<-chan types.WalletDTO)
	for _, dto := range wallets {
		wallet := wal.New(dto)
		ch, err := w.trader.Provider().WalletStream(w.stop, wallet.ToDTO())

		if err != nil {
			logrus.WithError(err).Panicf("could not get update stream for wallet %s", wallet.Currency().Name)
		}
		streams[wallet] = ch
	}

	w.streams = streams

	for wallet, stream := range streams {
		go func(stop <-chan bool, wallet internal.Wallet, stream <-chan types.WalletDTO) {
			for {
				select {
				case <-stop:
					return
				default:
				}

				select {
				case <-stop:
					return
				case data := <-stream:
					wallet.Update(data)
				}
			}
		}(w.stop, wallet, stream)
	}
}
