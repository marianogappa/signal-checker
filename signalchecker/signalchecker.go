// The signalchecker package contains the CheckSignal function that does the signal checking.
//
// Please review the docs on the common.SignalCheckInput and common.SignalCheckOutput common.
//
// If you are importing the package, use it like this:
//
// output, err := signalchecker.NewSignalChecker(input).Check()
//
// Note that the output contains richer information about the error than err itself, but you can still use err if it
// reads better in your code.
//
// Consider using this signal checker via de cli binary or as a server, which is also started from the binary.
// Latest release available at: https://github.com/marianogappa/signal-checker/releases
// You can also build locally using: go get github.com/marianogappa/signal-checker
package signalchecker

import (
	"log"
	"time"

	"github.com/marianogappa/signal-checker/binance"
	"github.com/marianogappa/signal-checker/binanceusdmfutures"
	"github.com/marianogappa/signal-checker/coinbase"
	"github.com/marianogappa/signal-checker/common"
	"github.com/marianogappa/signal-checker/fake"
	"github.com/marianogappa/signal-checker/ftx"
	"github.com/marianogappa/signal-checker/kraken"
	"github.com/marianogappa/signal-checker/kucoin"
	"github.com/marianogappa/signal-checker/profitcalculator"
)

var (
	exchanges = map[string]common.Exchange{
		common.BINANCE:              binance.NewBinance(),
		common.FTX:                  ftx.NewFTX(),
		common.COINBASE:             coinbase.NewCoinbase(),
		common.KRAKEN:               kraken.NewKraken(),
		common.KUCOIN:               kucoin.NewKucoin(),
		common.BINANCE_USDM_FUTURES: binanceusdmfutures.NewBinanceUSDMFutures(),
	}
)

// SignalChecker is the main struct does that the signal checking.
// Use it like this: output, err := signalchecker.NewSignalChecker(input).Check()
// Please review the docs on the common.SignalCheckInput and common.SignalCheckOutput.
type SignalChecker struct {
	input    common.SignalCheckInput
	exchange common.Exchange

	// For testing
	mockCandlesticks []common.Candlestick
	mockTrades       []common.Trade
	mockReturnErr    error
}

// NewSignalChecker is the constructor for SignalChecker.
// Use it like this: output, err := signalchecker.NewSignalChecker(input).Check()
// Please review the docs on the common.SignalCheckInput and common.SignalCheckOutput.
func NewSignalChecker(input common.SignalCheckInput) *SignalChecker {
	return &SignalChecker{input: input}
}

