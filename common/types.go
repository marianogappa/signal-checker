// The common package contains the input and output types of the signal checking function.
package common

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// SignalCheckInput is the input to the signal checking function.
//
// Notes:
//
// - All dates are ISO3610 e.g. "2021-07-04T14:14:18+00:00".
// - Durations are in seconds.
// - All prices are floating point numbers for the given asset pair on the given exchange.
type SignalCheckInput struct {
	// Exchange must be one of ['binance', 'ftx', 'coinbase', 'huobi', 'kraken', 'kucoin', 'binanceusdmfutures']; default is 'binance'
	Exchange string `json:"exchange"`

	// BaseAsset is LTC in LTCUSDT
	BaseAsset string `json:"baseAsset"`

	// QuoteAsset is USDT in LTCUSDT
	QuoteAsset string `json:"quoteAsset"`

	// Entries are the price ranges at which the checker should "buy in".
	//
	// There may be an arbitrary number of ranges, and ranges are subsequent to each other.
	//
	// Entries ranges may be specified in any order, as the order will be informed by whether the signal is a
	// LONG/SHORT, but it's recommended to match the EntryRatios order, because it will be confusing to you otherwise.
	//
	// Any overlap with TP/SL will be considered an error.
	//
	// Entries can be an empty array or nil, in which case the check "buys in" immediately.
	//
	// e.g. to enter immediately:                                             entries: []
	// e.g. to enter 100% between 0.1 and 0.5:                                entries: [0.1, 0.5]
	// e.g. to enter some between 0.5 and 0.3, and more between 0.3 and 0.1:  entries: [0.5, 0.3, 0.1]
	Entries []JsonFloat64 `json:"entries"`

	// EntryRatios are the ratios (i.e. array of 0 to 1) with respect to the capital to invest in this signal that the
	// checker should "buy in" with at each of the entry ranges.
	//
	// It defaults to [1], that is, to enter fully at the first entry.
	//
	// Ratios must be specified in the order that they would be entered, and they must add up to 1.
	//
	// e.g. to enter with 25% of capital at Entry 1 and 75% at Entry 2:  entryRatios: [0.25, 0.75]
	//
	// If the checker encounters Entry 2 before Entry 1, it will enter with the cumulative ratio of both entries.
	EntryRatios []JsonFloat64 `json:"entryRatios"`

	// IsShort defines if this signal is for a LONG or a SHORT. Defaults to LONG.
	IsShort bool `json:"isShort"`

	// TakeProfits are the prices at which to take profits (empty for no take profits)
	TakeProfits []JsonFloat64 `json:"takeProfits"`

	// TakeProfitRatios is used to calculate profitRatio, that is, how much would have been the profit/loss of
	// following the signal.
	// A value of [0.25, 0.5, 0.25] means taking out 25% of the total entered at TP1, 75% of the remaining at TP2 and
	// all the remaining at TP3.
	// TakeProfitRatios needn't match the array size of TakeProfits, e.g. [1.0] would be a valid value that takes all
	// on TP1, ignoring other TPs.
	// TakeProfitRatios values must sum up to 1.0.
	// If TakeProfitRatios is set so that everything is taken out early in the signal (e.g. TP1 on a signal with 3 TPs),
	// signal-checker will stop evaluating candlesticks once everything was taken out rather than checking if further
	// TPs were reached.
	TakeProfitRatios []JsonFloat64 `json:"takeProfitRatios"`

	// StopLoss is the price at which to stop loss (-1 for no stop loss)
	StopLoss JsonFloat64 `json:"stopLoss"`

	// InitialISO8601 is the ISO3601 datetime at which the signal becomes valid (e.g. 2021-07-04T14:14:18+00:00)
	InitialISO8601 ISO8601 `json:"initialISO8601"`

	// InvalidateISO8601 is the ISO3601 datetime at which the signal becomes invalid (empty for up to now)
	// (e.g. 2021-07-04T14:14:18+00:00)
	// Considering a signal invalid, if entered, means "selling", either at a profit or at a loss.
	InvalidateISO8601 ISO8601 `json:"invalidateISO8601"`

	// InvalidateAfterSeconds is the number of seconds from InitialISO8601 at which to consider the signal invalid.
	// Considering a signal invalid, if entered, means "selling", either at a profit or at a loss.
	InvalidateAfterSeconds int `json:"invalidateAfterSeconds"`

	// ReturnLogs decides whether to return logs on the output.
	ReturnLogs bool `json:"returnLogs"`

	// Debug decides whether to turn debug mode on, which means verbose stderr output.
	Debug bool `json:"debug"`

	// IfTP1StopAtEntry is a boolean that, if set, changes the stop loss to entry if TP1 is reached.
	IfTP1StopAtEntry bool `json:"ifTP1StopAtEntry"`

	// IfTP2StopAtTP1 is a boolean that, if set, changes the stop loss to TP1 if TP2 is reached.
	IfTP2StopAtTP1 bool `json:"ifTP2StopAtTP1"`

	// IfTP3StopAtTP2 is a boolean that, if set, changes the stop loss to TP2 if TP3 is reached.
	IfTP3StopAtTP2 bool `json:"ifTP3StopAtTP2"`

	// IfTP4StopAtTP3 is a boolean that, if set, changes the stop loss to TP3 if TP4 is reached.
	IfTP4StopAtTP3 bool `json:"ifTP4StopAtTP3"`

	// DontCalculateMaxEnterUSD prevents calculation of MaxEnterUSD, which can be expensive and lengthy.
	DontCalculateMaxEnterUSD bool `json:"dontCalculateMaxEnterUSD"`

	// ReturnCandlesticks decides if all input candlesticks should be returned with the output. This could span MBs,
	// so should only be set when needed, e.g. to plot a candlestick chart.
	ReturnCandlesticks bool `json:"returnCandlesticks"`
	// TODO add invalidateIfTPBeforeEntering
}

