package binanceusdmfutures

import (
	"github.com/marianogappa/signal-checker/common"
)

type BinanceUSDMFutures struct {
	apiURL string
}

func NewBinanceUSDMFutures() *BinanceUSDMFutures {
	return &BinanceUSDMFutures{apiURL: "https://api.binance.com/api/v3/"}
}

func (b *BinanceUSDMFutures) overrideAPIURL(url string) {
	b.apiURL = url
}

func (b BinanceUSDMFutures) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(b.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b BinanceUSDMFutures) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(b.newTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}

const ERR_INVALID_SYMBOL = -1121
