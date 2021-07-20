package coinbase

import (
	"errors"

	"github.com/marianogappa/signal-checker/common"
)

type coinbaseTradeIterator struct {
	coinbase              Coinbase
	baseAsset, quoteAsset string
	// trades                []common.Trade
	requestFromMillis int
}

func (c Coinbase) newTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *coinbaseTradeIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &coinbaseTradeIterator{
		coinbase:          c,
		baseAsset:         baseAsset,
		quoteAsset:        quoteAsset,
		requestFromMillis: int(initial.Unix()) * 1000,
	}
}

func (it *coinbaseTradeIterator) next() (common.Trade, error) {
	return common.Trade{}, errors.New("coinbase's trade iterator is not implemented yet")
}
