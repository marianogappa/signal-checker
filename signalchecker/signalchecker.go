// The signalchecker package contains the CheckSignal function that does the signal checking.
//
// Please review the docs on the common.SignalCheckInput and common.SignalCheckOutput common.
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
	"time"

	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/coinbase"
	"github.com/marianogappa/signal-checker/common"
	"github.com/marianogappa/signal-checker/ftx"
	"github.com/marianogappa/signal-checker/kraken"
	"github.com/marianogappa/signal-checker/kucoin"
	"github.com/marianogappa/signal-checker/profitcalculator"
)

var (
	exchanges = map[string]common.Exchange{
		common.BINANCE:  binance.NewBinance(),
		common.FTX:      ftx.NewFTX(),
		common.COINBASE: coinbase.NewCoinbase(),
		common.KRAKEN:   kraken.NewKraken(),
		common.KUCOIN:   kucoin.NewKucoin(),
	}
)

// CheckSignal is the main method of this project, and runs a signal check based on the provided SignalCheckInput.
// Please review the docs on the common.SignalCheckInput and common.SignalCheckOutput common.
func CheckSignal(input common.SignalCheckInput) (common.SignalCheckOutput, error) {
	validationResult, err := validateInput(input)
	if err != nil {
		return validationResult, err
	}
	input = validationResult.Input
	log.Printf("Input validation ok. Input: %+v\n", input)

	var (
		exchange            = exchanges[input.Exchange]
		candlestickIterator = exchange.BuildCandlestickIterator(input.BaseAsset, input.QuoteAsset, input.InitialISO8601)
		tickIterator        = buildTickIterator(candlestickIterator.Next)
	)

	return doCheckSignal(input, tickIterator)
}

func resolveInvalidAt(input common.SignalCheckInput) (time.Time, bool) {
	// N.B. already validated
	invalidate, _ := input.InvalidateISO8601.Time()
	initial, _ := input.InitialISO8601.Time()

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
	input                common.SignalCheckInput
	profitCalculator     profitcalculator.ProfitCalculator
	first                bool
	entered              bool
	reachedStopLoss      bool
	highestTakeProfit    int
	firstCandleOpenPrice common.JsonFloat64
	firstCandleAt        string
	invalidAt            time.Time
	hasInvalidAt         bool
	events               []common.SignalCheckOutputEvent
	stopLoss             common.JsonFloat64
	initialTime          time.Time
	priceCheckpoint      float64
	isEnded              bool
	enterAmounts         []common.JsonFloat64
}

