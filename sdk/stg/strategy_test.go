package stg

import (
	"testing"
)

func TestStrategy_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		percent float64
		want    uint
	}{
		{
			name:    "1% should return 9.9 rounded up to 10",
			percent: 1,
			want:    10,
		},
		{
			name:    "50% should return 5",
			percent: 50,
			want:    5,
		},
		{
			name:    "30% should return 7",
			percent: 30,
			want:    7,
		},
		{
			name:    "90% should return 1",
			percent: 90,
			want:    1,
		},
		{
			name:    "0% should return 10",
			percent: 0,
			want:    10,
		},
		{
			name:    "100% should return 0",
			percent: 100,
			want:    0,
		},
		{
			name:    "25% should return 7.5 rounded up to 8",
			percent: 25,
			want:    8,
		},
		{
			name:    "75% should return 2.5 rounded up to 3",
			percent: 75,
			want:    3,
		},
		{
			name:    "10% should return 9",
			percent: 10,
			want:    9,
		},
		{
			name:    "99% should return 0.1 rounded up to 1",
			percent: 99,
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Strategy[bool]{
				Percent: tt.percent,
			}
			got := s.Calculate()
			if got != tt.want {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrategy_SetQtdCall(t *testing.T) {
	tests := []struct {
		name           string
		initialQtdCall uint
		wantQtdCall    uint
	}{
		{
			name:           "should increment from 0 to 1",
			initialQtdCall: 0,
			wantQtdCall:    1,
		},
		{
			name:           "should increment from 5 to 6",
			initialQtdCall: 5,
			wantQtdCall:    6,
		},
		{
			name:           "should increment from 9 to 10",
			initialQtdCall: 9,
			wantQtdCall:    10,
		},
		{
			name:           "should reset from 10 to 0",
			initialQtdCall: 10,
			wantQtdCall:    0,
		},
		{
			name:           "should increment from 1 to 2",
			initialQtdCall: 1,
			wantQtdCall:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Strategy[bool]{
				QtdCall: tt.initialQtdCall,
			}
			s.SetQtdCall()
			if s.QtdCall != tt.wantQtdCall {
				t.Errorf("SetQtdCall() QtdCall = %v, want %v", s.QtdCall, tt.wantQtdCall)
			}
		})
	}
}

func TestStrategy_Bool(t *testing.T) {
	tests := []struct {
		name      string
		strategy  Strategy[bool]
		sessionID string
		want      bool
	}{
		{
			name: "empty SessionsID map - should use percent calculation (qtdCall=5, percent=50%, calculate=5)",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   map[string]bool{},
				Percent:      50,
				QtdCall:      5,
			},
			sessionID: "session-123",
			want:      true,
		},
		{
			name: "empty SessionsID map - should use percent calculation (qtdCall=3, percent=50%, calculate=5)",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   map[string]bool{},
				Percent:      50,
				QtdCall:      3,
			},
			sessionID: "session-123",
			want:      false,
		},
		{
			name: "nil SessionsID map - should use percent calculation",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   nil,
				Percent:      30,
				QtdCall:      7,
			},
			sessionID: "session-456",
			want:      true,
		},
		{
			name: "sessionID exists and is true",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID: map[string]bool{
					"session-abc": true,
					"session-def": false,
				},
				Percent: 50,
				QtdCall: 5,
			},
			sessionID: "session-abc",
			want:      true,
		},
		{
			name: "sessionID exists and is false",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID: map[string]bool{
					"session-abc": true,
					"session-def": false,
				},
				Percent: 50,
				QtdCall: 5,
			},
			sessionID: "session-def",
			want:      true,
		},
		{
			name: "sessionID does not exist in map",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID: map[string]bool{
					"session-abc": true,
					"session-def": false,
				},
				Percent: 50,
				QtdCall: 5,
			},
			sessionID: "session-xyz",
			want:      false,
		},
		{
			name: "empty SessionsID map - 1% percent with qtdCall=10 (should be active)",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   map[string]bool{},
				Percent:      1,
				QtdCall:      10,
			},
			sessionID: "session-test",
			want:      true,
		},
		{
			name: "empty SessionsID map - 90% percent with qtdCall=1 (should be active)",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   map[string]bool{},
				Percent:      90,
				QtdCall:      1,
			},
			sessionID: "session-test",
			want:      true,
		},
		{
			name: "empty SessionsID map - 90% percent with qtdCall=0 (should be inactive)",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   map[string]bool{},
				Percent:      90,
				QtdCall:      0,
			},
			sessionID: "session-test",
			want:      false,
		},
		{
			name: "empty SessionsID map - 100% percent with qtdCall=0 (should be active)",
			strategy: Strategy[bool]{
				WithStrategy: true,
				SessionsID:   map[string]bool{},
				Percent:      100,
				QtdCall:      0,
			},
			sessionID: "session-test",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.strategy.Bool(tt.sessionID)
			if got != tt.want {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrategy_SetQtdCall_Sequential(t *testing.T) {
	s := &Strategy[bool]{QtdCall: 0}

	// Test sequential calls to verify the cycle
	expected := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 0, 1, 2}

	for i, want := range expected {
		s.SetQtdCall()
		if s.QtdCall != want {
			t.Errorf("Call %d: SetQtdCall() QtdCall = %v, want %v", i+1, s.QtdCall, want)
		}
	}
}
