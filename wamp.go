package poloniex

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"gopkg.in/jcelliott/turnpike.v2"
)

const (
	wampEndpoint = "wss://api.poloniex.com"
	wampRealm    = "realm1"

	messageTypeOrderBookRemove = "orderBookRemove"
	messageTypeOrderBookModify = "orderBookModify"
	messageTypeNewTrade        = "newTrade"

	OrderUpdateTypeAsk = "ask"
	OrderUpdateTypeBid = "bid"
)

type OrderModification struct {
	Type     string
	Sequence uint
	Order
}

type NewTrade struct {
	Sequence uint
	Trade
}

type WampClient struct {
	client *turnpike.Client
}

func NewWampClient(tlsConfig *tls.Config, dial turnpike.DialFunc) (*WampClient, error) {
	client, err := turnpike.NewWebsocketClient(turnpike.JSON, wampEndpoint, tlsConfig, dial)
	if err != nil {
		return nil, err
	}

	_, err = client.JoinRealm(wampRealm, nil)
	if err != nil {
		return nil, err
	}

	return &WampClient{client: client}, nil
}

func (wampClient *WampClient) SubscribeToPair(pair string, messageChan chan interface{}, errChan chan error) error {
	if err := wampClient.client.Subscribe(pair, nil, marketMessageHandler(messageChan, errChan, pair)); err != nil {
		return err
	}

	return nil
}

func marketMessageHandler(messageChan chan interface{}, errChan chan error, pair string) turnpike.EventHandler {
	return func(args []interface{}, kwargs map[string]interface{}) {
		sequence, err := parseSequence(kwargs)

		if err != nil {
			errChan <- err
			return
		}

		for _, messageArg := range args {
			message := &marketMessage{Sequence: sequence, Pair: pair}
			if err := mapstructure.Decode(messageArg, &message); err != nil {
				errChan <- err
				continue
			}

			if message.Type == messageTypeNewTrade {
				newTrade, err := message.newTrade()
				if err != nil {
					errChan <- err
					continue
				}

				messageChan <- newTrade
				continue
			}

			orderModification, err := message.orderModification()
			if err != nil {
				errChan <- err
				continue
			}

			messageChan <- orderModification
		}
	}
}

type marketMessage struct {
	Type     string
	Sequence uint
	Pair     string
	Data     marketMessageData
}

type marketMessageData struct {
	TradeId string
	Type    string
	Rate    string
	Amount  string
	Total   string
	Date    string
}

func (message marketMessage) orderModification() (orderModification OrderModification, err error) {
	if message.Type != messageTypeOrderBookModify && message.Type != messageTypeOrderBookRemove {
		return orderModification, fmt.Errorf("can't convert marketMessage with type %s to OrderModification", message.Type)
	}

	orderModification.Type = message.Data.Type
	orderModification.Sequence = message.Sequence

	orderModification.Rate, err = decimal.NewFromString(message.Data.Rate)
	if err != nil {
		return orderModification, fmt.Errorf("marketMessage.orderModification(), rate: %s", err)
	}

	if message.Data.Amount != "" {
		orderModification.Amount, err = decimal.NewFromString(message.Data.Amount)
		if err != nil {
			return orderModification, fmt.Errorf("marketMessage.orderModification(), amount: %s", err)
		}

		orderModification.CalculateTotal()
	}

	return orderModification, nil
}

func (message marketMessage) newTrade() (newTrade NewTrade, err error) {
	newTrade.Sequence = message.Sequence
	newTrade.Trade = Trade{
		Date:         message.Data.Date,
		Type:         message.Data.Type,
		CurrencyPair: message.Pair,
	}

	newTrade.Rate, err = decimal.NewFromString(message.Data.Rate)
	if err != nil {
		return newTrade, fmt.Errorf("marketMessage.newTrade(), rate: %s", err)
	}

	newTrade.Amount, err = decimal.NewFromString(message.Data.Amount)
	if err != nil {
		return newTrade, fmt.Errorf("marketMessage.newTrade(), amount: %s", err)
	}

	newTrade.Total, err = decimal.NewFromString(message.Data.Total)
	if err != nil {
		return newTrade, fmt.Errorf("marketMessage.newTrade(), total: %s", err)
	}

	id, err := strconv.Atoi(message.Data.TradeId)
	if err != nil {
		return newTrade, fmt.Errorf("marketMessage.newTrade(), id: %s", err)
	}
	newTrade.Id = convertibleUint(id)

	return newTrade, nil
}

func parseSequence(kwargs map[string]interface{}) (uint, error) {
	var sequence uint

	sequenceArg, ok := kwargs["seq"]
	if !ok {
		return sequence, fmt.Errorf("key 'seq' was not found in kwargs: %v", kwargs)
	}

	sequenceFloat, ok := sequenceArg.(float64)
	if !ok {
		return sequence, fmt.Errorf("sequence value (%#v) type is %T, expected float64", sequenceArg, sequenceArg)
	}

	sequence = uint(sequenceFloat)

	return sequence, nil
}
