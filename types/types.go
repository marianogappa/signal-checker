// The types package contains the input and output types of the signal checking function.
package types

import (
	"errors"
	"math"
	"strconv"
)

// SignalCheckInput is the input to the signal checking function.
//
// Notes:
//
// - All dates are ISO3610 e.g. "2021-07-04T14:14:18+00:00".
// - Durations are in seconds.
// - All prices are floating point numbers for the given asset pair on the given exchange.
type SignalCheckInput struct {
	// Exchange must be one of ['binance', 'ftx', 'coinbase', 'huobi', 'kraken']; default is 'binance'
	Exchange string `json:"exchange"`

	// BaseAsset is LTC in LTCUSDT
	BaseAsset string `json:"baseAsset"`

	// QuoteAsset is USDT in LTCUSDT
	QuoteAsset string `json:"quoteAsset"`

	// EnterRangeLow is the minimum (inclusive) value at which to enter signal (-1 for enter immediately)
	EnterRangeLow JsonFloat64 `json:"enterRangeLow"`

	// EnterRangeHigh is the maximum (inclusive) value at which to enter signal (-1 for enter immediately)
	EnterRangeHigh JsonFloat64 `json:"enterRangeHigh"`

	// TakeProfits are the prices at which to take profits (empty for no take profits)
	TakeProfits []JsonFloat64 `json:"takeProfits"`

	// StopLoss is the price at which to stop loss (-1 for no stop loss)
	StopLoss JsonFloat64 `json:"stopLoss"`

	// InitialISO3601 is the ISO3601 datetime at which the signal becomes valid (e.g. 2021-07-04T14:14:18+00:00)
	InitialISO3601 string `json:"initialISO3601"`

	// InvalidateISO3601 is the ISO3601 datetime at which the signal becomes invalid (empty for up to now)
	// (e.g. 2021-07-04T14:14:18+00:00)
	// Considering a signal invalid, if entered, means "selling", either at a profit or at a loss.
	InvalidateISO3601 string `json:"invalidateISO3601"`

	// InvalidateAfterSeconds is the number of seconds from InitialISO3601 at which to consider the signal invalid.
	// Considering a signal invalid, if entered, means "selling", either at a profit or at a loss.
	InvalidateAfterSeconds int `json:"invalidateAfterSeconds"`

	// ReturnLogs decides whether to return logs on the output.
	ReturnLogs bool `json:"returnLogs"`

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

	// IfTP1StopAtEntry is a boolean that, if set, changes the stop loss to entry if TP1 is reached.
	IfTP1StopAtEntry bool `json:"ifTP1StopAtEntry"`

	// IfTP2StopAtTP1 is a boolean that, if set, changes the stop loss to TP1 if TP2 is reached.
	IfTP2StopAtTP1 bool `json:"ifTP2StopAtTP1"`

	// IfTP3StopAtTP2 is a boolean that, if set, changes the stop loss to TP2 if TP3 is reached.
	IfTP3StopAtTP2 bool `json:"ifTP3StopAtTP2"`

	// IfTP4StopAtTP3 is a boolean that, if set, changes the stop loss to TP3 if TP4 is reached.
	IfTP4StopAtTP3 bool `json:"ifTP4StopAtTP3"`

	// TODO for now only one request to Binance
	// TODO add invalidateIfTPBeforeEntering
}

const (
	ENTERED          = "entered"
	STOPPED_LOSS     = "stopped_loss"
	INVALIDATED      = "invalidated"
	FINISHED_DATASET = "finished_dataset"
	TAKEN_PROFIT_    = "taken_profit_"

	BINANCE  = "binance"
	FTX      = "ftx"
	COINBASE = "coinbase"
	HUOBI    = "huobi"
	KRAKEN   = "kraken"
)

// SignalCheckOutputEvent is an event that happened upon checking a signal.
type SignalCheckOutputEvent struct {
	// EventType is one of entered, taken_profit_1, taken_profit_2, ..., stopped_loss, invalidated, finished_dataset.
	EventType string `json:"eventType"`

	// Price is the floating point number for the given asset pair on the given exchange at the time of this event.
	Price JsonFloat64 `json:"price,omitempty"`

	// At is the ISO8601 datetime given by the exchange API at which this event happened.
	At string `json:"at,omitempty"`
}

// SignalCheckOutput is the result of the signal checking function.
type SignalCheckOutput struct {
	// Events is an ascendingly-ordered array of events that happened upon checking this signal.
	Events []SignalCheckOutputEvent `json:"events"`

	// Input is the input used for checking this signal.
	Input SignalCheckInput `json:"input"`

	// Entered is a boolean that answers if, upon checking this signal, the checker decided to buy the base asset.
	Entered bool `json:"entered"`

	// FirstCandleOpenPrice is the first price observed at the opening of the first checked candlestick.
	FirstCandleOpenPrice JsonFloat64 `json:"firstCandleOpenPrice"`

	// FirstCandleAt is the ISO8601 datetime at which the first checked candlestick opened.
	FirstCandleAt string `json:"firstCandleAt"`

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
	ErrorMessage string `json:"errorMessage"`

	// Logs returns logging information to debug the results. Logs is only returned when input.returnLogs is set.
	Logs []string `json:"logs"`
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
}

var ErrOutOfCandlesticks = errors.New("exchange ran out of candlesticks")

type JsonFloat64 float64

func (jf JsonFloat64) MarshalJSON() ([]byte, error) {
	f := float64(jf)
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, errors.New("unsupported value")
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		fmt = 'e'
	}
	b := strconv.AppendFloat(nil, f, fmt, -1, 64)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	return b, nil
}
