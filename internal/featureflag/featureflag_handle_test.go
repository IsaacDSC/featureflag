package featureflag

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/authutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
	"github.com/golang/mock/gomock"
)

// fakeChecker is a minimal ProjectAccessChecker test double: project access
// is whatever the test wires into allowed, no user lookup involved.
type fakeChecker struct {
	allowed map[string]bool
}

func (f fakeChecker) HasProjectAccess(ctx context.Context, email, project string) (bool, error) {
	return f.allowed[project], nil
}

func (f fakeChecker) MustChangePassword(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func newTestMux(t *testing.T, service *Service, checker ProjectAccessChecker) *http.ServeMux {
	t.Helper()
	handler := NewFeatureFlagHandler(service, checker)
	mux := http.NewServeMux()
	for pattern, h := range handler.GetRoutes() {
		mux.HandleFunc(pattern, h)
	}
	return mux
}

func loginCookie(t *testing.T, email string) *http.Cookie {
	t.Helper()
	token, err := authutils.CreateToken(map[string]string{"email": email})
	if err != nil {
		t.Fatalf("CreateToken() unexpected error = %v", err)
	}
	return &http.Cookie{Name: middlewares.SessionCookieName, Value: token}
}

// TestRequireProjectAccess_GetAll is a regression test: a logged-in user
// must not be able to read GET /featureflag/{project}/all for a project
// they were not granted access to. This route previously had no middleware
// at all (open to anyone, logged in or not) — this test would have caught
// that.
func TestRequireProjectAccess_GetAll(t *testing.T) {
	env.Override(env.Environment{SecretKey: "test-secret", ServiceClientAT: "service-token"})

	control := gomock.NewController(t)
	repo := NewMockFeatureFlagRepository(control)
	repo.EXPECT().GetAllFF(gomock.Any(), gomock.Any()).Return(map[string]Entity{}, nil).AnyTimes()

	service := NewFeatureflagService(repo, noopAuditor{})
	checker := fakeChecker{allowed: map[string]bool{"checkout": true}}
	mux := newTestMux(t, service, checker)

	t.Run("member without access to the project is denied", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/featureflag/billing/all", nil)
		req.AddCookie(loginCookie(t, "member@example.com"))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})

	t.Run("member with access to the project can read it", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/featureflag/checkout/all", nil)
		req.AddCookie(loginCookie(t, "member@example.com"))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("no session at all is rejected, not silently allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/featureflag/checkout/all", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})

	t.Run("static service token bypasses project scoping (automation, not a project-scoped human)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/featureflag/billing/all", nil)
		req.Header.Set("Authorization", "service-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})
}

// TestRequireProjectAccess_WriteAndReadByKey covers PATCH, DELETE/{key} and
// GET/{key} — the routes requireProjectAccess was originally added for.
func TestRequireProjectAccess_WriteAndReadByKey(t *testing.T) {
	env.Override(env.Environment{SecretKey: "test-secret", ServiceClientAT: "service-token"})

	control := gomock.NewController(t)
	repo := NewMockFeatureFlagRepository(control)
	repo.EXPECT().GetFF(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(Entity{}, nil).AnyTimes()
	repo.EXPECT().SaveFF(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	repo.EXPECT().DeleteFF(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	service := NewFeatureflagService(repo, noopAuditor{})
	checker := fakeChecker{allowed: map[string]bool{"checkout": true}}
	mux := newTestMux(t, service, checker)
	member := loginCookie(t, "member@example.com")

	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{name: "PATCH denied on ungranted project", method: http.MethodPatch, path: "/featureflag/billing", want: http.StatusForbidden},
		{name: "PATCH allowed on granted project", method: http.MethodPatch, path: "/featureflag/checkout", want: http.StatusNoContent},
		{name: "DELETE denied on ungranted project", method: http.MethodDelete, path: "/featureflag/billing/some-flag", want: http.StatusForbidden},
		{name: "DELETE allowed on granted project", method: http.MethodDelete, path: "/featureflag/checkout/some-flag", want: http.StatusNoContent},
		{name: "GET by key denied on ungranted project", method: http.MethodGet, path: "/featureflag/billing/some-flag", want: http.StatusForbidden},
		{name: "GET by key allowed on granted project", method: http.MethodGet, path: "/featureflag/checkout/some-flag", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPatch {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"flag_name":"some-flag","active":true}`))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req.AddCookie(member)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d, body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

// TestListProjects_FiltersToAccessibleProjects is a regression test for the
// same class of leak: GET /featureflag/projects must not reveal project
// names the caller has no access to.
func TestListProjects_FiltersToAccessibleProjects(t *testing.T) {
	env.Override(env.Environment{SecretKey: "test-secret", ServiceClientAT: "service-token"})

	control := gomock.NewController(t)
	repo := NewMockFeatureFlagRepository(control)
	repo.EXPECT().ListProjects(gomock.Any()).Return([]string{"billing", "checkout", "marketing"}, nil).AnyTimes()

	service := NewFeatureflagService(repo, noopAuditor{})
	checker := fakeChecker{allowed: map[string]bool{"checkout": true, "marketing": true}}
	mux := newTestMux(t, service, checker)

	t.Run("member only sees granted project names", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/featureflag/projects", nil)
		req.AddCookie(loginCookie(t, "member@example.com"))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		got := rec.Body.String()
		if !strings.Contains(got, "checkout") || !strings.Contains(got, "marketing") || strings.Contains(got, "billing") {
			t.Errorf("body = %s, want only checkout and marketing, never billing", got)
		}
	})

	t.Run("service token sees every project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/featureflag/projects", nil)
		req.Header.Set("Authorization", "service-token")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		got := rec.Body.String()
		if !strings.Contains(got, "billing") || !strings.Contains(got, "checkout") || !strings.Contains(got, "marketing") {
			t.Errorf("body = %s, want all three projects for a service account", got)
		}
	})
}
