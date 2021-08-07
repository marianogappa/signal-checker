package signalchecker

import (
	"errors"
	"reflect"
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestSignalChecker(t *testing.T) {
	type test struct {
		name         string
		input        common.SignalCheckInput
		candlesticks []common.Candlestick
		expected     common.SignalCheckOutput
	}

	ts := []common.ISO8601{
		common.ISO8601("2021-07-04T14:14:18Z"),
		common.ISO8601("2021-07-04T14:15:18Z"),
		common.ISO8601("2021-07-04T14:16:18Z"),
		common.ISO8601("2021-07-04T14:17:18Z"),
		common.ISO8601("2021-07-04T14:18:18Z"),
		common.ISO8601("2021-07-04T14:19:18Z"),
		common.ISO8601("2021-07-04T14:20:18Z"),
	}
	tsSec := []int{}
	for _, tmstmp := range ts {
		sec, _ := tmstmp.Seconds()
		tsSec = append(tsSec, sec)
	}
	earlierTsSec, _ := common.ISO8601("2021-07-04T14:13:18Z").Seconds()

	tss := []test{
		{
			name: "Does not enter",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				EntryRatios:              []common.JsonFloat64{f(1)},
				StopLoss:                 f(0.1),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.3), HighestPrice: f(0.3), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events:               []common.SignalCheckOutputEvent{{EventType: common.FINISHED_DATASET, At: ts[2], Price: f(0.3)}},
				Entered:              false,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "Enters",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				EntryRatios:              []common.JsonFloat64{f(1)},
				StopLoss:                 f(0.1),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.3), HighestPrice: f(0.3), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[1], Price: f(1.0), ProfitRatio: 0},
					{EventType: common.FINISHED_DATASET, At: ts[2], Price: f(0.3), ProfitRatio: -0.7},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "Enters immediately",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{},
				StopLoss:                 f(0.1),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.4), HighestPrice: f(0.4), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(0.2), ProfitRatio: f(0)},
					{EventType: common.FINISHED_DATASET, At: ts[2], Price: f(0.4), ProfitRatio: f(1)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "Stops losses",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				EntryRatios:              []common.JsonFloat64{f(1)},
				StopLoss:                 f(0.1),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.1), HighestPrice: f(0.1), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[1], Price: f(1.0), ProfitRatio: f(0)},
					{EventType: common.STOPPED_LOSS, At: ts[2], Price: f(0.1), ProfitRatio: f(-0.9)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      true,
				IsError:              false,
			},
		},
		{
			name: "Takes TP1",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				EntryRatios:              []common.JsonFloat64{f(1.0)},
				StopLoss:                 f(0.1),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(5.0), HighestPrice: f(5.0), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[1], Price: f(1.0), ProfitRatio: f(0)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[2], Price: f(5.0), ProfitRatio: f(4)},
					{EventType: common.FINISHED_DATASET, At: ts[2], Price: f(5.0), ProfitRatio: f(4)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(0.2),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "Takes TP1 and Stops Loss",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(5.0), HighestPrice: f(5.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(1.0), ProfitRatio: f(0)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[1], Price: f(5.0), ProfitRatio: f(4)},
					{EventType: common.STOPPED_LOSS, At: ts[2], Price: f(0.2), ProfitRatio: f(1.6)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(1.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      true,
				IsError:              false,
			},
		},
		{
			name: "Takes TP1 and Stops Loss, ignoring one earlier candlestick",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				IfTP1StopAtEntry:         false,
				IfTP2StopAtTP1:           false,
				IfTP3StopAtTP2:           false,
				IfTP4StopAtTP3:           false,
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: earlierTsSec, LowestPrice: f(1000.0), HighestPrice: f(1000.0), Volume: f(1.0)},
				{Timestamp: tsSec[0], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(5.0), HighestPrice: f(5.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(1.0), ProfitRatio: f(0)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[1], Price: f(5.0), ProfitRatio: f(4)},
					{EventType: common.STOPPED_LOSS, At: ts[2], Price: f(0.2), ProfitRatio: f(1.6)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(1.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      true,
				IsError:              false,
			},
		},
		{
			name: "Does not enter do to invalidate date",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        ts[0], // <-
				InvalidateAfterSeconds:   86400,
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.INVALIDATED, At: ts[0], Price: f(1.0), ProfitRatio: f(0)},
				},
				Entered:              false,
				FirstCandleOpenPrice: f(1.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "Does not enter do to invalidate in seconds",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.0), f(1.0)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        "",
				InvalidateAfterSeconds:   60, // <-
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(1.0), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.INVALIDATED, At: ts[1], Price: f(1.0), ProfitRatio: f(0)},
				},
				Entered:              false,
				FirstCandleOpenPrice: f(1.0),
				FirstCandleAt:        ts[1],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "enters twice and stops losses",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				TakeProfits:              []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.25, 0.25},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.STOPPED_LOSS, At: ts[2], Price: f(0.2), ProfitRatio: f(-0.85)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      true,
				IsError:              false,
			},
		},
		{
			name: "enters twice and takes tp1 (final)",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				TakeProfits:              []common.JsonFloat64{5.0},
				TakeProfitRatios:         []common.JsonFloat64{1.0},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(5), HighestPrice: f(6), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[2], Price: f(5), ProfitRatio: f(2.75)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "enters twice and takes tp1 (one remaining) and finishes dataset",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				TakeProfits:              []common.JsonFloat64{5.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.5},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(5), HighestPrice: f(6), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[2], Price: f(5), ProfitRatio: f(2.75)},
					{EventType: common.FINISHED_DATASET, At: ts[2], Price: f(6), ProfitRatio: f(3.125)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "enters twice and takes tp1 (one remaining) and invalidates",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        ts[3],
				TakeProfits:              []common.JsonFloat64{5.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.5},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(5), HighestPrice: f(6), Volume: f(1.0)},
				{Timestamp: tsSec[3], LowestPrice: f(5), HighestPrice: f(6), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[2], Price: f(5), ProfitRatio: f(2.75)},
					{EventType: common.INVALIDATED, At: ts[3], Price: f(5), ProfitRatio: f(2.75)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "enters twice and takes tp1 (one remaining) and stops losses",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				TakeProfits:              []common.JsonFloat64{5.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.5},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(5), HighestPrice: f(6), Volume: f(1.0)},
				{Timestamp: tsSec[3], LowestPrice: f(0.2), HighestPrice: f(0.3), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[2], Price: f(5), ProfitRatio: f(2.75)},
					{EventType: common.STOPPED_LOSS, At: ts[3], Price: f(0.2), ProfitRatio: f(0.95)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    1,
				ReachedStopLoss:      true,
				IsError:              false,
			},
		},
		{
			name: "enters twice and takes tp1 and tp2 (final)",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				TakeProfits:              []common.JsonFloat64{5.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.5},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(5), HighestPrice: f(6), Volume: f(1.0)},
				{Timestamp: tsSec[3], LowestPrice: f(7), HighestPrice: f(8), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.TOOK_PROFIT, Target: 1, At: ts[2], Price: f(5), ProfitRatio: f(2.75)},
					{EventType: common.TOOK_PROFIT, Target: 2, At: ts[3], Price: f(7), ProfitRatio: f(3.5)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    2,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		{
			name: "enters twice and finishes dataset",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				TakeProfits:              []common.JsonFloat64{5.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.5},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.FINISHED_DATASET, At: ts[1], Price: f(2), ProfitRatio: f(0.5)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
		// TODO enters twice and invalidates
		{
			name: "enters twice and finishes dataset",
			input: common.SignalCheckInput{
				Exchange:                 "fake",
				BaseAsset:                "BTC",
				QuoteAsset:               "USDT",
				Entries:                  []common.JsonFloat64{f(2.5), f(1.5), f(0.5)},
				EntryRatios:              []common.JsonFloat64{f(0.5), f(0.5)},
				StopLoss:                 f(0.2),
				InitialISO8601:           ts[0],
				InvalidateISO8601:        ts[2],
				TakeProfits:              []common.JsonFloat64{5.0, 7.0},
				TakeProfitRatios:         []common.JsonFloat64{0.5, 0.5},
				DontCalculateMaxEnterUSD: true,
			},
			candlesticks: []common.Candlestick{
				{Timestamp: tsSec[0], LowestPrice: f(2.0), HighestPrice: f(3.0), Volume: f(1.0)},
				{Timestamp: tsSec[1], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
				{Timestamp: tsSec[2], LowestPrice: f(1.0), HighestPrice: f(2.0), Volume: f(1.0)},
			},
			expected: common.SignalCheckOutput{
				Events: []common.SignalCheckOutputEvent{
					{EventType: common.ENTERED, Target: 1, At: ts[0], Price: f(2.0), ProfitRatio: f(0)},
					{EventType: common.ENTERED, Target: 2, At: ts[1], Price: f(1.0), ProfitRatio: f(-0.25)},
					{EventType: common.INVALIDATED, At: ts[2], Price: f(1), ProfitRatio: f(-0.25)},
				},
				Entered:              true,
				FirstCandleOpenPrice: f(2.0),
				FirstCandleAt:        ts[0],
				HighestTakeProfit:    0,
				ReachedStopLoss:      false,
				IsError:              false,
			},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			sChecker := NewSignalChecker(ts.input)
			sChecker.mockCandlesticks = ts.candlesticks
			actual, err := sChecker.Check()
			if actual.IsError && err == nil {
				t.Fatalf("Output says there is an error but function did not return an error!")
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
			if actual.IsError != ts.expected.IsError {
				t.Errorf("expected IsError = %v but got IsError = %v", ts.expected.IsError, actual.IsError)
			}
		})
	}

}

func TestValidationFailure(t *testing.T) {
	_, err := NewSignalChecker(common.SignalCheckInput{}).Check()
	if err == nil {
		t.Fatalf("validation should have failed")
	}
}

func TestCalculateInvalidateAtTime(t *testing.T) {
	ts := common.ISO8601("2021-07-04T14:14:18Z")
	sec, _ := ts.Seconds()
	tsSec := sec

	input := common.SignalCheckInput{
		Exchange:                 "fake",
		BaseAsset:                "BTC",
		QuoteAsset:               "USDT",
		InitialISO8601:           ts,
		InvalidateISO8601:        ts,
		InvalidateAfterSeconds:   86400,
		DontCalculateMaxEnterUSD: true,
	}
	sChecker := NewSignalChecker(input)
	sChecker.mockCandlesticks = []common.Candlestick{
		{Timestamp: tsSec, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
	}
	output, err := sChecker.Check()
	if err != nil {
		t.Fatalf("check should have succeeded, but failed with %v", err)
	}
	if len(output.Events) != 1 {
		t.Fatalf("there should only be 1 event, but there were %v", len(output.Events))
	}
	if output.Events[0].EventType != common.INVALIDATED {
		t.Fatalf("there only event should be 'invalidated' but was %v", output.Events[0].EventType)
	}
}

func TestCalculateInvalidateAfterSeconds(t *testing.T) {
	ts := common.ISO8601("2021-07-04T14:14:18Z")
	tsPlusOneMinute := common.ISO8601("2021-07-04T14:15:18Z")
	sec, _ := tsPlusOneMinute.Seconds()
	tsPlusOneMinuteSec := sec

	input := common.SignalCheckInput{
		Exchange:                 "fake",
		BaseAsset:                "BTC",
		QuoteAsset:               "USDT",
		InitialISO8601:           ts,
		InvalidateISO8601:        tsPlusOneMinute,
		InvalidateAfterSeconds:   180,
		DontCalculateMaxEnterUSD: true,
	}
	sChecker := NewSignalChecker(input)
	sChecker.mockCandlesticks = []common.Candlestick{
		{Timestamp: tsPlusOneMinuteSec, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
	}
	output, err := sChecker.Check()
	if err != nil {
		t.Fatalf("check should have succeeded, but failed with %v", err)
	}
	if len(output.Events) != 1 {
		t.Fatalf("there should only be 1 event, but there were %v", len(output.Events))
	}
	if output.Events[0].EventType != common.INVALIDATED {
		t.Fatalf("there only event should be 'invalidated' but was %v", output.Events[0].EventType)
	}
}

func TestCalculateNoInvalidateAts(t *testing.T) {
	ts := common.ISO8601("2021-07-04T14:14:18Z")

	input := common.SignalCheckInput{
		Exchange:                 "fake",
		BaseAsset:                "BTC",
		QuoteAsset:               "USDT",
		InitialISO8601:           ts,
		DontCalculateMaxEnterUSD: true,
	}
	sChecker := NewSignalChecker(input)
	sChecker.mockCandlesticks = []common.Candlestick{}
	output, err := sChecker.Check()
	if err != nil && err != common.ErrOutOfCandlesticks {
		t.Fatalf("check should have succeeded, but failed with %v", err)
	}
	if len(output.Events) != 1 {
		t.Fatalf("there should only be 1 event, but there were %v", len(output.Events))
	}
	if output.Events[0].EventType != common.FINISHED_DATASET {
		t.Fatalf("there only event should be 'finished dataset' but was %v", output.Events[0].EventType)
	}
}

func TestCalculateBothInvalidateAts(t *testing.T) {
	ts := common.ISO8601("2021-07-04T14:14:18Z")
	tsPlusOneMinute := common.ISO8601("2021-07-04T14:15:18Z")
	sec, _ := tsPlusOneMinute.Seconds()
	tsPlusOneMinuteSec := sec

	input := common.SignalCheckInput{
		Exchange:                 "fake",
		BaseAsset:                "BTC",
		QuoteAsset:               "USDT",
		InitialISO8601:           ts,
		InvalidateISO8601:        tsPlusOneMinute,
		InvalidateAfterSeconds:   30,
		DontCalculateMaxEnterUSD: true,
		Debug:                    true,
	}
	sChecker := NewSignalChecker(input)
	sChecker.mockCandlesticks = []common.Candlestick{
		{Timestamp: tsPlusOneMinuteSec, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
	}
	output, err := sChecker.Check()
	if err != nil {
		t.Fatalf("check should have succeeded, but failed with %v", err)
	}
	if len(output.Events) != 1 {
		t.Fatalf("there should only be 1 event, but there were %v", len(output.Events))
	}
	if output.Events[0].EventType != common.INVALIDATED {
		t.Fatalf("there only event should be 'invalidated' but was %v", output.Events[0].EventType)
	}
}

func TestSaveCandlesticks(t *testing.T) {
	ts := common.ISO8601("2021-07-04T14:14:18Z")
	sec, _ := ts.Seconds()
	tsSec := sec

	input := common.SignalCheckInput{
		Exchange:                 "fake",
		BaseAsset:                "BTC",
		QuoteAsset:               "USDT",
		Entries:                  []common.JsonFloat64{f(2), f(1)},
		InitialISO8601:           ts,
		DontCalculateMaxEnterUSD: true,
		ReturnCandlesticks:       true,
	}
	sChecker := NewSignalChecker(input)
	sChecker.mockCandlesticks = []common.Candlestick{
		{Timestamp: tsSec, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
	}
	output, err := sChecker.Check()
	if err != nil && err != common.ErrOutOfCandlesticks {
		t.Fatalf("check should have succeeded, but failed with %v", err)
	}
	if len(output.Events) != 1 {
		t.Fatalf("there should only be 1 event, but there were %v", len(output.Events))
	}
	if output.Events[0].EventType != common.FINISHED_DATASET {
		t.Fatalf("there only event should be 'finished_dataset' but was %v", output.Events[0].EventType)
	}
	if len(output.Candlesticks) != 1 {
		t.Fatalf("output candlesticks length should have been 2 but was %v", len(output.Candlesticks))
	}
}

func TestErrorsGettingCandlesticks(t *testing.T) {
	ts := common.ISO8601("2021-07-04T14:14:18Z")
	sec, _ := ts.Seconds()
	tsSec := sec

	testErr := errors.New("error for testing")

	input := common.SignalCheckInput{
		Exchange:                 "fake",
		BaseAsset:                "BTC",
		QuoteAsset:               "USDT",
		Entries:                  []common.JsonFloat64{f(2), f(1)},
		InitialISO8601:           ts,
		DontCalculateMaxEnterUSD: true,
	}
	sChecker := NewSignalChecker(input)
	sChecker.mockCandlesticks = []common.Candlestick{
		{Timestamp: tsSec, LowestPrice: f(0.2), HighestPrice: f(0.2), Volume: f(1.0)},
	}
	sChecker.mockReturnErr = testErr
	output, err := sChecker.Check()
	if err != testErr {
		t.Fatalf("check should have return a test error, but it returned %v", err)
	}
	if !output.IsError {
		t.Fatal("output.IsError should have been true")
	}
	if output.HttpStatus != 500 {
		t.Fatalf("output.HttpStatus should have been 500 but was %v", output.HttpStatus)
	}
	if output.ErrorMessage != testErr.Error() {
		t.Fatalf("output.ErrorMessage should have been %v but was %v", testErr.Error(), output.ErrorMessage)
	}
}

func f(fl float64) common.JsonFloat64 {
	return common.JsonFloat64(fl)
}