const (
	ENTERED          = "entered"
	STOPPED_LOSS     = "stopped_loss"
	INVALIDATED      = "invalidated"
	FINISHED_DATASET = "finished_dataset"
	TOOK_PROFIT      = "took_profit"

	BINANCE              = "binance"
	FTX                  = "ftx"
	COINBASE             = "coinbase"
	HUOBI                = "huobi"
	KRAKEN               = "kraken"
	KUCOIN               = "kucoin"
	BINANCE_USDM_FUTURES = "binanceusdmfutures"

	// Used for testing
	FAKE = "fake"
)

// SignalCheckOutputEvent is an event that happened upon checking a signal.
type SignalCheckOutputEvent struct {
	// EventType is one of entered, took_profit, stopped_loss, invalidated, finished_dataset.
	EventType string `json:"eventType"`

	// Target is, in the case of 'entered' and 'took_profit', which entry or take profit target, e.g. TP1, TP2.
	Target int `json:"target,omitempty"`

	// Price is the floating point number for the given asset pair on the given exchange at the time of this event.
	Price JsonFloat64 `json:"price"`

	// At is the ISO8601 datetime given by the exchange API at which this event happened.
	At ISO8601 `json:"at"`

	// ProfitRatio answers how much the profit/loss of this signal is up to this point.
	ProfitRatio JsonFloat64 `json:"takeProfitRatio"`
}

type ISO8601 string

func (t ISO8601) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(t))
}

func (t ISO8601) Seconds() (int, error) {
	tm, err := t.Time()
	if err != nil {
		return 0, fmt.Errorf("failed to convert %v to seconds because %v", string(t), err.Error())
	}
	return int(tm.Unix()), nil
}

func (t ISO8601) Millis() (int, error) {
	tm, err := t.Seconds()
	if err != nil {
		return 0, err
	}
	return tm * 100, nil
}

