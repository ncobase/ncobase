package service

import (
	"bytes"
	"context"
	"errors"
	"ncobase/resource/data/ent"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// readCloser wrapper for bytes.Reader
type readCloser struct {
	*bytes.Reader
}

func (r *readCloser) Close() error {
	return nil
}

// handleEntError handles ent errors consistently
func handleEntError(ctx context.Context, k string, err error) error {
	if ent.IsNotFound(err) {
		logger.Errorf(ctx, "Error not found in %s: %v", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if ent.IsConstraintError(err) {
		logger.Errorf(ctx, "Error constraint in %s: %v", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if ent.IsNotSingular(err) {
		logger.Errorf(ctx, "Error not singular in %s: %v", k, err)
		return errors.New(ecode.NotSingular(k))
	}
	if validator.IsNotNil(err) {
		logger.Errorf(ctx, "Error internal in %s: %v", k, err)
		return err
	}
	return err
}
