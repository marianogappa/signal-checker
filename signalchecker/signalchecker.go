// The signalchecker package contains the CheckSignal function that does the signal checking.
//
// Please review the docs on the types.SignalCheckInput and types.SignalCheckOutput types.
//
// If you are importing the package, use it like this:
//
// output, err := signalchecker.CheckSignal(input)
//
// Note that the output contains richer information about the error than err itself, but you can still use err if it
// reads better in your code.
//
// Consider using this signal checker via de cli binary or as a server, which is also started from the binary.
// Latest release available at: https://github.com/marianogappa/signal-checker/releases
// You can also build locally using: go get github.com/marianogappa/signal-checker
package signalchecker

import (
	"errors"
	"strings"
	"time"

	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/types"
)

// CheckSignal is the main method of this project, and runs a signal check based on the provided SignalCheckInput.
// Please review the docs on the types.SignalCheckInput and types.SignalCheckOutput types.
func CheckSignal(input types.SignalCheckInput) (types.SignalCheckOutput, error) {
	validationResult, err := validateInput(input)
	if err != nil {
		return validationResult, err
	}

	return binance.CheckSignal(validationResult.Input)
}

func invalidateWith(msg string, input types.SignalCheckInput) (types.SignalCheckOutput, error) {
	err := errors.New(msg)
	return types.SignalCheckOutput{IsError: true, HttpStatus: 400, ErrorMessage: err.Error(), Input: input}, err
}

func contains(needle string, haystack []string) bool {
	for _, hay := range haystack {
		if needle == hay {
			return true
		}
	}
	return false
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
	if input.CandlestickInterval == "" {
		input.CandlestickInterval = "15m"
	}
	if input.Exchange != "binance" {
		return invalidateWith("At the moment the only valid exchange is 'binance'", input)
	}
	if !contains(input.CandlestickInterval, []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d", "3d", "1w", "1M"}) {
		return invalidateWith("Only valid candlestick intervals are 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M", input)
	}
	if input.InitialISO3601 == "" {
		return invalidateWith("initialISO3601 is required", input)
	}
	if _, err := time.Parse(time.RFC3339, input.InitialISO3601); err != nil {
		return invalidateWith("initialISO3601 is formatted incorrectly, should be ISO3601 e.g. 2021-07-04T14:14:18+00:00", input)
	}
	if _, err := time.Parse(time.RFC3339, input.InvalidateISO3601); input.InvalidateISO3601 != "" && err != nil {
		return invalidateWith("invalidateISO3601 is formatted incorrectly, should be ISO3601 e.g. 2021-07-04T14:14:18+00:00", input)
	}
	if len(input.TakeProfitRatios) > 0 && sum(input.TakeProfitRatios) != 1.0 {
		return invalidateWith("takeProfitRatios must add up to 1 (but it does not need to match the takeProfits length)", input)
	}
	return types.SignalCheckOutput{Input: input}, nil
}
