package contenthub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

type ContenthubHandler struct {
	routes  map[string]func(w http.ResponseWriter, r *http.Request)
	service *Service
}

const contenthubRouterPrefix = "/contenthub"

func NewContenthubHandler(service *Service) *ContenthubHandler {
	handler := new(ContenthubHandler)
	handler.service = service
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		fmt.Sprintf("PATCH %s", contenthubRouterPrefix):         middlewares.Authorization(middlewares.CheckPermission(handler.patchContenthub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("DELETE %s/{key}", contenthubRouterPrefix):  middlewares.Authorization(middlewares.CheckPermission(handler.deleteContenthub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %s", contenthubRouterPrefix):           middlewares.Authorization(middlewares.CheckPermission(handler.getAllContenthub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %s/{key}", contenthubRouterPrefix):     middlewares.Authentication(middlewares.CheckPermission(handler.getContentHub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %s/sdk/{key}", contenthubRouterPrefix): middlewares.Authentication(middlewares.CheckPermission(handler.getContentHubBySDK, middlewares.USERNAME_SDK)),
	}

	return handler
}

func (h ContenthubHandler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h ContenthubHandler) patchContenthub(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload Dto
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	payloadEntity, err := payload.ToDomain()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.CreateOrUpdate(payloadEntity); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h ContenthubHandler) getContentHub(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	content, err := h.service.GetContentHub(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	payload := FromDomain(content)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h ContenthubHandler) getAllContenthub(w http.ResponseWriter, r *http.Request) {
	contents, err := h.service.GetAllContentHub()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(contents); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (h ContenthubHandler) deleteContenthub(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveContentHub(key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// TODO: validate and implement this method correctly
func (h ContenthubHandler) getContentHubBySDK(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	content, err := h.service.GetContentHub(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	payload := FromDomain(content)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
