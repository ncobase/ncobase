package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"ncobase/plugin/resource/data/repository"

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
	if repository.IsNotFound(err) {
		logger.Errorf(ctx, "Error not found in %s: %v", k, err)
		return errors.New(ecode.NotExist(k))
	}
	if repository.IsConstraintError(err) {
		logger.Errorf(ctx, "Error constraint in %s: %v", k, err)
		return errors.New(ecode.AlreadyExist(k))
	}
	if repository.IsNotSingular(err) {
		logger.Errorf(ctx, "Error not singular in %s: %v", k, err)
		return errors.New(ecode.NotSingular(k))
	}
	if validator.IsNotNil(err) {
		logger.Errorf(ctx, "Error internal in %s: %v", k, err)
		return err
	}
	return err
}

// calculateFileHash calculates SHA256 hash of file content
func calculateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("sha256:%x", hash)
}
