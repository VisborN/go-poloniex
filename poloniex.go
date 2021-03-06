//Package poloniex is an implementation of the Poloniex API in Golang.
package poloniex

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	API_BASE = "https://poloniex.com"   // Poloniex API endpoint
	API_WS   = "wss://api.poloniex.com" // Poloniex WS endpoint
)

// New returns an instantiated poloniex struct
func New(apiKey, apiSecret string) *Poloniex {
	client := NewClient(apiKey, apiSecret)
	return &Poloniex{client}
}

// New returns an instantiated poloniex struct with custom timeout
func NewWithCustomTimeout(apiKey, apiSecret string, timeout time.Duration) *Poloniex {
	client := NewClientWithCustomTimeout(apiKey, apiSecret, timeout)
	return &Poloniex{client}
}

// poloniex represent a poloniex client
type Poloniex struct {
	client *client
}

// set enable/disable http request/response dump
func (b *Poloniex) SetDebug(enable bool) {
	b.client.debug = enable
}

// GetTickers is used to get the ticker for all markets
func (b *Poloniex) GetTickers() (tickers map[string]Ticker, err error) {
	r, err := b.client.do("GET", "public?command=returnTicker", nil, false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &tickers); err != nil {
		return
	}
	return
}

// GetVolumes is used to get the volume for all markets
func (b *Poloniex) GetVolumes() (vc VolumeCollection, err error) {
	r, err := b.client.do("GET", "public?command=return24hVolume", nil, false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &vc); err != nil {
		return
	}
	return
}

func (b *Poloniex) GetCurrencies() (currencies Currencies, err error) {
	r, err := b.client.do("GET", "public?command=returnCurrencies", nil, false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &currencies.Pair); err != nil {
		return
	}
	return
}

// GetOrderBook is used to get retrieve the orderbook for a given market
// market: a string literal for the market (ex: BTC_NXT). 'all' in other function.
// cat: bid, ask or both to identify the type of orderbook to return.
// depth: how deep of an order book to retrieve
func (b *Poloniex) GetOrderBook(market, cat string, depth int) (orderBook OrderBook, err error) {
	// not implemented
	if cat != "bid" && cat != "ask" && cat != "both" {
		cat = "both"
	}
	if depth > 100 {
		depth = 100
	}
	if depth < 1 {
		depth = 1
	}

	r, err := b.client.do("GET", fmt.Sprintf("public?command=returnOrderBook&currencyPair=%s&depth=%d", strings.ToUpper(market), depth), nil, false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &orderBook); err != nil {
		return
	}
	if orderBook.Error != "" {
		err = errors.New(orderBook.Error)
		return
	}
	return
}

// GetAllOrderBook is used to get retrieve the orderbook for all markets
// cat: bid, ask or both to identify the type of orderbook to return.
// depth: how deep of an order book to retrieve
func (b *Poloniex) GetAllOrderBook(cat string, depth int) (orderBook map[string]OrderBook, err error) {
	// not implemented
	if cat != "bid" && cat != "ask" && cat != "both" {
		cat = "both"
	}
	if depth > 100 {
		depth = 100
	}
	if depth < 1 {
		depth = 1
	}

	r, err := b.client.do("GET", fmt.Sprintf("public?command=returnOrderBook&currencyPair=all&depth=%d", depth), nil, false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &orderBook); err != nil {
		return
	}
	return
}

// Returns candlestick chart data. Required GET parameters are "currencyPair",
// "period" (candlestick period in seconds; valid values are 300, 900, 1800,
// 7200, 14400, and 86400), "start", and "end". "Start" and "end" are given in
// UNIX timestamp format and used to specify the date range for the data
// returned.
func (b *Poloniex) ChartData(currencyPair string, period int, start, end time.Time) (candles []*CandleStick, err error) {
	r, err := b.client.do("GET", fmt.Sprintf(
		"public?command=returnChartData&currencyPair=%s&period=%d&start=%d&end=%d",
		strings.ToUpper(currencyPair),
		period,
		start.Unix(),
		end.Unix(),
	), nil, false)
	if err != nil {
		return
	}

	if err = json.Unmarshal(r, &candles); err != nil {
		return
	}

	return
}

// SubscribeOrderBook subscribes for trades and order book updates via WAMP.
//	symbol - a symbol you are interested in.
//	updatesCh - a channel for market updates.
//	stopCh - a channel to cancel or reset ws subscribtion.
//		close it or send 'true' to stop subscribtion.
//		send 'false' to reconnect. May be useful, if updates were stalled.
func (b *Poloniex) SubscribeOrderBook(symbol string, updatesCh chan<- MarketUpd, stopCh <-chan bool) error {
	for {
		if cont, err := b.client.wsConnect(symbol, makeOBookSubHandler(updatesCh), stopCh); !cont {
			return err
		}
	}
}

// UnsubscribeAll cancels all active subscriptions.
func (b *Poloniex) UnsubscribeAll() error {
	return b.client.wsReset()
}

