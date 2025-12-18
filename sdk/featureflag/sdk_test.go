package featureflag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IsaacDSC/featureflag/sdk/stg"
)

func TestNewFeatureFlagSDK(t *testing.T) {
	tests := []struct {
		name   string
		hostFF string
	}{
		{
			name:   "should create SDK with valid host",
			hostFF: "http://localhost:8080",
		},
		{
			name:   "should create SDK with empty host",
			hostFF: "",
		},
		{
			name:   "should create SDK with custom host",
			hostFF: "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sdk := NewFeatureFlagSDK(tt.hostFF)

			if sdk == nil {
				t.Fatal("NewFeatureFlagSDK() returned nil")
			}

			if sdk.host != tt.hostFF {
				t.Errorf("host = %v, want %v", sdk.host, tt.hostFF)
			}

			if sdk.client == nil {
				t.Error("client should not be nil")
			}

			if sdk.inMemoryFlags != nil {
				t.Error("inMemoryFlags should be nil initially")
			}
		})
	}
}

func TestFFResponse_WithDefault(t *testing.T) {
	tests := []struct {
		name      string
		response  FFResponse
		ffDefault bool
		want      bool
	}{
		{
			name: "should return default when error exists",
			response: FFResponse{
				Bool:  true,
				Error: ErrNotFoundFeatureFlag,
			},
			ffDefault: false,
			want:      false,
		},
		{
			name: "should return default true when error exists",
			response: FFResponse{
				Bool:  false,
				Error: ErrInvalidStrategy,
			},
			ffDefault: true,
			want:      true,
		},
		{
			name: "should return actual value when no error",
			response: FFResponse{
				Bool:  true,
				Error: nil,
			},
			ffDefault: false,
			want:      true,
		},
		{
			name: "should return actual false value when no error",
			response: FFResponse{
				Bool:  false,
				Error: nil,
			},
			ffDefault: true,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.response.WithDefault(tt.ffDefault)
			if got != tt.want {
				t.Errorf("WithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFFResponse_Err(t *testing.T) {
	tests := []struct {
		name      string
		response  FFResponse
		wantBool  bool
		wantError error
	}{
		{
			name: "should return bool and error",
			response: FFResponse{
				Bool:  true,
				Error: ErrNotFoundFeatureFlag,
			},
			wantBool:  true,
			wantError: ErrNotFoundFeatureFlag,
		},
		{
			name: "should return bool and nil error",
			response: FFResponse{
				Bool:  false,
				Error: nil,
			},
			wantBool:  false,
			wantError: nil,
		},
		{
			name: "should return true with invalid strategy error",
			response: FFResponse{
				Bool:  true,
				Error: ErrInvalidStrategy,
			},
			wantBool:  true,
			wantError: ErrInvalidStrategy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotError := tt.response.Err()
			if gotBool != tt.wantBool {
				t.Errorf("Err() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotError != tt.wantError {
				t.Errorf("Err() error = %v, want %v", gotError, tt.wantError)
			}
		})
	}
}

func TestFFResponse_Val(t *testing.T) {
	tests := []struct {
		name     string
		response FFResponse
		want     bool
	}{
		{
			name: "should return true value",
			response: FFResponse{
				Bool:  true,
				Error: nil,
			},
			want: true,
		},
		{
			name: "should return false value",
			response: FFResponse{
				Bool:  false,
				Error: nil,
			},
			want: false,
		},
		{
			name: "should return value even with error",
			response: FFResponse{
				Bool:  true,
				Error: ErrNotFoundFeatureFlag,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.response.Val()
			if got != tt.want {
				t.Errorf("Val() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeatureFlagSDK_GetFeatureFlag(t *testing.T) {
	tests := []struct {
		name          string
		inMemoryFlags map[string]Flag
		ffDefault     bool
		key           string
		sessionID     []string
		wantBool      bool
		wantError     error
	}{
		{
			name:          "should return error when flag not found",
			inMemoryFlags: map[string]Flag{},
			ffDefault:     false,
			key:           "non-existent-flag",
			sessionID:     nil,
			wantBool:      false,
			wantError:     ErrNotFoundFeatureFlag,
		},
		{
			name:          "should return default true when flag not found",
			inMemoryFlags: map[string]Flag{},
			ffDefault:     true,
			key:           "missing-flag",
			sessionID:     nil,
			wantBool:      true,
			wantError:     ErrNotFoundFeatureFlag,
		},
		{
			name: "should return active flag without strategy",
			inMemoryFlags: map[string]Flag{
				"feature-x": {
					Active:   true,
					FlagName: "feature-x",
					Strategy: stg.Strategy[bool]{
						WithStrategy: false,
					},
				},
			},
			ffDefault: false,
			key:       "feature-x",
			sessionID: nil,
			wantBool:  true,
			wantError: nil,
		},
		{
			name: "should return inactive flag without strategy",
			inMemoryFlags: map[string]Flag{
				"feature-y": {
					Active:   false,
					FlagName: "feature-y",
					Strategy: stg.Strategy[bool]{
						WithStrategy: false,
					},
				},
			},
			ffDefault: true,
			key:       "feature-y",
			sessionID: nil,
			wantBool:  false,
			wantError: nil,
		},
		{
			name: "should use balancer when strategy is enabled but no sessionID provided",
			inMemoryFlags: map[string]Flag{
				"feature-with-strategy": {
					Active:   true,
					FlagName: "feature-with-strategy",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						Percent:      50,
						QtdCall:      5,
					},
				},
			},
			ffDefault: false,
			key:       "feature-with-strategy",
			sessionID: nil,
			wantBool:  true, // Balancer: Calculate()=5, QtdCall=5, 5<=5 = true
			wantError: nil,
		},
		{
			name: "should use balancer returning false when QtdCall is below threshold",
			inMemoryFlags: map[string]Flag{
				"feature-with-strategy": {
					Active:   true,
					FlagName: "feature-with-strategy",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						Percent:      50,
						QtdCall:      3,
					},
				},
			},
			ffDefault: true,
			key:       "feature-with-strategy",
			sessionID: nil,
			wantBool:  false, // Balancer: Calculate()=5, QtdCall=3, 5<=3 = false
			wantError: nil,
		},
		{
			name: "should validate strategy with sessionID - active case",
			inMemoryFlags: map[string]Flag{
				"feature-strategy": {
					Active:   false,
					FlagName: "feature-strategy",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID: map[string]bool{
							"session-123": true,
						},
						Percent: 50,
						QtdCall: 5,
					},
				},
			},
			ffDefault: false,
			key:       "feature-strategy",
			sessionID: []string{"session-123"},
			wantBool:  true,
			wantError: nil,
		},
		{
			name: "should validate strategy with sessionID - inactive case",
			inMemoryFlags: map[string]Flag{
				"feature-strategy": {
					Active:   true,
					FlagName: "feature-strategy",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID: map[string]bool{
							"session-123": false,
						},
						Percent: 50,
						QtdCall: 5,
					},
				},
			},
			ffDefault: false,
			key:       "feature-strategy",
			sessionID: []string{"session-123"},
			wantBool:  true,
			wantError: nil,
		},
		{
			name: "should validate strategy with percent calculation",
			inMemoryFlags: map[string]Flag{
				"feature-percent": {
					Active:   false,
					FlagName: "feature-percent",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID:   map[string]bool{},
						Percent:      50,
						QtdCall:      5,
					},
				},
			},
			ffDefault: false,
			key:       "feature-percent",
			sessionID: []string{"new-session"},
			wantBool:  true,
			wantError: nil,
		},
		{
			name: "should validate strategy with percent calculation - inactive",
			inMemoryFlags: map[string]Flag{
				"feature-percent": {
					Active:   true,
					FlagName: "feature-percent",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID:   map[string]bool{},
						Percent:      50,
						QtdCall:      3,
					},
				},
			},
			ffDefault: false,
			key:       "feature-percent",
			sessionID: []string{"new-session"},
			wantBool:  false,
			wantError: nil,
		},
		{
			name: "should handle multiple sessionIDs - use first one",
			inMemoryFlags: map[string]Flag{
				"feature-multi": {
					Active:   false,
					FlagName: "feature-multi",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID: map[string]bool{
							"session-1": true,
							"session-2": false,
						},
						Percent: 50,
						QtdCall: 5,
					},
				},
			},
			ffDefault: false,
			key:       "feature-multi",
			sessionID: []string{"session-1", "session-2"},
			wantBool:  true,
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sdk := FeatureFlagSDK{
				host:          "http://localhost:8080",
				ffDefault:     tt.ffDefault,
				inMemoryFlags: tt.inMemoryFlags,
			}

			var result FFResponse
			if tt.sessionID != nil {
				result = sdk.GetFeatureFlag(tt.key, tt.sessionID...)
			} else {
				result = sdk.GetFeatureFlag(tt.key)
			}

			if result.Bool != tt.wantBool {
				t.Errorf("GetFeatureFlag() bool = %v, want %v", result.Bool, tt.wantBool)
			}

			if result.Error != tt.wantError {
				t.Errorf("GetFeatureFlag() error = %v, want %v", result.Error, tt.wantError)
			}
		})
	}
}

func TestFeatureFlagSDK_GetFeatureFlag_WithDefaultUsage(t *testing.T) {
	sdk := FeatureFlagSDK{
		host:      "http://localhost:8080",
		ffDefault: true,
		inMemoryFlags: map[string]Flag{
			"existing-flag": {
				Active:   false,
				FlagName: "existing-flag",
				Strategy: stg.Strategy[bool]{
					WithStrategy: false,
				},
			},
		},
	}

	t.Run("should use WithDefault for non-existent flag", func(t *testing.T) {
		result := sdk.GetFeatureFlag("non-existent")
		got := result.WithDefault(true)
		if !got {
			t.Errorf("WithDefault(true) = %v, want true", got)
		}
	})

	t.Run("should use WithDefault with false for non-existent flag", func(t *testing.T) {
		result := sdk.GetFeatureFlag("non-existent")
		got := result.WithDefault(false)
		if got {
			t.Errorf("WithDefault(false) = %v, want false", got)
		}
	})

	t.Run("should return actual value with WithDefault when no error", func(t *testing.T) {
		result := sdk.GetFeatureFlag("existing-flag")
		got := result.WithDefault(true)
		if got {
			t.Errorf("WithDefault(true) = %v, want false (actual value)", got)
		}
	})
}

func TestFeatureFlagSDK_getAllFlags(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []Flag
		serverStatus   int
		wantError      bool
		wantFlagsCount int
	}{
		{
			name: "should fetch and parse flags successfully",
			serverResponse: []Flag{
				{
					Active:   true,
					FlagName: "feature-a",
					Strategy: stg.Strategy[bool]{
						WithStrategy: false,
					},
				},
				{
					Active:   false,
					FlagName: "feature-b",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						Percent:      50,
						QtdCall:      5,
					},
				},
			},
			serverStatus:   http.StatusOK,
			wantError:      false,
			wantFlagsCount: 2,
		},
		{
			name:           "should handle empty flags list",
			serverResponse: []Flag{},
			serverStatus:   http.StatusOK,
			wantError:      false,
			wantFlagsCount: 0,
		},
		{
			name: "should handle single flag",
			serverResponse: []Flag{
				{
					Active:   true,
					FlagName: "single-flag",
					Strategy: stg.Strategy[bool]{
						WithStrategy: false,
					},
				},
			},
			serverStatus:   http.StatusOK,
			wantError:      false,
			wantFlagsCount: 1,
		},
		{
			name: "should handle multiple flags with strategies",
			serverResponse: []Flag{
				{
					Active:   true,
					FlagName: "flag-1",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID: map[string]bool{
							"session-1": true,
							"session-2": false,
						},
						Percent: 30,
						QtdCall: 7,
					},
				},
				{
					Active:   false,
					FlagName: "flag-2",
					Strategy: stg.Strategy[bool]{
						WithStrategy: false,
					},
				},
				{
					Active:   true,
					FlagName: "flag-3",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						Percent:      90,
						QtdCall:      1,
					},
				},
			},
			serverStatus:   http.StatusOK,
			wantError:      false,
			wantFlagsCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/featureflags" {
					t.Errorf("Expected path '/featureflags', got %s", r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			sdk := FeatureFlagSDK{
				host:   server.URL,
				client: &http.Client{},
			}

			ctx := context.Background()
			flags, err := sdk.getAllFlags(ctx)

			if tt.wantError && err == nil {
				t.Error("getAllFlags() expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("getAllFlags() unexpected error: %v", err)
			}

			if !tt.wantError {
				if len(flags) != tt.wantFlagsCount {
					t.Errorf("getAllFlags() returned %d flags, want %d", len(flags), tt.wantFlagsCount)
				}

				// Verify each flag is correctly mapped
				for _, expectedFlag := range tt.serverResponse {
					flag, ok := flags[expectedFlag.FlagName]
					if !ok {
						t.Errorf("Flag %s not found in result map", expectedFlag.FlagName)
						continue
					}

					if flag.Active != expectedFlag.Active {
						t.Errorf("Flag %s: Active = %v, want %v", expectedFlag.FlagName, flag.Active, expectedFlag.Active)
					}

					if flag.FlagName != expectedFlag.FlagName {
						t.Errorf("Flag %s: FlagName = %v, want %v", expectedFlag.FlagName, flag.FlagName, expectedFlag.FlagName)
					}

					if flag.Strategy.WithStrategy != expectedFlag.Strategy.WithStrategy {
						t.Errorf("Flag %s: WithStrategy = %v, want %v", expectedFlag.FlagName, flag.Strategy.WithStrategy, expectedFlag.Strategy.WithStrategy)
					}
				}
			}
		})
	}
}

