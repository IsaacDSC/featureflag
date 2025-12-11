package contenthub

import (
	"errors"
)

const MaxCalls = 10

type MultipleStrategy struct {
	Weight   uint `json:"weight" bson:"weight"`
	Response any  `json:"response" bson:"response"`
	Qtt      uint `bson:"qtt"`
}

type BalancerStrategy []MultipleStrategy

var ErrInvalidWeight = errors.New("invalid weight")

func (s *BalancerStrategy) Validate() error {
	total := uint(0)
	for _, strategy := range *s {
		total += strategy.Weight
	}

	if total > 100 || total < 100 {
		return ErrInvalidWeight
	}

	return nil
}

// Distribution implementa a lógica de distribuição weighted
// Distribui as respostas baseado no peso (weight) de cada estratégia
// 10 chamadas = 100%
func (s *BalancerStrategy) Distribution() any {
	if s == nil || len(*s) == 0 {
		return nil
	}

	strategies := *s

	var totalWeight uint
	for i := range strategies {
		totalWeight += strategies[i].Weight
	}

	if totalWeight == 0 {
		return nil
	}

	var totalCalls uint
	for i := range strategies {
		totalCalls += strategies[i].Qtt
	}

	if totalCalls >= MaxCalls {
		for i := range strategies {
			(*s)[i].Qtt = 0
		}
		totalCalls = 0
	}

	for i := range strategies {
		expectedCalls := uint(float64(strategies[i].Weight) / float64(totalWeight) * float64(MaxCalls))

		if strategies[i].Qtt < expectedCalls {
			strategies[i].Qtt++
			(*s)[i].Qtt = strategies[i].Qtt
			return strategies[i].Response
		}
	}

	for i := range strategies {
		if totalCalls < MaxCalls {
			strategies[i].Qtt++
			(*s)[i].Qtt = strategies[i].Qtt
			return strategies[i].Response
		}
	}

	if len(strategies) > 0 {
		return strategies[0].Response
	}

	return nil
}
