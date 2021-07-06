# signal-checker

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
	"github.com/marianogappa/signal-checker/types"
)

func main() {
        input := types.SignalCheckInput{<input data>}
        output, _ := signalchecker.CheckSignal(input)
  	byts, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(byts))
}
```

## Input and output JSON format

Check godoc: TODO