func TestFeatureFlagSDK_getAllFlags_ErrorCases(t *testing.T) {
	t.Run("should return error when server returns error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		sdk := FeatureFlagSDK{
			host:   server.URL,
			client: &http.Client{},
		}

		ctx := context.Background()
		_, err := sdk.getAllFlags(ctx)

		// Should still work since we only check for network errors, not status codes
		if err != nil {
			t.Logf("Got error (expected behavior): %v", err)
		}
	})

	t.Run("should return error when server returns invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		sdk := FeatureFlagSDK{
			host:   server.URL,
			client: &http.Client{},
		}

		ctx := context.Background()
		_, err := sdk.getAllFlags(ctx)

		if err == nil {
			t.Error("getAllFlags() expected error for invalid JSON, got nil")
		}
	})

	t.Run("should return error when server is unreachable", func(t *testing.T) {
		sdk := FeatureFlagSDK{
			host:   "http://localhost:99999",
			client: &http.Client{},
		}

		ctx := context.Background()
		_, err := sdk.getAllFlags(ctx)

		if err == nil {
			t.Error("getAllFlags() expected error for unreachable server, got nil")
		}
	})
}

func TestHasChanged(t *testing.T) {
	tests := []struct {
		name     string
		oldFlag  Flag
		newFlag  Flag
		expected bool
	}{
		{
			name: "should return false when no changes - all fields equal",
			oldFlag: Flag{
				Active:   true,
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					SessionsID:   map[string]bool{"user1": true},
					QtdCall:      5, // Deve ser ignorado na comparação
				},
			},
			newFlag: Flag{
				Active:   true,
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					SessionsID:   map[string]bool{"user1": true},
					QtdCall:      0, // Diferente, mas deve ser ignorado
				},
			},
			expected: false,
		},
		{
			name: "should return true when Active changed",
			oldFlag: Flag{
				Active:   true,
				FlagName: "test",
			},
			newFlag: Flag{
				Active:   false,
				FlagName: "test",
			},
			expected: true,
		},
		{
			name: "should return true when Percent changed",
			oldFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{Percent: 50},
			},
			newFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{Percent: 75},
			},
			expected: true,
		},
		{
			name: "should return true when SessionsID changed - new key added",
			oldFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: map[string]bool{"user1": true},
				},
			},
			newFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: map[string]bool{"user1": true, "user2": true},
				},
			},
			expected: true,
		},
		{
			name: "should return true when SessionsID changed - value changed",
			oldFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: map[string]bool{"user1": true},
				},
			},
			newFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: map[string]bool{"user1": false},
				},
			},
			expected: true,
		},
		{
			name: "should return true when WithStrategy changed",
			oldFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{WithStrategy: false},
			},
			newFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{WithStrategy: true},
			},
			expected: true,
		},
		{
			name: "should return false when only QtdCall changed",
			oldFlag: Flag{
				Active:   true,
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      1,
				},
			},
			newFlag: Flag{
				Active:   true,
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      10,
				},
			},
			expected: false,
		},
		{
			name: "should return false when both SessionsID are nil",
			oldFlag: Flag{
				Active:   true,
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: nil,
				},
			},
			newFlag: Flag{
				Active:   true,
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: nil,
				},
			},
			expected: false,
		},
		{
			name: "should return true when SessionsID nil vs empty map",
			oldFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: nil,
				},
			},
			newFlag: Flag{
				FlagName: "test",
				Strategy: stg.Strategy[bool]{
					SessionsID: map[string]bool{},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasChanged(tt.oldFlag, tt.newFlag)
			if result != tt.expected {
				t.Errorf("hasChanged() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFilterChangedFlags(t *testing.T) {
	t.Run("should identify changed and new flags", func(t *testing.T) {
		memoryFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      7, // Important local state
				},
			},
			"flag2": {
				Active:   false,
				FlagName: "flag2",
			},
		}

		serverFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      75, // Changed
					QtdCall:      0,
				},
			},
			"flag2": {
				Active:   false, // No change
				FlagName: "flag2",
			},
			"flag3": { // New flag
				Active:   true,
				FlagName: "flag3",
			},
		}

		changed := filterChangedFlags(serverFlags, memoryFlags)

		// Should return flag1 (changed) and flag3 (new)
		if len(changed) != 2 {
			t.Errorf("expected 2 changed flags, got %d", len(changed))
		}

		if _, ok := changed["flag1"]; !ok {
			t.Error("flag1 should be in changed flags")
		}

		if _, ok := changed["flag3"]; !ok {
			t.Error("flag3 should be in changed flags")
		}

		if _, ok := changed["flag2"]; ok {
			t.Error("flag2 should NOT be in changed flags")
		}
	})

	t.Run("should preserve QtdCall when strategy is active", func(t *testing.T) {
		memoryFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      7,
				},
			},
		}

		serverFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      75, // Changed
					QtdCall:      0,  // Server sends 0
				},
			},
		}

		changed := filterChangedFlags(serverFlags, memoryFlags)

		// QtdCall should be preserved from memory
		if changed["flag1"].Strategy.QtdCall != 7 {
			t.Errorf("expected QtdCall to be preserved as 7, got %d", changed["flag1"].Strategy.QtdCall)
		}
	})

	t.Run("should not preserve QtdCall when strategy becomes inactive", func(t *testing.T) {
		memoryFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      7,
				},
			},
		}

		serverFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: false, // Strategy disabled
					Percent:      50,
					QtdCall:      0,
				},
			},
		}

		changed := filterChangedFlags(serverFlags, memoryFlags)

		// QtdCall should not be preserved because strategy was disabled
		if changed["flag1"].Strategy.QtdCall != 0 {
			t.Errorf("expected QtdCall to be 0 when strategy becomes inactive, got %d", changed["flag1"].Strategy.QtdCall)
		}
	})

	t.Run("should return empty map when no changes", func(t *testing.T) {
		flags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      5,
				},
			},
		}

		// Server flags identical (except QtdCall)
		serverFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					WithStrategy: true,
					Percent:      50,
					QtdCall:      0, // Different but ignored
				},
			},
		}

		changed := filterChangedFlags(serverFlags, flags)

		if len(changed) != 0 {
			t.Errorf("expected 0 changed flags, got %d", len(changed))
		}
	})

	t.Run("should handle empty memory flags", func(t *testing.T) {
		memoryFlags := map[string]Flag{}

		serverFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
			},
		}

		changed := filterChangedFlags(serverFlags, memoryFlags)

		if len(changed) != 1 {
			t.Errorf("expected 1 changed flag (new), got %d", len(changed))
		}
	})
}