func newChecker(input common.SignalCheckInput) *checkSignalState {
	invalidAt, hasInvalidAt := resolveInvalidAt(input)
	initialTime, _ := input.InitialISO8601.Time()
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
func (s *checkSignalState) applyEvent(eventType string, tick common.Tick) bool {
	event := common.SignalCheckOutputEvent{EventType: eventType}
	switch eventType {
	case common.FINISHED_DATASET:
	default:
		event.At = common.ISO8601(time.Unix(int64(tick.Timestamp), 0).UTC().Format(time.RFC3339))
		event.Price = tick.Price
	}
	s.events = append(s.events, event)
	s.profitCalculator.ApplyEvent(event)
	s.isEnded = eventType == common.FINISHED_DATASET || eventType == common.STOPPED_LOSS || s.profitCalculator.IsFinished()
	return s.isEnded
}

func (s *checkSignalState) applyTick(tick common.Tick, err error) (bool, error) {
	if err == common.ErrOutOfCandlesticks {
		return s.applyEvent(common.FINISHED_DATASET, tick), err
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
		return s.applyEvent(common.INVALIDATED, tick), nil
	}
	if !s.entered && tick.Price >= s.input.EnterRangeLow && tick.Price <= s.input.EnterRangeHigh {
		s.entered = true
		return s.applyEvent(common.ENTERED, tick), nil
	}
	if s.entered && tick.Price <= s.stopLoss {
		s.reachedStopLoss = true
		return s.applyEvent(common.STOPPED_LOSS, tick), nil
	}
	if s.entered && s.highestTakeProfit < len(s.input.TakeProfits) && tick.Price >= s.input.TakeProfits[s.highestTakeProfit] {
		for i := len(s.input.TakeProfits) - 1; i >= s.highestTakeProfit; i-- {
			if tick.Price < s.input.TakeProfits[i] {
				continue
			}
			s.highestTakeProfit = i + 1
			break
		}
		s.applyEvent(fmt.Sprintf("%v%v", common.TAKEN_PROFIT_, s.highestTakeProfit), tick)
		if s.isEnded || s.highestTakeProfit == len(s.input.TakeProfits) {
			return true, nil
		}
		if (s.highestTakeProfit == 1 && s.input.IfTP1StopAtEntry) ||
			(s.highestTakeProfit == 2 && s.input.IfTP2StopAtTP1) ||
			(s.highestTakeProfit == 3 && s.input.IfTP3StopAtTP2) ||
			(s.highestTakeProfit == 4 && s.input.IfTP4StopAtTP3) {
			s.stopLoss = common.JsonFloat64(s.priceCheckpoint)
		}
		s.priceCheckpoint = float64(tick.Price)
	}
	return false, nil
}

func getEnteredEvent(events []common.SignalCheckOutputEvent) (common.SignalCheckOutputEvent, bool) {
	for _, event := range events {
		if event.EventType == common.ENTERED {
			return event, true
		}
	}
	return common.SignalCheckOutputEvent{}, false
}

func calculateMaxEnterUSD(exchange common.Exchange, input common.SignalCheckInput, events []common.SignalCheckOutputEvent) (common.JsonFloat64, error) {
	enteredEvent, ok := getEnteredEvent(events)
	if !ok {
		return common.JsonFloat64(0.0), errors.New("this signal did not enter so cannot calculate maxEnterUSD")
	}
	usdPricePerBaseAsset, err := common.GetUSDPricePerBaseAssetUnitAtEvent(exchange, input, enteredEvent)
	if err != nil {
		return common.JsonFloat64(0.0), err
	}
	tradeIterator := exchange.BuildTradeIterator(input.BaseAsset, input.QuoteAsset, enteredEvent.At)
	maxTrade, err := tradeIterator.GetMaxBaseAssetEnter(5 /* minuteCount */, 10 /* bucketCount */, 100000 /* maxTradeCount */)
	if err != nil {
		return common.JsonFloat64(0.0), err
	}
	return maxTrade.BaseAssetPrice * usdPricePerBaseAsset * maxTrade.BaseAssetQuantity, nil
}

func doCheckSignal(input common.SignalCheckInput, nextTick func() (common.Tick, error)) (common.SignalCheckOutput, error) {
	var (
		checker     = newChecker(input)
		err         error
		isEnded     bool
		maxEnterUSD common.JsonFloat64
	)
	for {
		if isEnded, err = checker.applyTick(nextTick()); isEnded || err != nil {
			break
		}
	}
	if isEnded && (err == nil || err == common.ErrOutOfCandlesticks) {
		maxEnterUSD, err = calculateMaxEnterUSD(exchanges[input.Exchange], input, checker.events)
		if err != nil {
			log.Println(err)
		}
	}
	output := common.SignalCheckOutput{
		Input:      input,
		HttpStatus: 200,
	}
	if err != nil {
		output.IsError = true
		output.HttpStatus = 500
		output.ErrorMessage = err.Error()
	}
	return common.SignalCheckOutput{
		Events:               checker.events,
		Input:                input,
		Entered:              checker.entered,
		FirstCandleOpenPrice: checker.firstCandleOpenPrice,
		FirstCandleAt:        checker.firstCandleAt,
		HighestTakeProfit:    checker.highestTakeProfit,
		ReachedStopLoss:      checker.reachedStopLoss,
		ProfitRatio:          common.JsonFloat64(checker.profitCalculator.CalculateTakeProfitRatio()),
		MaxEnterUSD:          maxEnterUSD,
	}, err
}
