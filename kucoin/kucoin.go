package kucoin

import (
	"github.com/marianogappa/signal-checker/common"
)

type Kucoin struct{}

func NewKucoin() Kucoin {
	return Kucoin{}
}

func (b Kucoin) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(newKucoinCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b Kucoin) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(newKucoinTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
