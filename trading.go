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

func (client *Client) DepositAddresses() (addresses map[string]string, err error) {
	err = client.tradingApiRequest(&addresses, "returnDepositAddresses")
	return
}

func (client *Client) NewAddress(currency string) (address string, err error) {
	params := Params{
		"currency": currency,
	}

	result := struct {
		Success  convertibleBool `json:"success"`
		Response string          `json:"response"`
	}{}

	if err = client.tradingApiRequest(&result, "generateNewAddress", params); err != nil {
		return result.Response, err
	}

	if !result.Success {
		return result.Response, fmt.Errorf("generateNewAddress for currency %s success = %v, response = %v", currency, result.Success, result.Response)
	}

	return result.Response, err
}

type Trade struct {
	GlobalTradeId uint64          `json:"globalTradeID"`
	Id            convertibleUint `json:"tradeID"`
	OrderNumber   uint64          `json:"orderNumber,string"`
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
	OrderNumber     convertibleUint `json:"orderNumber"`
	ResultingTrades []Trade
}

type UpdatedOrder struct {
	OrderNumber     convertibleUint `json:"orderNumber"`
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

// BuyFOK creates buy method with "Fill or Kill" option enabled
func (client *Client) BuyFOK(currencyPair string, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	err = client.tradingApiRequest(&placedOrder, "buy", Params{
		"currencyPair": currencyPair,
		"rate":         rate.String(),
		"amount":       amount.String(),
		"fillOrKill":   "1",
	})
	return
}

// SellFOK creates sell method with "Fill or Kill" option enabled
func (client *Client) SellFOK(currencyPair string, rate, amount decimal.Decimal) (placedOrder PlacedOrder, err error) {
	err = client.tradingApiRequest(&placedOrder, "sell", Params{
		"currencyPair": currencyPair,
		"rate":         rate.String(),
		"amount":       amount.String(),
		"fillOrKill":   "1",
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

	if amount.GreaterThan(decimal.Zero) {
		params["amount"] = amount.String()
	}

	err = client.tradingApiRequest(&result, "moveOrder", params)

	if !result.Success {
		err = errors.New("result is not successful")
	}

	updatedOrder = result.UpdatedOrder

	return
}

func (client *Client) Withdraw(currency, address string, amount decimal.Decimal) (response string, err error) {
	result := struct {
		Response string
	}{}

	params := Params{
		"currency": currency,
		"address":  address,
		"amount":   amount.String(),
	}

	err = client.tradingApiRequest(&result, "withdraw", params)
	return result.Response, err
}

type DepositsWithdrawalsResponse struct {
	Deposits []*struct {
		DepositNumber uint   `json:"depositNumber"`
		Currency      string `json:"currency"`
		Address       string `json:"address"`
		Amount        string `json:"amount"`
		Confirmations uint   `json:"confirmations"`
		Txid          string `json:"txid"`
		Timestamp     uint   `json:"timestamp"`
		Status        string `json:"status"`
	} `json:"deposits"`
	Withdrawals []*struct {
		WithdrawalNumber uint    `json:"withdrawalNumber"`
		Currency         string  `json:"currency"`
		Address          string  `json:"address"`
		Amount           string  `json:"amount"`
		Fee              string  `json:"fee"`
		Timestamp        uint    `json:"timestamp"`
		Status           string  `json:"status"`
		IPAddress        string  `json:"ipAddress"`
		PaymentID        *string `json:"paymentID"`
	} `json:"withdrawals"`
}

func (client *Client) DepositsWithdrawals(start, end int64) (response DepositsWithdrawalsResponse, err error) {
	params := Params{}
	if start > 0 {
		params["start"] = fmt.Sprintf("%d", start)
	}
	if end > 0 {
		params["end"] = fmt.Sprintf("%d", end)
	}

	err = client.tradingApiRequest(&response, "returnDepositsWithdrawals", params)
	return
}

type errorResponse struct {
	Error *string
}

type emptyArrayResponse []struct{}

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
		return err
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
		return err
	}

	errorResponse := errorResponse{}
	json.Unmarshal(response.Body(), &errorResponse)
	if errorResponse.Error != nil {
		return errors.New(*errorResponse.Error)
	}

	emptyArrayResponse := emptyArrayResponse{}
	err = json.Unmarshal(response.Body(), &emptyArrayResponse)
	if err == nil && len(emptyArrayResponse) == 0 {
		return nil
	}

	err = json.Unmarshal(response.Body(), result)
	if err != nil {
		err = fmt.Errorf("%s\nServer response: %s", err.Error(), string(response.Body()))
	}
	return err
}
