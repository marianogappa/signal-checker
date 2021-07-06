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
// - All dates are ISO3610 e.g. 2021-07-04T14:14:18+00:00
// - Durations are in seconds
// - All prices are floating point numbers for the given asset pair on the given exchange
//
// Parameters:
//
// - Exchange: must be "binance", default "binance"
// - CandlestickInterval: must be one of (for Binance, default 15m):
//   1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
// - BaseAsset: in LTCUSDT, baseAsset would be LTC
// - QuoteAsset: in LTCUSDT, quoteAsset would be USDT
// - EnterRangeLow: minimum (inclusive) value at which to enter signal (-1 for enter immediately)
// - EnterRangeHigh: maximum (inclusive) value at which to enter signal (-1 for enter immediately)
// - TakeProfits[]: values at which to take profits (empty for no take profits)
// - StopLoss: value at which to stop loss (-1 for no stop loss)
// - initialISO3601: datetime at which the signal becomes valid
// - invalidateISO3601: datetime at which the signal becomes invalid (empty for up to now)
// - InvalidateAfter: duration after which
type SignalCheckInput struct {
	Exchange               string        `json:"exchange"`
	CandlestickInterval    string        `json:"candlestickInterval"`
	BaseAsset              string        `json:"baseAsset"`
	QuoteAsset             string        `json:"quoteAsset"`
	EnterRangeLow          JsonFloat64   `json:"enterRangeLow"`
	EnterRangeHigh         JsonFloat64   `json:"enterRangeHigh"`
	TakeProfits            []JsonFloat64 `json:"takeProfits"`
	StopLoss               JsonFloat64   `json:"stopLoss"`
	InitialISO3601         string        `json:"initialISO3601"`
	InvalidateISO3601      string        `json:"invalidateISO3601"`
	InvalidateAfterSeconds int           `json:"invalidateAfterSeconds"`
	ReturnLogs             bool          `json:"returnLogs"`
	TakeProfitRatios       []JsonFloat64 `json:"takeProfitRatios"`
	IfTP1StopAtEntry       bool          `json:"ifTP1StopAtEntry"`
	IfTP2StopAtTP1         bool          `json:"ifTP2StopAtTP1"`
	IfTP3StopAtTP2         bool          `json:"ifTP3StopAtTP2"`
	IfTP4StopAtTP3         bool          `json:"ifTP4StopAtTP3"`
	// TODO for now only one request to Binance
	// TODO add invalidateIfTPBeforeEntering
	// TODO check if higher TPs reached first
}

const (
	ENTERED          = "entered"
	STOPPED_LOSS     = "stopped_loss"
	INVALIDATED      = "invalidated"
	FINISHED_DATASET = "finished_dataset"
	TAKEN_PROFIT_    = "taken_profit_"
)

// SignalCheckOutputEvent is an event that happened upon checking a signal.
//
// - eventType is one of entered, taken_profit_1, taken_profit_2, ..., stopped_loss, invalidated, finished_dataset
// - price is floating point number for the given asset pair on the given exchange
// - at is ISO8601 given by exchange API for candlestick start
type SignalCheckOutputEvent struct {
	EventType string      `json:"eventType"`
	Price     JsonFloat64 `json:"price"`
	At        string      `json:"at"`
}

// SignalCheckOutput is the result of the signal checking function.
type SignalCheckOutput struct {
	Input                SignalCheckInput         `json:"input"`
	Events               []SignalCheckOutputEvent `json:"events"`
	Entered              bool                     `json:"entered"`
	FirstCandleOpenPrice JsonFloat64              `json:"firstCandleOpenPrice"`
	FirstCandleAt        string                   `json:"firstCandleAt"`
	HighestTakeProfit    int                      `json:"highestTakeProfit"`
	ReachedStopLoss      bool                     `json:"reachedStopLoss"`
	// TODO cannot be done without takeProfitRatios
	ProfitRatio  JsonFloat64 `json:"profitRatio"`
	IsError      bool        `json:"isError"`
	HttpStatus   int         `json:"httpStatus"`
	ErrorMessage string      `json:"errorMessage"`
	// TODO No logs for now
	Logs []string `json:"logs"`
}

type JsonFloat64 float64

// Copied + small modifications from encoding/json/encode.go (1.8+).

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
