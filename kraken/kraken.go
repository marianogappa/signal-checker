package kraken

import (
	"github.com/marianogappa/signal-checker/common"
)

type Kraken struct {
	apiURL string
}

func NewKraken() *Kraken {
	return &Kraken{apiURL: "https://api.kraken.com/0/"}
}

func (k *Kraken) overrideAPIURL(apiURL string) {
	k.apiURL = apiURL
}

func (k Kraken) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(k.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (k Kraken) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(k.newTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
