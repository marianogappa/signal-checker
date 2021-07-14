package kraken

import (
	"github.com/marianogappa/signal-checker/common"
)

type krakenCandlestickIterator struct {
	baseAsset, quoteAsset string
	candlesticks          []common.Candlestick
	requestFromSecs       int
}

func newKrakenCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *krakenCandlestickIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &krakenCandlestickIterator{
		baseAsset:       baseAsset,
		quoteAsset:      quoteAsset,
		requestFromSecs: int(initial.Unix()),
	}
}

func (it *krakenCandlestickIterator) next() (common.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		c := it.candlesticks[0]
		it.candlesticks = it.candlesticks[1:]
		return c, nil
	}
	klinesResult, err := getKlines(it.baseAsset, it.quoteAsset, it.requestFromSecs)
	if err != nil {
		return common.Candlestick{}, err
	}
	it.candlesticks = klinesResult.candlesticks
	if len(it.candlesticks) <= 1 {
		return common.Candlestick{}, common.ErrOutOfCandlesticks
	}
	// Some exchanges return earlier candlesticks to the requested time. Prune them.
	// Note that this may remove all items, but this does not necessarily mean we are out of candlesticks.
	// In this case we just need to fetch again.
	for len(it.candlesticks) > 0 && it.candlesticks[0].Timestamp < it.requestFromSecs {
		it.candlesticks = it.candlesticks[1:]
	}
	it.requestFromSecs = klinesResult.nextSince
	return it.next()
}
