package poloniex

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/jcelliott/turnpike.v2"
)

func TestNewWampClient(t *testing.T) {
	Convey("Given fake WAMP server", t, func() {
		wampServer, httpServer, clear := newTestWebsocketServer(t)
		defer clear()

		localClient, err := wampServer.GetLocalClient(wampRealm, nil)
		if err != nil {
			t.Fatalf("error getting local client: %s", err)
		}

		client, err := NewWampClient(&tls.Config{InsecureSkipVerify: true}, func(network, addr string) (net.Conn, error) {
			So(network, ShouldEqual, "tcp")
			So(addr, ShouldEqual, "api.poloniex.com:443")
			url, err := url.Parse(httpServer.URL)
			if err != nil {
				t.Fatal(err)
			}
			return net.Dial("tcp", fmt.Sprintf("localhost:%s", url.Port()))
		})

		if err != nil {
			t.Fatalf("error creating client: %s", err)
		}

		messageChan := make(chan interface{})
		errChan := make(chan error)

		if err := client.SubscribeToPair("BTC_ETH", messageChan, errChan); err != nil {
			t.Fatalf("error subscribing to chat channel: %s", err)
		}

		Convey("It should recieve modify message", func() {
			modifyMarketMessage := map[string]interface{}{
				"type": messageTypeOrderBookModify,
				"data": map[string]interface{}{
					"type":   "bid",
					"rate":   "0.08529432",
					"amount": "5.00000000",
				},
			}

			args := []interface{}{interface{}(modifyMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("error publising to pair: %s", err)
			}

			message := <-messageChan
			orderModification, ok := message.(OrderModification)

			So(ok, ShouldBeTrue)
			So(orderModification.Type, ShouldEqual, OrderUpdateTypeBid)
		})

		Convey("It should recieve remove message", func() {
			modifyMarketMessage := map[string]interface{}{
				"type": messageTypeOrderBookRemove,
				"data": map[string]interface{}{
					"type": "ask",
					"rate": "0.08529432",
				},
			}

			args := []interface{}{interface{}(modifyMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("error publising to pair: %s", err)
			}

			message := <-messageChan
			orderRemoval, ok := message.(OrderModification)

			So(ok, ShouldBeTrue)
			So(orderRemoval.Type, ShouldEqual, OrderUpdateTypeAsk)
		})

		Convey("It should recieve trade message", func() {
			newTradeMarketMessage := map[string]interface{}{
				"type": messageTypeNewTrade,
				"data": map[string]interface{}{
					"total":   "0.01190544",
					"tradeID": "30132092",
					"type":    "sell",
					"amount":  "0.14080956",
					"date":    "2017-07-13 17:33:31",
					"rate":    "0.08455001",
				},
			}

			args := []interface{}{interface{}(newTradeMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("error publising to pair: %s", err)
			}

			message := <-messageChan
			newTrade, ok := message.(NewTrade)

			So(ok, ShouldBeTrue)
			So(newTrade.Type, ShouldEqual, TypeSell)
		})

		Convey("It should recieve errors on sequence", func() {
			args := []interface{}{}
			kwargs := map[string]interface{}{}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("error publising to pair: %s", err)
			}

			err := <-errChan

			So(err.Error(), ShouldEqual, `key 'seq' was not found in kwargs: map[]`)

			kwargs = map[string]interface{}{"seq": ""}
			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("error publising to pair: %s", err)
			}

			err = <-errChan

			So(err.Error(), ShouldEqual, `sequence value ("") type is string, expected float64`)
		})

		Convey("It should recieve error on structure decode", func() {
			badMarketMessage := map[string]interface{}{
				"type": float64(1),
			}

			args := []interface{}{interface{}(badMarketMessage)}
			kwargs := map[string]interface{}{"seq": float64(1)}

			if err := localClient.Publish("BTC_ETH", nil, args, kwargs); err != nil {
				t.Fatalf("error publising to pair: %s", err)
			}

			err := <-errChan

			So(err.Error(), ShouldContainSubstring, "1 error(s) decoding")
		})
	})
}

func newTestWebsocketServer(t *testing.T) (*turnpike.WebsocketServer, *httptest.Server, func()) {
	wampServer := turnpike.NewBasicWebsocketServer(wampRealm)
	httpServer := httptest.NewTLSServer(wampServer)
	return wampServer, httpServer, func() {
		httpServer.Close()
		wampServer.Close()
	}
}

func Test_marketMessage_newTrade(t *testing.T) {
	type fields struct {
		Sequence uint
		Pair     string
		Data     marketMessageData
	}
	tests := []struct {
		name         string
		fields       fields
		wantNewTrade NewTrade
		wantErr      bool
	}{
		{
			name: "rate parse error",
			fields: fields{
				Sequence: 1,
				Pair:     "ETH_BTC",
				Data: marketMessageData{
					Date: "30.08.2017",
					Type: TypeSell,
					Rate: "invalid",
				},
			},
			wantNewTrade: NewTrade{
				Sequence: 1,
				Trade: Trade{
					CurrencyPair: "ETH_BTC",
					Date:         "30.08.2017",
					Type:         TypeSell,
				},
			},
			wantErr: true,
		},
		{
			name: "amount parse error",
			fields: fields{
				Sequence: 1,
				Pair:     "ETH_BTC",
				Data: marketMessageData{
					Date:   "30.08.2017",
					Type:   TypeSell,
					Rate:   "0.1",
					Amount: "invalid",
				},
			},
			wantNewTrade: NewTrade{
				Sequence: 1,
				Trade: Trade{
					CurrencyPair: "ETH_BTC",
					Date:         "30.08.2017",
					Type:         TypeSell,
					Rate:         decimal.New(1, -1),
				},
			},
			wantErr: true,
		},
		{
			name: "total parse error",
			fields: fields{
				Sequence: 1,
				Pair:     "ETH_BTC",
				Data: marketMessageData{
					Date:   "30.08.2017",
					Type:   TypeSell,
					Rate:   "0.1",
					Amount: "0.01",
					Total:  "invalid",
				},
			},
			wantNewTrade: NewTrade{
				Sequence: 1,
				Trade: Trade{
					CurrencyPair: "ETH_BTC",
					Date:         "30.08.2017",
					Type:         TypeSell,
					Rate:         decimal.New(1, -1),
					Amount:       decimal.New(1, -2),
				},
			},
			wantErr: true,
		},
		{
			name: "tradeId parse error",
			fields: fields{
				Sequence: 1,
				Pair:     "ETH_BTC",
				Data: marketMessageData{
					Date:    "30.08.2017",
					Type:    TypeSell,
					Rate:    "0.1",
					Amount:  "0.01",
					Total:   "0.001",
					TradeId: "invalid",
				},
			},
			wantNewTrade: NewTrade{
				Sequence: 1,
				Trade: Trade{
					CurrencyPair: "ETH_BTC",
					Date:         "30.08.2017",
					Type:         TypeSell,
					Rate:         decimal.New(1, -1),
					Amount:       decimal.New(1, -2),
					Total:        decimal.New(1, -3),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := marketMessage{
				Sequence: tt.fields.Sequence,
				Pair:     tt.fields.Pair,
				Data:     tt.fields.Data,
			}
			gotNewTrade, err := message.newTrade()
			if (err != nil) != tt.wantErr {
				t.Errorf("marketMessage.newTrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNewTrade, tt.wantNewTrade) {
				t.Errorf("marketMessage.newTrade() = %v, want %v", gotNewTrade, tt.wantNewTrade)
			}
		})
	}
}

func Test_marketMessage_orderModification(t *testing.T) {
	type fields struct {
		Type     string
		Sequence uint
		Pair     string
		Data     marketMessageData
	}
	tests := []struct {
		name                  string
		fields                fields
		wantOrderModification OrderModification
		wantErr               bool
	}{
		{
			name: "wrong type",
			fields: fields{
				Type: "WRONG TYPE",
			},
			wantOrderModification: OrderModification{},
			wantErr:               true,
		},
		{
			name: "invalid rate",
			fields: fields{
				Type:     messageTypeOrderBookModify,
				Sequence: 1,
				Pair:     "ETH_BTC",
				Data: marketMessageData{
					Date: "30.08.2017",
					Type: TypeSell,
					Rate: "invalid",
				},
			},
			wantOrderModification: OrderModification{
				Sequence: 1,
				Type:     TypeSell,
			},
			wantErr: true,
		},
		{
			name: "invalid amount",
			fields: fields{
				Type:     messageTypeOrderBookModify,
				Sequence: 1,
				Pair:     "ETH_BTC",
				Data: marketMessageData{
					Date:   "30.08.2017",
					Type:   TypeSell,
					Rate:   "0.1",
					Amount: "invalid",
				},
			},
			wantOrderModification: OrderModification{
				Sequence: 1,
				Type:     TypeSell,
				Order: Order{
					Rate: decimal.New(1, -1),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := marketMessage{
				Type:     tt.fields.Type,
				Sequence: tt.fields.Sequence,
				Pair:     tt.fields.Pair,
				Data:     tt.fields.Data,
			}
			gotOrderModification, err := message.orderModification()
			if (err != nil) != tt.wantErr {
				t.Errorf("marketMessage.orderModification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOrderModification, tt.wantOrderModification) {
				t.Errorf("marketMessage.orderModification() = %v, want %v", gotOrderModification, tt.wantOrderModification)
			}
		})
	}
}
