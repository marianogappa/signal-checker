package binance

import (
	"fmt"
	"log"
	"time"

	"github.com/marianogappa/signal-checker/profitcalculator"
	"github.com/marianogappa/signal-checker/types"
)

func CheckSignal(input types.SignalCheckInput) (types.SignalCheckOutput, error) {
	validationResult, err := validateInput(input)
	if err != nil {
		return validationResult, err
	}
	input = validationResult.Input
	log.Printf("Binance input validation ok. Input: %+v\n", input)

	klinesResult, err := getKlines(input)
	if err != nil {
		return types.SignalCheckOutput{
			Input:        input,
			IsError:      true,
			ErrorMessage: err.Error(),
			HttpStatus:   klinesResult.httpStatus,
		}, err
	}
	candlesticks := klinesResult.candlesticks
	log.Printf("Got Klines successfully. Count: %v\n", len(candlesticks))

	return doCheckSignal(input, candlesticks)
}

func validateInput(input types.SignalCheckInput) (types.SignalCheckOutput, error) {
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

type stickSide struct {
	price types.JsonFloat64
	at    time.Time
	atStr string
}

func doCheckSignal(input types.SignalCheckInput, candlesticks []candlestick) (types.SignalCheckOutput, error) {
	var (
		profitCalculator        = profitcalculator.NewProfitCalculator(input)
		first                   = true
		invalidAt, hasInvalidAt = resolveInvalidAt(input)
		events                  = []types.SignalCheckOutputEvent{}
		stopLoss                = input.StopLoss
		output                  = types.SignalCheckOutput{
			Input:      input,
			HttpStatus: 200,
		}
		// N.B. this is used to move the stopLoss to previous step (e.g. enter, TP1, ...) as the signal enters new TPs
		priceCheckpoint = 0.0
	)
	for _, candlestick := range candlesticks {
		sides := []stickSide{
			{price: types.JsonFloat64(candlestick.openPrice), at: candlestick.openAt, atStr: candlestick.openAt.Format(time.RFC3339)},
			{price: types.JsonFloat64(candlestick.closePrice), at: candlestick.closeAt, atStr: candlestick.closeAt.Format(time.RFC3339)},
		}
		for _, side := range sides {
			if first {
				first = false
				// N.B. first is always an open side
				output.FirstCandleOpenPrice = side.price
				output.FirstCandleAt = side.atStr
			}
			if hasInvalidAt && (side.at.After(invalidAt) || side.at.Equal(invalidAt)) {
				event := types.SignalCheckOutputEvent{
					EventType: types.INVALIDATED,
					Price:     side.price,
					At:        side.atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
				break
			}
			if !output.Entered && side.price >= input.EnterRangeLow && side.price <= input.EnterRangeHigh {
				output.Entered = true
				event := types.SignalCheckOutputEvent{
					EventType: types.ENTERED,
					Price:     side.price,
					At:        side.atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
			}
			if output.Entered && side.price <= stopLoss {
				event := types.SignalCheckOutputEvent{
					EventType: types.STOPPED_LOSS,
					Price:     side.price,
					At:        side.atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
				output.ReachedStopLoss = true
				break
			}
			if output.Entered && output.HighestTakeProfit < len(input.TakeProfits) && side.price >= input.TakeProfits[output.HighestTakeProfit] {
				var event types.SignalCheckOutputEvent
				for i := len(input.TakeProfits) - 1; i >= output.HighestTakeProfit; i-- {
					if side.price < input.TakeProfits[i] {
						continue
					}
					output.HighestTakeProfit = i + 1
					break
				}
				event = types.SignalCheckOutputEvent{
					EventType: fmt.Sprintf("%v%v", types.TAKEN_PROFIT_, output.HighestTakeProfit),
					Price:     side.price,
					At:        side.atStr,
				}
				events = append(events, event)
				profitCalculator.ApplyEvent(event)
				if output.HighestTakeProfit == len(input.TakeProfits) || profitCalculator.IsFinished() {
					break
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
	return output, nil
}
