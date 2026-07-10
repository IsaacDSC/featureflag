package featureflag

import (
	"errors"
	"regexp"
)

// projectPattern define o formato aceito para um identificador de project:
// letras minúsculas/dígitos, opcionalmente separados por hífen, de 1 a 64 chars.
var projectPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// ErrInvalidProject é retornado quando o project está vazio ou fora do formato aceito.
var ErrInvalidProject = errors.New("project is required and must match ^[a-z0-9][a-z0-9-]{0,63}$")

// ValidateProject garante que o escopo (project) é não-vazio e possui formato válido.
func ValidateProject(project string) error {
	if !projectPattern.MatchString(project) {
		return ErrInvalidProject
	}

	return nil
}
