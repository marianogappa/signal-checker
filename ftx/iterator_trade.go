package ftx

import (
	"errors"

	"github.com/marianogappa/signal-checker/common"
)

type ftxTradeIterator struct {
	ftx                   FTX
	baseAsset, quoteAsset string
	// trades                []common.Trade
	requestFromMillis int
}

func (f FTX) newTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *ftxTradeIterator {
	// N.B. already validated
	initial, _ := initialISO8601.Time()
	return &ftxTradeIterator{
		ftx:               f,
		baseAsset:         baseAsset,
		quoteAsset:        quoteAsset,
		requestFromMillis: int(initial.Unix()) * 1000,
	}
}

func (it *ftxTradeIterator) next() (common.Trade, error) {
	return common.Trade{}, errors.New("ftx's trade iterator is not implemented yet")
}
