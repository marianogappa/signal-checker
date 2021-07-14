package coinbase

import (
	"github.com/marianogappa/signal-checker/common"
)

type Coinbase struct{}

func NewCoinbase() Coinbase {
	return Coinbase{}
}

func (b Coinbase) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(newCoinbaseCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b Coinbase) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(newCoinbaseTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