// SignalCheckOutput is the result of the signal checking function.
type SignalCheckOutput struct {
	// Events is an ascendingly-ordered array of events that happened upon checking this signal.
	Events []SignalCheckOutputEvent `json:"events"`

	// Input is the input used for checking this signal.
	Input SignalCheckInput `json:"input"`

	// Entered is a boolean that answers if, upon checking this signal, the checker decided to buy the base asset.
	Entered bool `json:"entered"`

	// HighestEntry is the highest entry that, upon checking this signal, the checker decided to buy the base asset.
	HighestEntry int `json:"highestEntry"`

	// FirstCandleOpenPrice is the first price observed at the opening of the first checked candlestick which was
	// in or after the signal's initial time.
	FirstCandleOpenPrice JsonFloat64 `json:"firstCandleOpenPrice"`

	// FirstCandleAt is the ISO8601 datetime at which the first checked candlestick (which was in or after the signal's
	// initial time) opened.
	FirstCandleAt ISO8601 `json:"firstCandleAt"`

	// HighestTakeProfit is the highest take profit reached from the signal input (0 means none were reached).
	HighestTakeProfit int `json:"highestTakeProfit"`

	// ReachedStopLoss is a boolean that answers if, upon checking this signal, a stop loss was reached.
	ReachedStopLoss bool `json:"reachedStopLoss"`

	// ProfitRatio answers how much the profit/loss of following this signal would have been.
	// A profit ratio of 0.0 means break even. A profit ratio of 1.0 means doubling your investment.
	// To calculate how much you would profit, multiply your investment times the profit ratio.
	ProfitRatio JsonFloat64 `json:"profitRatio"`

	// IsError is a boolean that answers if there was any error checking this signal. This boolean should always be
	// checked first, because if it is true, all other output values are meaningless, except for the ones that describe
	// the error.
	// Note that there are a variety of reasons why checking a signal would fail, so it's important to review the
	// errorMessage output value. For example, your input request could contain incorrect parameters, the exchange
	// might be having an outage, you could have triggered a temporary rate-limiting, there could be a bug in this
	// codebase.
	// If you suspect the cause of the error is a bug in this codebase, feel free to create an issue.
	IsError bool `json:"isError"`

	// HttpStatus returns the http status code that would have been returned if this check would have been done over
	// HTTP. This value can be used to get slightly more precise (but still automatable) information about the error.
	// For example, if httpStatus equals 400, a script should not attempt to retry, whereas if it is 429, it should
	// sleep and retry.
	HttpStatus int `json:"httpStatus"`

	// ErrorMessage returns a human-readable description of the error that occurred while checking the signal.
	ErrorMessage string `json:"errorMessage,omitempty"`

	// Logs returns logging information to debug the results. Logs is only returned when input.returnLogs is set.
	Logs []string `json:"logs,omitempty"`

	MaxEnterUSD JsonFloat64 `json:"maxEnterUSD,omitempty"`

	Candlesticks []Candlestick `json:"candlesticks,omitempty"`
}

// Candlestick is the generic struct for candlestick data for all supported exchanges.
type Candlestick struct {
	// Timestamp is the UNIX timestamp (i.e. seconds since UTC Epoch) at which the candlestick started.
	Timestamp int `json:"t"`

	// OpenPrice is the price at which the candlestick opened.
	OpenPrice JsonFloat64 `json:"o"`

	// ClosePrice is the price at which the candlestick closed.
	ClosePrice JsonFloat64 `json:"c"`

	// LowestPrice is the lowest price reached during the candlestick duration.
	LowestPrice JsonFloat64 `json:"l"`

	// HighestPrice is the highest price reached during the candlestick duration.
	HighestPrice JsonFloat64 `json:"h"`

	// Volume is the traded volume in base asset during this candlestick.
	Volume JsonFloat64 `json:"v"`

	// NumberOfTrades is the total number of filled order book orders in this candlestick.
	NumberOfTrades int `json:"n,omitempty"`
}

// ToTicks converts a Candlestick to two Ticks. Lowest value is put first, because since there's no way to tell
// which one happened first, this library chooses to be pessimistic.
func (c Candlestick) ToTicks() []Tick {
	return []Tick{
		{Timestamp: c.Timestamp, Volume: c.Volume, NumberOfTrades: c.NumberOfTrades, Price: c.LowestPrice},
		{Timestamp: c.Timestamp, Volume: c.Volume, NumberOfTrades: c.NumberOfTrades, Price: c.HighestPrice},
	}
}

