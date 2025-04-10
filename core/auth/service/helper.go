package service

import (
	"context"
	"errors"
	"ncobase/core/auth/data/ent"
	"time"

	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/jwt"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/types"
	"github.com/ncobase/ncore/pkg/validator"
)

// generateUserToken generates user access and refresh tokens.
func generateUserToken(signingKey, userID, tokenID string) (string, string) {
	accessPayload := types.JSON{
		"user_id": userID,
	}
	refreshPayload := types.JSON{
		"user_id": userID,
	}

	accessToken, _ := jwt.GenerateAccessToken(signingKey, tokenID, accessPayload)
	refreshToken, _ := jwt.GenerateRefreshToken(signingKey, tokenID, refreshPayload)

	return accessToken, refreshToken
}

// RefreshUserToken refreshes user access and refresh tokens.
func RefreshUserToken(signingKey, userID, tokenID, originalRefreshToken string, refreshTokenExp int64) (string, string) {
	now := time.Now().Unix()
	diff := refreshTokenExp - now

	refreshToken := originalRefreshToken
	accessPayload := types.JSON{
		"user_id": userID,
	}
	accessToken, _ := jwt.GenerateAccessToken(signingKey, tokenID, accessPayload)
	if diff < 60*60*24*15 {
		refreshPayload := types.JSON{
			"user_id": userID,
		}

		refreshToken, _ = jwt.GenerateRefreshToken(signingKey, tokenID, refreshPayload)
	}

	return accessToken, refreshToken
}

// handleEntError is a helper function to handle errors in a consistent manner.
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
