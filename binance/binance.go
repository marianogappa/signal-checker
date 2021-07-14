package binance

import (
	"github.com/marianogappa/signal-checker/common"
)

type Binance struct{}

func NewBinance() Binance {
	return Binance{}
}

func (b Binance) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(newBinanceCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b Binance) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(newBinanceTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}

const ERR_INVALID_SYMBOL = -1121
