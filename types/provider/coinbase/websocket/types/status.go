package types

import "github.com/shopspring/decimal"

type Status struct {
	Message
	Products []struct {
		ID             string          `json:"id"`
		BaseCurrency   string          `json:"base_currency"`
		QuoteCurrency  string          `json:"quote_currency"`
		BaseMinSize    decimal.Decimal `json:"base_min_size"`
		BaseMaxSize    decimal.Decimal `json:"base_max_size"`
		BaseIncrement  decimal.Decimal `json:"base_increment"`
		QuoteIncrement decimal.Decimal `json:"quote_increment"`
		DisplayName    string          `json:"display_name"`
		Status         string          `json:"status"`
		StatusMessage  string          `json:"status_message"`
		MinMarketFunds decimal.Decimal `json:"min_market_funds"`
		MaxMarketFunds decimal.Decimal `json:"max_market_funds"`
		PostOnly       bool            `json:"post_only"`
		LimitOnly      bool            `json:"limit_only"`
		CancelOnly     bool            `json:"cancel_only"`
	} `json:"products"`
	Currencies []struct {
		ID            string                 `json:"id"`
		Name          string                 `json:"name"`
		MinSize       string                 `json:"min_size"`
		Status        string                 `json:"status"`
		StatusMessage string                 `json:"status_message"`
		MaxPrecision  decimal.Decimal        `json:"max_precision"`
		ConvertableTo []string               `json:"convertible_to"`
		Details       map[string]interface{} `json:"details"`
	} `json:"currencies"`
}
