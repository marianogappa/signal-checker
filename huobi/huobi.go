package huobi

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
	return newHuobiCandlestickIterator(input).next
}

type huobiCandlestickIterator struct {
	input        types.SignalCheckInput
	candlesticks []types.Candlestick
	isDone       bool
}

func newHuobiCandlestickIterator(input types.SignalCheckInput) *huobiCandlestickIterator {
	return &huobiCandlestickIterator{input: input}
}

func (it *huobiCandlestickIterator) next() (types.Candlestick, error) {
	if len(it.candlesticks) > 0 {
		// N.B. Houbi returns candlesticks in descending order
		c := it.candlesticks[len(it.candlesticks)-1]
		it.candlesticks = it.candlesticks[:len(it.candlesticks)-1]
		return c, nil
	}
	if it.isDone {
		return types.Candlestick{}, types.ErrOutOfCandlesticks
	}

	klinesResult, err := getKlines(it.input.BaseAsset, it.input.QuoteAsset)
	if err != nil {
		return types.Candlestick{}, err
	}
	it.candlesticks = klinesResult.candlesticks
	it.isDone = true
	return it.next()
}

//{"id":1625500800,"open":33585.67,"close":34038.99,"low":33128.99,"high":35114.62,"amount":19722.38280792227,"vol":6.738773011373438E8,"count":506798}
type responseData struct {
	Timestamp int     `json:"id"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int     `json:"count"`
}

type response struct {
	Ch           string         `json:"ch"`
	Data         []responseData `json:"data"`
	Timestamp    int            `json:"ts"`
	Status       string         `json:"status"`
	ErrorCode    string         `json:"err-code"`
	ErrorMessage string         `json:"err-msg"`
}

func (r response) toCandlesticks() ([]types.Candlestick, error) {
	candlesticks := make([]types.Candlestick, len(r.Data))
	for i := 0; i < len(r.Data); i++ {
		raw := r.Data[i]
		candlestick := types.Candlestick{
			Timestamp:    raw.Timestamp,
			OpenPrice:    types.JsonFloat64(raw.Open),
			ClosePrice:   types.JsonFloat64(raw.Close),
			LowestPrice:  types.JsonFloat64(raw.Low),
			HighestPrice: types.JsonFloat64(raw.High),
			Volume:       types.JsonFloat64(raw.Amount), // N.B. Volume is quote asset volume; Amount is base asset volume
		}
		candlesticks[i] = candlestick
	}

	return candlesticks, nil
}

type klinesResult struct {
	candlesticks      []types.Candlestick
	err               error
	huobiErrorCode    string
	huobiErrorMessage string
	httpStatus        int
}

func getKlines(baseAsset string, quoteAsset string) (klinesResult, error) {
	req, err := http.NewRequest("GET", "https://api.huobi.pro/market/history/kline", nil)
	if err != nil {
		return klinesResult{err: err}, err
	}

	symbol := fmt.Sprintf("%v%v", strings.ToLower(baseAsset), strings.ToLower(quoteAsset))

	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("period", "1min")
	q.Add("size", "2000")

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
		err := fmt.Errorf("huobi returned %v status code with payload [%v]", resp.StatusCode, string(byts))
		return klinesResult{httpStatus: 500, err: err}, err
	}

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("huobi returned broken body response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	maybeResponse := response{}
	err = json.Unmarshal(byts, &maybeResponse)
	if err != nil {
		err := fmt.Errorf("huobi returned invalid JSON response! Was: %v", string(byts))
		return klinesResult{err: err, httpStatus: 500}, err
	}

	if maybeResponse.Status != "ok" {
		err := fmt.Errorf("huobi returned non-ok status response! Was: %v", string(byts))
		return klinesResult{err: err, huobiErrorMessage: maybeResponse.ErrorMessage, huobiErrorCode: maybeResponse.ErrorCode, httpStatus: 500}, err
	}

	candlesticks, err := maybeResponse.toCandlesticks()
	if err != nil {
		return klinesResult{
			httpStatus: 500,
			err:        fmt.Errorf("error unmarshalling successful JSON response from Huobi: %v", err),
		}, err
	}

	log.Printf("Huobi candlestick request successful! Candlestick count: %v\n", len(candlesticks))

	return klinesResult{
		candlesticks: candlesticks,
		httpStatus:   200,
	}, nil
}