// Check is the main method on SignalChecker that does the actual checking.
// Use it like this: output, err := signalchecker.NewSignalChecker(input).Check()
// Please review the docs on the common.SignalCheckInput and common.SignalCheckOutput.
func (c SignalChecker) Check() (common.SignalCheckOutput, error) {
	validationResult, err := validateInput(c.input)
	if err != nil {
		return validationResult, err
	}
	c.input = validationResult.Input
	if c.input.Debug {
		log.Printf("Input validation ok. Input: %+v\n", c.input)
	}
	c.exchange = exchanges[c.input.Exchange]

	if c.mockCandlesticks != nil || c.mockTrades != nil {
		c.exchange = fake.NewFake(c.mockCandlesticks, c.mockTrades, c.mockReturnErr)
	}
	c.exchange.SetDebug(c.input.Debug)

	return c.doCheck()
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
	reachedStopLoss      bool
	highestTakeProfit    int
	highestEntry         int
	firstCandleOpenPrice common.JsonFloat64
	firstCandleAt        common.ISO8601
	invalidAt            time.Time
	hasInvalidAt         bool
	events               []common.SignalCheckOutputEvent
	stopLoss             common.JsonFloat64
	initialTime          time.Time
	priceCheckpoint      float64
	isEnded              bool
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
func (s *checkSignalState) applyEvent(eventType string, target int, tick common.Tick) bool {
	event := common.SignalCheckOutputEvent{EventType: eventType}
	event.Target = target
	event.At = common.ISO8601(time.Unix(int64(tick.Timestamp), 0).UTC().Format(time.RFC3339))
	event.Price = tick.Price
	event.ProfitRatio = common.JsonFloat64(s.profitCalculator.ApplyEvent(event))
	s.events = append(s.events, event)
	s.isEnded = eventType == common.FINISHED_DATASET || eventType == common.STOPPED_LOSS || s.profitCalculator.IsFinished()
	return s.isEnded
}

func (s *checkSignalState) applyTick(tick common.Tick, err error) (bool, error) {
	if err == common.ErrOutOfCandlesticks {
		return s.applyEvent(common.FINISHED_DATASET, 0, tick), err
	}
	if err != nil {
		return true, err
	}
	tickTime := time.Unix(int64(tick.Timestamp), 0)

	// Ignore candlesticks before the signal's initial time.
	if tickTime.Before(s.initialTime) {
		return false, nil
	}

	// Save the first read candlestick WITHIN the signal's initial time (the first tick is the open price of the first
	// candlestick).
	if s.first {
		s.first = false
		s.firstCandleOpenPrice = tick.Price
		s.firstCandleAt = common.ISO8601(tickTime.UTC().Format(time.RFC3339))
	}

	// If the tick's time is >= the invalidation time, finish here.
	if s.hasInvalidAt && (tickTime.After(s.invalidAt) || tickTime.Equal(s.invalidAt)) {
		return s.applyEvent(common.INVALIDATED, 0, tick), nil
	}

	// If we haven't entered yet, or there are multiple entries and we're able to enter further, calculate so
	if (s.highestEntry == 0 && len(s.input.Entries) == 0) ||
		(len(s.input.Entries) >= s.highestEntry+2 && ((!s.input.IsShort && tick.Price >= s.input.Entries[s.highestEntry+1] && tick.Price < s.input.Entries[s.highestEntry]) ||
			(s.input.IsShort && tick.Price > s.input.Entries[s.highestEntry] && tick.Price <= s.input.Entries[s.highestEntry+1]))) {

		// Go backwards from furthest possible remaining entry, and enter the first range that the price is in
		for i := len(s.input.Entries) - 1; i >= s.highestEntry+1; i-- {
			if (!s.input.IsShort && tick.Price >= s.input.Entries[i] && tick.Price < s.input.Entries[i-1]) ||
				(s.input.IsShort && tick.Price <= s.input.Entries[i] && tick.Price > s.input.Entries[i-1]) {
				s.highestEntry = i
				break
			}
		}

		// If there are no entries at all, this must be the first and only entry
		if len(s.input.Entries) == 0 {
			s.highestEntry = 1
		}
		return s.applyEvent(common.ENTERED, s.highestEntry, tick), nil
	}

	// If we entered, and price <= stopLoss (for LONG) or >= stopLoss (for SHORT), then we reached stop loss.
	if s.highestEntry > 0 && ((!s.input.IsShort && tick.Price <= s.stopLoss) || (s.input.IsShort && tick.Price >= s.stopLoss)) {
		s.reachedStopLoss = true
		return s.applyEvent(common.STOPPED_LOSS, 0, tick), nil
	}

	// If we have entered and there are TPs and we're able to take profit further, calculate so
	if s.highestEntry > 0 && s.highestTakeProfit < len(s.input.TakeProfits) &&
		((!s.input.IsShort && tick.Price >= s.input.TakeProfits[s.highestTakeProfit]) || (s.input.IsShort && tick.Price <= s.input.TakeProfits[s.highestTakeProfit])) {

		// Go backwards from furthest possible TP, and take profit on the first range that the price is in
		for i := len(s.input.TakeProfits) - 1; i >= s.highestTakeProfit; i-- {
			if (!s.input.IsShort && tick.Price < s.input.TakeProfits[i]) || (s.input.IsShort && tick.Price > s.input.TakeProfits[i]) {
				continue
			}
			s.highestTakeProfit = i + 1
			break
		}
		s.applyEvent(common.TOOK_PROFIT, s.highestTakeProfit, tick)
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

func (c SignalChecker) doCheck() (common.SignalCheckOutput, error) {
	var (
		candlestickIterator = c.exchange.BuildCandlestickIterator(c.input.BaseAsset, c.input.QuoteAsset, c.input.InitialISO8601)
		checker             = newChecker(c.input)
		err                 error
		isEnded             bool
		maxEnterUSD         common.JsonFloat64
		nextTick            = buildTickIterator(candlestickIterator.Next)
	)
	if c.input.ReturnCandlesticks {
		candlestickIterator.SaveCandlesticks()
	}
	for {
		if isEnded, err = checker.applyTick(nextTick()); isEnded || err != nil {
			break
		}
	}
	if isEnded && (err == nil || err == common.ErrOutOfCandlesticks) && !c.input.DontCalculateMaxEnterUSD {
		maxEnterUSD, err = calculateMaxEnterUSD(c.exchange, c.input, checker.events)
		if err != nil {
			log.Println(err)
		}
	}
	output := common.SignalCheckOutput{
		Input:      c.input,
		HttpStatus: 200,
	}
	if err != nil && err != common.ErrOutOfCandlesticks {
		output.IsError = true
		output.HttpStatus = 500
		output.ErrorMessage = err.Error()
	}
	output.Events = checker.events
	output.Input = c.input
	output.Entered = checker.highestEntry > 0
	output.HighestEntry = checker.highestEntry
	output.FirstCandleOpenPrice = checker.firstCandleOpenPrice
	output.FirstCandleAt = checker.firstCandleAt
	output.HighestTakeProfit = checker.highestTakeProfit
	output.ReachedStopLoss = checker.reachedStopLoss
	output.ProfitRatio = common.JsonFloat64(checker.profitCalculator.CalculateTakeProfitRatio())
	output.MaxEnterUSD = maxEnterUSD
	output.Candlesticks = candlestickIterator.SavedCandlesticks
	return output, err
}
