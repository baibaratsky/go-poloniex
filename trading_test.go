package poloniex

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
	resty "gopkg.in/resty.v0"
)

func TestClient_FeeInfo(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"makerFee": "0.00140000", "takerFee": "0.00240000", "thirtyDayVolume": "612.00248891", "nextTier": "1200.00000000"}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return fee info", func() {
			fee, err := client.FeeInfo()
			So(err, ShouldBeNil)

			So(fee.MakerFee.String(), ShouldEqual, "0.0014")
		})
	})
}

func TestClient_Balances(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC":"0.59098578","LTC":"3.31117268"}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return balances", func() {
			balances, err := client.Balances()
			So(err, ShouldBeNil)

			So(len(balances), ShouldEqual, 2)
			So(balances["BTC"].Equal(decimal.New(59098578, -8)), ShouldBeTrue)
		})
	})
}

func TestClient_DepositAddresses(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC":"19YqztHmspv2egyD6jQM3yn81x5t5krVdJ","LTC":"LPgf9kjv9H1Vuh4XSaKhzBe8JHdou1WgUB"}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return deposit_addresses", func() {
			addresses, err := client.DepositAddresses()
			So(err, ShouldBeNil)

			So(len(addresses), ShouldEqual, 2)
			So(addresses["BTC"], ShouldEqual, "19YqztHmspv2egyD6jQM3yn81x5t5krVdJ")
		})
	})
}

func TestClient_NewAddress(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"success":1,"response":"CKXbbs8FAVbtEa397gJHSutmrdrBrhUMxe"}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return deposit_addresses", func() {
			address, err := client.NewAddress("BTC")
			So(err, ShouldBeNil)

			So(address, ShouldEqual, "CKXbbs8FAVbtEa397gJHSutmrdrBrhUMxe")
		})

		Convey("Should return error on tradingApiRequest", func() {
			handler.HandleFunc = func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `///`)
			}
			_, err := client.NewAddress("BTC")
			So(err, ShouldBeError)
		})

		Convey("Should return error on failured response", func() {
			handler.HandleFunc = func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"success":0,"response":"some error"}`)
			}
			_, err := client.NewAddress("BTC")
			So(err, ShouldBeError)
		})
	})
}

func TestClient_TradeHistory(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[{ "globalTradeID": 25129732, "tradeID": "6325758", "date": "2016-04-05 08:08:40", "rate": "0.02565498", "amount": "0.10000000", "total": "0.00256549", "fee": "0.00200000", "orderNumber": "34225313575", "type": "sell", "category": "exchange" }, { "globalTradeID": 25129628, "tradeID": "6325741", "date": "2016-04-05 08:07:55", "rate": "0.02565499", "amount": "0.10000000", "total": "0.00256549", "fee": "0.00200000", "orderNumber": "34225195693", "type": "buy", "category": "exchange" }]`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return trade history", func() {
			tradeHistory, err := client.TradeHistory("ANY", 1, 2)
			So(err, ShouldBeNil)

			So(len(tradeHistory), ShouldEqual, 2)
			So(tradeHistory[0].Rate.String(), ShouldEqual, "0.02565498")
			So(tradeHistory[0].CurrencyPair, ShouldEqual, "ANY")
		})
	})
}
func TestClient_TradeHistoryAll(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC_MAID": [ { "globalTradeID": 29251512, "tradeID": "1385888", "date": "2016-05-03 01:29:55", "rate": "0.00014243", "amount": "353.74692925", "total": "0.05038417", "fee": "0.00200000", "orderNumber": "12603322113", "type": "buy", "category": "settlement" }, { "globalTradeID": 29251511, "tradeID": "1385887", "date": "2016-05-03 01:29:55", "rate": "0.00014111", "amount": "311.24262497", "total": "0.04391944", "fee": "0.00200000", "orderNumber": "12603319116", "type": "sell", "category": "marginTrade" }]}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return all trade history", func() {
			tradeHistoryMap, err := client.TradeHistoryAll(1, 2)
			So(err, ShouldBeNil)

			So(len(tradeHistoryMap), ShouldEqual, 1)
			So(len(tradeHistoryMap["BTC_MAID"]), ShouldEqual, 2)
			So(tradeHistoryMap["BTC_MAID"][0].Rate.String(), ShouldEqual, "0.00014243")
			So(tradeHistoryMap["BTC_MAID"][0].CurrencyPair, ShouldEqual, "BTC_MAID")
		})
	})
}

func TestClient_OrderTrades(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[{"globalTradeID": 20825863, "tradeID": 147142, "currencyPair": "BTC_XVC", "type": "buy", "rate": "0.00018500", "amount": "455.34206390", "total": "0.08423828", "fee": "0.00200000", "date": "2016-03-14 01:04:36"}]`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return order trades", func() {
			orderTrades, err := client.OrderTrades(0)
			So(err, ShouldBeNil)

			So(len(orderTrades), ShouldEqual, 1)
			So(orderTrades[0].Rate.String(), ShouldEqual, "0.000185")
		})
	})
}

func TestClient_OpenOrders(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[{"orderNumber":"120466","type":"sell","rate":"0.025","amount":"100","total":"2.5"},{"orderNumber":"120467","type":"sell","rate":"0.04","amount":"100","total":"4"}]`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return open orders", func() {
			orders, err := client.OpenOrders("ALT1")
			So(err, ShouldBeNil)

			So(len(orders), ShouldEqual, 2)
			So(orders[0].Rate.String(), ShouldEqual, "0.025")
		})
	})
}

