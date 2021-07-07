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
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/coinbase"
	"github.com/marianogappa/signal-checker/ftx"
	"github.com/marianogappa/signal-checker/profitcalculator"
	"github.com/marianogappa/signal-checker/types"
)

// CheckSignal is the main method of this project, and runs a signal check based on the provided SignalCheckInput.
// Please review the docs on the types.SignalCheckInput and types.SignalCheckOutput types.
func CheckSignal(input types.SignalCheckInput) (types.SignalCheckOutput, error) {
	validationResult, err := validateInput(input)
	if err != nil {
		return validationResult, err
	}
	input = validationResult.Input
	log.Printf("Input validation ok. Input: %+v\n", input)

	var candlestickIterator func() (types.Candlestick, error)
	switch input.Exchange {
	case types.BINANCE:
		candlestickIterator = binance.BuildCandlestickIterator(input)
	case types.FTX:
		candlestickIterator = ftx.BuildCandlestickIterator(input)
	case types.COINBASE:
		candlestickIterator = coinbase.BuildCandlestickIterator(input)
	}

	return doCheckSignal(input, candlestickIterator)
}

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
	if input.Exchange != "binance" && input.Exchange != "ftx" && input.Exchange != "coinbase" {
		return invalidateWith("At the moment the only valid exchanges are 'binance',  'ftx' and 'coinbase'", input)
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

func resolveInvalidAt(input types.SignalCheckInput) (time.Time, bool) {
	// N.B. already validated
	invalidate, _ := time.Parse(time.RFC3339, input.InvalidateISO3601)
	initial, _ := time.Parse(time.RFC3339, input.InitialISO3601)

	invalidAts := []time.Time{}
	if input.InvalidateISO3601 != "" {
		invalidAts = append(invalidAts, invalidate)
	}
	if input.InvalidateAfterSeconds > 0 {
		invalidAts = append(invalidAts, initial.Add(time.Duration(input.InvalidateAfterSeconds)*time.Second))
	}
	if len(invalidAts) == 0 {
		return time.Time{}, false
	}
	invalidAt := invalidAts[0]
	for i := 1; i < len(invalidAts); i++ {
		if invalidAt.After(invalidAts[i]) {
			invalidAt = invalidAts[i]
		}
	}
	return invalidAt, true
}

func doCheckSignal(input types.SignalCheckInput, nextCandlestick func() (types.Candlestick, error)) (types.SignalCheckOutput, error) {
	var (
		profitCalculator        = profitcalculator.NewProfitCalculator(input)
		first                   = true
		invalidAt, hasInvalidAt = resolveInvalidAt(input)
		events                  = []types.SignalCheckOutputEvent{}
		stopLoss                = input.StopLoss
		err                     error
		candlestick             types.Candlestick
		output                  = types.SignalCheckOutput{
			Input:      input,
			HttpStatus: 200,
		}
		// N.B. this is used to move the stopLoss to previous step (e.g. enter, TP1, ...) as the signal enters new TPs
		priceCheckpoint = 0.0
	)
out:
	for {
		candlestick, err = nextCandlestick()
		if err == types.ErrOutOfCandlesticks {
			event := types.SignalCheckOutputEvent{EventType: types.FINISHED_DATASET}
			events = append(events, event)
			profitCalculator.ApplyEvent(event)
			break out
		}
		if err != nil {
			output.IsError = true
			output.HttpStatus = 500
			output.ErrorMessage = err.Error()
			break out
		}
		prices := []types.JsonFloat64{types.JsonFloat64(candlestick.LowestPrice), types.JsonFloat64(candlestick.HighestPrice)}
		at := time.Unix(int64(candlestick.Timestamp), 0)
		atStr := at.Format(time.RFC3339)
		for _, price := range prices {
			if first {
				first = false
				// N.B. first is always an open side
				output.FirstCandleOpenPrice = price
				output.FirstCandleAt = atStr
			}
			if hasInvalidAt && (at.After(invalidAt) || at.Equal(invalidAt)) {
				event := types.SignalCheckOutputEvent{
					EventType: types.INVALIDATED,
					Price:     price,
					At:        atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
				break out
			}
			if !output.Entered && price >= input.EnterRangeLow && price <= input.EnterRangeHigh {
				output.Entered = true
				event := types.SignalCheckOutputEvent{
					EventType: types.ENTERED,
					Price:     price,
					At:        atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
			}
			if output.Entered && price <= stopLoss {
				event := types.SignalCheckOutputEvent{
					EventType: types.STOPPED_LOSS,
					Price:     price,
					At:        atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
				output.ReachedStopLoss = true
				break out
			}
			if output.Entered && output.HighestTakeProfit < len(input.TakeProfits) && price >= input.TakeProfits[output.HighestTakeProfit] {
				var event types.SignalCheckOutputEvent
				for i := len(input.TakeProfits) - 1; i >= output.HighestTakeProfit; i-- {
					if price < input.TakeProfits[i] {
						continue
					}
					output.HighestTakeProfit = i + 1
					break
				}
				event = types.SignalCheckOutputEvent{
					EventType: fmt.Sprintf("%v%v", types.TAKEN_PROFIT_, output.HighestTakeProfit),
					Price:     price,
					At:        atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
				if output.HighestTakeProfit == len(input.TakeProfits) || profitCalculator.IsFinished() {
					break out
				}
				if (output.HighestTakeProfit == 1 && input.IfTP1StopAtEntry) ||
					(output.HighestTakeProfit == 2 && input.IfTP2StopAtTP1) ||
					(output.HighestTakeProfit == 3 && input.IfTP3StopAtTP2) ||
					(output.HighestTakeProfit == 4 && input.IfTP4StopAtTP3) {
					stopLoss = types.JsonFloat64(priceCheckpoint)
				}
				priceCheckpoint = float64(event.Price)
			}
		}
	}
	output.ProfitRatio = types.JsonFloat64(profitCalculator.CalculateTakeProfitRatio())
	output.Events = events
	return output, err
}
