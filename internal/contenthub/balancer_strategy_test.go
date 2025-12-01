package contenthub

import (
	"testing"
)

func TestStrategies_Distribution(t *testing.T) {
	tests := []struct {
		name                 string
		strategies           BalancerStrategy
		expectedCalls        int
		expectedDistribution map[string]int
	}{
		{
			name: "Distribuição 50-10-40",
			strategies: BalancerStrategy{
				{Weight: 50, Response: "1"},
				{Weight: 10, Response: "2"},
				{Weight: 40, Response: "3"},
			},
			expectedCalls: 10,
			expectedDistribution: map[string]int{
				"1": 5, // 50% de 10 = 5 chamadas
				"2": 1, // 10% de 10 = 1 chamada
				"3": 4, // 40% de 10 = 4 chamadas
			},
		},
		{
			name: "Distribuição 30-30-40",
			strategies: BalancerStrategy{
				{Weight: 30, Response: "A"},
				{Weight: 30, Response: "B"},
				{Weight: 40, Response: "C"},
			},
			expectedCalls: 10,
			expectedDistribution: map[string]int{
				"A": 3, // 30% de 10 = 3 chamadas
				"B": 3, // 30% de 10 = 3 chamadas
				"C": 4, // 40% de 10 = 4 chamadas
			},
		},
		{
			name: "Distribuição 100-0",
			strategies: BalancerStrategy{
				{Weight: 100, Response: "1"},
				{Weight: 0, Response: "2"},
			},
			expectedCalls: 10,
			expectedDistribution: map[string]int{
				"1": 10, // 100% de 10 = 10 chamadas
				"2": 0,  // 0% de 10 = 0 chamadas
			},
		},
		{
			name: "Distribuição 20-20-30-30",
			strategies: BalancerStrategy{
				{Weight: 20, Response: "1"},
				{Weight: 20, Response: "2"},
				{Weight: 30, Response: "3"},
				{Weight: 30, Response: "4"},
			},
			expectedCalls: 10,
			expectedDistribution: map[string]int{
				"1": 2, // 20% de 10 = 2 chamadas
				"2": 2, // 20% de 10 = 2 chamadas
				"3": 3, // 30% de 10 = 3 chamadas
				"4": 3, // 30% de 10 = 3 chamadas
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make(map[string]int)

			for i := 0; i < tt.expectedCalls; i++ {
				result := tt.strategies.Distribution()
				if result != nil {
					resultStr, ok := result.(string)
					if ok {
						results[resultStr]++
					}
				}
			}

			for response, expectedCount := range tt.expectedDistribution {
				actualCount := results[response]
				if actualCount != expectedCount {
					t.Errorf("Response %q: esperado %d chamadas, obteve %d chamadas",
						response, expectedCount, actualCount)
				}
			}

			totalCalls := 0
			for _, count := range results {
				totalCalls += count
			}

			if totalCalls > tt.expectedCalls {
				t.Errorf("Total de chamadas excedeu o esperado: esperado %d, obteve %d",
					tt.expectedCalls, totalCalls)
			}
		})
	}
}

func TestStrategies_Distribution_NilAndEmpty(t *testing.T) {
	tests := []struct {
		name       string
		strategies *BalancerStrategy
		want       any
	}{
		{
			name:       "Strategies nil",
			strategies: nil,
			want:       nil,
		},
		{
			name:       "Strategies vazio",
			strategies: &BalancerStrategy{},
			want:       nil,
		},
		{
			name: "Weight zero",
			strategies: &BalancerStrategy{
				{Weight: 0, Response: "1"},
				{Weight: 0, Response: "2"},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.strategies == nil {
				var s *BalancerStrategy
				got := s.Distribution()
				if got != tt.want {
					t.Errorf("Distribution() = %v, want %v", got, tt.want)
				}
			} else {
				got := tt.strategies.Distribution()
				if got != tt.want {
					t.Errorf("Distribution() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestValidateStrategies(t *testing.T) {
	tests := []struct {
		name       string
		strategies BalancerStrategy
		want       error
	}{
		{
			name: "Invalid weight with less than 100",
			strategies: BalancerStrategy{
				{Weight: 30, Response: "1"},
				{Weight: 20, Response: "2"},
			},
			want: ErrInvalidWeight,
		},
		{
			name: "Invalid weight with more than 100",
			strategies: BalancerStrategy{
				{Weight: 90, Response: "1"},
				{Weight: 100, Response: "2"},
			},
			want: ErrInvalidWeight,
		},
		{
			name: "Valid strategies",
			strategies: BalancerStrategy{
				{Weight: 90, Response: "1"},
				{Weight: 10, Response: "2"},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.strategies.Validate()
			if err != tt.want {
				t.Errorf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}
