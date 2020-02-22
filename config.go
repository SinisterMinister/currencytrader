package currencytrader

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("currencytrader.tickersvc.streamBufferSize", 64)
	viper.SetDefault("currencytrader.tickersvc.streamBufferSize", 4)
}