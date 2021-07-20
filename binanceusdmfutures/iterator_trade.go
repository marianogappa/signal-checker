package binanceusdmfutures

import "github.com/marianogappa/signal-checker/common"

type binanceTradeIterator struct {
	binance                           BinanceUSDMFutures
	baseAsset, quoteAsset             string
	trades                            []common.Trade
	requestFromMillis, initialSeconds int
}

func (b BinanceUSDMFutures) newTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *binanceTradeIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	initialSeconds := int(initial.Unix())
	return &binanceTradeIterator{
		baseAsset:         baseAsset,
		quoteAsset:        quoteAsset,
		requestFromMillis: initialSeconds * 1000,
		initialSeconds:    initialSeconds,
	}
}

func (it *binanceTradeIterator) next() (common.Trade, error) {
	if len(it.trades) > 0 {
		c := it.trades[0]
		it.trades = it.trades[1:]
		return c, nil
	}
	aggTradesResult, err := it.binance.getTrades(it.baseAsset, it.quoteAsset, it.requestFromMillis)
	if err != nil {
		return common.Trade{}, err
	}
	it.trades = aggTradesResult.trades
	if len(it.trades) == 0 {
		return common.Trade{}, common.ErrOutOfTrades
	}
	// Some exchanges return earlier trades to the requested time. Prune them.
	// Note that this may remove all items, but this does not necessarily mean we are out of trades.
	// In this case we just need to fetch again.
	for len(it.trades) > 0 && it.trades[0].Timestamp < it.initialSeconds {
		it.trades = it.trades[1:]
	}
	if len(it.trades) > 0 {
		it.requestFromMillis = it.trades[len(it.trades)-1].Timestamp*1000 + 1
	}
	return it.next()
}
