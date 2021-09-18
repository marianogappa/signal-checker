package profitcalculator

import (
	"log"

	"github.com/marianogappa/signal-checker/common"
)

type ProfitCalculator struct {
	input             common.SignalCheckInput
	tpCumRatios       []float64 // 0-based
	entryCumRatios    []float64 // 0-based
	appliedEventCount int
	highestEntered    int // 1-based
	lastEventType     string

	lastPrice          float64
	entryPrice         float64
	positionSize       float64
	ratioAwaitingEnter float64
	ratioOut           float64
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
		input:              input,
		tpCumRatios:        calculateCumulativeRatios(len(input.TakeProfits), input.TakeProfitRatios),
		entryCumRatios:     calculateCumulativeRatios(max(1, len(input.Entries)), input.EntryRatios),
		ratioAwaitingEnter: 1.0,
	}
}

func (p *ProfitCalculator) updatePositionSize(event common.SignalCheckOutputEvent) float64 {
	p.positionSize *= float64(event.Price) / p.lastPrice
	return p.positionSize
}

func (p *ProfitCalculator) ApplyEvent(event common.SignalCheckOutputEvent) float64 {
	if p.input.Debug {
		log.Printf("ProfitCalculator: applying event '%v' with price %v\n", event.EventType, event.Price)
	}
	p.appliedEventCount++

	switch event.EventType {
	case common.ENTERED:
		// Rarely, there's an entry after all entryRatio has been used. In this case, entry is ignored.
		if p.ratioAwaitingEnter == 0 {
			p.updatePositionSize(event)
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

		oldPositionSize := p.positionSize
		newPositionSize := enterWith / float64(event.Price)
		if p.positionSize == 0 {
			p.entryPrice = float64(event.Price)
		} else {
			oldPositionSize = p.updatePositionSize(event)
			p.entryPrice = (p.entryPrice*oldPositionSize + float64(event.Price)*newPositionSize) / (oldPositionSize + newPositionSize)
		}

		p.highestEntered = event.Target
		p.ratioAwaitingEnter -= enterWith
		p.positionSize = oldPositionSize + newPositionSize
	case common.STOPPED_LOSS, common.INVALIDATED, common.FINISHED_DATASET:
		// Empty ratio awaiting enter, so that isFinished returns true
		p.ratioOut += p.ratioAwaitingEnter
		p.ratioAwaitingEnter = 0
		p.updatePositionSize(event)
		result := p.positionSize * p.entryPrice
		if p.input.IsShort {
			result *= -1
		}
		p.ratioOut += result
		p.positionSize = 0

		if p.positionSize == 0 && event.EventType == common.STOPPED_LOSS {
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
	case common.TOOK_PROFIT:
		// Once having taken profit, all balance awaiting enter should never enter again, so empty it
		p.ratioOut += p.ratioAwaitingEnter
		p.ratioAwaitingEnter = 0
		p.updatePositionSize(event)

		if p.positionSize == 0 {
			if p.input.Debug {
				log.Println("ProfitCalculator: took profit without entering. This is likely a bug!")
			}
			break
		}
		if len(p.tpCumRatios)-1 < event.Target-1 {
			if p.input.Debug {
				log.Println("ProfitCalculator: took profit above existing take profit targets. This is likely a bug!")
			}
			break
		}

		ratioToTakeOut := p.positionSize * p.tpCumRatios[event.Target-1]
		result := ratioToTakeOut * p.entryPrice
		if p.input.IsShort {
			result *= -1
		}

		p.positionSize -= ratioToTakeOut
		p.ratioOut += result
	default:
		if p.input.Debug {
			log.Println("ProfitCalculator: found invalid event type. This is likely a bug!")
		}
	}
	p.lastEventType = event.EventType
	p.lastPrice = float64(event.Price)
	return p.CalculateTakeProfitRatio()
}

func (p ProfitCalculator) IsFinished() bool {
	return p.ratioAwaitingEnter+p.positionSize == 0.0
}

func (p ProfitCalculator) CalculateTakeProfitRatio() float64 {
	if p.entryPrice == 0 {
		return 0
	}
	resultIn := p.positionSize * p.entryPrice
	if p.input.IsShort {
		resultIn *= -1
	}
	tpr := resultIn + p.ratioOut - 1 + p.ratioAwaitingEnter
	if p.input.Debug {
		log.Printf("ProfitCalculator: awaiting enter = %v, taken out = %v, position size = %v, entry price = %v (PS*EP = %v). Take profit ratio =  %v\n",
			p.ratioAwaitingEnter, p.ratioOut, p.positionSize, p.entryPrice, resultIn, tpr,
		)
	}
	return tpr
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
