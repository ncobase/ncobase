package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/common/validator"
	"ncobase/feature/resource/data/ent"
)

// handleEntError is a helper function to handle errors in a consistent manner.
func handleEntError(k string, err error) (*resp.Exception, error) {
	if ent.IsNotFound(err) {
		log.Errorf(context.Background(), "Error not found in %s: %v\n", k, err)
		return resp.NotFound(ecode.NotExist(k)), nil
	}
	if ent.IsConstraintError(err) {
		log.Errorf(context.Background(), "Error constraint in %s: %v\n", k, err)
		return resp.Conflict(ecode.AlreadyExist(k)), nil
	}
	if ent.IsNotSingular(err) {
		log.Errorf(context.Background(), "Error not singular in %s: %v\n", k, err)
		return resp.BadRequest(ecode.NotSingular(k)), nil
	}
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "Error internal in %s: %v\n", k, err)
		return resp.InternalServer(err.Error()), nil
	}
	return nil, err
}
