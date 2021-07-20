package kucoin

import (
	"errors"

	"github.com/marianogappa/signal-checker/common"
)

type kucoinTradeIterator struct {
	kucoin                Kucoin
	baseAsset, quoteAsset string
	// trades                []common.Trade
	requestFromMillis int
}

func (k Kucoin) newTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *kucoinTradeIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &kucoinTradeIterator{
		kucoin:            k,
		baseAsset:         baseAsset,
		quoteAsset:        quoteAsset,
		requestFromMillis: int(initial.Unix()) * 1000,
	}
}

func (it *kucoinTradeIterator) next() (common.Trade, error) {
	return common.Trade{}, errors.New("kucoin's trade iterator is not implemented yet")
}
