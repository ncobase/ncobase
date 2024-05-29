package middleware

// // GenerateUserToken generates user access and refresh tokens.
// func GenerateUserToken(user *ent.User, authToken *ent.AuthToken) (string, string) {
// 	accessPayload := types.JSON{
// 		"user_id": user.ID,
// 	}
// 	refreshPayload := types.JSON{
// 		"user_id":  user.ID,
// 		"token_id": authToken.ID,
// 	}
//
// 	accessToken, _ := jwt.GenerateAccessToken(signingKey, accessPayload)
// 	refreshToken, _ := jwt.GenerateRefreshToken(signingKey, refreshPayload)
//
// 	return accessToken, refreshToken
// }
//
// // RefreshUserToken refreshes user access and refresh tokens.
// func RefreshUserToken(user *ent.User, tokenID string, originalRefreshToken string, refreshTokenExp int64) (string, string) {
// 	now := time.Now().Unix()
// 	diff := refreshTokenExp - now
//
// 	refreshToken := originalRefreshToken
// 	accessPayload := types.JSON{
// 		"user_id": user.ID,
// 	}
// 	accessToken, _ := jwt.GenerateAccessToken(signingKey, accessPayload)
// 	if diff < 60*60*24*15 {
// 		refreshPayload := types.JSON{
// 			"user_id":  user.ID,
// 			"token_id": tokenID,
// 		}
//
// 		refreshToken, _ = jwt.GenerateRefreshToken(signingKey, refreshPayload)
// 	}
//
// 	return accessToken, refreshToken
// }
