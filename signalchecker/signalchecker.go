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
	"fmt"
	"log"
	"time"

	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/coinbase"
	"github.com/marianogappa/signal-checker/ftx"
	"github.com/marianogappa/signal-checker/huobi"
	"github.com/marianogappa/signal-checker/kraken"
	"github.com/marianogappa/signal-checker/kucoin"
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
	case types.HUOBI:
		candlestickIterator = huobi.BuildCandlestickIterator(input)
	case types.KRAKEN:
		candlestickIterator = kraken.BuildCandlestickIterator(input)
	case types.KUCOIN:
		candlestickIterator = kucoin.BuildCandlestickIterator(input)
	}

	return doCheckSignal(input, buildTickIterator(candlestickIterator))
}

func resolveInvalidAt(input types.SignalCheckInput) (time.Time, bool) {
	// N.B. already validated
	invalidate, _ := time.Parse(time.RFC3339, input.InvalidateISO8601)
	initial, _ := time.Parse(time.RFC3339, input.InitialISO8601)

	invalidAts := []time.Time{}
	if input.InvalidateISO8601 != "" {
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

type checkSignalState struct {
	input                types.SignalCheckInput
	profitCalculator     profitcalculator.ProfitCalculator
	first                bool
	entered              bool
	reachedStopLoss      bool
	highestTakeProfit    int
	firstCandleOpenPrice types.JsonFloat64
	firstCandleAt        string
	invalidAt            time.Time
	hasInvalidAt         bool
	events               []types.SignalCheckOutputEvent
	stopLoss             types.JsonFloat64
	initialTime          time.Time
	priceCheckpoint      float64
	isEnded              bool
}

func newChecker(input types.SignalCheckInput) *checkSignalState {
	invalidAt, hasInvalidAt := resolveInvalidAt(input)
	initialTime, _ := time.Parse(time.RFC3339, input.InitialISO8601)
	return &checkSignalState{
		input:            input,
		profitCalculator: profitcalculator.NewProfitCalculator(input),
		first:            true,
		invalidAt:        invalidAt,
		hasInvalidAt:     hasInvalidAt,
		stopLoss:         input.StopLoss,
		initialTime:      initialTime,
		priceCheckpoint:  0.0,
	}
}

// N.B. appyEvent returns "isEnded" boolean, to decide whether to continue.
func (s *checkSignalState) applyEvent(eventType string, tick types.Tick) bool {
	event := types.SignalCheckOutputEvent{EventType: eventType}
	switch eventType {
	case types.FINISHED_DATASET:
	default:
		event.At = time.Unix(int64(tick.Timestamp), 0).UTC().Format(time.RFC3339)
		event.Price = tick.Price
	}
	s.events = append(s.events, event)
	s.profitCalculator.ApplyEvent(event)
	s.isEnded = eventType == types.FINISHED_DATASET || eventType == types.STOPPED_LOSS || s.profitCalculator.IsFinished()
	return s.isEnded
}

func (s *checkSignalState) applyTick(tick types.Tick, err error) (bool, error) {
	if err == types.ErrOutOfCandlesticks {
		return s.applyEvent(types.FINISHED_DATASET, tick), err
	}
	if err != nil {
		return true, err
	}
	tickTime := time.Unix(int64(tick.Timestamp), 0)
	if tickTime.Before(s.initialTime) {
		return false, nil
	}
	if s.first {
		s.first = false
		s.firstCandleOpenPrice = tick.Price
		s.firstCandleAt = tickTime.UTC().Format(time.RFC3339)
	}
	if s.hasInvalidAt && (tickTime.After(s.invalidAt) || tickTime.Equal(s.invalidAt)) {
		return s.applyEvent(types.INVALIDATED, tick), nil
	}
	if !s.entered && tick.Price >= s.input.EnterRangeLow && tick.Price <= s.input.EnterRangeHigh {
		s.entered = true
		return s.applyEvent(types.ENTERED, tick), nil
	}
	if s.entered && tick.Price <= s.stopLoss {
		s.reachedStopLoss = true
		return s.applyEvent(types.STOPPED_LOSS, tick), nil
	}
	if s.entered && s.highestTakeProfit < len(s.input.TakeProfits) && tick.Price >= s.input.TakeProfits[s.highestTakeProfit] {
		for i := len(s.input.TakeProfits) - 1; i >= s.highestTakeProfit; i-- {
			if tick.Price < s.input.TakeProfits[i] {
				continue
			}
			s.highestTakeProfit = i + 1
			break
		}
		s.applyEvent(fmt.Sprintf("%v%v", types.TAKEN_PROFIT_, s.highestTakeProfit), tick)
		if s.isEnded || s.highestTakeProfit == len(s.input.TakeProfits) {
			return true, nil
		}
		if (s.highestTakeProfit == 1 && s.input.IfTP1StopAtEntry) ||
			(s.highestTakeProfit == 2 && s.input.IfTP2StopAtTP1) ||
			(s.highestTakeProfit == 3 && s.input.IfTP3StopAtTP2) ||
			(s.highestTakeProfit == 4 && s.input.IfTP4StopAtTP3) {
			s.stopLoss = types.JsonFloat64(s.priceCheckpoint)
		}
		s.priceCheckpoint = float64(tick.Price)
	}
	return false, nil
}

func doCheckSignal(input types.SignalCheckInput, nextTick func() (types.Tick, error)) (types.SignalCheckOutput, error) {
	var (
		checker = newChecker(input)
		err     error
		isEnded bool
	)
	for {
		if isEnded, err = checker.applyTick(nextTick()); isEnded || err != nil {
			break
		}
	}
	output := types.SignalCheckOutput{
		Input:      input,
		HttpStatus: 200,
	}
	if err != nil {
		output.IsError = true
		output.HttpStatus = 500
		output.ErrorMessage = err.Error()
	}
	return types.SignalCheckOutput{
		Events:               checker.events,
		Input:                input,
		Entered:              checker.entered,
		FirstCandleOpenPrice: checker.firstCandleOpenPrice,
		FirstCandleAt:        checker.firstCandleAt,
		HighestTakeProfit:    checker.highestTakeProfit,
		ReachedStopLoss:      checker.reachedStopLoss,
		ProfitRatio:          types.JsonFloat64(checker.profitCalculator.CalculateTakeProfitRatio()),
	}, err
}
