package common

import "testing"

//"USDT", "USDC", "BUSD", "DAI", "USD"

func TestUSDPrice(t *testing.T) {
	eventAtISO8601 := ISO8601("2021-07-04T14:14:18Z")
	eventAtTimestamp := 1625408058

	tss := []struct {
		name              string
		baseAsset         string
		quoteAsset        string
		eventPrice        JsonFloat64
		eventAt           ISO8601
		candlestickGroups [][]Candlestick
		expectedPrice     JsonFloat64
		expectedErr       bool
	}{
		{
			name:          "Trivial USDT base asset case",
			baseAsset:     "USDT",
			eventPrice:    JsonFloat64(1.0),
			expectedPrice: JsonFloat64(1.0),
		},
		{
			name:          "Trivial USDC base asset case",
			baseAsset:     "USDC",
			eventPrice:    JsonFloat64(1.0),
			expectedPrice: JsonFloat64(1.0),
		},
		{
			name:          "Trivial BUSD base asset case",
			baseAsset:     "BUSD",
			eventPrice:    JsonFloat64(1.0),
			expectedPrice: JsonFloat64(1.0),
		},
		{
			name:          "Trivial DAI base asset case",
			baseAsset:     "DAI",
			eventPrice:    JsonFloat64(1.0),
			expectedPrice: JsonFloat64(1.0),
		},
		{
			name:          "Trivial USD base asset case",
			baseAsset:     "USD",
			eventPrice:    JsonFloat64(1.0),
			expectedPrice: JsonFloat64(1.0),
		},
		{
			name:          "Trivial USDT quote asset case",
			quoteAsset:    "USDT",
			eventPrice:    JsonFloat64(10.0),
			expectedPrice: JsonFloat64(0.1),
		},
		{
			name:          "Trivial USDC quote asset case",
			quoteAsset:    "USDC",
			eventPrice:    JsonFloat64(10.0),
			expectedPrice: JsonFloat64(0.1),
		},
		{
			name:          "Trivial BUSD quote asset case",
			quoteAsset:    "BUSD",
			eventPrice:    JsonFloat64(10.0),
			expectedPrice: JsonFloat64(0.1),
		},
		{
			name:          "Trivial DAI quote asset case",
			quoteAsset:    "DAI",
			eventPrice:    JsonFloat64(10.0),
			expectedPrice: JsonFloat64(0.1),
		},
		{
			name:          "Trivial USD quote asset case",
			quoteAsset:    "USD",
			eventPrice:    JsonFloat64(10.0),
			expectedPrice: JsonFloat64(0.1),
		},
		{
			name:          "Trivial USD quote asset case",
			quoteAsset:    "USD",
			eventPrice:    JsonFloat64(10.0),
			expectedPrice: JsonFloat64(0.1),
		},
		{
			name:       "UNI/BTC to UNI/stable (works on first stablecoin attempt)",
			baseAsset:  "UNI",
			quoteAsset: "BTC",
			eventPrice: JsonFloat64(2.0),
			eventAt:    eventAtISO8601,
			candlestickGroups: [][]Candlestick{
				{
					{Timestamp: eventAtTimestamp, OpenPrice: 4.0},
				},
			},
			expectedPrice: JsonFloat64(0.25),
		},
		{
			name:       "UNI/BTC to UNI/stable (works on 5th stablecoin attempt)",
			baseAsset:  "UNI",
			quoteAsset: "BTC",
			eventPrice: JsonFloat64(2.0),
			eventAt:    eventAtISO8601,
			candlestickGroups: [][]Candlestick{
				{},
				{},
				{},
				{},
				{
					{Timestamp: eventAtTimestamp, OpenPrice: 4.0},
				},
			},
			expectedPrice: JsonFloat64(0.25),
		},
		{
			name:       "UNI/BTC to UNI/stable (works on 2nd stablecoin attempt), second candlestick (first < date)",
			baseAsset:  "UNI",
			quoteAsset: "BTC",
			eventPrice: JsonFloat64(2.0),
			eventAt:    eventAtISO8601,
			candlestickGroups: [][]Candlestick{
				{},
				{},
				{
					{Timestamp: eventAtTimestamp - 1, OpenPrice: 10.0},
					{Timestamp: eventAtTimestamp, OpenPrice: 4.0},
				},
			},
			expectedPrice: JsonFloat64(0.25),
		},
		{
			name:       "BAKE/BNB, failed against stable, transitive to BNB/BUSD",
			baseAsset:  "BAKE",
			quoteAsset: "BNB",
			eventPrice: JsonFloat64(2.0),
			eventAt:    eventAtISO8601,
			candlestickGroups: [][]Candlestick{
				{},
				{},
				{},
				{},
				{}, // Last stablecoin
				{}, // First transitive
				{ // Second transitive
					{Timestamp: eventAtTimestamp - 1, OpenPrice: 10.0},
					{Timestamp: eventAtTimestamp, OpenPrice: 4.0},
				},
				{ // transitive to stablecoin
					{Timestamp: eventAtTimestamp - 1, OpenPrice: 10.0},
					{Timestamp: eventAtTimestamp, OpenPrice: 5.0},
				},
			},
			expectedPrice: JsonFloat64(1 / (4.0 * 5.0)),
		},
	}

	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			exchange := NewTestExchange(ts.candlestickGroups)
			input := SignalCheckInput{BaseAsset: ts.baseAsset, QuoteAsset: ts.quoteAsset, Debug: true}
			event := SignalCheckOutputEvent{EventType: ENTERED, At: ts.eventAt, Price: ts.eventPrice}
			actualPrice, actualErr := GetUSDPricePerBaseAssetUnitAtEvent(exchange, input, event)
			if actualErr != nil && !ts.expectedErr {
				t.Fatalf("Expected no error, but failed with %v", actualErr)
			}
			if actualErr == nil && ts.expectedErr {
				t.Fatal("Expected to error, but it didn't")
			}
			if actualPrice != ts.expectedPrice {
				t.Fatalf("Expected price %v but was %v", ts.expectedPrice, actualPrice)
			}
		})
	}
}

type testExchange struct {
	candlestickI          int
	mockCandlestickGroups [][]Candlestick
}

func NewTestExchange(mockCandlestickGroups [][]Candlestick) *testExchange {
	return &testExchange{mockCandlestickGroups: mockCandlestickGroups, candlestickI: -1}
}

func (t *testExchange) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 ISO8601) *CandlestickIterator {
	t.candlestickI++
	testCandlestickIterator := func(cs []Candlestick) func() (Candlestick, error) {
		i := 0
		return func() (Candlestick, error) {
			if i >= len(cs) {
				return Candlestick{}, ErrOutOfCandlesticks
			}
			i++
			return cs[i-1], nil
		}
	}
	return NewCandlestickIterator(testCandlestickIterator(t.mockCandlestickGroups[t.candlestickI]))
}
func (t testExchange) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 ISO8601) *TradeIterator {
	return nil
}
func (t testExchange) SetDebug(debug bool) {}
