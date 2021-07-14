# signal-checker

[![Go Reference](https://pkg.go.dev/badge/github.com/marianogappa/signal-checker.svg)](https://pkg.go.dev/github.com/marianogappa/signal-checker)

cli tool, server & library to check the results of crypto signals against an exchange's historical data.

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
