package kucoin

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestHappyToCandlesticks(t *testing.T) {
	testCandlestick := `[["1566789720","10411.5","10401.9","10411.5","10396.3","29.11357276","302889.301529914"]]`

	sr := [][]string{}
	err := json.Unmarshal([]byte(testCandlestick), &sr)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	cs, err := responseToCandlesticks(sr)
	if err != nil {
		t.Fatalf("Candlestick should have converted successfully but returned: %v", err)
	}
	if len(cs) != 1 {
		t.Fatalf("Should have converted 1 candlesticks but converted: %v", len(cs))
	}
	expected := common.Candlestick{
		Timestamp:      1566789720,
		OpenPrice:      f(10411.5),
		ClosePrice:     f(10401.9),
		LowestPrice:    f(10396.3),
		HighestPrice:   f(10411.5),
		Volume:         f(29.11357276),
		NumberOfTrades: 0,
	}
	if cs[0] != expected {
		t.Fatalf("Candlestick should have been %v but was %v", expected, cs[0])
	}
}

func TestUnhappyToCandlesticks(t *testing.T) {
	tests := []string{
		// candlestick %v has len != 7! Invalid syntax from Kucoin", i)
		`[["1566789720"]]`,
		// candlestick %v has non-int open time! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["INVALID","10411.5","10401.9","10411.5","10396.3","29.11357276","302889.301529914"]]`,
		// candlestick %v has non-float open! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["1566789720","INVALID","10401.9","10411.5","10396.3","29.11357276","302889.301529914"]]`,
		// candlestick %v has non-float close! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["1566789720","10411.5","INVALID","10411.5","10396.3","29.11357276","302889.301529914"]]`,
		// candlestick %v has non-float high! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["1566789720","10411.5","10401.9","INVALID","10396.3","29.11357276","302889.301529914"]]`,
		// candlestick %v has non-float low! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["1566789720","10411.5","10401.9","10411.5","INVALID","29.11357276","302889.301529914"]]`,
		// candlestick %v has non-float volume! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["1566789720","10411.5","10401.9","10411.5","10396.3","INVALID","302889.301529914"]]`,
		// candlestick %v has non-float turnover! Err was %v. Invalid syntax from Kucoin", i, err)
		`[["1566789720","10411.5","10401.9","10411.5","10396.3","29.11357276","INVALID"]]`,
	}

	for i, ts := range tests {
		t.Run(fmt.Sprintf("Unhappy toCandlesticks %v", i), func(t *testing.T) {
			sr := [][]string{}
			err := json.Unmarshal([]byte(ts), &sr)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			cs, err := responseToCandlesticks(sr)
			if err == nil {
				t.Fatalf("Candlestick should have failed to convert but converted successfully to: %v", cs)
			}
		})
	}
}

func f(fl float64) common.JsonFloat64 {
	return common.JsonFloat64(fl)
}
