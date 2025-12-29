package featureflag

import (
	"testing"

	"github.com/IsaacDSC/featureflag/internal/strategy"
)

func TestToDomain(t *testing.T) {
	tests := []struct {
		name    string
		input   Dto
		wantErr bool
		errMsg  string
		check   func(t *testing.T, entity Entity)
	}{
		{
			name: "should return error when flag name is empty",
			input: Dto{
				FlagName: "",
				Active:   true,
			},
			wantErr: true,
			errMsg:  "flag name is required",
		},
		{
			name: "should return error when flag name is whitespace only",
			input: Dto{
				FlagName: "   ",
				Active:   true,
			},
			wantErr: true,
			errMsg:  "flag name is required",
		},
		{
			name: "should create entity with valid flag name and no strategy",
			input: Dto{
				FlagName:   "test-flag",
				Active:     true,
				Strategies: strategy.StrategyDto{},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.FlagName != "test-flag" {
					t.Errorf("expected FlagName 'test-flag', got '%s'", entity.FlagName)
				}
				if !entity.Active {
					t.Error("expected Active to be true")
				}
				if entity.Strategies.WithStrategy {
					t.Error("expected WithStrategy to be false")
				}
				if entity.ID.String() == "" {
					t.Error("expected ID to be generated")
				}
				if entity.CreatedAt.IsZero() {
					t.Error("expected CreatedAt to be set")
				}
			},
		},
		{
			name: "should create entity with inactive flag",
			input: Dto{
				FlagName:   "inactive-flag",
				Active:     false,
				Strategies: strategy.StrategyDto{},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.FlagName != "inactive-flag" {
					t.Errorf("expected FlagName 'inactive-flag', got '%s'", entity.FlagName)
				}
				if entity.Active {
					t.Error("expected Active to be false")
				}
			},
		},
		{
			name: "should create entity with percent strategy",
			input: Dto{
				FlagName: "percent-flag",
				Active:   true,
				Strategies: strategy.StrategyDto{
					Percent: 50.0,
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.FlagName != "percent-flag" {
					t.Errorf("expected FlagName 'percent-flag', got '%s'", entity.FlagName)
				}
				if !entity.Strategies.WithStrategy {
					t.Error("expected WithStrategy to be true")
				}
				if entity.Strategies.Percent != 50.0 {
					t.Errorf("expected Percent 50.0, got %f", entity.Strategies.Percent)
				}
			},
		},
		{
			name: "should create entity with session ID strategy",
			input: Dto{
				FlagName: "session-flag",
				Active:   true,
				Strategies: strategy.StrategyDto{
					SessionsID: []string{"session-1", "session-2"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.FlagName != "session-flag" {
					t.Errorf("expected FlagName 'session-flag', got '%s'", entity.FlagName)
				}
				if !entity.Strategies.WithStrategy {
					t.Error("expected WithStrategy to be true")
				}
				if len(entity.Strategies.SessionsID) != 2 {
					t.Errorf("expected 2 sessions, got %d", len(entity.Strategies.SessionsID))
				}
				if !entity.Strategies.SessionsID["session-1"] {
					t.Error("expected session-1 to be in SessionsID")
				}
				if !entity.Strategies.SessionsID["session-2"] {
					t.Error("expected session-2 to be in SessionsID")
				}
			},
		},
		{
			name: "should create entity with both percent and session strategy",
			input: Dto{
				FlagName: "combined-flag",
				Active:   true,
				Strategies: strategy.StrategyDto{
					Percent:    75.0,
					SessionsID: []string{"session-1"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.FlagName != "combined-flag" {
					t.Errorf("expected FlagName 'combined-flag', got '%s'", entity.FlagName)
				}
				if !entity.Strategies.WithStrategy {
					t.Error("expected WithStrategy to be true")
				}
				if entity.Strategies.Percent != 75.0 {
					t.Errorf("expected Percent 75.0, got %f", entity.Strategies.Percent)
				}
				if !entity.Strategies.SessionsID["session-1"] {
					t.Error("expected session-1 to be in SessionsID")
				}
			},
		},
		{
			name: "should create entity with special characters in flag name",
			input: Dto{
				FlagName: "test-flag_v2.0",
				Active:   true,
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.FlagName != "test-flag_v2.0" {
					t.Errorf("expected FlagName 'test-flag_v2.0', got '%s'", entity.FlagName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToDomain(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestDtoFromDomain(t *testing.T) {
	tests := []struct {
		name   string
		entity Entity
		check  func(t *testing.T, dto Dto)
	}{
		{
			name: "should convert entity without strategy to dto",
			entity: Entity{
				FlagName: "test-flag",
				Active:   true,
				Strategies: strategy.Strategy{
					WithStrategy: false,
				},
			},
			check: func(t *testing.T, dto Dto) {
				if dto.FlagName != "test-flag" {
					t.Errorf("expected FlagName 'test-flag', got '%s'", dto.FlagName)
				}
				if !dto.Active {
					t.Error("expected Active to be true")
				}
			},
		},
		{
			name: "should convert entity with percent strategy to dto",
			entity: Entity{
				FlagName: "percent-flag",
				Active:   true,
				Strategies: strategy.Strategy{
					WithStrategy: true,
					Percent:      50.0,
				},
			},
			check: func(t *testing.T, dto Dto) {
				if dto.FlagName != "percent-flag" {
					t.Errorf("expected FlagName 'percent-flag', got '%s'", dto.FlagName)
				}
				if dto.Strategies.Percent != 50.0 {
					t.Errorf("expected Percent 50.0, got %f", dto.Strategies.Percent)
				}
			},
		},
		{
			name: "should convert entity with session strategy to dto",
			entity: Entity{
				FlagName: "session-flag",
				Active:   true,
				Strategies: strategy.Strategy{
					WithStrategy: true,
					SessionsID:   map[string]bool{"session-1": true, "session-2": true},
				},
			},
			check: func(t *testing.T, dto Dto) {
				if dto.FlagName != "session-flag" {
					t.Errorf("expected FlagName 'session-flag', got '%s'", dto.FlagName)
				}
				if len(dto.Strategies.SessionsID) != 2 {
					t.Errorf("expected 2 sessions, got %d", len(dto.Strategies.SessionsID))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DtoFromDomain(tt.entity)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
