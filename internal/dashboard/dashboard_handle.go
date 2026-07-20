package dashboard

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

//go:embed static
var staticFiles embed.FS

type Handler struct {
	routes map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewDashboardHandler() *Handler {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	fileServer := http.StripPrefix("/dashboard/", http.FileServerFS(static))

	handler := new(Handler)
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"GET /dashboard": func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/dashboard/", http.StatusMovedPermanently)
		},
		"GET /dashboard/": middlewares.RequireLogin(fileServer.ServeHTTP),
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}
