package user

import (
	"unicode"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

const minPasswordLength = 8

// ValidatePassword enforces the minimum security criteria for a password set
// via ChangePassword: at least 8 characters, one uppercase letter, one
// lowercase letter, one digit and one special character. It intentionally
// does not apply to Register/Seed — those set a temporary password chosen by
// an admin (or the ADMIN_USERS bootstrap), and the user is forced to replace
// it with a policy-compliant one on first login (Entity.MustChangePassword).
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return errorutils.NewValidationError("password must have at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errorutils.NewValidationError("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errorutils.NewValidationError("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errorutils.NewValidationError("password must contain at least one digit")
	}
	if !hasSpecial {
		return errorutils.NewValidationError("password must contain at least one special character")
	}

	return nil
}
