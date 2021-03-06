# signal-checker

[![Go Reference](https://pkg.go.dev/badge/github.com/marianogappa/signal-checker.svg)](https://pkg.go.dev/github.com/marianogappa/signal-checker)
[![Coverage Status](https://coveralls.io/repos/github/marianogappa/signal-checker/badge.svg?branch=main)](https://coveralls.io/github/marianogappa/signal-checker?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/marianogappa/signal-checker)](https://goreportcard.com/report/github.com/marianogappa/signal-checker)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

cli tool, server & importable Go library to check the results of crypto signals against an exchange's historical data.

<img width="1667" alt="Signal Checker UI" src="https://user-images.githubusercontent.com/1078546/158064239-d5405657-c456-4bfe-8367-a6c2b2684e54.png">

Note: this is not signal-checker, but a frontend for it (e.g. does not support Take Profit Ratios).

⚠️ Warning: API unstable (e.g. SHORT signal support not properly tested). Contact me if you want to use this. ⚠️

## What can you do with this?

Many "businesses" offer "crypto signals" (i.e. cryptocurrency market predictions) for a fee, but how can you trust them? Most cryptocurrency exchanges offer historical candlestick and trade data for free. This tool allows you to validate if past predictions were met or not, so with a reasonable "signals dataset", you can fact-check if the "business" is legit or not.

It's optimised for compatibility and composability: it works on your terminal, as a server and as a Go module you can import, and its inputs and outputs are JSON.

## Supported exchanges (all major supported)

- Binance
- Binance Futures (USD-M) *is being implemented*
- Coinbase
- FTX
- Kraken
- KuCoin

NOTE: Huobi does not provide historical data with sufficient granularity, so it cannot be supported.

## Feature support

- Multiple entries with configurable ratios.
- Multiple take profits with configurable ratios.
- Adjustable stop losses on price checkpoints.
- Calculates maximum amount (in stablecoin USD) that could have been invested in the signal.

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
        output, _ := signalchecker.NewSignalChecker(input).Check()
  	byts, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(byts))
}
```

## Input and output JSON format

[![Go Reference](https://pkg.go.dev/badge/github.com/marianogappa/signal-checker.svg)](https://pkg.go.dev/github.com/marianogappa/signal-checker)
