package poloniex

import "github.com/shopspring/decimal"

type OrderBook struct {
	Asks     [][]decimal.Decimal `json:"asks"`
	Bids     [][]decimal.Decimal `json:"bids"`
	IsFrozen int                 `json:"isFrozen,string"`
	Error    string              `json:"error"`
	Seq      int                 `json:"seq"`
}

// This can probably be implemented using UnmarshalJSON
/*
type OrderBook struct {
	Bids     []Orderb `json:"bids"`
	Asks     []Orderb `json:"asks"`
	IsFrozen int      `json:"isFrozen,string"`
}
type Orderb struct {
	Rate     string
	Quantity float64
}
*/

type OpenOrder struct {
	OrderNumber int     `json:"orderNumber,string"`
	Type        string  `json:"type"`
	Rate        float64 `json:"rate,string"`
	Amount      float64 `json:"amount,string"`
	Total       float64 `json:"total,string"`
}
