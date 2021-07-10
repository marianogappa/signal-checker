package coinbase

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/marianogappa/signal-checker/types"
)

func BuildCandlestickIterator(input types.SignalCheckInput) func() (types.Candlestick, error) {
	return newCoinbaseCandlestickIterator(input).next
}

type coinbaseCandlestickIterator struct {
	input           types.SignalCheckInput
	candlesticks    []types.Candlestick
	requestFromTime time.Time
}

func newCoinbaseCandlestickIterator(input types.SignalCheckInput) *coinbaseCandlestickIterator {
	// N.B. already validated
	initial, _ := time.Parse(time.RFC3339, input.InitialISO8601)
	return &coinbaseCandlestickIterator{input: input, requestFromTime: initial}
}

func (it *coinbaseCandlestickIterator) next() (types.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		c := it.candlesticks[0]
		it.candlesticks = it.candlesticks[1:]
		return c, nil
	}
	if it.requestFromTime.After(time.Now().Add(-1 * time.Minute)) {
		return types.Candlestick{}, types.ErrOutOfCandlesticks
	}
	startTimeISO8601 := it.requestFromTime.Format(time.RFC3339)
	endTimeISO8601 := it.requestFromTime.Add(299 * 60 * time.Second).Format(time.RFC3339)

	klinesResult, err := getKlines(it.input.BaseAsset, it.input.QuoteAsset, startTimeISO8601, endTimeISO8601)
	if err != nil {
		return types.Candlestick{}, err
	}
	it.candlesticks = klinesResult.candlesticks
	if len(it.candlesticks) == 0 {
		return types.Candlestick{}, types.ErrOutOfCandlesticks
	}
	it.requestFromTime = it.requestFromTime.Add(299 * 60 * time.Second)
	return it.next()
}

type successResponse = [][]interface{}

type errorResponse struct {
	Message string `json:"message"`
}

func coinbaseToCandlesticks(response successResponse) ([]types.Candlestick, error) {
	candlesticks := make([]types.Candlestick, len(response))
	for i := 0; i < len(response); i++ {
		raw := response[i]
		timestampMillisFloat, ok := raw[0].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had timestampMillis = %v! Invalid syntax from Coinbase", i, timestampMillisFloat)
		}
		timestampMillis := int(timestampMillisFloat)
		lowestPrice, ok := raw[1].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had lowestPrice = %v! Invalid syntax from Coinbase", i, lowestPrice)
		}
		highestPrice, ok := raw[2].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had highestPrice = %v! Invalid syntax from Coinbase", i, highestPrice)
		}
		openPrice, ok := raw[3].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had openPrice = %v! Invalid syntax from Coinbase", i, openPrice)
		}
		closePrice, ok := raw[4].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had closePrice = %v! Invalid syntax from Coinbase", i, closePrice)
		}
		volume, ok := raw[5].(float64)
		if !ok {
			return candlesticks, fmt.Errorf("candlestick %v had volume = %v! Invalid syntax from Coinbase", i, volume)
		}

		candlestick := types.Candlestick{
			Timestamp:    timestampMillis / 1000,
			LowestPrice:  types.JsonFloat64(lowestPrice),
			HighestPrice: types.JsonFloat64(highestPrice),
			OpenPrice:    types.JsonFloat64(openPrice),
			ClosePrice:   types.JsonFloat64(closePrice),
			Volume:       types.JsonFloat64(volume),
		}
		candlesticks[i] = candlestick
	}

	return candlesticks, nil
}

type klinesResult struct {
	candlesticks         []types.Candlestick
	err                  error
	coinbaseErrorMessage string
	httpStatus           int
}

func getKlines(baseAsset string, quoteAsset string, startTimeISO8601, endTimeISO8601 string) (klinesResult, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.pro.coinbase.com/products/%v-%v/candles", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset)), nil)
	if err != nil {
		return klinesResult{err: err}, err
	}

	q := req.URL.Query()
	q.Add("granularity", "60")
	q.Add("start", fmt.Sprintf("%v", startTimeISO8601))
	q.Add("end", fmt.Sprintf("%v", endTimeISO8601))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Printf("Making request: %v\n", string(requestDump))
	}

	resp, err := client.Do(req)
	if err != nil {
		return klinesResult{err: err}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		byts, _ := ioutil.ReadAll(resp.Body)
		err := fmt.Errorf("coinbase returned %v status code with payload [%v]", resp.StatusCode, string(byts))
		return klinesResult{httpStatus: 500, err: err}, err
	}

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("coinbase returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	maybeErrorResponse := errorResponse{}
	err = json.Unmarshal(byts, &maybeErrorResponse)
	if err == nil && (maybeErrorResponse.Message != "") {
		err := fmt.Errorf("coinbase returned error code! Message: %v", maybeErrorResponse.Message)
		return klinesResult{
			coinbaseErrorMessage: maybeErrorResponse.Message,
			httpStatus:           500,
		}, err
	}

	maybeResponse := successResponse{}
	err = json.Unmarshal(byts, &maybeResponse)
	if err != nil {
		err := fmt.Errorf("coinbase returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	candlesticks, err := coinbaseToCandlesticks(maybeResponse)
	if err != nil {
		return klinesResult{
			httpStatus: 500,
			err:        fmt.Errorf("error unmarshalling successful JSON response from Coinbase: %v", err),
		}, err
	}

	log.Printf("Coinbase candlestick request successful! Candlestick count: %v\n", len(candlesticks))

	return klinesResult{
		candlesticks: candlesticks,
		httpStatus:   200,
	}, nil
}
