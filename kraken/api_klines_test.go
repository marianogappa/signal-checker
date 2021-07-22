package kraken

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestHappyToCandlesticks(t *testing.T) {
	testCandlestick := `{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`

	sr := response{}
	err := json.Unmarshal([]byte(testCandlestick), &sr)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	cs, err := sr.toCandlesticks()
	if err != nil {
		t.Fatalf("Candlestick should have converted successfully but returned: %v", err)
	}
	if len(cs) != 1 {
		t.Fatalf("Should have converted 1 candlesticks but converted: %v", len(cs))
	}
	expected := common.Candlestick{
		Timestamp:      1625623260,
		OpenPrice:      f(34221.6),
		ClosePrice:     f(34215.7),
		LowestPrice:    f(34215.7),
		HighestPrice:   f(34221.6),
		Volume:         f(0.25998804),
		NumberOfTrades: 7,
	}
	if cs[0] != expected {
		t.Fatalf("Candlestick should have been %v but was %v", expected, cs[0])
	}
}

func TestUnhappyToCandlesticks(t *testing.T) {
	tests := []string{
		// data key [%v] did not contain an array of datapoints
		`{"error":[],"result":{"XBTUSDT":"INVALID","last":1626869340}}`,
		// candlestick [%v] did not contain an array of data fields, instead: [%v]
		`{"error":[],"result":{"XBTUSDT":["INVALID"],"last":1626869340}}`,
		// candlestick %v has non-int open time! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[["INVALID","34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string open! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,34221.6,"34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had open = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"INVALID","34221.6","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string high! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6",34221.6,"34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had high = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","INVALID","34215.7","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string low! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6",34215.7,"34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had low = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","INVALID","34215.7","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string close! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7",34215.7,"34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had close = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","INVALID","34215.7","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string vwap! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7",34215.7,"0.25998804",7]],"last":1626869340}}`,
		// candlestick %v had vwap = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","INVALID","0.25998804",7]],"last":1626869340}}`,
		// candlestick %v has non-string volume! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7",0.25998804,7]],"last":1626869340}}`,
		// candlestick %v had volume = %v! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","INVALID",7]],"last":1626869340}}`,
		// candlestick %v has non-int trade count! Invalid syntax from Kraken
		`{"error":[],"result":{"XBTUSDT":[[1625623260,"34221.6","34221.6","34215.7","34215.7","34215.7","0.25998804","7"]],"last":1626869340}}`,
	}

	for i, ts := range tests {
		t.Run(fmt.Sprintf("Unhappy toCandlesticks %v", i), func(t *testing.T) {
			sr := response{}
			err := json.Unmarshal([]byte(ts), &sr)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			cs, err := sr.toCandlesticks()
			if err == nil {
				t.Fatalf("Candlestick should have failed to convert but converted successfully to: %v", cs)
			}
		})
	}
}

func f(fl float64) common.JsonFloat64 {
	return common.JsonFloat64(fl)
}
