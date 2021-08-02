package signalchecker

import (
	"github.com/marianogappa/signal-checker/common"
)

func buildTickIterator(f func() (common.Candlestick, error)) func() (common.Tick, error) {
	return newTickIterator(f).next
}

type tickIterator struct {
	f        func() (common.Candlestick, error)
	ticks    []common.Tick
	lastTick common.Tick
}

func newTickIterator(f func() (common.Candlestick, error)) *tickIterator {
	return &tickIterator{f: f}
}

// N.B. When next() hits an error, it returns the previous (last) tick.
// This is because when ErrOutOfCandlesticks happens, we want the previous
// candlestick to calculate things for the finished_dataset event.
func (it *tickIterator) next() (common.Tick, error) {
	if len(it.ticks) > 0 {
		c := it.ticks[0]
		it.ticks = it.ticks[1:]
		it.lastTick = c
		return c, nil
	}
	candlestick, err := it.f()
	if err != nil {
		return it.lastTick, err
	}
	it.ticks = append(it.ticks, candlestick.ToTicks()...)
	return it.next()
}
