package repository

import "ncobase/core/organization/data/ent"

// IsNotFound reports whether err indicates a missing entity.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}

// IsConstraintError reports whether err indicates a constraint violation.
func IsConstraintError(err error) bool {
	return ent.IsConstraintError(err)
}

// IsNotSingular reports whether err indicates a non-singular result.
func IsNotSingular(err error) bool {
	return ent.IsNotSingular(err)
}
