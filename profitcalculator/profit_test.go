package profitcalculator

import (
	"testing"

	"github.com/marianogappa/signal-checker/common"
)

func TestProfitCalculator(t *testing.T) {
	type test struct {
		name     string
		input    common.SignalCheckInput
		events   []common.SignalCheckOutputEvent
		expected []float64
	}

	tss := []test{
		{
			name: "Happy",
			input: common.SignalCheckInput{
				BaseAsset:        "BTC",
				QuoteAsset:       "USDT",
				TakeProfits:      []common.JsonFloat64{1.0},
				TakeProfitRatios: []common.JsonFloat64{1.0},
			},
			events: []common.SignalCheckOutputEvent{
				{
					EventType: common.ENTERED,
					Price:     0.1,
					At:        "2020-01-02T03:04:05+00:00",
				},
				{
					EventType: common.TAKEN_PROFIT_ + "1",
					Price:     1,
					At:        "2020-01-02T03:04:05+00:00",
				},
			},
			expected: []float64{0.0, 9.0},
		},
	}
	for _, ts := range tss {
		t.Run(ts.name, func(t *testing.T) {
			profitCalculator := NewProfitCalculator(ts.input)
			for i, ev := range ts.events {
				actual := profitCalculator.ApplyEvent(ev)
				if actual != ts.expected[i] {
					t.Errorf("On event %v expected %v to equal %v", i, actual, ts.expected[i])
					t.FailNow()
				}
			}
			actual := profitCalculator.CalculateTakeProfitRatio()
			if actual != ts.expected[len(ts.events)-1] {
				t.Errorf("On final calculation expected %v to equal %v", actual, ts.expected[len(ts.events)-1])
				t.FailNow()
			}
		})
	}
}
