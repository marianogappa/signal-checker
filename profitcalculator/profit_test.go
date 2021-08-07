package profitcalculator

import (
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestProfitCalculator(t *testing.T) {
	type test struct {
		name               string
		input              common.SignalCheckInput
		events             []common.SignalCheckOutputEvent
		expected           []float64
		expectedIsFinished bool
	}

	tss := []test{
		{
			name: "Do not enter, invalidate",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{1.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.INVALIDATED,
					Price:     1,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "Do not enter, finish dataset",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{1.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.FINISHED_DATASET,
					Price:     1,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "Do not enter, incorrect take profit",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{1.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     1,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "Do not enter, incorrect stop loss",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{1.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.STOPPED_LOSS,
					Target:    1,
					Price:     0.1,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), invalidate",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{1.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     0.1,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.INVALIDATED,
					Target:    1,
					Price:     1,
					At:        "2020-01-02T04:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9.0},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), finish dataset",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{1.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     0.1,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.FINISHED_DATASET,
					Target:    1,
					Price:     1,
					At:        "2020-01-02T04:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9.0},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), stop loss",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.STOPPED_LOSS,
					Price:     1,
					At:        "2020-01-02T04:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.9},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9.0},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit (one remaining), invalidate",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0},
				TakeProfitRatios: []common.JsonFloat64{0.5, 0.5},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.INVALIDATED,
					Price:     10,
					At:        "2020-01-02T05:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 4.5},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit (one remaining), finish dataset",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0},
				TakeProfitRatios: []common.JsonFloat64{0.5, 0.5},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.FINISHED_DATASET,
					Price:     10,
					At:        "2020-01-02T05:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 4.5},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit (one remaining), stop loss",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0},
				TakeProfitRatios: []common.JsonFloat64{0.5, 0.5},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.STOPPED_LOSS,
					Price:     1,
					At:        "2020-01-02T05:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 4.05},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit, take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0},
				TakeProfitRatios: []common.JsonFloat64{0.5, 0.5},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    2,
					Price:     1000,
					At:        "2020-01-02T05:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 54},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit, take profit (one remaining), invalidate",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0, 10000.0},
				TakeProfitRatios: []common.JsonFloat64{0.25, 0.50, 0.25},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    2,
					Price:     1000,
					At:        "2020-01-02T05:04:05+00:00",
				},
				{
					EventType: common.INVALIDATED,
					Price:     1000,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 76.5, 76.5},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit, take profit (one remaining), finish dataset",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0, 10000.0},
				TakeProfitRatios: []common.JsonFloat64{0.25, 0.50, 0.25},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    2,
					Price:     1000,
					At:        "2020-01-02T05:04:05+00:00",
				},
				{
					EventType: common.FINISHED_DATASET,
					Price:     1000,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 76.5, 76.5},
			expectedIsFinished: true,
		},
		{
			name: "enter (last), take profit, take profit (one remaining), stop loss",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{10.0},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0, 10000.0},
				TakeProfitRatios: []common.JsonFloat64{0.20, 0.40, 0.40},
			},

			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     10,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    2,
					Price:     1000,
					At:        "2020-01-02T05:04:05+00:00",
				},
				{
					EventType: common.STOPPED_LOSS,
					Price:     1,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 9, 81, 49.032000000000004},
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), invalidate",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.2, 0.8},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0, 10000.0},
				TakeProfitRatios: []common.JsonFloat64{0.2, 0.50, 0.25},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.INVALIDATED,
					Price:     10,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.09999999999999998}, // N.B. argument to stay away from floating point?
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), finish dataset",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.2, 0.8},
				StopLoss:         common.JsonFloat64(1.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0, 10000.0},
				TakeProfitRatios: []common.JsonFloat64{0.2, 0.50, 0.25},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.FINISHED_DATASET,
					Price:     10,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.09999999999999998}, // N.B. argument to stay away from floating point?
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), stop loss",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.2, 0.8},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0, 1000.0, 10000.0},
				TakeProfitRatios: []common.JsonFloat64{0.2, 0.50, 0.25},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.STOPPED_LOSS,
					Price:     10,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.09999999999999998}, // N.B. argument to stay away from floating point?
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 2},
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), enter (last), invalidate",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     40,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.INVALIDATED,
					Price:     20,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.25, -0.25},
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), enter (last), finish dataset",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     40,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.FINISHED_DATASET,
					Price:     20,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.25, -0.25},
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), enter (last), stop loss",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     40,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.STOPPED_LOSS,
					Price:     10,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.25, -0.625},
			expectedIsFinished: true,
		},
		{
			name: "enter (one remaining), enter (last), take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     40,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.25, 2.75},
			expectedIsFinished: true,
		},
		{
			name: "enter (second, last before entering first), take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 4},
			expectedIsFinished: true,
		},
		{
			name: "enter (second, one remaining), take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.3, 0.2},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 3.2},
			expectedIsFinished: true,
		},
		{
			name: "enter first, enter third (last), take profit (last)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.3, 0.2},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    3,
					Price:     10,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.25, 6.5},
			expectedIsFinished: true,
		},
		{
			name: "enter immediately (no entries)",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{},
				EntryRatios:      []common.JsonFloat64{1.0},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.FINISHED_DATASET,
					Target:    1,
					Price:     40,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, 1},
			expectedIsFinished: true,
		},
		{
			name: "bug: unknown event",
			input: common.SignalCheckInput{
				Debug: true,
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: "unknown event",
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: false,
		},
		{
			name: "bug: stop loss without entering",
			input: common.SignalCheckInput{
				Debug: true,
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.STOPPED_LOSS,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "bug: take profit without entering",
			input: common.SignalCheckInput{
				Debug: true,
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "out of sync: invalidating at first event",
			input: common.SignalCheckInput{
				// BaseAsset:        "BTC",
				// QuoteAsset:       "USDT",
				// Entries:          []common.JsonFloat64{40.0, 20.0, 10.0},
				// EntryRatios:      []common.JsonFloat64{0.5, 0.3, 0.2},
				// StopLoss:         common.JsonFloat64(10.0),
				// TakeProfits:      []common.JsonFloat64{100.0},
				// TakeProfitRatios: []common.JsonFloat64{1},
				Debug: true,
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.INVALIDATED,
					Price:     20,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected:           []float64{0.0},
			expectedIsFinished: true,
		},
		{
			name: "more entries than entry ratios",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				Entries:          []common.JsonFloat64{40.0, 20.0, 10.0},
				EntryRatios:      []common.JsonFloat64{0.5, 0.5},
				StopLoss:         common.JsonFloat64(10.0),
				TakeProfits:      []common.JsonFloat64{100.0},
				TakeProfitRatios: []common.JsonFloat64{1},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Target:    1,
					Price:     40,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    2,
					Price:     20,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.ENTERED,
					Target:    3,
					Price:     10,
					At:        "2020-01-02T04:04:05+00:00",
				},
				{
					EventType: common.TOOK_PROFIT,
					Target:    1,
					Price:     100,
					At:        "2020-01-02T06:04:05+00:00",
				},
			},
			expected:           []float64{0.0, -0.25, -0.625, 2.75},
			expectedIsFinished: true,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			profitCalculator := NewProfitCalculator(ts.input)
			if ts.expected == nil {
				return
			}
			for i, ev := range ts.events {
				actual := profitCalculator.ApplyEvent(ev)
				if actual != ts.expected[i] {
					t.Fatalf("On event %v expected %v to equal %v", i, actual, ts.expected[i])
				}
			}
			actual := profitCalculator.CalculateTakeProfitRatio()
			if actual != ts.expected[len(ts.events)-1] {
				t.Fatalf("On final calculation expected %v to equal %v", actual, ts.expected[len(ts.events)-1])
			}
			actualIsFinished := profitCalculator.IsFinished()
			if actualIsFinished != ts.expectedIsFinished {
				t.Fatalf("Expected isFinished = %v to be = %v", actualIsFinished, ts.expectedIsFinished)
			}
		})
	}
}
