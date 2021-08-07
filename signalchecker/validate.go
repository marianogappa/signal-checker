package signalchecker

import (
	"sort"
	"strings"

	"github.com/marianogappa/signal-checker/common"
)

func invalidateWith(err error, input common.SignalCheckInput) (common.SignalCheckOutput, error) {
	return common.SignalCheckOutput{IsError: true, HttpStatus: 400, ErrorMessage: err.Error(), Input: input}, err
}

func sum(ss []common.JsonFloat64) float64 {
	sum := 0.0
	for _, s := range ss {
		sum += float64(s)
	}
	return sum
}

func validateInput(input common.SignalCheckInput) (common.SignalCheckOutput, error) {
	if input.BaseAsset == "" {
		return invalidateWith(common.ErrBaseAssetRequired, input)
	}
	if input.QuoteAsset == "" {
		return invalidateWith(common.ErrQuoteAssetRequired, input)
	}
	input.Exchange = strings.ToLower(input.Exchange)
	input.BaseAsset = strings.ToUpper(input.BaseAsset)
	input.QuoteAsset = strings.ToUpper(input.QuoteAsset)
	if len(input.EntryRatios) == 0 {
		input.EntryRatios = []common.JsonFloat64{1.0}
	}
	if sum(input.EntryRatios) != 1.0 {
		return invalidateWith(common.ErrEntryRatiosMustAddUpToOne, input)
	}
	if len(input.Entries) == 1 {
		return invalidateWith(common.ErrInvalidEntriesLength, input)
	}
	if !input.IsShort {
		sort.Slice(input.TakeProfits, func(i, j int) bool { return input.TakeProfits[i] < input.TakeProfits[j] })
		sort.Slice(input.Entries, func(i, j int) bool { return input.Entries[i] > input.Entries[j] })
	} else {
		sort.Slice(input.TakeProfits, func(i, j int) bool { return input.TakeProfits[i] > input.TakeProfits[j] })
		sort.Slice(input.Entries, func(i, j int) bool { return input.Entries[i] < input.Entries[j] })
	}
	if !input.IsShort && input.StopLoss != -1 && len(input.Entries) > 0 && input.StopLoss >= input.Entries[len(input.Entries)-1] {
		return invalidateWith(common.ErrStopLossIsGreaterThanOrEqualToEnterRangeLow, input)
	}
	if input.IsShort && input.StopLoss != -1 && len(input.Entries) > 0 && input.StopLoss <= input.Entries[len(input.Entries)-1] {
		return invalidateWith(common.ErrStopLossIsLessThanOrEqualToEnterRangeHigh, input)
	}
	if !input.IsShort && len(input.Entries) > 0 && len(input.TakeProfits) > 0 && input.TakeProfits[0] <= input.Entries[0] {
		return invalidateWith(common.ErrFirstTPIsLessThanOrEqualToEnterRangeHigh, input)
	}
	if input.IsShort && len(input.Entries) > 0 && len(input.TakeProfits) > 0 && input.TakeProfits[0] >= input.Entries[0] {
		return invalidateWith(common.ErrFirstTPIsGreaterThanOrEqualToEnterRangeLow, input)
	}
	if input.Exchange == "" {
		input.Exchange = "binance"
	}
	if input.Exchange != "binance" && input.Exchange != "ftx" && input.Exchange != "coinbase" &&
		input.Exchange != "huobi" && input.Exchange != "kraken" && input.Exchange != "kucoin" &&
		input.Exchange != "binanceusdmfutures" && input.Exchange != "fake" {
		return invalidateWith(common.ErrInvalidExchange, input)
	}
	if input.InitialISO8601 == "" {
		return invalidateWith(common.ErrInitialISO8601Required, input)
	}
	if _, err := input.InitialISO8601.Time(); err != nil {
		return invalidateWith(common.ErrInitialISO8601FormattedIncorrectly, input)
	}
	if _, err := input.InvalidateISO8601.Time(); input.InvalidateISO8601 != "" && err != nil {
		return invalidateWith(common.ErrInvalidateISO8601FormattedIncorrectly, input)
	}
	if len(input.TakeProfitRatios) > 0 && sum(input.TakeProfitRatios) != 1.0 {
		return invalidateWith(common.ErrTakeProfitRatiosMustAddUpToOne, input)
	}
	return common.SignalCheckOutput{Input: input}, nil
}
