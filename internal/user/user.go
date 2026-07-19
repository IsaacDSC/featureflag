package user

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// IsValidRole reports whether role is a known value. Anything else is
// rejected at the API boundary (POST/PATCH /users).
func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleMember
}

type Entity struct {
	// ID vira o próprio _id do documento no MongoDB (bson:"_id") — evita o
	// ObjectId redundante que o driver geraria por conta própria, já que já
	// temos um identificador único gerado na aplicação (uuid.New()).
	ID           uuid.UUID `json:"id" bson:"_id"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-" bson:"password_hash"`
	Role         string    `json:"role" bson:"role"`
	// Projects só é relevante para RoleMember — um admin tem acesso a todos
	// os projects independente do que estiver aqui.
	Projects  []string  `json:"projects" bson:"projects"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	// MustChangePassword força a troca de senha no próximo login — setado
	// para true em Register (admin cria a conta com uma senha temporária) e
	// em Seed (bootstrap do primeiro admin). Zero value (false) preserva o
	// comportamento de usuários já existentes antes deste campo existir.
	MustChangePassword bool `json:"must_change_password" bson:"must_change_password"`
}

func (e Entity) HasProjectAccess(project string) bool {
	if e.Role == RoleAdmin {
		return true
	}
	return slices.Contains(e.Projects, project)
}
