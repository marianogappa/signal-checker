package binanceusdmfutures

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestHappyToCandlesticks(t *testing.T) {
	testCandlestick := `[
		[
		1499040000000,
		"0.01634790",
		"0.80000000",
		"0.01575800",
		"0.01577100",
		"148976.11427815",
		1499644799999,
		"2434.19055334",
		308,
		"1756.87402397",
		"28.46694368",
		"17928899.62484339"
		]
	]`

	sr := successfulResponse{}
	err := json.Unmarshal([]byte(testCandlestick), &sr.ResponseCandlesticks)
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
		Timestamp:      1499040000,
		OpenPrice:      f(0.01634790),
		ClosePrice:     f(0.01577100),
		LowestPrice:    f(0.01575800),
		HighestPrice:   f(0.80000000),
		Volume:         f(148976.11427815),
		NumberOfTrades: 308,
	}
	if cs[0] != expected {
		t.Fatalf("Candlestick should have been %v but was %v", expected, cs[0])
	}
}

func TestUnhappyToCandlesticks(t *testing.T) {
	tests := []string{
		// candlestick %v has len != 12! Invalid syntax from Binance
		`[
			[
				1499040000000
			]
		]`,
		// candlestick %v has non-int open time! Invalid syntax from Binance
		`[
			[
				"1499040000000",
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string open! Invalid syntax from Binance
		`[
			[
				1499040000000,
				0.01634790,
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had open = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"INVALID",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string high! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				0.80000000,
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had high = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"INVALID",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string low! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				0.01575800,
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had low = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"INVALID",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string close! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				0.01577100,
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had close = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"INVALID",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string volume! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				148976.11427815,
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had volume = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"INVALID",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-int close time! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				"1499644799999",
				"2434.19055334",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string quote asset volume! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				2434.19055334,
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had quote asset volume = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"INVALID",
				308,
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-int number of trades! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				"308",
				"1756.87402397",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string taker base asset volume! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				1756.87402397,
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v had taker base asset volume = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"INVALID",
				"28.46694368",
				"17928899.62484339"
			]
		]`,
		// candlestick %v has non-string taker quote asset volume! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				28.46694368,
				"17928899.62484339"
			]
		]`,
		// candlestick %v had taker quote asset volume = %v! Invalid syntax from Binance
		`[
			[
				1499040000000,
				"0.01634790",
				"0.80000000",
				"0.01575800",
				"0.01577100",
				"148976.11427815",
				1499644799999,
				"2434.19055334",
				308,
				"1756.87402397",
				"INVALID",
				"17928899.62484339"
			]
		]`,
	}

	for i, ts := range tests {
		t.Run(fmt.Sprintf("Unhappy toCandlesticks %v", i), func(t *testing.T) {
			sr := successfulResponse{}
			err := json.Unmarshal([]byte(ts), &sr.ResponseCandlesticks)
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
