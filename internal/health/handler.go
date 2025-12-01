package health

import (
	"fmt"
	"net/http"
)

type Handler struct {
	routes map[string]func(w http.ResponseWriter, r *http.Request)
}

const featureFlagPrefix = "/ping"

func NewHandler() *Handler {
	handler := new(Handler)
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		fmt.Sprintf("GET %s", featureFlagPrefix): handler.Ping,
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}
