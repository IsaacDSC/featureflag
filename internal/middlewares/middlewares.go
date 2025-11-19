package middlewares

import (
	"errors"
	"net/http"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/authutils"
	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
)

const (
	KEY              = "client"
	USERNAME_SERVICE = "SERVICE_CLIENT"
	USERNAME_SDK     = "SDK_CLIENT"
)

func getClientPermission(key string) (string, error) {
	cfg := env.Get()
	if client, ok := map[string]string{
		cfg.ServiceClientAT: USERNAME_SERVICE,
		cfg.SDKClientAT:     USERNAME_SDK,
	}[key]; ok {
		return client, nil
	}

	return "", errors.New("client not found")
}

func CheckPermission(h http.HandlerFunc, permission string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := ctxutils.GetValueCtx(r.Context(), KEY)
		if key == permission {
			h.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusForbidden)
	}
}

func Authorization(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		clientName, err := getClientPermission(authorization)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		request := r.WithContext(ctxutils.SetContext(r.Context(), KEY, clientName))

		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, request)
	}
}

func Authentication(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := authutils.VerifyToken(cookie.Value); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	}
}
