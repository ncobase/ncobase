package errorsx

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/ncobase/ncore/logging/logger"
)

// HandleEntError handles Ent ORM errors with consistent logging and messaging
func HandleEntError(ctx context.Context, entity string, err error) error {
	if err == nil {
		return nil
	}

	// Check for common Ent errors
	if sql.IsNotFound(err) {
		return fmt.Errorf("%s not found", entity)
	}

	if sql.IsConstraintError(err) {
		return fmt.Errorf("%s constraint violation: %w", entity, err)
	}

	// Log the error for debugging
	logger.Errorf(ctx, "Database error for %s: %v", entity, err)

	// Return a generic error message
	return fmt.Errorf("%s operation failed", entity)
}
