package web

import (
	"errors"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/utils/authutils"
	"github.com/IsaacDSC/featureflag/utils/ctxutils"
	"net/http"
)

const (
	KEY              = "client"
	USERNAME_SERVICE = "SERVICE_CLIENT"
	USERNAME_SDK     = "SDK_CLIENT"
)

func SDKPermission(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("token")
		data, err := authutils.GetDataJWT(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		username := data.(string)
		key := ctxutils.GetValueCtx(r.Context(), KEY)
		if key != USERNAME_SDK || username != USERNAME_SDK {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func ClientServicePermission(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("token")
		data, err := authutils.GetDataJWT(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		username := data.(string)
		key := ctxutils.GetValueCtx(r.Context(), KEY)
		if key != USERNAME_SERVICE || username != USERNAME_SERVICE {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func Authorization(h http.HandlerFunc) http.HandlerFunc {
	cfg := env.Get()
	var users = map[string]string{
		cfg.ServiceClientAT: USERNAME_SERVICE,
		cfg.SDKClientAT:     USERNAME_SDK,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		clientName, ok := users[authorization]
		if !ok {
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
