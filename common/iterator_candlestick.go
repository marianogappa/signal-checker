package common

import (
	"time"
)

type CandlestickIterator struct {
	SavedCandlesticks []Candlestick
	next              func() (Candlestick, error)
	calmDuration      time.Duration
}

func NewCandlestickIterator(next func() (Candlestick, error)) *CandlestickIterator {
	return &CandlestickIterator{next: next, SavedCandlesticks: nil, calmDuration: 1 * time.Second}
}

func (ci *CandlestickIterator) SaveCandlesticks() {
	ci.SavedCandlesticks = []Candlestick{}
}

func (ci *CandlestickIterator) Next() (Candlestick, error) {
	cs, err := ci.next()
	if ci.SavedCandlesticks != nil && err == nil {
		ci.SavedCandlesticks = append(ci.SavedCandlesticks, cs)
	}
	return cs, err
}

func (ci *CandlestickIterator) GetPriceAt(at ISO8601) (JsonFloat64, error) {
	rateLimitAttempts := 5
	atTimestamp, err := at.Seconds()
	if err != nil {
		return JsonFloat64(0.0), err
	}
	for {
		candlestick, err := ci.next()
		if err == ErrRateLimit && rateLimitAttempts > 0 {
			time.Sleep(ci.calmDuration)
			rateLimitAttempts--
			continue
		}
		if err != nil {
			return JsonFloat64(0.0), err
		}
		if candlestick.Timestamp < atTimestamp {
			continue
		}
		return candlestick.OpenPrice, nil
	}
}
