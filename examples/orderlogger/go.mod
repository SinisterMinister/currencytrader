module github.com/sinisterminister/currencyexchange/examples/orderlogger

go 1.13

require (
	github.com/go-playground/log/v7 v7.0.2
	github.com/preichenberger/go-coinbasepro/v2 v2.0.5
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/sinisterminister/currencytrader v0.0.0
	github.com/spf13/viper v1.6.2
)

replace github.com/sinisterminister/currencytrader => ../../../currencytrader
