package common

import (
	"math"
	"sort"
)

type TradeIterator struct {
	next func() (Trade, error)
}

func NewTradeIterator(next func() (Trade, error)) *TradeIterator {
	return &TradeIterator{next}
}

func (ci *TradeIterator) Next() (Trade, error) {
	return ci.next()
}

func (ci *TradeIterator) GetMaxBaseAssetEnter(minuteCount, bucketCount, maxTotalTrades int) (Trade, error) {
	// Get all trades from the beginning of this iterator and up to {minuteCount} minutes after, unless {maxTotalTrades} reached.
	secondCount := minuteCount * 60
	trades := []Trade{}
	for {
		if len(trades) >= maxTotalTrades {
			break
		}
		trade, err := ci.Next()
		if err == ErrOutOfTrades {
			break
		}
		if err != nil {
			return Trade{}, err
		}
		if len(trades) > 0 && trade.Timestamp > trades[0].Timestamp+secondCount {
			break
		}
		trades = append(trades, trade)
	}

	// Sort them by quantity
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].BaseAssetQuantity < trades[j].BaseAssetQuantity
	})

	// Pick the trade with quantity at 80% of the way to the largest
	chosen := int(math.Round(float64(len(trades)) * 0.8))

	return trades[chosen], nil
}
