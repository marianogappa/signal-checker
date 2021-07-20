package kucoin

import (
	"github.com/marianogappa/signal-checker/common"
)

type kucoinCandlestickIterator struct {
	kucoin                Kucoin
	baseAsset, quoteAsset string
	candlesticks          []common.Candlestick
	requestFromSecs       int
}

func (k Kucoin) newCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *kucoinCandlestickIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &kucoinCandlestickIterator{
		kucoin:          k,
		baseAsset:       baseAsset,
		quoteAsset:      quoteAsset,
		requestFromSecs: int(initial.Unix()),
	}
}

func (it *kucoinCandlestickIterator) next() (common.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		// N.B. KuCoin returns data in descending order
		c := it.candlesticks[len(it.candlesticks)-1]
		it.candlesticks = it.candlesticks[:len(it.candlesticks)-1]
		return c, nil
	}
	klinesResult, err := it.kucoin.getKlines(it.baseAsset, it.quoteAsset, it.requestFromSecs)
	if err != nil {
		return common.Candlestick{}, err
	}
	it.candlesticks = klinesResult.candlesticks
	if len(it.candlesticks) == 0 {
		return common.Candlestick{}, common.ErrOutOfCandlesticks
	}
	// Some exchanges return earlier candlesticks to the requested time. Prune them.
	// Note that this may remove all items, but this does not necessarily mean we are out of candlesticks.
	// In this case we just need to fetch again.
	for len(it.candlesticks) > 0 && it.candlesticks[0].Timestamp < it.requestFromSecs {
		it.candlesticks = it.candlesticks[1:]
	}
	if len(it.candlesticks) > 0 {
		it.requestFromSecs = it.candlesticks[0].Timestamp + 60
	}

	return it.next()
}
