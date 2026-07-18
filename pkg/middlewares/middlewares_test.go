package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/authutils"
	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
)

func TestRedactSensitiveFields(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "redacts password, keeps other fields",
			body: `{"email":"alice@example.com","password":"s3cret"}`,
			want: `{"email":"alice@example.com","password":"***REDACTED***"}`,
		},
		{
			name: "no sensitive field returns the body unchanged",
			body: `{"flag_name":"checkout","active":true}`,
			want: `{"flag_name":"checkout","active":true}`,
		},
		{
			name: "empty body returns unchanged",
			body: "",
			want: "",
		},
		{
			name: "non-JSON body returns unchanged, no panic",
			body: "not-json-at-all",
			want: "not-json-at-all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactSensitiveFields(tt.body)

			if tt.body == tt.want {
				// nada a redigir: deve devolver exatamente o mesmo texto, sem
				// depender de reserialização (preserva ordem/formatação).
				if got != tt.want {
					t.Errorf("redactSensitiveFields() = %q, want unchanged %q", got, tt.want)
				}
				return
			}

			var gotMap map[string]any
			if err := json.Unmarshal([]byte(got), &gotMap); err != nil {
				t.Fatalf("redactSensitiveFields() produced invalid JSON: %v", err)
			}
			if gotMap["password"] != "***REDACTED***" {
				t.Errorf("password field = %v, want ***REDACTED***", gotMap["password"])
			}
			if gotMap["email"] != "alice@example.com" {
				t.Errorf("email field = %v, want untouched", gotMap["email"])
			}
		})
	}
}

func TestRequireLogin(t *testing.T) {
	env.Override(env.Environment{SecretKey: "test-secret", ServiceClientAT: "service-token"})

	handlerCalled := false
	var gotEmail any
	next := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		gotEmail = ctxutils.GetValueCtx(r.Context(), EMAIL_KEY)
		w.WriteHeader(http.StatusOK)
	}

	t.Run("missing cookie returns 401 for API calls", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodPatch, "/featureflag/checkout", nil)
		rec := httptest.NewRecorder()

		RequireLogin(next)(rec, req)

		if handlerCalled {
			t.Error("handler should not be called without a session cookie")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("missing cookie redirects to /auth/login for page navigation", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodGet, "/dashboard/", nil)
		req.Header.Set("Accept", "text/html")
		rec := httptest.NewRecorder()

		RequireLogin(next)(rec, req)

		if handlerCalled {
			t.Error("handler should not be called without a session cookie")
		}
		if rec.Code != http.StatusFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
		}
		if loc := rec.Header().Get("Location"); loc != "/auth/login" {
			t.Errorf("Location = %q, want /auth/login", loc)
		}
	})

	t.Run("valid session cookie propagates the email in context", func(t *testing.T) {
		handlerCalled = false
		token, err := authutils.CreateToken(map[string]string{"email": "alice@example.com"})
		if err != nil {
			t.Fatalf("CreateToken() unexpected error = %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, "/featureflag/checkout", nil)
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
		rec := httptest.NewRecorder()

		RequireLogin(next)(rec, req)

		if !handlerCalled {
			t.Fatal("handler should be called with a valid session cookie")
		}
		if gotEmail != "alice@example.com" {
			t.Errorf("email in context = %v, want alice@example.com", gotEmail)
		}
	})
}

func TestRequireServiceOrLogin(t *testing.T) {
	env.Override(env.Environment{SecretKey: "test-secret", ServiceClientAT: "service-token"})

	var gotEmail any
	next := func(w http.ResponseWriter, r *http.Request) {
		gotEmail = ctxutils.GetValueCtx(r.Context(), EMAIL_KEY)
		w.WriteHeader(http.StatusOK)
	}

	t.Run("static service token bypasses the cookie check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/featureflag/checkout", nil)
		req.Header.Set("Authorization", "service-token")
		rec := httptest.NewRecorder()

		RequireServiceOrLogin(next)(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if gotEmail != "service-account" {
			t.Errorf("email in context = %v, want service-account", gotEmail)
		}
	})

	t.Run("missing service token falls back to RequireLogin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/featureflag/checkout", nil)
		rec := httptest.NewRecorder()

		RequireServiceOrLogin(next)(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}
