package types

import "time"

type Level2Update struct {
	Message
	ProductID string     `json:"product_id"`
	Time      time.Time  `json:"time"`
	Changes   [][]string `json:"changes"`
}
