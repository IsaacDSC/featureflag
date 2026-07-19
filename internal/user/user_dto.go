package user

type CreateUserDto struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     string   `json:"role"`
	Projects []string `json:"projects"`
}

type UpdateAccessDto struct {
	Role     string   `json:"role"`
	Projects []string `json:"projects"`
}
