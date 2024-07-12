package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/validator"
	"ncobase/feature/system/data/ent"
)

// handleEntError is a helper function to handle errors in a consistent manner.
func handleEntError(k string, err error) error {
	if ent.IsNotFound(err) {
		log.Errorf(context.Background(), "Error not found in %s: %v\n", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if ent.IsConstraintError(err) {
		log.Errorf(context.Background(), "Error constraint in %s: %v\n", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if ent.IsNotSingular(err) {
		log.Errorf(context.Background(), "Error not singular in %s: %v\n", k, err)
		return errors.New(ecode.NotSingular(k))
	}
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "Error internal in %s: %v\n", k, err)
		return err
	}
	return err
}
