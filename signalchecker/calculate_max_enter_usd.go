package signalchecker

import (
	"errors"
	"log"

	"github.com/marianogappa/signal-checker/common"
)

func getEnteredEvent(events []common.SignalCheckOutputEvent) (common.SignalCheckOutputEvent, bool) {
	for _, event := range events {
		if event.EventType == common.ENTERED {
			return event, true
		}
	}
	return common.SignalCheckOutputEvent{}, false
}

func calculateMaxEnterUSD(exchange common.Exchange, input common.SignalCheckInput, events []common.SignalCheckOutputEvent) (common.JsonFloat64, error) {
	enteredEvent, ok := getEnteredEvent(events)
	if !ok {
		return common.JsonFloat64(0.0), errors.New("this signal did not enter so cannot calculate maxEnterUSD")
	}
	usdPricePerBaseAsset, err := common.GetUSDPricePerBaseAssetUnitAtEvent(exchange, input, enteredEvent)
	if err != nil {
		return common.JsonFloat64(0.0), err
	}
	tradeIterator := exchange.BuildTradeIterator(input.BaseAsset, input.QuoteAsset, enteredEvent.At)
	maxTrade, err := tradeIterator.GetMaxBaseAssetEnter(5 /* minuteCount */, 10 /* bucketCount */, 10000 /* maxTradeCount */)
	if err != nil {
		return common.JsonFloat64(0.0), err
	}
	maxEnterUSD := usdPricePerBaseAsset * maxTrade.BaseAssetQuantity
	if input.Debug {
		log.Printf("calculateMaxEnterUSD: best-ish quantity trade was %v units of %v/%v at a price of %.6f (but entered price was %.6f!!), which is a USD price of ~$%.6f per unit, totalling ~$%.6f\n",
			maxTrade.BaseAssetQuantity, input.BaseAsset, input.QuoteAsset, maxTrade.BaseAssetPrice, enteredEvent.Price, usdPricePerBaseAsset, maxEnterUSD)
	}
	return maxEnterUSD, nil
}
