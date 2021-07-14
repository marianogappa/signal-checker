package kraken

import (
	"errors"

	"github.com/marianogappa/signal-checker/common"
)

type krakenTradeIterator struct {
	baseAsset, quoteAsset string
	// trades                []common.Trade
	requestFromMillis int
}

func newKrakenTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *krakenTradeIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &krakenTradeIterator{
		baseAsset:         baseAsset,
		quoteAsset:        quoteAsset,
		requestFromMillis: int(initial.Unix()) * 1000,
	}
}

func (it *krakenTradeIterator) next() (common.Trade, error) {
	return common.Trade{}, errors.New("kraken's trade iterator is not implemented yet")
}
