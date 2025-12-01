package contenthub

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		fmt.Sprintf("PATCH %s", contenthubRouterPrefix):         handler.patchContenthub,    //middlewares.Authorization(middlewares.CheckPermission(handler.patchContenthub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("DELETE %s/{key}", contenthubRouterPrefix):  handler.deleteContenthub,   //middlewares.Authorization(middlewares.CheckPermission(handler.deleteContenthub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %ss", contenthubRouterPrefix):          handler.getAllContenthub,   //middlewares.Authorization(middlewares.CheckPermission(handler.getAllContenthub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %s/{key}", contenthubRouterPrefix):     handler.getContentHub,      //middlewares.Authentication(middlewares.CheckPermission(handler.getContentHub, middlewares.USERNAME_SERVICE)),
		fmt.Sprintf("GET %s/sdk/{key}", contenthubRouterPrefix): handler.getContentHubBySDK, //middlewares.Authentication(middlewares.CheckPermission(handler.getContentHubBySDK, middlewares.USERNAME_SDK)),
	}

	return handler
}

func (h ContenthubHandler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h ContenthubHandler) patchContenthub(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()

	var payload Dto
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Write([]byte("error on decode body"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payloadEntity, err := payload.ToDomain()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := h.service.CreateOrUpdate(ctx, payloadEntity); err != nil {
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

	var result []Entity
	for _, entity := range contents {
		result = append(result, entity)
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
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