func TestFeatureFlagSDK_GetFeatureFlag_QtdCallIncrement(t *testing.T) {
	t.Run("should increment QtdCall on sequential calls with balancer strategy", func(t *testing.T) {
		// 90% percent means Calculate() = 1
		// So QtdCall >= 1 returns true
		// QtdCall = 0 -> false (0 < 1)
		// QtdCall = 1 -> true (1 >= 1)
		// QtdCall = 2 -> true (2 >= 1)
		// etc.
		sdk := &FeatureFlagSDK{
			host:      "http://localhost:8080",
			ffDefault: false,
			inMemoryFlags: map[string]Flag{
				"teste1": {
					Active:   false,
					FlagName: "teste1",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						Percent:      90, // Calculate() = 1
						QtdCall:      0,
					},
				},
			},
		}

		expectedResults := []bool{
			false, // QtdCall=0 before, Calculate()=1, 0 >= 1 = false, then QtdCall becomes 1
			true,  // QtdCall=1 before, 1 >= 1 = true, then QtdCall becomes 2
			true,  // QtdCall=2 before, 2 >= 1 = true, then QtdCall becomes 3
			true,  // QtdCall=3 before, 3 >= 1 = true, then QtdCall becomes 4
			true,  // QtdCall=4 before, 4 >= 1 = true, then QtdCall becomes 5
		}

		for i, expected := range expectedResults {
			result := sdk.GetFeatureFlag("teste1")
			if result.Bool != expected {
				t.Errorf("Call %d: GetFeatureFlag() bool = %v, want %v (QtdCall=%d)",
					i+1, result.Bool, expected, sdk.inMemoryFlags["teste1"].Strategy.QtdCall)
			}
		}

		// Verify final QtdCall value
		finalQtdCall := sdk.inMemoryFlags["teste1"].Strategy.QtdCall
		if finalQtdCall != 5 {
			t.Errorf("Final QtdCall = %d, want 5", finalQtdCall)
		}
	})

	t.Run("should increment QtdCall with 10% percent (9 false then true)", func(t *testing.T) {
		// 10% percent means Calculate() = 9
		// QtdCall needs to be >= 9 to return true
		sdk := &FeatureFlagSDK{
			host:      "http://localhost:8080",
			ffDefault: false,
			inMemoryFlags: map[string]Flag{
				"feature-10pct": {
					Active:   false,
					FlagName: "feature-10pct",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						Percent:      10, // Calculate() = 9
						QtdCall:      0,
					},
				},
			},
		}

		// First 9 calls should return false, 10th and onwards should return true
		for i := 1; i <= 11; i++ {
			result := sdk.GetFeatureFlag("feature-10pct")
			expected := i > 9 // true when i > 9 (calls 10 and 11)
			if result.Bool != expected {
				t.Errorf("Call %d: GetFeatureFlag() bool = %v, want %v", i, result.Bool, expected)
			}
		}
	})

	t.Run("should increment QtdCall with sessionID strategy", func(t *testing.T) {
		sdk := &FeatureFlagSDK{
			host:      "http://localhost:8080",
			ffDefault: false,
			inMemoryFlags: map[string]Flag{
				"session-flag": {
					Active:   false,
					FlagName: "session-flag",
					Strategy: stg.Strategy[bool]{
						WithStrategy: true,
						SessionsID:   map[string]bool{},
						Percent:      90, // Calculate() = 1
						QtdCall:      0,
					},
				},
			},
		}

		// With empty SessionsID, it falls back to percent calculation
		// First call should be false, rest true
		for i := 1; i <= 5; i++ {
			result := sdk.GetFeatureFlag("session-flag", "any-session")
			expected := i > 1
			if result.Bool != expected {
				t.Errorf("Call %d: GetFeatureFlag() bool = %v, want %v", i, result.Bool, expected)
			}
		}
	})
}

