package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"ncobase/resource/data/ent"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
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

// calculateFileHash calculates SHA256 hash of file content
func calculateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("sha256:%x", hash)
}

// findFileByHash finds existing file by hash for deduplication
func findFileByHash(ctx context.Context, ownerID, hash string) (*ent.File, error) {
	// This would require a repository method to find by hash
	// For now, return nil (no deduplication)
	return nil, fmt.Errorf("not implemented")
}

// getExtrasFromFile returns the extras from the file
func getExtrasFromFile(file *ent.File) types.JSON {
	if file == nil || file.Extras == nil {
		return make(types.JSON)
	}

	extras := make(types.JSON)
	for k, v := range file.Extras {
		extras[k] = v
	}
	return extras
}
