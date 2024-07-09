package web

import (
	"encoding/json"
	"github.com/IsaacDSC/featureflag/internal/domain"
	"github.com/IsaacDSC/featureflag/internal/dto"
	"github.com/IsaacDSC/featureflag/internal/errorutils"
	"io"
	"net/http"
)

type Handler struct {
	routes  map[string]func(w http.ResponseWriter, r *http.Request)
	service *domain.FeatureflagService
}

func NewHandler(service *domain.FeatureflagService) *Handler {
	handler := new(Handler)
	handler.service = service
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"PATCH /":       handler.createOrUpdate,
		"DELETE /{key}": handler.delete,
		"GET /":         handler.getAll,
		"GET /{key}":    handler.get,
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h *Handler) createOrUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload dto.FeatureflagDTO
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	featureflag, err := dto.FeatureFlagToDomain(payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := h.service.CreateOrUpdate(featureflag); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveFeatureFlag(key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if key == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("required key params"))
		return
	}

	sessionID := r.Header.Get("session_id")
	ff, err := h.service.GetFeatureFlag(key, sessionID)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			statusError := err.(errorutils.NotFoundError)
			w.WriteHeader(statusError.GetStatusCode())
			w.Write([]byte(err.Error()))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	output, err := json.Marshal(dto.FeatureFlagFromDomain(ff))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	database, err := h.service.GetAllFeatureFlag()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(database)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
