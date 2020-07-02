package client

import (
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/shopspring/decimal"
)

type Fees struct {
	MakerRate decimal.Decimal `json:"maker_fee_rate"`
	TakerRate decimal.Decimal `json:"taker_fee_rate"`
	Volume    decimal.Decimal `json:"usd_volume"`
}

type Client struct {
	*coinbasepro.Client
}

func NewClient() *Client {
	return &Client{coinbasepro.NewClient()}
}

func (c *Client) GetFees() (fees Fees, err error) {
	// Fetch the fees
	_, err = c.Request("GET", "/fees", nil, &fees)
	if err != nil {
		return
	}

	return
}
