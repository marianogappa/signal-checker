package common

import (
	"time"
)

type CandlestickIterator struct {
	next func() (Candlestick, error)
}

func NewCandlestickIterator(next func() (Candlestick, error)) *CandlestickIterator {
	return &CandlestickIterator{next}
}

func (ci *CandlestickIterator) Next() (Candlestick, error) {
	return ci.next()
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
			time.Sleep(1 * time.Second)
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
