package poloniex

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/shopspring/decimal"
)

const (
	publicApiEndpoint = "https://poloniex.com/public"

	// Order unmarshalling constants
	orderParamsCount = 2
	orderRateIndex   = 0
	orderAmountIndex = 1
)

type Market struct {
	Id            uint32
	Last          decimal.Decimal
	LowestAsk     decimal.Decimal
	HighestBid    decimal.Decimal
	PercentChange decimal.Decimal
	BaseVolume    decimal.Decimal
	QuoteVolume   decimal.Decimal
	IsFrozen      convertibleBool
	High24hr      decimal.Decimal
	Low24hr       decimal.Decimal
}

type Ticker map[string]Market

func (client *Client) Ticker() (ticker Ticker, err error) {
	err = client.publicApiRequest(&ticker, "returnTicker")
	return
}

type Order struct {
	Rate   decimal.Decimal
	Amount decimal.Decimal
	Total  decimal.Decimal
}

func (order *Order) UnmarshalJSON(data []byte) error {
	var orderParams [orderParamsCount]decimal.Decimal

	if err := json.Unmarshal(data, &orderParams); err != nil {
		return err
	}

	order.Rate = orderParams[orderRateIndex]
	order.Amount = orderParams[orderAmountIndex]
	order.CalculateTotal()

	return nil
}

func (order *Order) CalculateTotal() {
	if order.Amount.Equal(decimal.Zero) {
		return
	}

	order.Total = order.Rate.Mul(order.Amount)
}

type OrderBook struct {
	Asks     []Order         `json:"asks"`
	Bids     []Order         `json:"bids"`
	IsFrozen convertibleBool `json:"isFrozen"`
	Sequence uint            `json:"seq"`
}

func (client *Client) OrderBook(currencyPair string) (orderBook OrderBook, err error) {
	err = client.publicApiRequest(&orderBook, "returnOrderBook", Params{
		"currencyPair": currencyPair,
	})

	return orderBook, err
}

func (client *Client) OrderBookAll() (orderBooks map[string]OrderBook, err error) {
	err = client.publicApiRequest(&orderBooks, "returnOrderBook", Params{
		"currencyPair": "all",
	})

	return orderBooks, err
}

func (client *Client) Currencies() (currencies map[string]Currency, err error) {
	err = client.publicApiRequest(&currencies, "returnCurrencies", Params{})

	return currencies, err
}

type Currency struct {
	Id             uint            `json:"id"`
	Name           string          `json:"name"`
	TaxFee         decimal.Decimal `json:"txFee"`
	MinimumConf    uint            `json:"minConf"`
	DepositAddress string          `json:"depositAddress"`
	Disabled       convertibleBool `json:"disabled"`
	Frozen         convertibleBool `json:"frozen"`
	Delisted       convertibleBool `json:"delisted"`
}

func (client *Client) publicApiRequest(result interface{}, method string, params ...Params) error {
	if len(params) > 1 {
		return errors.New("too much arguments")
	}

	queryParams := Params{"command": method}
	if len(params) == 1 {
		for name, value := range params[0] {
			queryParams[name] = value
		}
	}

	err := client.limiter.Wait(context.TODO())
	if err != nil {
		return err
	}

	response, err := client.resty.R().
		SetQueryParams(queryParams).
		Get(publicApiEndpoint)
	if err != nil {
		return err
	}

	errorResponse := errorResponse{}
	err = json.Unmarshal(response.Body(), &errorResponse)
	if err != nil {
		return err
	}

	if errorResponse.Error != nil {
		return errors.New(*errorResponse.Error)
	}

	err = json.Unmarshal(response.Body(), result)
	return err
}
