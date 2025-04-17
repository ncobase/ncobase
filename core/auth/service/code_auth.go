package service

import (
	"context"
	"errors"
	accessService "ncobase/core/access/service"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/ent"
	codeAuthEnt "ncobase/core/auth/data/ent/codeauth"
	"ncobase/core/auth/structs"
	userService "ncobase/core/user/service"
	userStructs "ncobase/core/user/structs"
	"strings"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
)

// CodeAuthServiceInterface is the interface for the service.
type CodeAuthServiceInterface interface {
	SendCode(ctx context.Context, body *structs.SendCodeBody) (*types.JSON, error)
	CodeAuth(ctx context.Context, code string) (*types.JSON, error)
}

// codeAuth is the struct for the service.
type codeAuthService struct {
	d   *data.Data
	jtm *jwt.TokenManager
	as  *accessService.Service
	us  *userService.Service
}

// NewCodeAuthService creates a new service.
func NewCodeAuthService(d *data.Data, jtm *jwt.TokenManager, as *accessService.Service, us *userService.Service) CodeAuthServiceInterface {
	return &codeAuthService{
		d:   d,
		jtm: jtm,
		as:  as,
		us:  us,
	}
}

// CodeAuth code auth service
func (s *codeAuthService) CodeAuth(ctx context.Context, code string) (*types.JSON, error) {
	client := s.d.GetEntClient()

	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if err = handleEntError(ctx, "Code", err); err != nil {
		return nil, err
	}
	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return nil, errors.New("code expired")
	}

	user, err := s.us.User.FindUser(ctx, &userStructs.FindUser{Email: codeAuth.Email})
	if ent.IsNotFound(err) {
		return sendRegisterMail(ctx, s.jtm, codeAuth)
	}

	tenantID, roleSlugs, permissionCodes, isAdmin, _ := GetUserTenantsRolesPermissions(ctx, s.as, user.ID)

	payload := types.JSON{
		"user_id":     user.ID,
		"roles":       roleSlugs,
		"permissions": permissionCodes,
		"is_admin":    isAdmin,
		"tenant_id":   tenantID,
	}

	return generateTokensForUser(ctx, s.jtm, client, payload)
}

// Helper functions for codeAuthService
func isCodeExpired(createdAt int64) bool {
	createdTime := time.UnixMilli(createdAt)
	expirationTime := createdTime.Add(24 * time.Hour)
	return time.Now().After(expirationTime)
}

// SendCode send code service
func (s *codeAuthService) SendCode(ctx context.Context, body *structs.SendCodeBody) (*types.JSON, error) {
	client := s.d.GetEntClient()

	user, _ := s.us.User.FindUser(ctx, &userStructs.FindUser{Email: body.Email, Phone: body.Phone})
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	authCode := nanoid.String(6)
	_, err = tx.CodeAuth.Create().SetEmail(strings.ToLower(body.Email)).SetCode(authCode).Save(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := sendAuthEmail(ctx, body.Email, authCode, user != nil); err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		logger.Errorf(ctx, "send mail error: %v", err)
		return nil, errors.New("send mail failed, please try again or contact support")
	}

	return &types.JSON{"registered": user != nil}, tx.Commit()
}
