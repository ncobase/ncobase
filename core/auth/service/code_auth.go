package service

import (
	"context"
	"errors"
	"ncobase/auth/data"
	"ncobase/auth/data/ent"
	codeAuthEnt "ncobase/auth/data/ent/codeauth"
	"ncobase/auth/structs"
	"ncobase/auth/wrapper"
	userStructs "ncobase/user/structs"
	"strings"
	"time"

	"github.com/ncobase/ncore/ctxutil"
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

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.TenantServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// NewCodeAuthService creates a new service.
func NewCodeAuthService(d *data.Data, jtm *jwt.TokenManager, usw *wrapper.UserServiceWrapper, tsw *wrapper.TenantServiceWrapper, asw *wrapper.AccessServiceWrapper) CodeAuthServiceInterface {
	return &codeAuthService{
		d:   d,
		jtm: jtm,
		usw: usw,
		tsw: tsw,
		asw: asw,
	}
}

// CodeAuth code auth service
func (s *codeAuthService) CodeAuth(ctx context.Context, code string) (*types.JSON, error) {
	client := s.d.GetMasterEntClient()

	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if err = handleEntError(ctx, "Code", err); err != nil {
		return nil, err
	}
	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return nil, errors.New("code expired")
	}

	user, err := s.usw.FindUser(ctx, &userStructs.FindUser{Email: codeAuth.Email})
	if ent.IsNotFound(err) {
		return sendRegisterMail(ctx, s.jtm, codeAuth)
	}

	// Get all tenants the user belongs to
	userTenants, _ := s.tsw.GetUserTenants(ctx, user.ID)
	var tenantIDs []string
	if len(userTenants) > 0 {
		for _, t := range userTenants {
			tenantIDs = append(tenantIDs, t.ID)
		}
	}

	// Set tenant ID in context if there's a default tenant
	defaultTenant, err := s.tsw.GetTenantByUser(ctx, user.ID)
	if err == nil && defaultTenant != nil {
		ctx = ctxutil.SetTenantID(ctx, defaultTenant.ID)
	}

	// Create token payload
	payload, err := CreateUserTokenPayload(ctx, s.asw, user, tenantIDs)
	if err != nil {
		return nil, err
	}

	tokens, err := generateTokensForUser(ctx, s.jtm, client, payload)
	if err != nil {
		return nil, err
	}

	// Include tenant information in the response
	(*tokens)["tenant_ids"] = tenantIDs
	(*tokens)["default_tenant"] = defaultTenant

	return tokens, nil
}

// isCodeExpired checks if the code has expired
func isCodeExpired(createdAt int64) bool {
	createdTime := time.UnixMilli(createdAt)
	expirationTime := createdTime.Add(24 * time.Hour)
	return time.Now().After(expirationTime)
}

// SendCode send code service
func (s *codeAuthService) SendCode(ctx context.Context, body *structs.SendCodeBody) (*types.JSON, error) {
	client := s.d.GetMasterEntClient()

	user, _ := s.usw.FindUser(ctx, &userStructs.FindUser{Email: body.Email, Phone: body.Phone})
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
