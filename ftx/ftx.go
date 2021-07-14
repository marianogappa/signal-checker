package ftx

import (
	"github.com/marianogappa/signal-checker/common"
)

type FTX struct{}

func NewFTX() FTX {
	return FTX{}
}

func (b FTX) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(newFTXCandlestickIterator(baseAsset, quoteAsset, initialISO8601).next)
}

func (b FTX) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(newFTXTradeIterator(baseAsset, quoteAsset, initialISO8601).next)
}
