package contenthub

import (
	"reflect"
	"testing"
)

func TestDto_ToDomain(t *testing.T) {
	tests := []struct {
		name    string
		dto     Dto
		wantErr bool
		errMsg  string
		check   func(t *testing.T, entity Entity)
	}{
		{
			name: "should return error when variable/key is empty",
			dto: Dto{
				Variable: "",
				Active:   true,
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "should return error when variable/key is whitespace only",
			dto: Dto{
				Variable: "   ",
				Active:   true,
			},
			wantErr: true,
			errMsg:  "key is required",
		},
		{
			name: "should return error when balancer strategy weights sum to less than 100",
			dto: Dto{
				Variable: "test-key",
				Active:   true,
				BalancerStrategy: BalancerStrategy{
					{Weight: 30, Response: "response-a"},
					{Weight: 20, Response: "response-b"},
				},
			},
			wantErr: true,
			errMsg:  ErrInvalidWeight.Error(),
		},
		{
			name: "should return error when balancer strategy weights sum to more than 100",
			dto: Dto{
				Variable: "test-key",
				Active:   true,
				BalancerStrategy: BalancerStrategy{
					{Weight: 60, Response: "response-a"},
					{Weight: 60, Response: "response-b"},
				},
			},
			wantErr: true,
			errMsg:  ErrInvalidWeight.Error(),
		},
		{
			name: "should create entity with valid balancer strategy summing to 100",
			dto: Dto{
				Variable: "test-key",
				Active:   true,
				Value:    "test-value",
				BalancerStrategy: BalancerStrategy{
					{Weight: 70, Response: "response-a"},
					{Weight: 30, Response: "response-b"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.Variable != "test-key" {
					t.Errorf("expected Variable 'test-key', got '%s'", entity.Variable)
				}
				if !entity.Active {
					t.Error("expected Active to be true")
				}
				if entity.Value != "test-value" {
					t.Errorf("expected Value 'test-value', got '%s'", entity.Value)
				}
				if len(entity.BalancerStrategy) != 2 {
					t.Errorf("expected 2 balancer strategies, got %d", len(entity.BalancerStrategy))
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
			name: "should create entity with session strategies",
			dto: Dto{
				Variable: "session-key",
				Active:   true,
				SessionsStrategies: SessionsStrategies{
					{SessionID: "session-1", Response: "response-1"},
					{SessionID: "session-2", Response: "response-2"},
				},
				BalancerStrategy: BalancerStrategy{
					{Weight: 100, Response: "default"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.Variable != "session-key" {
					t.Errorf("expected Variable 'session-key', got '%s'", entity.Variable)
				}
				if len(entity.SessionsStrategies) != 2 {
					t.Errorf("expected 2 session strategies, got %d", len(entity.SessionsStrategies))
				}
				if entity.SessionsStrategies[0].SessionID != "session-1" {
					t.Errorf("expected first session ID 'session-1', got '%s'", entity.SessionsStrategies[0].SessionID)
				}
			},
		},
		{
			name: "should create entity with description",
			dto: Dto{
				Variable:    "desc-key",
				Active:      true,
				Description: "This is a test description",
				BalancerStrategy: BalancerStrategy{
					{Weight: 100, Response: "default"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.Description != "This is a test description" {
					t.Errorf("expected Description 'This is a test description', got '%s'", entity.Description)
				}
			},
		},
		{
			name: "should create inactive entity",
			dto: Dto{
				Variable: "inactive-key",
				Active:   false,
				BalancerStrategy: BalancerStrategy{
					{Weight: 100, Response: "default"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.Active {
					t.Error("expected Active to be false")
				}
			},
		},
		{
			name: "should create entity with single balancer strategy at 100%",
			dto: Dto{
				Variable: "single-strategy-key",
				Active:   true,
				BalancerStrategy: BalancerStrategy{
					{Weight: 100, Response: "only-response"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if len(entity.BalancerStrategy) != 1 {
					t.Errorf("expected 1 balancer strategy, got %d", len(entity.BalancerStrategy))
				}
				if entity.BalancerStrategy[0].Weight != 100 {
					t.Errorf("expected weight 100, got %d", entity.BalancerStrategy[0].Weight)
				}
			},
		},
		{
			name: "should create entity with multiple balancer strategies",
			dto: Dto{
				Variable: "multi-strategy-key",
				Active:   true,
				BalancerStrategy: BalancerStrategy{
					{Weight: 50, Response: "response-a"},
					{Weight: 30, Response: "response-b"},
					{Weight: 20, Response: "response-c"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if len(entity.BalancerStrategy) != 3 {
					t.Errorf("expected 3 balancer strategies, got %d", len(entity.BalancerStrategy))
				}
				totalWeight := uint(0)
				for _, s := range entity.BalancerStrategy {
					totalWeight += s.Weight
				}
				if totalWeight != 100 {
					t.Errorf("expected total weight 100, got %d", totalWeight)
				}
			},
		},
		{
			name: "should create entity with special characters in variable name",
			dto: Dto{
				Variable: "test-key_v2.0",
				Active:   true,
				BalancerStrategy: BalancerStrategy{
					{Weight: 100, Response: "default"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, entity Entity) {
				if entity.Variable != "test-key_v2.0" {
					t.Errorf("expected Variable 'test-key_v2.0', got '%s'", entity.Variable)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dto.ToDomain()

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

func TestFromDomain(t *testing.T) {
	tests := []struct {
		name   string
		entity Entity
		check  func(t *testing.T, dto Dto)
	}{
		{
			name: "should convert entity to dto",
			entity: Entity{
				Variable:    "test-key",
				Value:       "test-value",
				Description: "test description",
				Active:      true,
			},
			check: func(t *testing.T, dto Dto) {
				if dto.Variable != "test-key" {
					t.Errorf("expected Variable 'test-key', got '%s'", dto.Variable)
				}
				if dto.Value != "test-value" {
					t.Errorf("expected Value 'test-value', got '%s'", dto.Value)
				}
				if dto.Description != "test description" {
					t.Errorf("expected Description 'test description', got '%s'", dto.Description)
				}
				if !dto.Active {
					t.Error("expected Active to be true")
				}
			},
		},
		{
			name: "should convert entity with balancer strategy to dto",
			entity: Entity{
				Variable: "balancer-key",
				Active:   true,
				BalancerStrategy: BalancerStrategy{
					{Weight: 70, Response: "response-a"},
					{Weight: 30, Response: "response-b"},
				},
			},
			check: func(t *testing.T, dto Dto) {
				if len(dto.BalancerStrategy) != 2 {
					t.Errorf("expected 2 balancer strategies, got %d", len(dto.BalancerStrategy))
				}
			},
		},
		{
			name: "should convert entity with session strategies to dto",
			entity: Entity{
				Variable: "session-key",
				Active:   true,
				SessionsStrategies: SessionsStrategies{
					{SessionID: "session-1", Response: "response-1"},
				},
			},
			check: func(t *testing.T, dto Dto) {
				if len(dto.SessionsStrategies) != 1 {
					t.Errorf("expected 1 session strategy, got %d", len(dto.SessionsStrategies))
				}
				if dto.SessionsStrategies[0].SessionID != "session-1" {
					t.Errorf("expected SessionID 'session-1', got '%s'", dto.SessionsStrategies[0].SessionID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromDomain(tt.entity)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestManyFromDomain(t *testing.T) {
	tests := []struct {
		name     string
		entities map[string]Entity
		wantLen  int
	}{
		{
			name:     "should return empty slice for empty map",
			entities: map[string]Entity{},
			wantLen:  0,
		},
		{
			name: "should convert single entity",
			entities: map[string]Entity{
				"key1": {Variable: "key1", Active: true},
			},
			wantLen: 1,
		},
		{
			name: "should convert multiple entities",
			entities: map[string]Entity{
				"key1": {Variable: "key1", Active: true},
				"key2": {Variable: "key2", Active: false},
				"key3": {Variable: "key3", Active: true},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ManyFromDomain(tt.entities)
			if len(got) != tt.wantLen {
				t.Errorf("expected length %d, got %d", tt.wantLen, len(got))
			}

			// Verify all entities are properly converted
			for _, dto := range got {
				entity, exists := tt.entities[dto.Variable]
				if !exists {
					t.Errorf("unexpected Variable '%s' in result", dto.Variable)
					continue
				}
				if dto.Active != entity.Active {
					t.Errorf("Active mismatch for '%s': expected %v, got %v", dto.Variable, entity.Active, dto.Active)
				}
			}
		})
	}
}

func TestFromDomain_PreservesAllFields(t *testing.T) {
	entity := Entity{
		Variable:    "complete-key",
		Value:       "complete-value",
		Description: "complete description",
		Active:      true,
		SessionsStrategies: SessionsStrategies{
			{SessionID: "sess-1", Response: "resp-1"},
		},
		BalancerStrategy: BalancerStrategy{
			{Weight: 100, Response: "default"},
		},
	}

	dto := FromDomain(entity)

	if !reflect.DeepEqual(dto.SessionsStrategies, entity.SessionsStrategies) {
		t.Error("SessionsStrategies not properly preserved")
	}
	if !reflect.DeepEqual(dto.BalancerStrategy, entity.BalancerStrategy) {
		t.Error("BalancerStrategy not properly preserved")
	}
	if dto.ID != entity.ID {
		t.Error("ID not properly preserved")
	}
	if dto.CreatedAt != entity.CreatedAt {
		t.Error("CreatedAt not properly preserved")
	}
}


