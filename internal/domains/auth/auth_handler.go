package auth

import (
	"github.com/IsaacDSC/featureflag/internal/middlewares"
	"github.com/IsaacDSC/featureflag/utils/authutils"
	"github.com/IsaacDSC/featureflag/utils/ctxutils"
	"net/http"
	"time"
)

type AuthHandler struct {
	routes map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewAuthHandler() *AuthHandler {
	handler := new(AuthHandler)
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"POST /auth": middlewares.Authorization(handler.auth),
	}

	return handler
}

func (h *AuthHandler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h *AuthHandler) auth(w http.ResponseWriter, r *http.Request) {
	username := ctxutils.GetValueCtx(r.Context(), middlewares.KEY)
	token, err := authutils.CreateToken(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
	})
}
