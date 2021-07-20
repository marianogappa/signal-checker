package signalchecker

import (
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestValidate(t *testing.T) {
	type test struct {
		name        string
		input       common.SignalCheckInput
		expectedErr error
	}

	startISO8601 := common.ISO8601("2021-07-04T14:14:18Z")

	tss := []test{
		{
			name: "enterRangeHigh is < enterRangeLow",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(1.0), // Lower than EnterRangeLow!
				StopLoss:               f(0.1),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []common.JsonFloat64{0.5, 0.25, 0.25},
			},
			expectedErr: common.ErrEnterRangeHighIsLessThanEnterRangeLow,
		},
		{
			name: "stoploss is >= enterRangeLow",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(2.1), // Greater than EnterRangeLow!
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []common.JsonFloat64{0.5, 0.25, 0.25},
			},
			expectedErr: common.ErrStopLossIsGreaterThanOrEqualToEnterRangeLow,
		},
		{
			name: "(SHORT) stoploss is <= enterRangeHigh",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                true,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(2.9), // Less than EnterRangeHigh!
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{},
				TakeProfitRatios:       []common.JsonFloat64{},
			},
			expectedErr: common.ErrStopLossIsLessThanOrEqualToEnterRangeHigh,
		},
		{
			name: "(LONG) stoploss is >= enterRangeLow",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(2.1), // Greater than EnterRangeLow!
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{5.0, 6.0, 7.0},
				TakeProfitRatios:       []common.JsonFloat64{0.5, 0.25, 0.25},
			},
			expectedErr: common.ErrStopLossIsGreaterThanOrEqualToEnterRangeLow,
		},
		{
			name: "(LONG) TP1 is <= enterRangeHigh",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(1.0),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{2.9}, // TP1 <= EnterRangeHigh!
				TakeProfitRatios:       []common.JsonFloat64{1},
			},
			expectedErr: common.ErrFirstTPIsLessThanOrEqualToEnterRangeHigh,
		},
		{
			name: "(SHORT) TP1 is <= enterRangeHigh",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                true,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(4.0),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{2.1}, // TP1 >= EnterRangeLow!
				TakeProfitRatios:       []common.JsonFloat64{1},
			},
			expectedErr: common.ErrFirstTPIsGreaterThanOrEqualToEnterRangeLow,
		},
		{
			name: "Invalid exchange",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				Exchange:               "invalid exchange",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(1.0),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{},
				TakeProfitRatios:       []common.JsonFloat64{},
			},
			expectedErr: common.ErrInvalidExchange,
		},
		{
			name: "InitialISO8601 empty",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(1.0),
				InitialISO8601:         "",
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{},
				TakeProfitRatios:       []common.JsonFloat64{},
			},
			expectedErr: common.ErrInitialISO8601Required,
		},
		{
			name: "InitialISO8601 formatted incorrectly",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(1.0),
				InitialISO8601:         "01 Jan 2021 10:31:00",
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{},
				TakeProfitRatios:       []common.JsonFloat64{},
			},
			expectedErr: common.ErrInitialISO8601FormattedIncorrectly,
		},
		{
			name: "InvalidateISO8601 formatted incorrectly",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(1.0),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "01 Jan 2021 10:31:00",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{},
				TakeProfitRatios:       []common.JsonFloat64{},
			},
			expectedErr: common.ErrInvalidateISO8601FormattedIncorrectly,
		},
		{
			name: "TakeProfitRatios must add up to 1",
			input: common.SignalCheckInput{
				BaseAsset:              "BTC",
				QuoteAsset:             "USDT",
				IsShort:                false,
				EnterRangeLow:          f(2.0),
				EnterRangeHigh:         f(3.0),
				StopLoss:               f(1.0),
				InitialISO8601:         startISO8601,
				InvalidateISO8601:      "",
				InvalidateAfterSeconds: 10,
				TakeProfits:            []common.JsonFloat64{},
				TakeProfitRatios:       []common.JsonFloat64{0.99},
			},
			expectedErr: common.ErrTakeProfitRatiosMustAddUpToOne,
		},
		{
			name: "base asset is required",
			input: common.SignalCheckInput{
				QuoteAsset:     "USDT",
				EnterRangeLow:  f(2.0),
				EnterRangeHigh: f(3.0),
				StopLoss:       f(1.0),
				InitialISO8601: startISO8601,
			},
			expectedErr: common.ErrBaseAssetRequired,
		},
		{
			name: "quote asset is required",
			input: common.SignalCheckInput{
				BaseAsset:      "BTC",
				EnterRangeLow:  f(2.0),
				EnterRangeHigh: f(3.0),
				StopLoss:       f(1.0),
				InitialISO8601: startISO8601,
			},
			expectedErr: common.ErrQuoteAssetRequired,
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			_, actualErr := validateInput(ts.input)
			if actualErr != ts.expectedErr {
				t.Errorf("Expected error %v, but got error %v", ts.expectedErr, actualErr)
				t.FailNow()
			}
		})
	}

}

func TestUppercasesAssets(t *testing.T) {
	startISO8601 := common.ISO8601("2021-07-04T14:14:18Z")

	validatedInput, err := validateInput(common.SignalCheckInput{
		BaseAsset:      "btc",
		QuoteAsset:     "usdt",
		EnterRangeLow:  f(2.0),
		EnterRangeHigh: f(3.0),
		StopLoss:       f(1.0),
		InitialISO8601: startISO8601,
	})
	if err != nil {
		t.Errorf("validation returned error %v", err)
	}
	if validatedInput.Input.BaseAsset != "BTC" {
		t.Errorf("validation did not uppercase base asset")
	}
	if validatedInput.Input.QuoteAsset != "USDT" {
		t.Errorf("validation did not uppercase quote asset")
	}
}

func TestLowercasesExchangeNames(t *testing.T) {
	startISO8601 := common.ISO8601("2021-07-04T14:14:18Z")

	validatedInput, err := validateInput(common.SignalCheckInput{
		BaseAsset:      "BTC",
		QuoteAsset:     "USDT",
		Exchange:       "BINANCE",
		EnterRangeLow:  f(2.0),
		EnterRangeHigh: f(3.0),
		StopLoss:       f(1.0),
		InitialISO8601: startISO8601,
	})
	if err != nil {
		t.Errorf("validation returned error %v", err)
	}
	if validatedInput.Input.Exchange != "binance" {
		t.Errorf("validation did not lowercase exchange")
	}
}
