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
		if err != nil {
			return Trade{}, err
		}
		if len(trades) > 0 && trade.Timestamp > trades[0].Timestamp+secondCount {
			break
		}
		trades = append(trades, trade)
	}

	// Calculate min & max quantities to define the bucket size.
	quantityMin, priceMin := 0.0, 0.0
	quantityMax, priceMax := math.MaxFloat64, math.MaxFloat64
	for _, trade := range trades {
		quantityMin = math.Min(quantityMin, float64(trade.BaseAssetQuantity))
		quantityMax = math.Max(quantityMax, float64(trade.BaseAssetQuantity))
		priceMin = math.Min(priceMin, float64(trade.BaseAssetQuantity))
		priceMax = math.Max(priceMax, float64(trade.BaseAssetQuantity))
	}
	bucketSize := JsonFloat64(quantityMax - quantityMin/float64(bucketCount))

	// Place quantities & prices in buckets, length being the frequency.
	quantityBuckets := make([][]JsonFloat64, bucketCount)
	priceBuckets := make([][]JsonFloat64, bucketCount)
	for _, trade := range trades {
		bucket := int(math.Round(float64((trade.BaseAssetQuantity - JsonFloat64(quantityMin)) / bucketSize)))
		quantityBuckets[bucket] = append(quantityBuckets[bucket], trade.BaseAssetQuantity)
		priceBuckets[bucket] = append(priceBuckets[bucket], trade.BaseAssetPrice)
	}

	// Sweep buckets from right to left until 20% of all frequency is reached: choose that bucket.
	runningLength := 0
	chosenBucket := len(quantityBuckets) - 1
	for ; chosenBucket >= 0; chosenBucket-- {
		runningLength += len(quantityBuckets[chosenBucket])
		if runningLength >= len(trades)/5 {
			break
		}
	}

	// Take medians of quantity and price of that bucket.
	sort.Slice(quantityBuckets[chosenBucket], func(i, j int) bool {
		return quantityBuckets[chosenBucket][i] < quantityBuckets[chosenBucket][j]
	})
	sort.Slice(priceBuckets[chosenBucket], func(i, j int) bool {
		return priceBuckets[chosenBucket][i] < priceBuckets[chosenBucket][j]
	})

	return Trade{
		BaseAssetQuantity: median(quantityBuckets[chosenBucket]),
		BaseAssetPrice:    median(quantityBuckets[chosenBucket]),
	}, nil
}

func median(fs []JsonFloat64) JsonFloat64 {
	if len(fs)%2 == 1 {
		return fs[len(fs)/2]
	}
	return (fs[len(fs)/2-1] + fs[len(fs)/2]) / 2
}
