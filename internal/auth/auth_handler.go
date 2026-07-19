package auth

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"github.com/IsaacDSC/featureflag/internal/user"
	"github.com/IsaacDSC/featureflag/pkg/authutils"
	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

//go:embed static
var staticFiles embed.FS

type AuthHandler struct {
	routes      map[string]func(w http.ResponseWriter, r *http.Request)
	userService *user.Service
}

func NewAuthHandler(userService *user.Service) *AuthHandler {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	fileServer := http.StripPrefix("/auth/login/", http.FileServerFS(static))

	handler := new(AuthHandler)
	handler.userService = userService
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"POST /auth": middlewares.Authorization(handler.auth),
		"GET /auth/login": func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/auth/login/", http.StatusMovedPermanently)
		},
		"GET /auth/login/":  fileServer.ServeHTTP,
		"POST /auth/login":  handler.login,
		"POST /auth/logout": middlewares.RequireLogin(handler.logout),
	}

	return handler
}

func (h *AuthHandler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

// auth mantém o fluxo legado (SERVICE_CLIENT/SDK_CLIENT -> cookie "token"),
// usado por automação, não pelo login humano descrito abaixo.
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

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	var payload loginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !h.userService.Authenticate(ctx, payload.Email, payload.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid email or password"))
		return
	}

	token, err := authutils.CreateToken(map[string]string{"email": payload.Email})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}
