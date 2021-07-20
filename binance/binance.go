package binance

import (
	"github.com/marianogappa/signal-checker/common"
)

type Binance struct {
	apiURL string
}

func NewBinance() *Binance {
	return &Binance{apiURL: "https://api.binance.com/api/v3/"}
}

func (b *Binance) overrideAPIURL(url string) {
	b.apiURL = url
}

func (b Binance) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(b.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b Binance) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(b.newTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}

const ERR_INVALID_SYMBOL = -1121
