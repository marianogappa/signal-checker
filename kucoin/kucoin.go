package kucoin

import (
	"github.com/marianogappa/signal-checker/common"
)

type Kucoin struct {
	apiURL string
	debug  bool
}

func NewKucoin() *Kucoin {
	return &Kucoin{apiURL: "https://api.kucoin.com/api/v1/"}
}

func (k *Kucoin) overrideAPIURL(apiURL string) {
	k.apiURL = apiURL
}

func (b *Kucoin) SetDebug(debug bool) {
	b.debug = debug
}

func (k Kucoin) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(k.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (k Kucoin) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(k.newTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
