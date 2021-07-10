package signalchecker

import (
	"errors"
	"strings"
	"time"

	"github.com/marianogappa/signal-checker/types"
)

func invalidateWith(msg string, input types.SignalCheckInput) (types.SignalCheckOutput, error) {
	err := errors.New(msg)
	return types.SignalCheckOutput{IsError: true, HttpStatus: 400, ErrorMessage: err.Error(), Input: input}, err
}

func sum(ss []types.JsonFloat64) float64 {
	sum := 0.0
	for _, s := range ss {
		sum += float64(s)
	}
	return sum
}

func validateInput(input types.SignalCheckInput) (types.SignalCheckOutput, error) {
	input.Exchange = strings.ToLower(input.Exchange)
	input.BaseAsset = strings.ToUpper(input.BaseAsset)
	input.QuoteAsset = strings.ToUpper(input.QuoteAsset)
	if input.Exchange == "" {
		input.Exchange = "binance"
	}
	if input.Exchange != "binance" && input.Exchange != "ftx" && input.Exchange != "coinbase" && input.Exchange != "huobi" && input.Exchange != "kraken" && input.Exchange != "kucoin" {
		return invalidateWith("The only valid exchanges are 'binance', 'ftx', 'coinbase', 'huobi', 'kraken' and 'kucoin'", input)
	}
	if input.InitialISO8601 == "" {
		return invalidateWith("InitialISO8601 is required", input)
	}
	if _, err := time.Parse(time.RFC3339, input.InitialISO8601); err != nil {
		return invalidateWith("InitialISO8601 is formatted incorrectly, should be ISO3601 e.g. 2021-07-04T14:14:18+00:00", input)
	}
	if _, err := time.Parse(time.RFC3339, input.InvalidateISO8601); input.InvalidateISO8601 != "" && err != nil {
		return invalidateWith("InvalidateISO8601 is formatted incorrectly, should be ISO3601 e.g. 2021-07-04T14:14:18+00:00", input)
	}
	if len(input.TakeProfitRatios) > 0 && sum(input.TakeProfitRatios) != 1.0 {
		return invalidateWith("takeProfitRatios must add up to 1 (but it does not need to match the takeProfits length)", input)
	}
	// TODO check price targets don't match
	return types.SignalCheckOutput{Input: input}, nil
}
