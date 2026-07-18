package user

import (
	"encoding/json"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

type Handler struct {
	routes  map[string]func(w http.ResponseWriter, r *http.Request)
	service *Service
}

func NewUserHandler(service *Service) *Handler {
	handler := new(Handler)
	handler.service = service
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"POST /users":           middlewares.RequireLogin(handler.create),
		"DELETE /users/{email}": middlewares.RequireLogin(handler.delete),
		"GET /users":            middlewares.RequireLogin(handler.list),
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	var payload CreateUserDto
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Email == "" || payload.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("email and password are required"))
		return
	}

	if err := h.service.Register(ctx, payload.Email, payload.Password); err != nil {
		switch err.(type) {
		case *errorutils.ConflictError:
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	emailInput := r.PathValue("email")
	if emailInput == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users, err := h.service.List(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if len(users) <= 1 {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("cannot remove the last remaining user"))
		return
	}

	if err := h.service.Remove(ctx, emailInput); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.service.List(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
