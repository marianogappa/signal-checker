package common

import (
	"testing"
)

func TestGetMaxBaseAssetEnter(t *testing.T) {
	type test struct {
		name           string
		trades         []Trade
		expectedErr    error
		expectedTrade  Trade
		minuteCount    int
		bucketCount    int
		maxTotalTrades int
	}

	startISO8601, _ := ISO8601("2021-07-04T14:14:18Z").Seconds()
	anHourLaterISO8601, _ := ISO8601("2021-07-04T15:14:19Z").Seconds()

	tss := []test{
		{
			name:           "Takes the q in the 99%",
			minuteCount:    5,
			bucketCount:    10,
			maxTotalTrades: 10,
			trades: []Trade{
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(1.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(2.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(3.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(4.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(5.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(6.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(7.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(8.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(9.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0)},
			},
			expectedErr:   nil,
			expectedTrade: Trade{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10)},
		},
		{
			name:           "Same result because it ignores a trade outside maxTotalTrades",
			minuteCount:    5,
			bucketCount:    10,
			maxTotalTrades: 10,
			trades: []Trade{
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(1.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(2.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(3.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(4.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(5.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(6.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(7.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(8.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(9.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(100000000000.0)},
			},
			expectedErr:   nil,
			expectedTrade: Trade{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0)},
		},
		{
			name:           "Same result because even inside maxTotalTrades, exceeding 5 minute count",
			minuteCount:    5,
			bucketCount:    10,
			maxTotalTrades: 11,
			trades: []Trade{
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(1.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(2.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(3.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(4.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(5.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(6.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(7.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(8.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(9.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0), Timestamp: startISO8601},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(100000000000.0), Timestamp: anHourLaterISO8601},
			},
			expectedErr:   nil,
			expectedTrade: Trade{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0), Timestamp: startISO8601},
		},
		{
			name:           "Does not fail when running out of trades",
			minuteCount:    5,
			bucketCount:    10,
			maxTotalTrades: 100,
			trades: []Trade{
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(1.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(2.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(3.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(4.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(5.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(6.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(7.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(8.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(9.0)},
				{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0)},
			},
			expectedErr:   nil,
			expectedTrade: Trade{BaseAssetPrice: f(1.0), BaseAssetQuantity: f(10.0)},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			tradeIterator := NewTradeIterator(testTradeIterator(ts.trades))

			actualTrade, err := tradeIterator.GetMaxBaseAssetEnter(ts.minuteCount, ts.bucketCount, ts.maxTotalTrades)
			if err != ts.expectedErr {
				t.Errorf("expected error to be %v but was %v\n", ts.expectedErr, err)
				t.FailNow()
			}
			if err == nil && ts.expectedTrade != actualTrade {
				t.Errorf("expected trade to be %v but was %v\n", ts.expectedTrade, actualTrade)
				t.FailNow()
			}
		})
	}
}

func f(fl float64) JsonFloat64 {
	return JsonFloat64(fl)
}

func testTradeIterator(cs []Trade) func() (Trade, error) {
	i := 0
	return func() (Trade, error) {
		if i >= len(cs) {
			return Trade{}, ErrOutOfTrades
		}
		i++
		return cs[i-1], nil
	}
}
