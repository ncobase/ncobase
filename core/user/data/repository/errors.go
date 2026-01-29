package repository

import "ncobase/core/user/data/ent"

// IsNotFound reports whether the error means the row does not exist.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}

// IsConstraintError reports whether the error is a constraint violation.
func IsConstraintError(err error) bool {
	return ent.IsConstraintError(err)
}

// IsNotSingular reports whether the error is a non-singular result.
func IsNotSingular(err error) bool {
	return ent.IsNotSingular(err)
}
