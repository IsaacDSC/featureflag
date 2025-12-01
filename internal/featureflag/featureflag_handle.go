package featureflag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

type Handler struct {
	routes  map[string]func(w http.ResponseWriter, r *http.Request)
	service *Service
}

const featureFlagPrefix = "/featureflag"

func NewFeatureFlagHandler(service *Service) *Handler {
	handler := new(Handler)
	handler.service = service
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		fmt.Sprintf("PATCH %s", featureFlagPrefix):         handler.createOrUpdate,
		fmt.Sprintf("DELETE %s/{key}", featureFlagPrefix):  middlewares.Authorization(middlewares.CheckPermission(handler.delete, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %ss", featureFlagPrefix):          handler.getAll,
		fmt.Sprintf("GET %s/{key}", featureFlagPrefix):     middlewares.Authorization(middlewares.CheckPermission(handler.get, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %s/sdk/{key}", featureFlagPrefix): middlewares.Authorization(middlewares.CheckPermission(handler.getFeatureFlagBySDK, middlewares.USERNAME_SDK)),
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h *Handler) createOrUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload Dto
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	featureflag, err := ToDomain(payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := h.service.CreateOrUpdate(ctx, featureflag); err != nil {
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
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("feature flag not found"))
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	output, err := json.Marshal(DtoFromDomain(ff))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (h *Handler) getFeatureFlagBySDK(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if key == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("required key params"))
		return
	}

	sessionID := r.Header.Get("session_id")
	statusFF, err := h.service.GetFeatureFlagBySDK(key, sessionID)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("feature flag not found"))
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": "%t"}`, statusFF)))
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	// TODO: Possibilitar receber um parametro de query para filtrar por status
	// status := r.URL.Query().Get("status")

	database, err := h.service.GetAllFeatureFlag()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var result []Entity
	for _, entity := range database {
		result = append(result, entity)
	}

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
