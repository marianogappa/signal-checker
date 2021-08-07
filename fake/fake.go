package fake

import (
	"github.com/marianogappa/signal-checker/common"
)

type Fake struct {
	candlesticks []common.Candlestick
	trades       []common.Trade
	returnErr    error
}

func NewFake(candlesticks []common.Candlestick, trades []common.Trade, returnErr error) *Fake {
	return &Fake{candlesticks: candlesticks, trades: trades, returnErr: returnErr}
}

func (b *Fake) SetDebug(debug bool) {}

func (b Fake) BuildCandlestickIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.CandlestickIterator {
	return common.NewCandlestickIterator(b.testCandlestickIterator(b.candlesticks))
}

func (b Fake) BuildTradeIterator(baseAsset, quoteAsset string, initialISO8601 common.ISO8601) *common.TradeIterator {
	return common.NewTradeIterator(b.testTradeIterator(b.trades))
}

func (b Fake) testCandlestickIterator(cs []common.Candlestick) func() (common.Candlestick, error) {
	i := 0
	last := common.Candlestick{}
	return func() (common.Candlestick, error) {
		if i >= len(cs) {
			return last, common.ErrOutOfCandlesticks
		}
		i++
		last = cs[i-1]
		return cs[i-1], b.returnErr
	}
}
func (b Fake) testTradeIterator(ts []common.Trade) func() (common.Trade, error) {
	i := 0
	last := common.Trade{}
	return func() (common.Trade, error) {
		if i >= len(ts) {
			return last, common.ErrOutOfTrades
		}
		i++
		last = ts[i-1]
		return ts[i-1], b.returnErr
	}
}
