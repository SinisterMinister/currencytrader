package coinbase

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("coinbase.websocketURL", "wss://ws-feed.pro.coinbase.com")
	viper.SetDefault("coinbase.websocket.incomingDataBufferSize", 1024)

	viper.SetDefault("coinbase.websocket.tickerHandlerInputBufferSize", 8)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerInputBufferSize", 8)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerInputBufferSize", 8)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerInputBufferSize", 8)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerInputBufferSize", 8)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerInputBufferSize", 8)

	viper.SetDefault("coinbase.websocket.tickerHandlerOutputBufferSize", 1)
	viper.SetDefault("coinbase.websocket.orderReceivedHandlerOutputBufferSize", 1)
	viper.SetDefault("coinbase.websocket.orderOpenHandlerOutputBufferSize", 1)
	viper.SetDefault("coinbase.websocket.orderDoneHandlerOutputBufferSize", 1)
	viper.SetDefault("coinbase.websocket.orderMatchHandlerOutputBufferSize", 1)
	viper.SetDefault("coinbase.websocket.orderChangeHandlerOutputBufferSize", 1)

	viper.SetDefault("coinbase.streams.tickerStreamBufferSize", 1)
	viper.SetDefault("coinbase.streams.orderStreamBufferSize", 1)
}
