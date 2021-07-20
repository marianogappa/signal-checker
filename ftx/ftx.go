package ftx

import (
	"github.com/marianogappa/signal-checker/common"
)

type FTX struct {
	apiURL string
}

func NewFTX() *FTX {
	return &FTX{apiURL: "https://ftx.com/api/"}
}

func (f *FTX) overrideAPIURL(apiURL string) {
	f.apiURL = apiURL
}

func (f FTX) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(f.newCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (f FTX) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(f.newTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
