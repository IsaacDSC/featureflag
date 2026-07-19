package user

import (
	"encoding/json"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
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
		// Gestão de usuários (criar/remover/listar/alterar acesso) é
		// restrita a admins — membros não gerenciam outros usuários. Também
		// exige que o admin já tenha trocado a senha do primeiro acesso
		// (requirePasswordChanged), senão poderia contornar a exigência
		// simplesmente não trocando a senha e seguindo usando o sistema.
		"POST /users":           middlewares.RequireLogin(handler.requirePasswordChanged(handler.requireAdmin(handler.create))),
		"DELETE /users/{email}": middlewares.RequireLogin(handler.requirePasswordChanged(handler.requireAdmin(handler.delete))),
		"GET /users":            middlewares.RequireLogin(handler.requirePasswordChanged(handler.requireAdmin(handler.list))),
		"PATCH /users/{email}":  middlewares.RequireLogin(handler.requirePasswordChanged(handler.requireAdmin(handler.updateAccess))),
		// POST /users/{email}/require-password-change força a troca de senha
		// no próximo login de uma conta já existente — usado para migrar
		// contas criadas antes desta feature (sem must_change_password
		// persistido) ou após suspeita de senha comprometida.
		"POST /users/{email}/require-password-change": middlewares.RequireLogin(handler.requirePasswordChanged(handler.requireAdmin(handler.requirePasswordChangeForUser))),
		// GET /me não exige admin nem senha trocada — o dashboard usa essa
		// rota logo após o login para descobrir se deve exibir o prompt de
		// troca de senha obrigatória (must_change_password).
		"GET /me": middlewares.RequireLogin(handler.me),
		// PATCH /me/password é a própria via para satisfazer
		// requirePasswordChanged — nunca pode ficar atrás dele.
		"PATCH /me/password": middlewares.RequireLogin(handler.changePassword),
	}

	return handler
}

func (h *Handler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

// requireAdmin exige que o usuário autenticado (já resolvido por
// RequireLogin, email no contexto) tenha o papel admin.
func (h *Handler) requireAdmin(inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		emailValue, _ := ctxutils.GetValueCtx(ctx, middlewares.EMAIL_KEY).(string)

		isAdmin, err := h.service.IsAdmin(ctx, emailValue)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !isAdmin {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("admin role required"))
			return
		}

		inner(w, r)
	}
}

// requirePasswordChanged bloqueia o acesso do usuário autenticado enquanto
// MustChangePassword estiver true — força a troca de senha do primeiro
// acesso antes de liberar qualquer outra ação. Nunca envolve GET /me nem
// PATCH /me/password, senão o usuário ficaria sem como sair desse estado.
func (h *Handler) requirePasswordChanged(inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		emailValue, _ := ctxutils.GetValueCtx(ctx, middlewares.EMAIL_KEY).(string)

		if emailValue == middlewares.ServiceAccountEmail {
			inner(w, r)
			return
		}

		must, err := h.service.MustChangePassword(ctx, emailValue)
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

type ChangePasswordDto struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()
	emailValue, _ := ctxutils.GetValueCtx(ctx, middlewares.EMAIL_KEY).(string)

	var payload ChangePasswordDto
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.ChangePassword(ctx, emailValue, payload.CurrentPassword, payload.NewPassword); err != nil {
		switch err.(type) {
		case *errorutils.ValidationError:
			w.WriteHeader(http.StatusBadRequest)
		case *errorutils.NotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	if payload.Role == "" {
		payload.Role = RoleMember
	}
	if !IsValidRole(payload.Role) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("role must be \"admin\" or \"member\""))
		return
	}

	if err := h.service.Register(ctx, payload.Email, payload.Password, payload.Role, payload.Projects); err != nil {
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

func (h *Handler) updateAccess(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	emailInput := r.PathValue("email")
	if emailInput == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload UpdateAccessDto
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidRole(payload.Role) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("role must be \"admin\" or \"member\""))
		return
	}

	if payload.Role == RoleMember {
		users, err := h.service.List(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if isLastAdmin(users, emailInput) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("cannot demote the last remaining admin"))
			return
		}
	}

	if err := h.service.UpdateAccess(ctx, emailInput, payload.Role, payload.Projects); err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requirePasswordChangeForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	emailInput := r.PathValue("email")
	if emailInput == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.RequirePasswordChange(ctx, emailInput); err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	if isLastAdmin(users, emailInput) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("cannot remove the last remaining admin"))
		return
	}

	if err := h.service.Remove(ctx, emailInput); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isLastAdmin reports whether emailInput is an admin in users and no other
// admin exists — used to block both deletion and self-demotion of the last
// remaining admin, so the system never ends up with nobody able to manage
// users or grant project access (same spirit as the last-user-delete guard).
func isLastAdmin(users []Entity, emailInput string) bool {
	adminCount := 0
	targetIsAdmin := false
	for _, u := range users {
		if u.Role == RoleAdmin {
			adminCount++
		}
		if u.Email == emailInput && u.Role == RoleAdmin {
			targetIsAdmin = true
		}
	}
	return targetIsAdmin && adminCount <= 1
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

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	emailValue, _ := ctxutils.GetValueCtx(ctx, middlewares.EMAIL_KEY).(string)

	u, err := h.service.GetByEmail(ctx, emailValue)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
