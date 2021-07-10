package signalchecker

import (
	"reflect"
	"testing"

	"github.com/marianogappa/signal-checker/types"
)

func TestSignalChecker(t *testing.T) {
	type test struct {
		name         string
		input        types.SignalCheckInput
		candlesticks []types.Candlestick
		expected     types.SignalCheckOutput
	}

	startISO8601 := "2021-07-04T14:14:18Z"
	tick2 := "2021-07-04T14:14:19Z"
	tick3 := "2021-07-04T14:14:20Z"
	startTs := 1625408058

	tss := []test{
		{
			name: "Does not enter",
			input: types.SignalCheckInput{
				EnterRangeLow:          f(1.0),
				EnterRangeHigh:         f(2.0),
				StopLoss:               f(0.1),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []types.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []types.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:       false,
				IfTP2StopAtTP1:         false,
				IfTP3StopAtTP2:         false,
				IfTP4StopAtTP3:         false,
			},
			candlesticks: []types.Candlestick{
				{Timestamp: startTs + 0, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: startTs + 1, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: startTs + 2, LowestPrice: f(0.3), HighestPrice: f(0.3), Volume: f(1.0)},
			},
			expected: types.SignalCheckOutput{
				Events:               []types.SignalCheckOutputEvent{{EventType: types.FINISHED_DATASET}},
				Entered:              false,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        startISO8601,
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				ProfitRatio:          f(0.0),
				IsError:              false,
			},
		},
		{
			name: "Enters",
			input: types.SignalCheckInput{
				EnterRangeLow:          f(1.0),
				EnterRangeHigh:         f(2.0),
				StopLoss:               f(0.1),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []types.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []types.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:       false,
				IfTP2StopAtTP1:         false,
				IfTP3StopAtTP2:         false,
				IfTP4StopAtTP3:         false,
			},
			candlesticks: []types.Candlestick{
				{Timestamp: startTs + 0, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: startTs + 1, LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: startTs + 2, LowestPrice: f(0.3), HighestPrice: f(0.3), Volume: f(1.0)},
			},
			expected: types.SignalCheckOutput{
				Events: []types.SignalCheckOutputEvent{
					{EventType: types.ENTERED, At: tick2, Price: f(1.0)},
					{EventType: types.FINISHED_DATASET},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        startISO8601,
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				ProfitRatio:          f(0.0),
				IsError:              false,
			},
		},
		{
			name: "Stops losses",
			input: types.SignalCheckInput{
				EnterRangeLow:          f(1.0),
				EnterRangeHigh:         f(2.0),
				StopLoss:               f(0.1),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []types.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []types.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:       false,
				IfTP2StopAtTP1:         false,
				IfTP3StopAtTP2:         false,
				IfTP4StopAtTP3:         false,
			},
			candlesticks: []types.Candlestick{
				{Timestamp: startTs + 0, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: startTs + 1, LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: startTs + 2, LowestPrice: f(0.1), HighestPrice: f(0.1), Volume: f(1.0)},
			},
			expected: types.SignalCheckOutput{
				Events: []types.SignalCheckOutputEvent{
					{EventType: types.ENTERED, At: tick2, Price: f(1.0)},
					{EventType: types.STOPPED_LOSS, At: tick3, Price: f(0.1)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        startISO8601,
				HighestTakeProfit:    0,
				ReachedStopLoss:      true,
				ProfitRatio:          f(-0.9),
				IsError:              false,
			},
		},
		{
			name: "Takes TP1",
			input: types.SignalCheckInput{
				EnterRangeLow:          f(1.0),
				EnterRangeHigh:         f(2.0),
				StopLoss:               f(0.1),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []types.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []types.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:       false,
				IfTP2StopAtTP1:         false,
				IfTP3StopAtTP2:         false,
				IfTP4StopAtTP3:         false,
			},
			candlesticks: []types.Candlestick{
				{Timestamp: startTs + 0, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: startTs + 1, LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: startTs + 2, LowestPrice: f(5.0), HighestPrice: f(5.0), Volume: f(1.0)},
			},
			expected: types.SignalCheckOutput{
				Events: []types.SignalCheckOutputEvent{
					{EventType: types.ENTERED, At: tick2, Price: f(1.0)},
					{EventType: types.TAKEN_PROFIT_ + "1", At: tick3, Price: f(5.0)},
					{EventType: types.FINISHED_DATASET},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        startISO8601,
				HighestTakeProfit:    1,
				ReachedStopLoss:      false,
				ProfitRatio:          f(4.0),
				IsError:              false,
			},
		},
		{
			name: "Takes TP1 and Stops Loss",
			input: types.SignalCheckInput{
				EnterRangeLow:          f(1.0),
				EnterRangeHigh:         f(2.0),
				StopLoss:               f(0.1),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []types.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []types.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:       false,
				IfTP2StopAtTP1:         false,
				IfTP3StopAtTP2:         false,
				IfTP4StopAtTP3:         false,
				Debug:                  true,
			},
			candlesticks: []types.Candlestick{
				{Timestamp: startTs + 0, LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: startTs + 1, LowestPrice: f(5.0), HighestPrice: f(5.0), Volume: f(1.0)},
				{Timestamp: startTs + 2, LowestPrice: f(0.1), HighestPrice: f(0.1), Volume: f(1.0)},
			},
			expected: types.SignalCheckOutput{
				Events: []types.SignalCheckOutputEvent{
					{EventType: types.ENTERED, At: startISO8601, Price: f(1.0)},
					{EventType: types.TAKEN_PROFIT_ + "1", At: tick2, Price: f(5.0)},
					{EventType: types.STOPPED_LOSS, At: tick3, Price: f(0.1)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(1.0),
				FirstCandleAt:        startISO8601,
				HighestTakeProfit:    1,
				ReachedStopLoss:      true,
				ProfitRatio:          f(1.5499999999999998),
				IsError:              false,
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			actual, err := doCheckSignal(ts.input, buildTickIterator(testCandlestickIterator(ts.candlesticks)))
			if actual.IsError && err == nil {
				t.Error("Output says there is an error but function did not return an error!")
				t.FailNow()
			}
			if !reflect.DeepEqual(actual.Events, ts.expected.Events) {
				t.Errorf("expected Events = %v but got Events = %v", ts.expected.Events, actual.Events)
			}
			if actual.Entered != ts.expected.Entered {
				t.Errorf("expected Entered = %v but got Entered = %v", ts.expected.Entered, actual.Entered)
			}
			if actual.FirstCandleOpenPrice != ts.expected.FirstCandleOpenPrice {
				t.Errorf("expected FirstCandleOpenPrice = %v but got FirstCandleOpenPrice = %v", ts.expected.FirstCandleOpenPrice, actual.FirstCandleOpenPrice)
			}
			if actual.FirstCandleAt != ts.expected.FirstCandleAt {
				t.Errorf("expected FirstCandleAt = %v but got FirstCandleAt = %v", ts.expected.FirstCandleAt, actual.FirstCandleAt)
			}
			if actual.HighestTakeProfit != ts.expected.HighestTakeProfit {
				t.Errorf("expected HighestTakeProfit = %v but got HighestTakeProfit = %v", ts.expected.HighestTakeProfit, actual.HighestTakeProfit)
			}
			if actual.ReachedStopLoss != ts.expected.ReachedStopLoss {
				t.Errorf("expected ReachedStopLoss = %v but got ReachedStopLoss = %v", ts.expected.ReachedStopLoss, actual.ReachedStopLoss)
			}
			if actual.ProfitRatio != ts.expected.ProfitRatio {
				t.Errorf("expected ProfitRatio = %v but got ProfitRatio = %v", ts.expected.ProfitRatio, actual.ProfitRatio)
			}
			if actual.IsError != ts.expected.IsError {
				t.Errorf("expected IsError = %v but got IsError = %v", ts.expected.IsError, actual.IsError)
			}
		})
	}

}

func f(fl float64) types.JsonFloat64 {
	return types.JsonFloat64(fl)
}

func testCandlestickIterator(cs []types.Candlestick) func() (types.Candlestick, error) {
	i := 0
	return func() (types.Candlestick, error) {
		if i >= len(cs) {
			return types.Candlestick{}, types.ErrOutOfCandlesticks
		}
		i++
		return cs[i-1], nil
	}
}
