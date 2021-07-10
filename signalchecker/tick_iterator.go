package signalchecker

import "github.com/marianogappa/signal-checker/types"

func buildTickIterator(f func() (types.Candlestick, error)) func() (types.Tick, error) {
	return newTickIterator(f).next
}

type tickIterator struct {
	f     func() (types.Candlestick, error)
	ticks []types.Tick
	err   error
}

func newTickIterator(f func() (types.Candlestick, error)) *tickIterator {
	return &tickIterator{f: f}
}

func (it *tickIterator) next() (types.Tick, error) {
	if len(it.ticks) > 0 {
		c := it.ticks[0]
		it.ticks = it.ticks[1:]
		return c, nil
	}
	candlestick, err := it.f()
	if err != nil {
		return types.Tick{}, err
	}
	it.ticks = append(it.ticks, candlestick.ToTicks()...)
	return it.next()
}
