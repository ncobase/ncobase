package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/validator"
	"ncobase/core/access/data/ent"
)

// handleEntError is a helper function to handle errors in a consistent manner.
func handleEntError(ctx context.Context, k string, err error) error {
	if ent.IsNotFound(err) {
		log.Errorf(ctx, "Error not found in %s: %v", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if ent.IsConstraintError(err) {
		log.Errorf(ctx, "Error constraint in %s: %v", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if ent.IsNotSingular(err) {
		log.Errorf(ctx, "Error not singular in %s: %v", k, err)
		return errors.New(ecode.NotSingular(k))
	}
	if validator.IsNotNil(err) {
		log.Errorf(ctx, "Error internal in %s: %v", k, err)
		return err
	}
	return err
}
