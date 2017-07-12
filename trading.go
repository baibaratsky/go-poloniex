package poloniex

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const tradingApiEndpoint = "https://poloniex.com/tradingApi"

const (
	TypeBuy  = "buy"
	TypeSell = "sell"
)

type FeeInfo struct {
	MakerFee        decimal.Decimal
	TakerFee        decimal.Decimal
	ThirtyDayVolume decimal.Decimal
	NextTier        decimal.Decimal
}

func (client *Client) FeeInfo() (feeInfo FeeInfo, err error) {
	err = client.tradingApiRequest(&feeInfo, "returnFeeInfo")
	return
}

func (client *Client) Balances() (balances map[string]decimal.Decimal, err error) {
	err = client.tradingApiRequest(&balances, "returnBalances")
	return
}

type Trade struct {
	GlobalTradeId uint64 `json:"globalTradeID"`
	Id            uint64 `json:"tradeID,string"`
	OrderNumber   uint64 `json:"orderNumber,string"`
	CurrencyPair  string
	Type          string
	Rate          decimal.Decimal
	Amount        decimal.Decimal
	Total         decimal.Decimal
	Fee           decimal.Decimal
	Date          string
}

func (client *Client) TradeHistory(currencyPair string, start, end int64) (trades []Trade, err error) {
	params := Params{
		"currencyPair": currencyPair,
	}

	if start > 0 {
		params["start"] = fmt.Sprintf("%d", start)
	}
	if end > 0 {
		params["end"] = fmt.Sprintf("%d", end)
	}

	err = client.tradingApiRequest(&trades, "returnTradeHistory", params)

	for i := range trades {
		trades[i].CurrencyPair = currencyPair
	}

	return
}

func (client *Client) TradeHistoryAll(start, end int64) (trades map[string][]Trade, err error) {
	params := Params{
		"currencyPair": "all",
	}

	if start > 0 {
		params["start"] = fmt.Sprintf("%d", start)
	}
	if end > 0 {
		params["end"] = fmt.Sprintf("%d", end)
	}

	err = client.tradingApiRequest(&trades, "returnTradeHistory", params)

	for pair := range trades {
		for i := range trades[pair] {
			trades[pair][i].CurrencyPair = pair
		}
	}

	return
}

func (client *Client) OrderTrades(orderNumber uint64) (trades []Trade, err error) {
	err = client.tradingApiRequest(&trades, "returnOrderTrades",
		Params{"orderNumber": fmt.Sprintf("%d", orderNumber)})

	for i := range trades {
		trades[i].OrderNumber = orderNumber
	}

	return
}

type OwnOrder struct {
	OrderNumber uint64 `json:"orderNumber,string"`
	Type        string
	Rate        decimal.Decimal
	Amount      decimal.Decimal
	Total       decimal.Decimal
}

func (client *Client) OpenOrders(currencyPair string) (orders []OwnOrder, err error) {
	err = client.tradingApiRequest(&orders, "returnOpenOrders", Params{
		"currencyPair": currencyPair,
	})
	return
}

func (client *Client) OpenOrdersAll() (orders map[string][]OwnOrder, err error) {
	err = client.tradingApiRequest(&orders, "returnOpenOrders", Params{
		"currencyPair": "all",
	})
	return
}

type PlacedOrder struct {
	OrderNumber     uint64 `json:"orderNumber,string"`
	ResultingTrades []Trade
}

type UpdatedOrder struct {
	OrderNumber     uint64 `json:"orderNumber,string"`
	ResultingTrades map[string][]Trade
}

func (client *Client) Buy(currencyPair string, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	err = client.tradingApiRequest(&placedOrder, "buy", Params{
		"currencyPair": currencyPair,
		"rate":         rate.String(),
		"amount":       amount.String(),
	})
	return
}

func (client *Client) Sell(currencyPair string, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	err = client.tradingApiRequest(&placedOrder, "sell", Params{
		"currencyPair": currencyPair,
		"rate":         rate.String(),
		"amount":       amount.String(),
	})
	return
}

func (client *Client) CancelOrder(orderNumber uint64) (success bool, err error) {
	result := struct {
		Success convertibleBool
	}{}
	err = client.tradingApiRequest(&result, "cancelOrder",
		Params{"orderNumber": fmt.Sprintf("%d", orderNumber)})
	success = bool(result.Success)
	return
}

func (client *Client) MoveOrder(orderNumber uint64, rate, amount decimal.Decimal) (updatedOrder UpdatedOrder, err error) {
	result := struct {
		Success convertibleBool
		UpdatedOrder
	}{}

	params := Params{
		"orderNumber": fmt.Sprintf("%d", orderNumber),
		"rate":        rate.String(),
	}

	// amount > 0
	if amount.Cmp(decimal.Zero) == 1 {
		params["amount"] = amount.String()
	}

	err = client.tradingApiRequest(&result, "moveOrder", params)

	if !result.Success {
		err = errors.New("result is not successful")
	}

	updatedOrder = result.UpdatedOrder

	return
}

type errorResponse struct {
	Error *string
}

func (client *Client) tradingApiRequest(result interface{}, method string, params ...Params) (err error) {
	if len(params) > 1 {
		return errors.New("too much arguments")
	}

	formData := Params{
		"command": method,
	}

	if len(params) == 1 {
		for name, value := range params[0] {
			formData[name] = value
		}
	}

	err = client.limiter.Wait(context.TODO())
	if err != nil {
		return
	}

	key := client.keyPool.Get()

	nonce := time.Now().UnixNano()
	formData["nonce"] = fmt.Sprintf("%d", nonce)

	request := client.resty.R().
		SetFormData(formData)

	signature := hmac.New(sha512.New, []byte(key.Secret))
	signature.Write([]byte(request.FormData.Encode()))

	request.SetHeader("Key", key.Key).
		SetHeader("Sign", hex.EncodeToString(signature.Sum(nil)))

	response, err := request.Post(tradingApiEndpoint)
	client.keyPool.Put(key)
	if err != nil {
		return
	}

	errorResponse := errorResponse{}
	json.Unmarshal(response.Body(), &errorResponse)
	if errorResponse.Error != nil {
		return errors.New(*errorResponse.Error)
	}

	err = json.Unmarshal(response.Body(), result)
	if err != nil {
		err = fmt.Errorf("%s\nServer response: %s", err.Error(), string(response.Body()))
	}
	return
}
