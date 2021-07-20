# signal-checker

[![Go Reference](https://pkg.go.dev/badge/github.com/marianogappa/signal-checker.svg)](https://pkg.go.dev/github.com/marianogappa/signal-checker)

cli tool, server & library to check the results of crypto signals against an exchange's historical data.

## What can you do with this?

Many "businesses" offer "crypto signals" (i.e. cryptocurrency market predictions) for a fee. Most cryptocurrency exchanges offer historical candlestick and trade data for free. This tool allows you to validate if the predictions are met or not.

It's optimised for compatibility and composability: it works on your terminal, as a server and as a Go module you can import, and its inputs and outputs are JSON.

## Supported exchanges

- Binance
- Binance Futures (USD-M) *is being implemented*
- Coinbase
- FTX
- Kraken
- KuCoin

## Feature support

- Low/High enter range, Stop Loss and arbitrary number of Take Profits with configurable Take Profit ratios.
- Configurable adjusting Stop Loss to entry, to TP1, etc, as Take profit checkpoints are met.
- Calculates maximum amount (in stablecoin USD) that could be invested in the signal.

## Installation

Either get latest release: https://github.com/marianogappa/signal-checker/releases

Or build locally

```
$ go get github.com/marianogappa/signal-checker
```

## CLI usage

```bash
$ signal-checker '<JSON input data>'
```

## Server usage

```bash
$ signal-checker serve 8080
... other terminal ...
$ curl "localhost:8080/run" -d '<JSON input data>'
```

## Import library usage

```go
import (
	"github.com/marianogappa/signal-checker/signalchecker"
	"github.com/marianogappa/signal-checker/common"
)

func main() {
        input := common.SignalCheckInput{<input data>}
        output, _ := signalchecker.CheckSignal(input)
  	byts, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(byts))
}
```

## Input and output JSON format

[![Go Reference](https://pkg.go.dev/badge/github.com/marianogappa/signal-checker.svg)](https://pkg.go.dev/github.com/marianogappa/signal-checker)
