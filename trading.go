package poloniex

import (
	"time"
	"crypto/sha512"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"github.com/shopspring/decimal"
	"errors"
	"fmt"
	"context"
)

const tradingApiEndpoint = "https://poloniex.com/tradingApi"

const (
	TypeBuy = "buy"
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

func (client *Client) OrderTrades(orderNumber uint64) (trades []Trade, err error) {
	err = client.tradingApiRequest(&trades, "returnOrderTrades",
		Params{"orderNumber": fmt.Sprintf("%d", orderNumber)})

	for i := range trades {
		trades[i].OrderNumber = orderNumber
	}

	return
}

type Order struct {
	OrderNumber uint64 `json:"orderNumber,string"`
	Type        string
	Rate        decimal.Decimal
	Amount      decimal.Decimal
	Total       decimal.Decimal
}

func (client *Client) OpenOrders(currencyPair string) (orders []Order, err error) {
	err = client.tradingApiRequest(&orders, "returnOpenOrders", Params{
		"currencyPair": currencyPair,
	})
	return
}

type PlacedOrder struct {
	OrderNumber     uint64 `json:"orderNumber,string"`
	ResultingTrades []Trade
}

func (client *Client) Buy(currencyPair string, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	err = client.tradingApiRequest(&placedOrder, "buy", Params{
		"currencyPair": currencyPair,
		"rate": rate.String(),
		"amount": amount.String(),
	})
	return
}

func (client *Client) Sell(currencyPair string, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	err = client.tradingApiRequest(&placedOrder, "sell", Params{
		"currencyPair": currencyPair,
		"rate": rate.String(),
		"amount": amount.String(),
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

func (client *Client) MoveOrder(orderNumber uint64, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	result := struct {
		Success convertibleBool
		PlacedOrder
	}{}

	params := Params{
		"orderNumber": fmt.Sprintf("%d", orderNumber),
		"rate": rate.String(),
	}

	// amount > 0
	if amount.Cmp(decimal.Zero) == 1 {
		params["amount"] = amount.String()
	}

	err = client.tradingApiRequest(&result, "moveOrder", params)

	if !result.Success {
		err = errors.New("Result is not successful")
	}

	placedOrder = result.PlacedOrder

	return
}

type errorResponse struct {
	Error *string
}

func (client *Client) tradingApiRequest(result interface{}, method string, params ...Params) (err error) {
	if len(params) > 1 {
		return errors.New("Too much arguments")
	}

	formData := Params{
		"command": method,
	}

	if len(params) == 1 {
		for name, value := range params[0] {
			formData[name] = value
		}
	}

	client.nonceMutex.Lock()
	nonce := time.Now().UnixNano()
	formData["nonce"] = fmt.Sprintf("%d", nonce)

	request := client.resty.R().
		SetFormData(formData)

	signature := hmac.New(sha512.New, []byte(client.secret))
	signature.Write([]byte(request.FormData.Encode()))

	request.SetHeader("Key", client.key).
		SetHeader("Sign", hex.EncodeToString(signature.Sum(nil)))

	err = client.limiter.Wait(context.TODO())
	if err != nil {
		return
	}

	response, err := request.Post(tradingApiEndpoint)
	client.nonceMutex.Unlock()
	if err != nil {
		return
	}

	errorResponse := errorResponse{}
	json.Unmarshal(response.Body(), &errorResponse)
	if errorResponse.Error != nil {
		return errors.New(*errorResponse.Error)
	}

	err = json.Unmarshal(response.Body(), result)
	return
}