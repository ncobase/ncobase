package middleware

import (
	"time"

	"ncobase/common/jwt"
	"ncobase/common/types"
)

// GenerateUserToken generates user access and refresh tokens.
func GenerateUserToken(signingKey, userID, tokenID string) (string, string) {
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