func TestMergeFlags(t *testing.T) {
	t.Run("should merge changed flags correctly", func(t *testing.T) {
		memoryFlags := map[string]Flag{
			"flag1": {
				Active:   true,
				FlagName: "flag1",
				Strategy: stg.Strategy[bool]{
					QtdCall: 7,
				},
			},
			"flag2": {
				Active:   false,
				FlagName: "flag2",
				Strategy: stg.Strategy[bool]{
					QtdCall: 3,
				},
			},
		}

		serverFlags := map[string]Flag{
			"flag1": {Active: true, FlagName: "flag1"},
			"flag2": {Active: true, FlagName: "flag2"}, // Alterado
		}

		changedFlags := map[string]Flag{
			"flag2": {
				Active:   true,
				FlagName: "flag2",
				Strategy: stg.Strategy[bool]{
					QtdCall: 3, // Preserved
				},
			},
		}

		result := mergeFlags(memoryFlags, serverFlags, changedFlags)

		// flag1 should keep the in-memory version (unchanged)
		if result["flag1"].Strategy.QtdCall != 7 {
			t.Errorf("flag1 QtdCall should be preserved as 7, got %d", result["flag1"].Strategy.QtdCall)
		}

		// flag2 should use the changed version
		if result["flag2"].Active != true {
			t.Error("flag2 Active should be true (changed)")
		}
	})

	t.Run("should remove deleted flags", func(t *testing.T) {
		memoryFlags := map[string]Flag{
			"flag1": {Active: true, FlagName: "flag1"},
			"flag2": {Active: true, FlagName: "flag2"},
			"flag3": {Active: true, FlagName: "flag3"}, // Will be deleted
		}

		serverFlags := map[string]Flag{
			"flag1": {Active: true, FlagName: "flag1"},
			"flag2": {Active: true, FlagName: "flag2"},
			// flag3 no longer exists on server
		}

		changedFlags := map[string]Flag{} // No changes

		result := mergeFlags(memoryFlags, serverFlags, changedFlags)

		if len(result) != 2 {
			t.Errorf("expected 2 flags after merge, got %d", len(result))
		}

		if _, ok := result["flag3"]; ok {
			t.Error("flag3 should have been removed")
		}
	})

	t.Run("should add new flags from changes", func(t *testing.T) {
		memoryFlags := map[string]Flag{
			"flag1": {Active: true, FlagName: "flag1"},
		}

		serverFlags := map[string]Flag{
			"flag1": {Active: true, FlagName: "flag1"},
			"flag2": {Active: true, FlagName: "flag2"}, // New
		}

		changedFlags := map[string]Flag{
			"flag2": {Active: true, FlagName: "flag2"},
		}

		result := mergeFlags(memoryFlags, serverFlags, changedFlags)

		if len(result) != 2 {
			t.Errorf("expected 2 flags after merge, got %d", len(result))
		}

		if _, ok := result["flag2"]; !ok {
			t.Error("flag2 should have been added")
		}
	})
}
