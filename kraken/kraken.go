package kraken

import (
	"github.com/marianogappa/signal-checker/common"
)

type Kraken struct{}

func NewKraken() Kraken {
	return Kraken{}
}

func (b Kraken) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(newKrakenCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b Kraken) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(newKrakenTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
