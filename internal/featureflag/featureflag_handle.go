package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/IsaacDSC/featureflag/pkg/middlewares"
)

// ProjectAccessChecker decide se o usuário identificado por email pode
// acessar project. Satisfeita por *user.Service (admin sempre pode; membro
// só nos projects liberados para ele).
type ProjectAccessChecker interface {
	HasProjectAccess(ctx context.Context, email, project string) (bool, error)
	// MustChangePassword reporta se email ainda não trocou a senha do
	// primeiro acesso — usado por requirePasswordChanged para bloquear as
	// rotas de feature flag até essa troca acontecer.
	MustChangePassword(ctx context.Context, email string) (bool, error)
}

type Handler struct {
	routes  map[string]func(w http.ResponseWriter, r *http.Request)
	service *Service
	checker ProjectAccessChecker
}

const featureFlagPrefix = "/featureflag/{project}"

func NewFeatureFlagHandler(service *Service, checker ProjectAccessChecker) *Handler {
	handler := new(Handler)
	handler.service = service
	handler.checker = checker
	handler.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"GET /featureflag/projects":                        middlewares.RequireServiceOrLogin(handler.requirePasswordChanged(handler.listProjects)),
		fmt.Sprintf("PATCH %s", featureFlagPrefix):         middlewares.RequireServiceOrLogin(handler.requirePasswordChanged(handler.requireProjectAccess(handler.createOrUpdate))),
		fmt.Sprintf("DELETE %s/{key}", featureFlagPrefix):  middlewares.RequireServiceOrLogin(handler.requirePasswordChanged(handler.requireProjectAccess(handler.delete))),
		fmt.Sprintf("GET %s/all", featureFlagPrefix):       middlewares.RequireServiceOrLogin(handler.requirePasswordChanged(handler.requireProjectAccess(handler.getAll))),
		fmt.Sprintf("GET %s/{key}", featureFlagPrefix):     middlewares.RequireServiceOrLogin(handler.requirePasswordChanged(handler.requireProjectAccess(handler.get))),
		fmt.Sprintf("GET %s/sdk/{key}", featureFlagPrefix): middlewares.Authorization(middlewares.CheckPermission(handler.getFeatureFlagBySDK, middlewares.USERNAME_SDK)),
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

// requireProjectAccess exige que o usuário identificado por RequireLogin (ou
// pelo token de serviço, via RequireServiceOrLogin) tenha permissão sobre o
// project da própria URL. Chamadas via SERVICE_CLIENT_AT (email sentinela
// middlewares.ServiceAccountEmail) continuam irrestritas, como já eram antes
// desta checagem existir — automação não é escopada por project.
func (h *Handler) requireProjectAccess(inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailValue, _ := ctxutils.GetValueCtx(r.Context(), middlewares.EMAIL_KEY).(string)
		if emailValue == middlewares.ServiceAccountEmail {
			inner(w, r)
			return
		}

		project := r.PathValue("project")
		ok, err := h.checker.HasProjectAccess(r.Context(), emailValue, project)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("no access to this project"))
			return
		}

		inner(w, r)
	}
}

// requirePasswordChanged bloqueia rotas de feature flag para um usuário que
// ainda não trocou a senha do primeiro acesso (mesmo critério de
// requireProjectAccess: contas de serviço, sem usuário humano por trás,
// ficam de fora dessa checagem).
func (h *Handler) requirePasswordChanged(inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailValue, _ := ctxutils.GetValueCtx(r.Context(), middlewares.EMAIL_KEY).(string)
		if emailValue == middlewares.ServiceAccountEmail {
			inner(w, r)
			return
		}

		must, err := h.checker.MustChangePassword(r.Context(), emailValue)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if must {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("password change required"))
			return
		}

		inner(w, r)
	}
}

func (h *Handler) createOrUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	project := r.PathValue("project")

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

	featureflag, err := ToDomain(project, payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := h.service.CreateOrUpdate(ctx, project, featureflag); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project := r.PathValue("project")
	if err := ValidateProject(project); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	key := r.PathValue("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveFeatureFlag(ctx, project, key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project := r.PathValue("project")
	if err := ValidateProject(project); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	key := r.PathValue("key")

	if key == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("required key params"))
		return
	}

	sessionID := r.Header.Get("session_id")
	ff, err := h.service.GetFeatureFlag(ctx, project, key, sessionID)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (h *Handler) getFeatureFlagBySDK(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project := r.PathValue("project")
	if err := ValidateProject(project); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	key := r.PathValue("key")

	if key == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("required key params"))
		return
	}

	sessionID := r.Header.Get("session_id")
	statusFF, err := h.service.GetFeatureFlagBySDK(ctx, project, key, sessionID)

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

// filterAccessibleProjects reduz projects à sublista que emailValue pode
// acessar (contas de serviço veem tudo, sem filtro — mesmo critério de
// requireProjectAccess).
func (h *Handler) filterAccessibleProjects(ctx context.Context, emailValue string, projects []string) ([]string, error) {
	if emailValue == middlewares.ServiceAccountEmail {
		return projects, nil
	}

	filtered := make([]string, 0, len(projects))
	for _, project := range projects {
		ok, err := h.checker.HasProjectAccess(ctx, emailValue, project)
		if err != nil {
			return nil, err
		}
		if ok {
			filtered = append(filtered, project)
		}
	}
	return filtered, nil
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.service.ListProjects(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	emailValue, _ := ctxutils.GetValueCtx(ctx, middlewares.EMAIL_KEY).(string)
	projects, err = h.filterAccessibleProjects(ctx, emailValue, projects)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(projects)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project := r.PathValue("project")
	if err := ValidateProject(project); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO: Possibilitar receber um parametro de query para filtrar por status
	// status := r.URL.Query().Get("status")

	database, err := h.service.GetAllFeatureFlag(ctx, project)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
