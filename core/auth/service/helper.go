package service

import (
	"context"
	"errors"
	"fmt"
	accessService "ncobase/access/service"
	accessStructs "ncobase/access/structs"
	"ncobase/auth/data/ent"
	userStructs "ncobase/user/structs"
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

// GetUserTenantsRolesPermissions retrieves user's roles and permissions with improved logic
func GetUserTenantsRolesPermissions(
	ctx context.Context,
	as *accessService.Service,
	userID string,
) (tenantID string, roleSlugs []string, permissionCodes []string, isAdmin bool, err error) {
	tenantID = ctxutil.GetTenantID(ctx)
	logger.Debugf(ctx, "Getting permissions for user %s, tenant %s", userID, tenantID)

	// Get global roles first (primary roles)
	globalRoles, err := as.UserRole.GetUserRoles(ctx, userID)
	if err != nil {
		logger.Warnf(ctx, "Failed to get global roles for user %s: %v", userID, err)
	} else {
		for _, role := range globalRoles {
			roleSlugs = append(roleSlugs, role.Slug)
		}
		logger.Debugf(ctx, "Found %d global roles for user", len(globalRoles))
	}

	// Get tenant-specific roles if tenant context exists
	if tenantID != "" {
		roleIDs, roleErr := as.UserTenantRole.GetUserRolesInTenant(ctx, userID, tenantID)
		if roleErr == nil && len(roleIDs) > 0 {
			tenantRoles, _ := as.Role.GetByIDs(ctx, roleIDs)
			for _, role := range tenantRoles {
				// Avoid duplicates
				if !utils.Contains(roleSlugs, role.Slug) {
					roleSlugs = append(roleSlugs, role.Slug)
				}
			}
			logger.Debugf(ctx, "Found %d tenant roles for user", len(tenantRoles))
		}
	}

	// Check admin status
	isAdmin = isAdminRole(roleSlugs)
	if isAdmin {
		logger.Debugf(ctx, "User has admin privileges")
	}

	// Get permissions for all roles
	permissionCodes, err = getPermissionsForRoles(ctx, as, globalRoles, isAdmin)
	if err != nil {
		logger.Warnf(ctx, "Failed to get permissions: %v", err)
	}

	// Remove duplicates
	roleSlugs = utils.RemoveDuplicates(roleSlugs)
	permissionCodes = utils.RemoveDuplicates(permissionCodes)

	logger.Infof(ctx, "User %s has %d roles, %d permissions, isAdmin: %v",
		userID, len(roleSlugs), len(permissionCodes), isAdmin)
	return
}

// isAdminRole checks if any role has admin privileges
func isAdminRole(roleSlugs []string) bool {
	adminRoles := []string{"super-admin", "system-admin", "enterprise-admin", "tenant-admin"}
	for _, roleSlug := range roleSlugs {
		for _, adminRole := range adminRoles {
			if roleSlug == adminRole {
				return true
			}
		}
	}
	return false
}

// getPermissionsForRoles gets all permissions for the given roles
func getPermissionsForRoles(ctx context.Context, as *accessService.Service, roles []*accessStructs.ReadRole, isAdmin bool) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}

	var permissionCodes []string
	// var hasSuperAdminAccess bool

	for _, role := range roles {
		rolePermissions, err := as.RolePermission.GetRolePermissions(ctx, role.ID)
		if err != nil {
			logger.Warnf(ctx, "Failed to get permissions for role %s: %v", role.Slug, err)
			continue
		}

		for _, perm := range rolePermissions {
			// Check for super admin wildcard permission
			if perm.Action == "*" && perm.Subject == "*" {
				// hasSuperAdminAccess = true
				permissionCodes = []string{"*:*"}
				logger.Infof(ctx, "User has super admin wildcard permission")
				return permissionCodes, nil
			}

			// Add regular permission
			permCode := fmt.Sprintf("%s:%s", perm.Action, perm.Subject)
			permissionCodes = append(permissionCodes, permCode)
		}

		logger.Debugf(ctx, "Role %s has %d permissions", role.Slug, len(rolePermissions))
	}

	return permissionCodes, nil
}

// CreateUserTokenPayload creates token payload with user permissions
func CreateUserTokenPayload(
	ctx context.Context,
	as *accessService.Service,
	user *userStructs.ReadUser,
	tenantIDs []string,
) (types.JSON, error) {
	if user.ID == "" {
		return nil, errors.New("userID is required")
	}

	tenantID, roleSlugs, permissionCodes, isAdmin, err := GetUserTenantsRolesPermissions(ctx, as, user.ID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get user roles and permissions: %v", err)
		// Continue with empty roles and permissions rather than failing
		roleSlugs = []string{}
		permissionCodes = []string{}
		isAdmin = false
	}

	return types.JSON{
		"user_id":      user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"is_admin":     isAdmin,
		"tenant_id":    tenantID,
		"tenant_ids":   tenantIDs,
		"roles":        roleSlugs,
		"permissions":  permissionCodes,
		"user_status":  user.Status,
		"is_certified": user.IsCertified,
	}, nil
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

// generateTokensForUser generates access and refresh tokens
func generateTokensForUser(ctx context.Context, jtm *jwt.TokenManager, client *ent.Client, payload map[string]any) (*types.JSON, error) {
	userID, ok := payload["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.New("user_id not found in payload")
	}

	// Create auth token record in transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Errorf(ctx, "Failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	authToken, err := createAuthToken(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken := generateUserToken(jtm, payload, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		return nil, errors.New("failed to generate tokens")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &types.JSON{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

// createAuthToken creates a new auth token for a user.
func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
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
	template := email.EmailTemplate{
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