func TestClient_OpenOrdersAll(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"BTC_1CR":[],"BTC_AC":[{"orderNumber":"120466","type":"sell","rate":"0.025","amount":"100","total":"2.5"},{"orderNumber":"120467","type":"sell","rate":"0.04","amount":"100","total":"4"}]}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should return all open orders", func() {
			ordersMap, err := client.OpenOrdersAll()
			So(err, ShouldBeNil)

			So(len(ordersMap), ShouldEqual, 2)
			So(ordersMap["BTC_AC"][0].Rate.String(), ShouldEqual, "0.025")
		})
	})
}

func TestClient_Buy(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"orderNumber":31226040,"resultingTrades":[{"amount":"338.8732","date":"2014-10-18 23:03:21","rate":"0.00000173","total":"0.00058625","tradeID":"16164","type":"buy"}]}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should buy", func() {
			placedOrder, err := client.Buy("ANY", decimal.Zero, decimal.Zero)
			So(err, ShouldBeNil)

			So(placedOrder.ResultingTrades[0].Rate.String(), ShouldEqual, "0.00000173")
		})
	})
}

func TestClient_Sell(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"orderNumber":31226040,"resultingTrades":[{"amount":"338.8732","date":"2014-10-18 23:03:21","rate":"0.00000173","total":"0.00058625","tradeID":"16164","type":"sell"}]}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should sell", func() {
			placedOrder, err := client.Sell("ANY", decimal.Zero, decimal.Zero)
			So(err, ShouldBeNil)

			So(placedOrder.ResultingTrades[0].Rate.String(), ShouldEqual, "0.00000173")
		})
	})
}

func TestClient_CancellOrder(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"success":1}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should sell", func() {
			success, err := client.CancelOrder(0)
			So(err, ShouldBeNil)

			So(success, ShouldBeTrue)
		})
	})
}

func TestClient_MoveOrder(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"success":1,"orderNumber":"239574176","resultingTrades":{"BTC_BTS":[{"amount":"338.8732","date":"2014-10-18 23:03:21","rate":"0.00000173","total":"0.00058625","tradeID":"16164","type":"buy"}]}}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))

		Convey("Should sell", func() {
			updatedOrder, err := client.MoveOrder(0, decimal.Zero, decimal.New(1, 0))
			So(err, ShouldBeNil)

			So(len(updatedOrder.ResultingTrades), ShouldEqual, 1)
		})

		Convey("Should return error", func() {
			handler.HandleFunc = func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"success":0,"orderNumber":"239574176","resultingTrades":{"BTC_BTS":[{"amount":"338.8732","date":"2014-10-18 23:03:21","rate":"0.00000173","total":"0.00058625","tradeID":"16164","type":"buy"}]}}`)
			}
			_, err := client.MoveOrder(0, decimal.Zero, decimal.New(1, 0))
			So(err, ShouldNotBeEmpty)
		})
	})
}

func TestClient_Withdraw(t *testing.T) {
	Convey("Setup correct server", t, func() {
		handler := &fakeHandler{
			HandleFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"response":"Withdrew 10 BTC."}`)
			},
		}
		server := createFakeServer(handler)
		defer server.Close()

		client := NewClient([]Key{NewKey("key", "secret")})
		client.SetTransport(transportForTesting(server))
		response, err := client.Withdraw("BTC", "xyz", decimal.New(1, 0))
		So(err, ShouldBeNil)

		So(response, ShouldEqual, "Withdrew 10 BTC.")
	})
}

func TestClient_tradingApiRequest(t *testing.T) {
	type fields struct {
		resty      *resty.Client
		limiter    *rate.Limiter
		handleFunc http.HandlerFunc
	}
	type args struct {
		result interface{}
		method string
		params []Params
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "too much arguments",
			fields: fields{
				resty:      resty.DefaultClient,
				limiter:    nil,
				handleFunc: func(w http.ResponseWriter, r *http.Request) {},
			},
			args: args{
				result: make(map[string]string),
				method: "any",
				params: []Params{
					Params{"first": "first"},
					Params{"second": "second"},
				},
			},
			wantErr: true,
		},
		{
			name: "limiter error",
			fields: fields{
				resty:      resty.DefaultClient,
				limiter:    rate.NewLimiter(1, 0),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {},
			},
			args: args{
				result: make(map[string]string),
				method: "any",
				params: []Params{
					Params{"first": "first"},
				},
			},
			wantErr: true,
		},
		{
			name: "unmarshal error",
			fields: fields{
				resty:   resty.DefaultClient,
				limiter: rate.NewLimiter(1, 1),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, `///`)
				},
			},
			args: args{
				result: make(map[string]string),
				method: "POST",
				params: []Params{},
			},
			wantErr: true,
		},
		{
			name: "empty array",
			fields: fields{
				resty:   resty.DefaultClient,
				limiter: rate.NewLimiter(1, 1),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, `[]`)
				},
			},
			args: args{
				result: make(map[string]string),
				method: "POST",
				params: []Params{},
			},
			wantErr: false,
		},
		{
			name: "response error",
			fields: fields{
				resty:   resty.DefaultClient,
				limiter: rate.NewLimiter(1, 1),
				handleFunc: func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, `{"error": "some"}`)
				},
			},
			args: args{
				result: make(map[string]string),
				method: "ERROR",
				params: []Params{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &fakeHandler{
				HandleFunc: tt.fields.handleFunc,
			}
			server := createFakeServer(handler)
			defer server.Close()
			client := &Client{
				keyPool: keyPool{
					keys: make(chan *Key, 1),
				},
				resty:   tt.fields.resty,
				limiter: tt.fields.limiter,
			}
			key := NewKey("KEY", "SECRET")
			client.keyPool.Put(&key)
			transport := transportForTesting(server)
			client.SetTransport(transport)
			if err := client.tradingApiRequest(tt.args.result, tt.args.method, tt.args.params...); (err != nil) != tt.wantErr {
				t.Errorf("Client.tradingApiRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
