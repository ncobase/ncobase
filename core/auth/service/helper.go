package service

import (
	"context"
	"errors"
	"fmt"
	accessService "ncobase/core/access/service"
	accessStructs "ncobase/core/access/structs"
	"ncobase/core/auth/data/ent"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/messaging/email"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils"
	"github.com/ncobase/ncore/validation/validator"
)

// GetUserTenantsRolesPermissions retrieves user's tenants, roles, and permissions.
func GetUserTenantsRolesPermissions(
	ctx context.Context,
	as *accessService.Service,
	userID string,
) (tenantID string, roleSlugs []string, permissionCodes []string, isAdmin bool, err error) {
	tenantID = ctxutil.GetTenantID(ctx)

	var roles []*accessStructs.ReadRole

	if len(tenantID) > 0 {
		roleIDs, _ := as.UserTenantRole.GetUserRolesInTenant(ctx, userID, tenantID)
		roles, _ = as.Role.GetByIDs(ctx, roleIDs)
		for _, role := range roles {
			roleSlugs = append(roleSlugs, role.Slug)
		}
	} else {
		roles, _ = as.UserRole.GetUserRoles(ctx, userID)
		for _, role := range roles {
			roleSlugs = append(roleSlugs, role.Slug)
		}
	}

	for _, slug := range roleSlugs {
		if slug == "admin" || slug == "super-admin" {
			isAdmin = true
			break
		}
	}
	roleSlugs = utils.RemoveDuplicates(roleSlugs)

	if len(roles) > 0 {
		for _, role := range roles {
			rolePermissions, _ := as.RolePermission.GetRolePermissions(ctx, role.ID)
			for _, perm := range rolePermissions {
				permCode := fmt.Sprintf("%s:%s", perm.Action, perm.Subject)
				permissionCodes = append(permissionCodes, permCode)
			}
		}
	}
	permissionCodes = utils.RemoveDuplicates(permissionCodes)
	return
}

// generateUserToken generates user access and refresh tokens.
func generateUserToken(jtm *jwt.TokenManager, payload map[string]any, tokenID string) (string, string) {
	// Get user ID from payload
	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return "", ""
	}
	// Generate tokens
	// Use shorter expiry for access token (2 hours) and longer for refresh (7 days)
	accessToken, _ := jtm.GenerateAccessTokenWithExpiry(tokenID, payload, 2*time.Hour)
	refreshToken, _ := jtm.GenerateRefreshTokenWithExpiry(tokenID, types.JSON{
		"user_id": userID,
	}, 7*24*time.Hour)

	return accessToken, refreshToken
}

// RefreshUserToken refreshes user access and refresh tokens.
func RefreshUserToken(jtm *jwt.TokenManager, userID, tokenID, originalRefreshToken string, refreshTokenExp int64) (string, string) {
	now := time.Now().Unix()
	diff := refreshTokenExp - now

	refreshToken := originalRefreshToken
	accessPayload := types.JSON{
		"user_id": userID,
	}
	accessToken, _ := jtm.GenerateAccessToken(tokenID, accessPayload)
	if diff < 60*60*24*15 {
		refreshPayload := types.JSON{
			"user_id": userID,
		}

		refreshToken, _ = jtm.GenerateRefreshToken(tokenID, refreshPayload)
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

// sendAuthEmail sends an email with a code for authentication.
func sendAuthEmail(ctx context.Context, e, code string, registered bool) error {
	conf := ctxutil.GetConfig(ctx)
	template := email.AuthEmailTemplate{
		Subject:  "Email authentication",
		Template: "auth-email",
		Keyword:  "Sign in",
	}
	if registered {
		template.URL = conf.Frontend.SignInURL + "?code=" + code
	} else {
		template.Keyword = "Sign Up"
		template.URL = conf.Frontend.SignUpURL + "?code=" + code
	}
	_, err := ctxutil.SendEmailWithTemplate(ctx, e, template)
	return err
}

// sendRegisterMail sends an email with a code for register.
func sendRegisterMail(_ context.Context, jtm *jwt.TokenManager, codeAuth *ent.CodeAuth) (*types.JSON, error) {
	subject := "email-register"
	payload := types.JSON{"email": codeAuth.Email, "id": codeAuth.ID}
	registerToken, err := jtm.GenerateRegisterToken(codeAuth.ID, payload, subject)
	if err != nil {
		return nil, err
	}
	return &types.JSON{"email": codeAuth.Email, "register_token": registerToken}, nil
}

// generateTokensForUser generates tokens for a user.
func generateTokensForUser(ctx context.Context, jtm *jwt.TokenManager, client *ent.Client, payload map[string]any) (*types.JSON, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.New("user_id is not found")
	}

	authToken, err := createAuthToken(ctx, tx, userID)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}
	accessToken, refreshToken := generateUserToken(jtm, payload, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, errors.New("authorize is not created")
	}
	return &types.JSON{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, tx.Commit()
}
