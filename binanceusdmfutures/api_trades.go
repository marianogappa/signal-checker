package binanceusdmfutures

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/marianogappa/signal-checker/common"
)

// [
//   {
//     "a": 26129,         // Aggregate tradeId
//     "p": "0.01633102",  // Price
//     "q": "4.70443515",  // Quantity
//     "f": 27781,         // First tradeId
//     "l": 27781,         // Last tradeId
//     "T": 1498793709153, // Timestamp
//     "m": true,          // Was the buyer the maker?
//     "M": true           // Was the trade the best price match?
//   }
// ]
type binanceTrade struct {
	AggregateTradeId      int    `json:"a"`
	Price                 string `json:"p"`
	Quantity              string `json:"q"`
	FirstTradeId          int    `json:"f"`
	LastTradeId           int    `json:"l"`
	TimestampMillis       int    `json:"T"`
	IsBuyerMaker          bool   `json:"m"`
	IsTradeBestPriceMatch bool   `json:"M"`
}

func (t binanceTrade) toTrade() (common.Trade, error) {
	price, err := strconv.ParseFloat(t.Price, 64)
	if err != nil {
		return common.Trade{}, err
	}
	quantity, err := strconv.ParseFloat(t.Quantity, 64)
	if err != nil {
		return common.Trade{}, err
	}
	return common.Trade{
		BaseAssetPrice:    common.JsonFloat64(price),
		BaseAssetQuantity: common.JsonFloat64(quantity),
		Timestamp:         t.TimestampMillis / 1000,
	}, nil
}

type aggTradesResponse = []binanceTrade

func binanceTradesToTrades(r aggTradesResponse) ([]common.Trade, error) {
	trades := []common.Trade{}
	for _, binanceTrade := range r {
		trade, err := binanceTrade.toTrade()
		if err != nil {
			return trades, err
		}
		trades = append(trades, trade)
	}
	return trades, nil
}

type aggTradesResult struct {
	trades              []common.Trade
	err                 error
	binanceErrorCode    int
	binanceErrorMessage string
	httpStatus          int
}

func (b BinanceUSDMFutures) getTrades(baseAsset string, quoteAsset string, startTimeMillis int) (aggTradesResult, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%vaggTrades", b.apiURL), nil)
	if err != nil {
		return aggTradesResult{err: err}, err
	}

	symbol := fmt.Sprintf("%v%v", strings.ToUpper(baseAsset), strings.ToUpper(quoteAsset))

	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("limit", "1000")
	q.Add("startTime", fmt.Sprintf("%v", startTimeMillis))
	q.Add("endTime", fmt.Sprintf("%v", startTimeMillis+50*6*1000))

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}

	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Printf("Making request: %v\n", string(requestDump))
	}

	resp, err := client.Do(req)
	if err != nil {
		return aggTradesResult{err: err}, err
	}
	defer resp.Body.Close()

	// N.B. commenting this out, because 400 returns valid JSON with error description, which we need!
	// if resp.StatusCode != http.StatusOK {
	// 	err := fmt.Errorf("binance returned %v status code", resp.StatusCode)
	// 	return aggTradesResult{httpStatus: 500, err: err}, err
	// }

	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("binance returned broken body response! Was: %v", string(byts))
		return aggTradesResult{err: err, httpStatus: 500}, err
	}

	maybeErrorResponse := errorResponse{}
	err = json.Unmarshal(byts, &maybeErrorResponse)
	errResp := maybeErrorResponse.toError()
	if err == nil && errResp != nil {
		return aggTradesResult{
			binanceErrorCode:    maybeErrorResponse.Code,
			binanceErrorMessage: maybeErrorResponse.Msg,
			httpStatus:          500,
			err:                 errResp,
		}, errResp
	}

	maybeResponse := aggTradesResponse([]binanceTrade{})
	err = json.Unmarshal(byts, &maybeResponse)
	if err != nil {
		err := fmt.Errorf("binance returned invalid JSON response! Was: %v", string(byts))
		return aggTradesResult{err: err, httpStatus: 500}, err
	}

	trades, err := binanceTradesToTrades(maybeResponse)
	if err != nil {
		return aggTradesResult{
			httpStatus: 500,
			err:        err,
		}, err
	}

	if len(trades) == 0 {
		return aggTradesResult{
			httpStatus: 200,
			err:        common.ErrOutOfCandlesticks,
		}, common.ErrOutOfCandlesticks
	}

	return aggTradesResult{
		trades:     trades,
		httpStatus: 200,
	}, nil
}
