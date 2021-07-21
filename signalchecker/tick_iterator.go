package signalchecker

import (
	"github.com/marianogappa/signal-checker/common"
)

func buildTickIterator(f func() (common.Candlestick, error)) func() (common.Tick, error) {
	return newTickIterator(f).next
}

type tickIterator struct {
	f     func() (common.Candlestick, error)
	ticks []common.Tick
}

func newTickIterator(f func() (common.Candlestick, error)) *tickIterator {
	return &tickIterator{f: f}
}

func (it *tickIterator) next() (common.Tick, error) {
	if len(it.ticks) > 0 {
		c := it.ticks[0]
		it.ticks = it.ticks[1:]
		return c, nil
	}
	candlestick, err := it.f()
	if err != nil {
		return common.Tick{}, err
	}
	it.ticks = append(it.ticks, candlestick.ToTicks()...)
	return it.next()
}
