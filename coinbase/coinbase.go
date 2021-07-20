package coinbase

import (
	"github.com/marianogappa/signal-checker/common"
)

type Coinbase struct {
	apiURL string
}

func NewCoinbase() *Coinbase {
	return &Coinbase{apiURL: "https://api.pro.coinbase.com/"}
}

func (c *Coinbase) overrideAPIURL(apiURL string) {
	c.apiURL = apiURL
}

func (c Coinbase) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(c.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (c Coinbase) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(c.newTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
