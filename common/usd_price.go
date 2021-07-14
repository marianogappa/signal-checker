package common

import "fmt"

func GetUSDPricePerBaseAssetUnitAtEvent(exchange Exchange, input SignalCheckInput, event SignalCheckOutputEvent) (JsonFloat64, error) {
	// If the base asset is a stablecoin that tracks the US dollar, then that's the USD price.
	stablecoins := []string{"USDT", "USDC", "BUSD", "DAI", "USD"}
	for _, stablecoin := range stablecoins {
		if input.BaseAsset == stablecoin {
			return event.Price, nil
		}
	}
	// If the quote asset is a stablecoin that tracks the US dollar, then the price is 1/base_asset.
	for _, stablecoin := range stablecoins {
		if input.QuoteAsset == stablecoin {
			return 1.0 / event.Price, nil
		}
	}
	// If there is a market pair with the base asset against a stablecoin, get its price.
	for _, stablecoin := range stablecoins {
		candlestickIterator := exchange.BuildCandlestickIterator(input.BaseAsset, stablecoin, event.At)
		price, err := candlestickIterator.GetPriceAt(event.At)
		if err != nil {
			continue
		}
		return price, nil
	}
	// Otherwise, check if there's a market pair with the base asset against known assets that go against stablecoins.
	transitives := map[string]string{
		"BTC": "USDT",
		"BNB": "BUSD",
	}
	for transitiveAsset, stablecoin := range transitives {
		candlestickIterator1 := exchange.BuildCandlestickIterator(input.BaseAsset, transitiveAsset, event.At)
		transitivePrice, err := candlestickIterator1.GetPriceAt(event.At)
		if err != nil {
			continue
		}
		candlestickIterator2 := exchange.BuildCandlestickIterator(transitiveAsset, stablecoin, event.At)
		stablecoinPrice, err := candlestickIterator2.GetPriceAt(event.At)
		if err != nil {
			continue
		}
		return transitivePrice * stablecoinPrice, nil
	}
	return JsonFloat64(0.0), fmt.Errorf("could not calculate USD price per unit of %v at event '%v'", input.BaseAsset, event.At)
}
