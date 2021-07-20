package profitcalculator

import (
	"log"
	"strconv"
	"strings"

	"github.com/marianogappa/signal-checker/common"
)

type ProfitCalculator struct {
	input             common.SignalCheckInput
	putIn             float64
	price             float64
	tookOut           float64
	accumRatios       []float64
	appliedEventCount int
}

func NewProfitCalculator(input common.SignalCheckInput) ProfitCalculator {
	accum := 0.0
	accums := []float64{}
	for i := 0; i < len(input.TakeProfits); i++ {
		if i >= len(input.TakeProfitRatios) {
			accums = append(accums, accum)
			break
		}
		accum += float64(input.TakeProfitRatios[i])
		accums = append(accums, accum)
	}
	return ProfitCalculator{input: input, accumRatios: accums}
}

func (p *ProfitCalculator) ApplyEvent(event common.SignalCheckOutputEvent) float64 {
	if p.input.Debug {
		log.Printf("ProfitCalculator: applying event '%v' with price %v\n", event.EventType, event.Price)
	}
	p.appliedEventCount++
	switch event.EventType {
	case common.ENTERED:
		p.putIn = 1.0
		p.tookOut = 0.0
		p.price = float64(event.Price)
	case common.STOPPED_LOSS, common.INVALIDATED:
		if p.appliedEventCount == 1 {
			if p.input.Debug {
				log.Println("ProfitCalculator: invalidating at first event. Likely signal out-of-sync with data.")
			}
			return 0.0
		}
		if !p.input.IsShort {
			p.putIn *= float64(event.Price) / p.price
		} else {
			p.putIn *= p.price / float64(event.Price)
		}
		p.tookOut += p.putIn
		p.putIn = 0
		p.price = float64(event.Price)
	default:
		if len(event.EventType) <= 13 || event.EventType[:13] != common.TAKEN_PROFIT_ {
			// N.B. invalid event types are not considered possible
			if p.input.Debug {
				log.Println("ProfitCalculator: found invalid event type. This is likely a bug!")
			}
			return 0.0
		}
		n, err := strconv.Atoi(strings.Split(event.EventType, common.TAKEN_PROFIT_)[1])
		if err != nil {
			// N.B. invalid event types are not considered possible
			if p.input.Debug {
				log.Println("ProfitCalculator: found invalid event type. This is likely a bug!")
			}
			return 0.0
		}
		if !p.input.IsShort {
			p.putIn *= float64(event.Price) / p.price
		} else {
			p.putIn *= p.price / float64(event.Price)
		}
		takeOut := p.putIn * p.accumRatios[n-1]
		p.putIn -= takeOut
		p.tookOut += takeOut
		p.price = float64(event.Price)
	}
	return p.CalculateTakeProfitRatio()
}

func (p ProfitCalculator) IsFinished() bool {
	return p.putIn == 0.0
}

func (p ProfitCalculator) CalculateTakeProfitRatio() float64 {
	if p.putIn+p.tookOut == 0 {
		return 0
	}
	tpr := (p.putIn + p.tookOut) - 1
	if p.input.Debug {
		log.Printf("ProfitCalculator: still in = %v, taken out = %v. Take profit ratio =  %v\n", p.putIn, p.tookOut, tpr)
	}
	return tpr
}
