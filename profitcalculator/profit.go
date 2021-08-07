package profitcalculator

import (
	"log"

	"github.com/marianogappa/signal-checker/common"
)

type ProfitCalculator struct {
	input             common.SignalCheckInput
	price             float64
	tpCumRatios       []float64 // 0-based
	entryCumRatios    []float64 // 0-based
	appliedEventCount int
	highestEntered    int // 1-based
	lastEventType     string

	awaitingEnter float64
	putIn         float64
	tookOut       float64
}

func calculateCumulativeRatios(requiredLen int, ratios []common.JsonFloat64) []float64 {
	cum := 0.0
	cums := []float64{}
	for i := 0; i < requiredLen; i++ {
		if i >= len(ratios) {
			cums = append(cums, cum)
			continue
		}
		cum += float64(ratios[i])
		cums = append(cums, cum)
	}
	return cums
}

func NewProfitCalculator(input common.SignalCheckInput) ProfitCalculator {
	return ProfitCalculator{
		input:          input,
		tpCumRatios:    calculateCumulativeRatios(len(input.TakeProfits), input.TakeProfitRatios),
		entryCumRatios: calculateCumulativeRatios(max(1, len(input.Entries)), input.EntryRatios),
		awaitingEnter:  1.0,
	}
}

func (p *ProfitCalculator) ApplyEvent(event common.SignalCheckOutputEvent) float64 {
	if p.input.Debug {
		log.Printf("ProfitCalculator: applying event '%v' with price %v\n", event.EventType, event.Price)
	}
	p.appliedEventCount++

	// TODO: LONG-only!
	if p.putIn > 0 {
		p.putIn *= float64(event.Price) / p.price
	}

	switch event.EventType {
	case common.ENTERED:
		// Rarely, there's an entry after all entryRatio has been used. In this case, entry is ignored.
		if p.awaitingEnter == 0 {
			break
		}
		// On most cases, signal will enter the first entry target.
		// However, sometimes the first entry will be a different target.
		// In this case, the cumulative entry ratio has to be calculated.
		// This is done by first calculating the cumulative of the last entry:
		cumLastEntry := 0.0
		if p.highestEntered > 0 {
			cumLastEntry = p.entryCumRatios[p.highestEntered-1]
		}

		// Then calculating the cumulative of the current entry:
		cumCurrentEntry := p.entryCumRatios[event.Target-1]

		// And calculating the difference between them:
		enterWith := cumCurrentEntry - cumLastEntry

		p.highestEntered = event.Target
		p.awaitingEnter -= enterWith
		p.putIn += enterWith
	case common.STOPPED_LOSS, common.INVALIDATED, common.FINISHED_DATASET:
		// All balance awaiting enter should never enter again, so take it out
		p.tookOut += p.awaitingEnter
		p.awaitingEnter = 0

		if p.putIn == 0 && event.EventType == common.STOPPED_LOSS {
			if p.input.Debug {
				log.Println("ProfitCalculator: stopped loss without entering. This is likely a bug!")
			}
			break
		}
		if p.appliedEventCount == 1 {
			if p.input.Debug {
				log.Println("ProfitCalculator: invalidating at first event. Likely signal out-of-sync with data.")
			}
			break
		}
		p.tookOut += p.putIn
		p.putIn = 0
	case common.TOOK_PROFIT:
		// Once having taken profit, all balance awaiting enter should never enter again, so take it out
		p.tookOut += p.awaitingEnter
		p.awaitingEnter = 0

		if p.putIn == 0 {
			if p.input.Debug {
				log.Println("ProfitCalculator: took profit without entering. This is likely a bug!")
			}
			break
		}
		takeOut := p.putIn * p.tpCumRatios[event.Target-1]
		p.putIn -= takeOut
		p.tookOut += takeOut
	default:
		if p.input.Debug {
			log.Println("ProfitCalculator: found invalid event type. This is likely a bug!")
		}
	}
	p.lastEventType = event.EventType
	p.price = float64(event.Price)
	return p.CalculateTakeProfitRatio()
}

func (p ProfitCalculator) IsFinished() bool {
	return p.awaitingEnter+p.putIn == 0.0
}

func (p ProfitCalculator) CalculateTakeProfitRatio() float64 {
	tpr := (p.awaitingEnter + p.putIn + p.tookOut) - 1
	if p.input.Debug {
		log.Printf("ProfitCalculator: awaiting enter = %v, still in = %v, taken out = %v. Take profit ratio =  %v\n", p.awaitingEnter, p.putIn, p.tookOut, tpr)
	}
	return tpr
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
