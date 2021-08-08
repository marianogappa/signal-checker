package binanceusdmfutures

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

type expectedTrade struct {
	trade common.Trade
	err   error
}

func TestTrades(t *testing.T) {
	i := 0
	replies := []string{
		`[
			{"a":850187608,"p":"29781.76000000","q":"0.00055700","f":961525512,"l":961525512,"T":1626798722486,"m":true,"M":true},
			{"a":850187609,"p":"29781.77000000","q":"0.00380000","f":961525513,"l":961525513,"T":1626798723004,"m":false,"M":true}
		]`,
		`[
			{"a":850187610,"p":"29781.77000000","q":"0.00100600","f":961525514,"l":961525514,"T":1626798723257,"m":false,"M":true},
			{"a":850187611,"p":"29781.77000000","q":"0.00619400","f":961525515,"l":961525515,"T":1626798723257,"m":false,"M":true}
		]`,
		`[]`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewBinanceUSDMFutures()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildTradeIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")

	expectedResults := []expectedTrade{
		{
			trade: common.Trade{
				BaseAssetPrice:    29781.76000000,
				BaseAssetQuantity: 0.00055700,
				Timestamp:         1626798722,
			},
			err: nil,
		},
		{
			trade: common.Trade{
				BaseAssetPrice:    29781.77000000,
				BaseAssetQuantity: 0.00380000,
				Timestamp:         1626798723,
			},
			err: nil,
		},
		{
			trade: common.Trade{
				BaseAssetPrice:    29781.77000000,
				BaseAssetQuantity: 0.00100600,
				Timestamp:         1626798723,
			},
			err: nil,
		},
		{
			trade: common.Trade{
				BaseAssetPrice:    29781.77000000,
				BaseAssetQuantity: 0.00619400,
				Timestamp:         1626798723,
			},
			err: nil,
		},
		{
			trade: common.Trade{},
			err:   common.ErrOutOfTrades,
		},
	}
	for i, expectedResult := range expectedResults {
		actualTrade, actualErr := ci.Next()
		if actualTrade != expectedResult.trade {
			t.Errorf("on trade %v expected %v but got %v", i, expectedResult.trade, actualTrade)
			t.FailNow()
		}
		if actualErr != expectedResult.err {
			t.Errorf("on trade %v expected no errors but this error happened %v", i, actualErr)
			t.FailNow()
		}
	}
}

func TestBinanceTradeToTradeFailsPrice(t *testing.T) {
	_, err := binanceTrade{
		AggregateTradeId:      12345,
		Price:                 "invalid",
		Quantity:              "123",
		FirstTradeId:          12345,
		LastTradeId:           12345,
		TimestampMillis:       1626798723000,
		IsBuyerMaker:          false,
		IsTradeBestPriceMatch: true,
	}.toTrade()
	if err == nil {
		t.Fatalf("should have failed with invalid price")
	}
}

func TestBinanceTradeToTradeFailsQuantity(t *testing.T) {
	_, err := binanceTrade{
		AggregateTradeId:      12345,
		Price:                 "0.1",
		Quantity:              "invalid",
		FirstTradeId:          12345,
		LastTradeId:           12345,
		TimestampMillis:       1626798723000,
		IsBuyerMaker:          false,
		IsTradeBestPriceMatch: true,
	}.toTrade()
	if err == nil {
		t.Fatalf("should have failed with invalid quantity")
	}
}

func TestBinanceTradesToTradeFails(t *testing.T) {
	_, err := binanceTradesToTrades([]binanceTrade{{
		AggregateTradeId:      12345,
		Price:                 "0.1",
		Quantity:              "invalid",
		FirstTradeId:          12345,
		LastTradeId:           12345,
		TimestampMillis:       1626798723000,
		IsBuyerMaker:          false,
		IsTradeBestPriceMatch: true,
	}})
	if err == nil {
		t.Fatalf("should have failed with invalid quantity")
	}
}

func TestInvalidUrl(t *testing.T) {
	i := 0
	replies := []string{
		`[
			{"a":850187608,"p":"29781.76000000","q":"0.00055700","f":961525512,"l":961525512,"T":1626798722486,"m":true,"M":true},
			{"a":850187609,"p":"29781.77000000","q":"0.00380000","f":961525513,"l":961525513,"T":1626798723004,"m":false,"M":true}
		]`,
		`[
			{"a":850187610,"p":"29781.77000000","q":"0.00100600","f":961525514,"l":961525514,"T":1626798723257,"m":false,"M":true},
			{"a":850187611,"p":"29781.77000000","q":"0.00619400","f":961525515,"l":961525515,"T":1626798723257,"m":false,"M":true}
		]`,
		`[]`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, replies[i%len(replies)])
		i++
	}))
	defer ts.Close()

	b := NewBinanceUSDMFutures()
	b.overrideAPIURL("invalid url")
	ci := b.BuildTradeIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")
	_, err := ci.Next()
	if err == nil {
		t.Fatalf("should have failed due to invalid url")
	}
}

func TestErrReadingResponseBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	defer ts.Close()

	b := NewBinanceUSDMFutures()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildTradeIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")
	_, err := ci.Next()
	if err == nil {
		t.Fatalf("should have failed due to invalid response body")
	}
}

func TestErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"code":-1100,"msg":"Illegal characters found in parameter 'symbol'; legal range is '^[A-Z0-9-_.]{1,20}$'."}`)
	}))
	defer ts.Close()

	b := NewBinanceUSDMFutures()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildTradeIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")
	_, err := ci.Next()
	if err == nil {
		t.Fatalf("should have failed due to error response")
	}
}
func TestInvalidJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid json`)
	}))
	defer ts.Close()

	b := NewBinanceUSDMFutures()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildTradeIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")
	_, err := ci.Next()
	if err == nil {
		t.Fatalf("should have failed due to invalid json")
	}
}

func TestInvalidFloatsInJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[
			{"a":850187608,"p":"invalid float","q":"0.00055700","f":961525512,"l":961525512,"T":1626798722486,"m":true,"M":true}
		]`)
	}))
	defer ts.Close()

	b := NewBinanceUSDMFutures()
	b.overrideAPIURL(ts.URL + "/")
	ci := b.BuildTradeIterator("BTC", "USDT", "2021-07-04T14:14:18+00:00")
	_, err := ci.Next()
	if err == nil {
		t.Fatalf("should have failed due to invalid floats in json")
	}
}