// Tick is either one side of the candlestick, or the return type of a Ticker.
type Tick struct {
	// Timestamp is the UNIX timestamp (i.e. seconds since UTC Epoch) at which tick happened.
	Timestamp int `json:"t"`

	// Price is the price of this tick.
	Price JsonFloat64 `json:"p"`

	// Volume (may not exist) is the traded volume in base asset during the candlestick of this tick.
	Volume JsonFloat64 `json:"v,omitempty"`

	// NumberOfTrades (may not exist) is the total number of filled order book orders in the candlestick of this tick.
	NumberOfTrades int `json:"n,omitempty"`
}

// Trade is an order filled on an exchange for some price & quantity of a base asset.
type Trade struct {
	// BaseAssetPrice is the price of the base asset at which this trade was filled.
	BaseAssetPrice JsonFloat64

	// BaseAssetQuantity is the quantity of the base asset at which this trade was filled.
	BaseAssetQuantity JsonFloat64

	// Timestamp is the UNIX timestamp (i.e. seconds since UTC Epoch) at which this trade was filled.
	Timestamp int `json:"t"`
}

var (
	ErrOutOfCandlesticks                           = errors.New("exchange ran out of candlesticks")
	ErrOutOfTrades                                 = errors.New("exchange ran out of trades")
	ErrInvalidMarketPair                           = errors.New("market pair does not exist on exchange")
	ErrRateLimit                                   = errors.New("exchange asked us to enhance our calm")
	ErrInvalidEntriesLength                        = errors.New("entries must either be empty or have two values or more (because a range is made of at least 2 numbers)")
	ErrEntryRatiosMustAddUpToOne                   = errors.New("entryRatios must add up to 1")
	ErrStopLossIsGreaterThanOrEqualToEnterRangeLow = errors.New("stopLoss is >= enterRangeLow; if you want no stopLoss, set the value to -1")
	ErrStopLossIsLessThanOrEqualToEnterRangeHigh   = errors.New("stopLoss is <= enterRangeHigh; if you want no stopLoss, set the value to -1")
	ErrFirstTPIsLessThanOrEqualToEnterRangeHigh    = errors.New("first take profit is <= enterRangeHigh")
	ErrFirstTPIsGreaterThanOrEqualToEnterRangeLow  = errors.New("first take profit is >= enterRangeLow")
	ErrInvalidExchange                             = errors.New("the only valid exchanges are 'binance', 'ftx', 'coinbase', 'huobi', 'kraken', 'kucoin' and 'binanceusdmfutures'")
	ErrInitialISO8601Required                      = errors.New("InitialISO8601 is required")
	ErrInitialISO8601FormattedIncorrectly          = errors.New("InitialISO8601 is formatted incorrectly, should be ISO3601 e.g. 2021-07-04T14:14:18+00:00")
	ErrInvalidateISO8601FormattedIncorrectly       = errors.New("InvalidateISO8601 is formatted incorrectly, should be ISO3601 e.g. 2021-07-04T14:14:18+00:00")
	ErrTakeProfitRatiosMustAddUpToOne              = errors.New("takeProfitRatios must add up to 1 (but it does not need to match the takeProfits length)")
	ErrBaseAssetRequired                           = errors.New("base asset is required (e.g. BTC)")
	ErrQuoteAssetRequired                          = errors.New("quote asset is required (e.g. USDT)")
)

type JsonFloat64 float64

func (jf JsonFloat64) MarshalJSON() ([]byte, error) {
	f := float64(jf)
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, errors.New("unsupported value")
	}
	bs := []byte(fmt.Sprintf("%.12f", f))
	var i int
	for i = len(bs) - 1; i >= 0; i-- {
		if bs[i] == '0' {
			continue
		}
		if bs[i] == '.' {
			return bs[:i], nil
		}
		break
	}
	return bs[:i+1], nil
}

type Exchange interface {
	BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 ISO8601) *CandlestickIterator
	BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 ISO8601) *TradeIterator
	SetDebug(debug bool)
}
