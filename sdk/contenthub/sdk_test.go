package contenthub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IsaacDSC/featureflag/internal/contenthub"
)

func TestContenthubSDK_Content(t *testing.T) {
	tests := []struct {
		name        string
		db          map[string]Content
		key         string
		sessionID   []string
		ffDefault   Value
		wantValue   string
		wantErr     error
		description string
	}{
		{
			name:        "content not found",
			db:          map[string]Content{},
			key:         "non-existent",
			sessionID:   nil,
			ffDefault:   Value(`false`),
			wantValue:   "false",
			wantErr:     ErrNotFoundContenthub,
			description: "should return error when content key doesn't exist",
		},
		{
			name: "content found without sessionID - uses balancer",
			db: map[string]Content{
				"test-content": {
					Key:         "test-content",
					Description: "test",
					CreatedAt:   time.Now(),
					BalancerStrategy: contenthub.BalancerStrategy{
						{Weight: 100, Response: "response-a", Qtt: 0},
					},
				},
			},
			key:         "test-content",
			sessionID:   nil,
			ffDefault:   Value(`false`),
			wantValue:   `"response-a"`,
			wantErr:     nil,
			description: "should return balancer distribution when no sessionID",
		},
		{
			name: "content found with sessionID - uses session strategy",
			db: map[string]Content{
				"test-content": {
					Key:         "test-content",
					Description: "test",
					CreatedAt:   time.Now(),
					SessionStrategy: contenthub.SessionsStrategies{
						{SessionID: "session-123", Response: "session-response"},
						{SessionID: "default", Response: "default-response"},
					},
				},
			},
			key:         "test-content",
			sessionID:   []string{"session-123"},
			ffDefault:   Value(`false`),
			wantValue:   `"session-response"`,
			wantErr:     nil,
			description: "should return session response when sessionID matches",
		},
		{
			name: "content with sessionID not found - uses default",
			db: map[string]Content{
				"test-content": {
					Key:         "test-content",
					Description: "test",
					CreatedAt:   time.Now(),
					SessionStrategy: contenthub.SessionsStrategies{
						{SessionID: "session-456", Response: "other-session"},
						{SessionID: "default", Response: "default-response"},
					},
				},
			},
			key:         "test-content",
			sessionID:   []string{"session-123"},
			ffDefault:   Value(`false`),
			wantValue:   `"default-response"`,
			wantErr:     nil,
			description: "should return default response when sessionID doesn't match",
		},
		{
			name: "content with complex response object",
			db: map[string]Content{
				"test-content": {
					Key:         "test-content",
					Description: "test",
					CreatedAt:   time.Now(),
					BalancerStrategy: contenthub.BalancerStrategy{
						{Weight: 100, Response: map[string]any{"enabled": true, "value": "test"}, Qtt: 0},
					},
				},
			},
			key:         "test-content",
			sessionID:   nil,
			ffDefault:   Value(`{}`),
			wantValue:   `{"enabled":true,"value":"test"}`,
			wantErr:     nil,
			description: "should handle complex response objects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sdk := &ContenthubSDK{
				host:      "http://localhost:8080",
				client:    &http.Client{},
				ffDefault: tt.ffDefault,
				db:        tt.db,
			}

			result := sdk.Content(tt.key, tt.sessionID...)

			value, err := result.Err()

			if err != tt.wantErr {
				t.Errorf("Content() Error = %v, want %v", err, tt.wantErr)
			}

			if err == nil {
				if string(value) != tt.wantValue {
					t.Errorf("Content() Value = %v, want %v", string(value), tt.wantValue)
				}
			}
		})
	}
}

