package binanceusdmfutures

import "github.com/marianogappa/signal-checker/common"

type binanceCandlestickIterator struct {
	baseAsset, quoteAsset string
	candlesticks          []common.Candlestick
	requestFromMillis     int
	initialSeconds        int
}

func newBinanceCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *binanceCandlestickIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	initialSeconds := int(initial.Unix())
	return &binanceCandlestickIterator{
		baseAsset:         baseAsset,
		quoteAsset:        quoteAsset,
		requestFromMillis: initialSeconds * 1000,
		initialSeconds:    initialSeconds,
	}
}

func (it *binanceCandlestickIterator) next() (common.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		c := it.candlesticks[0]
		it.candlesticks = it.candlesticks[1:]
		return c, nil
	}
	klinesResult, err := getKlines(it.baseAsset, it.quoteAsset, it.requestFromMillis)
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
	for len(it.candlesticks) > 0 && it.candlesticks[0].Timestamp < it.initialSeconds {
		it.candlesticks = it.candlesticks[1:]
	}
	if len(it.candlesticks) > 0 {
		it.requestFromMillis = (it.candlesticks[len(it.candlesticks)-1].Timestamp + 60) * 1000
	}
	return it.next()
}
