package common

import (
	"fmt"
	"log"
)

func GetUSDPricePerBaseAssetUnitAtEvent(exchange Exchange, input SignalCheckInput, event SignalCheckOutputEvent) (JsonFloat64, error) {
	// If the base asset is a stablecoin that tracks the US dollar, then that's the USD price.
	stablecoins := []string{"USDT", "USDC", "BUSD", "DAI", "USD"}
	for _, stablecoin := range stablecoins {
		if input.BaseAsset == stablecoin {
			if input.Debug {
				log.Printf("GetUSDPricePerBaseAssetUnitAtEvent: base asset is %v (USD-based stablecoin), so no calculation needed: price is $%v.\n", input.BaseAsset, event.Price)
			}
			return event.Price, nil
		}
	}
	// If the quote asset is a stablecoin that tracks the US dollar, then the price is 1/base_asset.
	for _, stablecoin := range stablecoins {
		if input.QuoteAsset == stablecoin {
			price := 1.0 / event.Price
			if input.Debug {
				log.Printf("GetUSDPricePerBaseAssetUnitAtEvent: quote asset is %v (USD-based stablecoin), so price is $%v (which is 1/base asset price).\n", input.BaseAsset, price)
			}
			return price, nil
		}
	}
	// If there is a market pair with the base asset against a stablecoin, get its price.
	for _, stablecoin := range stablecoins {
		candlestickIterator := exchange.BuildCandlestickIterator(input.BaseAsset, stablecoin, event.At)
		price, err := candlestickIterator.GetPriceAt(event.At)
		if err != nil {
			continue
		}
		if input.Debug {
			log.Printf("GetUSDPricePerBaseAssetUnitAtEvent: found market pair %v/%v and checked base asset price in USD to be $%v\n", input.BaseAsset, stablecoin, price)
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
		price := transitivePrice * stablecoinPrice
		if input.Debug {
			log.Printf("GetUSDPricePerBaseAssetUnitAtEvent: found transitive market pairs %v/%v -> %v/%v and checked base asset price in USD to be $%v\n", input.BaseAsset, transitiveAsset, transitiveAsset, stablecoin, price)
		}
		return price, nil
	}
	return JsonFloat64(0.0), fmt.Errorf("GetUSDPricePerBaseAssetUnitAtEvent: could not calculate USD price per unit of %v at event '%v'", input.BaseAsset, event.At)
}