func TestResult_Err(t *testing.T) {
	tests := []struct {
		name    string
		result  Result
		wantVal string
		wantErr error
	}{
		{
			name: "result with value and no error",
			result: Result{
				value: Value(`"test-value"`),
				error: nil,
			},
			wantVal: `"test-value"`,
			wantErr: nil,
		},
		{
			name: "result with value and error",
			result: Result{
				value: Value(`false`),
				error: ErrNotFoundContenthub,
			},
			wantVal: "false",
			wantErr: ErrNotFoundContenthub,
		},
		{
			name: "result with empty value and error",
			result: Result{
				value: Value(``),
				error: ErrNotFoundContenthub,
			},
			wantVal: "",
			wantErr: ErrNotFoundContenthub,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotErr := tt.result.Err()

			if string(gotVal) != tt.wantVal {
				t.Errorf("Err() value = %v, want %v", string(gotVal), tt.wantVal)
			}

			if gotErr != tt.wantErr {
				t.Errorf("Err() error = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestResult_Val(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "returns value from result",
			result: Result{
				value: Value(`"test-value"`),
				error: nil,
			},
			want: `"test-value"`,
		},
		{
			name: "returns value even with error",
			result: Result{
				value: Value(`false`),
				error: ErrNotFoundContenthub,
			},
			want: "false",
		},
		{
			name: "returns empty value",
			result: Result{
				value: Value(``),
				error: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.Val()

			if string(got) != tt.want {
				t.Errorf("Val() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "converts value to string",
			result: Result{
				value: Value(`"test-value"`),
				error: nil,
			},
			want: `"test-value"`,
		},
		{
			name: "converts boolean value to string",
			result: Result{
				value: Value(`true`),
				error: nil,
			},
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.String()

			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_DecodeJson(t *testing.T) {
	tests := []struct {
		name    string
		result  Result
		target  any
		wantErr bool
		check   func(t *testing.T, target any)
	}{
		{
			name: "decode simple string",
			result: Result{
				value: Value(`"test-value"`),
				error: nil,
			},
			target:  new(string),
			wantErr: false,
			check: func(t *testing.T, target any) {
				str := target.(*string)
				if *str != "test-value" {
					t.Errorf("DecodeJson() decoded = %v, want %v", *str, "test-value")
				}
			},
		},
		{
			name: "decode boolean",
			result: Result{
				value: Value(`true`),
				error: nil,
			},
			target:  new(bool),
			wantErr: false,
			check: func(t *testing.T, target any) {
				b := target.(*bool)
				if *b != true {
					t.Errorf("DecodeJson() decoded = %v, want %v", *b, true)
				}
			},
		},
		{
			name: "decode object",
			result: Result{
				value: Value(`{"enabled":true,"value":"test"}`),
				error: nil,
			},
			target:  new(map[string]any),
			wantErr: false,
			check: func(t *testing.T, target any) {
				m := target.(*map[string]any)
				if (*m)["enabled"] != true || (*m)["value"] != "test" {
					t.Errorf("DecodeJson() decoded object incorrectly: %v", *m)
				}
			},
		},
		{
			name: "decode invalid JSON",
			result: Result{
				value: Value(`invalid json`),
				error: nil,
			},
			target:  new(string),
			wantErr: true,
			check:   func(t *testing.T, target any) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.DecodeJson(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.check(t, tt.target)
			}
		})
	}
}

func TestContenthubSDK_getAllContents(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		serverError    bool
		wantErr        bool
		wantCount      int
		description    string
	}{
		{
			name: "successfully gets all contents",
			serverResponse: `[
				{
					"key": "feature1",
					"description": "Test feature 1",
					"created_at": "2024-01-01T00:00:00Z",
					"strategy": {"with_strategy": false, "session_id": {}, "percent": 0, "qtd_call": 0},
					"session_strategy": [],
					"balancer_strategy": [{"weight": 100, "response": "value1"}]
				},
				{
					"key": "feature2",
					"description": "Test feature 2",
					"created_at": "2024-01-01T00:00:00Z",
					"strategy": {"with_strategy": false, "session_id": {}, "percent": 0, "qtd_call": 0},
					"session_strategy": [],
					"balancer_strategy": [{"weight": 100, "response": "value2"}]
				}
			]`,
			statusCode:  http.StatusOK,
			serverError: false,
			wantErr:     false,
			wantCount:   2,
			description: "should parse valid JSON response with multiple contents",
		},
		{
			name:           "handles empty response",
			serverResponse: `[]`,
			statusCode:     http.StatusOK,
			serverError:    false,
			wantErr:        false,
			wantCount:      0,
			description:    "should handle empty array response",
		},
		{
			name:           "handles invalid JSON",
			serverResponse: `invalid json`,
			statusCode:     http.StatusOK,
			serverError:    false,
			wantErr:        true,
			wantCount:      0,
			description:    "should return error for invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/contenthubs" {
					t.Errorf("Expected path /contenthubs, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			sdk := NewContenthubSDK(server.URL)
			ctx := context.Background()

			contents, err := sdk.getAllContents(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("getAllContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(contents) != tt.wantCount {
				t.Errorf("getAllContents() returned %d contents, want %d", len(contents), tt.wantCount)
			}

			if !tt.wantErr && tt.wantCount > 0 {
				for key, content := range contents {
					if content.Key != key {
						t.Errorf("Content key mismatch: map key = %s, content.Key = %s", key, content.Key)
					}
				}
			}
		})
	}
}

func TestContenthubSDK_getAllContents_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid": "not an array"}`))
	}))
	defer server.Close()

	sdk := NewContenthubSDK(server.URL)
	ctx := context.Background()

	_, err := sdk.getAllContents(ctx)
	if err == nil {
		t.Error("getAllContents() expected error for invalid JSON structure, got nil")
	}
}

func TestContenthubSDK_Integration(t *testing.T) {
	content1 := Content{
		Key:         "feature1",
		Description: "Integration test feature",
		CreatedAt:   time.Now(),
		BalancerStrategy: contenthub.BalancerStrategy{
			{Weight: 50, Response: "response-a", Qtt: 0},
			{Weight: 50, Response: "response-b", Qtt: 0},
		},
		SessionStrategy: contenthub.SessionsStrategies{
			{SessionID: "user-123", Response: "user-specific"},
			{SessionID: "default", Response: "default-response"},
		},
	}

	responseJSON, _ := json.Marshal([]Content{content1})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}))
	defer server.Close()

	sdk := NewContenthubSDK(server.URL)
	ctx := context.Background()

	contents, err := sdk.getAllContents(ctx)
	if err != nil {
		t.Fatalf("getAllContents() failed: %v", err)
	}

	sdk.db = contents

	// Test without sessionID - should use balancer
	result1 := sdk.Content("feature1")
	val1, err1 := result1.Err()
	if err1 != nil {
		t.Errorf("Content() without sessionID failed: %v", err1)
	}
	if len(val1) == 0 {
		t.Error("Content() without sessionID returned empty value")
	}

	// Test with sessionID - should use session strategy
	result2 := sdk.Content("feature1", "user-123")
	val2, err2 := result2.Err()
	if err2 != nil {
		t.Errorf("Content() with sessionID failed: %v", err2)
	}
	if string(val2) != `"user-specific"` {
		t.Errorf("Content() with sessionID = %s, want %s", string(val2), `"user-specific"`)
	}

	// Test with non-existent sessionID - should use default
	result3 := sdk.Content("feature1", "unknown-user")
	val3, err3 := result3.Err()
	if err3 != nil {
		t.Errorf("Content() with unknown sessionID failed: %v", err3)
	}
	if string(val3) != `"default-response"` {
		t.Errorf("Content() with unknown sessionID = %s, want %s", string(val3), `"default-response"`)
	}

	// Test with non-existent key
	result4 := sdk.Content("non-existent")
	_, err4 := result4.Err()
	if err4 != ErrNotFoundContenthub {
		t.Errorf("Content() with non-existent key error = %v, want %v", err4, ErrNotFoundContenthub)
	}
}
