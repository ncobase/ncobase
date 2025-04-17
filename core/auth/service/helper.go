package service

import (
	"context"
	"errors"
	"ncobase/core/auth/data/ent"
	"time"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// generateUserToken generates user access and refresh tokens.
func generateUserToken(signingKey, userID, tokenID string) (string, string) {
	accessPayload := types.JSON{
		"user_id": userID,
		// We can add more data here like roles and permissions
		// These would need to be passed as parameters to this function
	}
	refreshPayload := types.JSON{
		"user_id": userID,
	}

	// Use shorter expiry for access token (2 hours) and longer for refresh (7 days)
	accessToken, _ := jwt.GenerateAccessTokenWithExpiry(signingKey, tokenID, accessPayload, 2*time.Hour)
	refreshToken, _ := jwt.GenerateRefreshTokenWithExpiry(signingKey, tokenID, refreshPayload, 7*24*time.Hour)

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
