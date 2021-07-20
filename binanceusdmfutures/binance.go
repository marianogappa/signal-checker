package binanceusdmfutures

import (
	"github.com/marianogappa/signal-checker/common"
)

type BinanceUSDMFutures struct{}

func NewBinanceUSDMFutures() BinanceUSDMFutures {
	return BinanceUSDMFutures{}
}

func (b BinanceUSDMFutures) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(newBinanceCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b BinanceUSDMFutures) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(newBinanceTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}

const ERR_INVALID_SYMBOL = -1121