// Close closes ws connections.
func (b *Poloniex) Close() error {
	return b.client.close()
}

// SubscribeTicker subscribes for ticker via WAMP.
// Send to, or close stopCh to cancel subscribtion.
//	updatesCh - a channel for ticker updates.
//	stopCh - a channel to cancel or reset ws subscribtion.
//		close it or send 'true' to stop subscribtion.
//		send 'false' to reconnect. May be useful, if updates were stalled.
func (b *Poloniex) SubscribeTicker(updatesCh chan<- TickerUpd, stopCh <-chan bool) error {
	for {
		if cont, err := b.client.wsConnect("ticker", makeTickerSubHandler(updatesCh), stopCh); !cont {
			return err
		}
	}
}

func (b *Poloniex) GetBalances() (balances map[string]Balance, err error) {
	balances = make(map[string]Balance)
	r, err := b.client.doCommand("returnCompleteBalances", nil)
	if err != nil {
		return
	}

	if err = json.Unmarshal(r, &balances); err != nil {
		return
	}

	return
}

func (b *Poloniex) GetTradeHistory(pair string, start uint32) (trades map[string][]Trade, err error) {
	trades = make(map[string][]Trade)
	r, err := b.client.doCommand("returnTradeHistory", map[string]string{"currencyPair": pair, "start": strconv.FormatUint(uint64(start), 10)})
	if err != nil {
		return
	}

	if pair == "all" {
		if err = json.Unmarshal(r, &trades); err != nil {
			return
		}
	} else {
		var pairTrades []Trade
		if err = json.Unmarshal(r, &pairTrades); err != nil {
			return
		}
		trades[pair] = pairTrades
	}

	return
}

type responseDepositsWithdrawals struct {
	Deposits    []Deposit    `json:"deposits"`
	Withdrawals []Withdrawal `json:"withdrawals"`
}

func (b *Poloniex) GetDepositsWithdrawals(start uint32, end uint32) (deposits []Deposit, withdrawals []Withdrawal, err error) {
	deposits = make([]Deposit, 0)
	withdrawals = make([]Withdrawal, 0)
	r, err := b.client.doCommand("returnDepositsWithdrawals", map[string]string{"start": strconv.FormatUint(uint64(start), 10), "end": strconv.FormatUint(uint64(end), 10)})
	if err != nil {
		return
	}
	var response responseDepositsWithdrawals
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}

	return response.Deposits, response.Withdrawals, nil
}

func (b *Poloniex) Buy(pair string, rate float64, amount float64, tradeType string) (TradeOrder, error) {
	reqParams := map[string]string{
		"currencyPair": pair, "rate": strconv.FormatFloat(rate, 'f', -1, 64),
		"amount": strconv.FormatFloat(amount, 'f', -1, 64)}
	if tradeType != "" {
		reqParams[tradeType] = "1"
	}
	r, err := b.client.doCommand("buy", reqParams)
	if err != nil {
		return TradeOrder{}, err
	}
	var orderResponse TradeOrder
	if err = json.Unmarshal(r, &orderResponse); err != nil {
		return TradeOrder{}, err
	}

	return orderResponse, nil
}

func (b *Poloniex) Sell(pair string, rate float64, amount float64, tradeType string) (TradeOrder, error) {
	reqParams := map[string]string{
		"currencyPair": pair, "rate": strconv.FormatFloat(rate, 'f', -1, 64),
		"amount": strconv.FormatFloat(amount, 'f', -1, 64)}
	if tradeType != "" {
		reqParams[tradeType] = "1"
	}
	r, err := b.client.doCommand("sell", reqParams)
	if err != nil {
		return TradeOrder{}, err
	}
	var orderResponse TradeOrder
	if err = json.Unmarshal(r, &orderResponse); err != nil {
		return TradeOrder{}, err
	}

	return orderResponse, nil
}

func (b *Poloniex) GetOpenOrders(pair string) (openOrders map[string][]OpenOrder, err error) {
	openOrders = make(map[string][]OpenOrder)
	r, err := b.client.doCommand("returnOpenOrders", map[string]string{"currencyPair": pair})
	if err != nil {
		return
	}
	if pair == "all" {
		if err = json.Unmarshal(r, &openOrders); err != nil {
			return
		}
	} else {
		var onePairOrders []OpenOrder
		if err = json.Unmarshal(r, &onePairOrders); err != nil {
			return
		}
		openOrders[pair] = onePairOrders
	}
	return
}

func (b *Poloniex) GetFees() (Fees, error) {
	reqParams := map[string]string{}
	r, err := b.client.doCommand("returnFeeInfo", reqParams)
	if err != nil {
		return Fees{}, err
	}
	var orderResponse Fees
	if err = json.Unmarshal(r, &orderResponse); err != nil {
		return Fees{}, err
	}

	return orderResponse, nil
}
