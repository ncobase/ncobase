package service

import (
	"context"
	"errors"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/validator"
	"ncobase/plugin/counter/data/ent"
)

// handleEntError is a helper function to handle errors in a consistent manner.
func handleEntError(ctx context.Context, k string, err error) error {
	if ent.IsNotFound(err) {
		logger.Errorf(context.Background(), "Error not found in %s: %v", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if ent.IsConstraintError(err) {
		logger.Errorf(context.Background(), "Error constraint in %s: %v", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if ent.IsNotSingular(err) {
		logger.Errorf(context.Background(), "Error not singular in %s: %v", k, err)
		return errors.New(ecode.NotSingular(k))
	}
	if validator.IsNotNil(err) {
		logger.Errorf(context.Background(), "Error internal in %s: %v", k, err)
		return err
	}
	return err
}
