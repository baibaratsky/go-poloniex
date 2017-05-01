package poloniex

import (
	"github.com/shopspring/decimal"
	"gopkg.in/resty.v0"
	"encoding/json"
	"errors"
)

const publicApiEndpoint = "https://poloniex.com/public"

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

func (client *Client) publicApiRequest(result interface{}, method string, params ...Params) (err error) {
	if len(params) > 1 {
		return errors.New("Too much arguments")
	}

	queryParams := Params{"command": method}
	if len(params) == 1 {
		for name, value := range params[0] {
			queryParams[name] = value
		}
	}

	response, err := resty.R().
		SetQueryParams(queryParams).
		Get(publicApiEndpoint)
	if err != nil {
		return
	}

	errorResponse := errorResponse{}
	err = json.Unmarshal(response.Body(), &errorResponse)
	if err != nil {
		return
	}

	if errorResponse.Error != nil {
		return errors.New(*errorResponse.Error)
	}

	err = json.Unmarshal(response.Body(), result)
	return
}