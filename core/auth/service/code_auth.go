package service

import (
	"context"
	"errors"
	"ncobase/auth/data"
	"ncobase/auth/data/ent"
	codeAuthEnt "ncobase/auth/data/ent/codeauth"
	"ncobase/auth/event"
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

// CodeAuthServiceInterface is the interface for the service
type CodeAuthServiceInterface interface {
	SendCode(ctx context.Context, body *structs.SendCodeBody) (*types.JSON, error)
	CodeAuth(ctx context.Context, code string) (*AuthResponse, error)
}

// codeAuthService is the struct for the service
type codeAuthService struct {
	d   *data.Data
	jtm *jwt.TokenManager
	ep  event.PublisherInterface

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.TenantServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// NewCodeAuthService creates a new service
func NewCodeAuthService(d *data.Data, jtm *jwt.TokenManager, ep event.PublisherInterface, usw *wrapper.UserServiceWrapper, tsw *wrapper.TenantServiceWrapper, asw *wrapper.AccessServiceWrapper) CodeAuthServiceInterface {
	return &codeAuthService{
		d:   d,
		jtm: jtm,
		ep:  ep,
		usw: usw,
		tsw: tsw,
		asw: asw,
	}
}

// CodeAuth handles code authentication
func (s *codeAuthService) CodeAuth(ctx context.Context, code string) (*AuthResponse, error) {
	client := s.d.GetMasterEntClient()

	// Find and validate code
	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if err = handleEntError(ctx, "Code", err); err != nil {
		return nil, err
	}

	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return nil, errors.New("code expired or already used")
	}

	// Check if user exists
	user, err := s.usw.FindUser(ctx, &userStructs.FindUser{Email: codeAuth.Email})
	if ent.IsNotFound(err) {
		// User doesn't exist, return register token
		registerResult, err := sendRegisterMail(ctx, s.jtm, codeAuth)
		if err != nil {
			return nil, err
		}

		return &AuthResponse{
			AccessToken: (*registerResult)["register_token"].(string),
			TokenType:   "Register",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// User exists, proceed with login
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
	payload, err := CreateUserTokenPayload(ctx, user, tenantIDs, s.asw, s.tsw)
	if err != nil {
		return nil, err
	}

	// Generate authentication response
	authResp, err := generateAuthResponse(ctx, s.jtm, client, payload, nil, "email_code")
	if err != nil {
		return nil, err
	}

	// Set additional response data
	authResp.TenantIDs = tenantIDs
	if defaultTenant != nil {
		authResp.DefaultTenant = &types.JSON{
			"id":   defaultTenant.ID,
			"name": defaultTenant.Name,
		}
	}

	// Mark code as used
	if err := s.markCodeAsUsed(ctx, codeAuth.ID); err != nil {
		logger.Warnf(ctx, "Failed to mark code as used: %v", err)
	}

	// Publish login event
	if s.ep != nil {
		ip, userAgent, sessionID := ctxutil.GetClientInfo(ctx)
		uaInfo := ctxutil.GetParsedUserAgent(ctx)

		metadata := &types.JSON{
			"ip_address":   ip,
			"user_agent":   userAgent,
			"session_id":   sessionID,
			"login_method": "email_code",
			"browser":      uaInfo.Browser,
			"os":           uaInfo.OS,
			"mobile":       uaInfo.Mobile,
			"referer":      ctxutil.GetReferer(ctx),
			"timestamp":    time.Now().UnixMilli(),
		}

		s.ep.PublishUserLogin(ctx, user.ID, metadata)
	}

	return authResp, nil
}

// SendCode sends verification code
func (s *codeAuthService) SendCode(ctx context.Context, body *structs.SendCodeBody) (*types.JSON, error) {
	client := s.d.GetMasterEntClient()

	// Check if user exists
	user, _ := s.usw.FindUser(ctx, &userStructs.FindUser{Email: body.Email, Phone: body.Phone})

	// Create transaction for code creation
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	// Generate auth code
	authCode := nanoid.String(6)
	_, err = tx.CodeAuth.Create().SetEmail(strings.ToLower(body.Email)).SetCode(authCode).Save(ctx)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	// Send email with code
	if err := sendAuthEmail(ctx, body.Email, authCode, user != nil); err != nil {
		if err = tx.Rollback(); err != nil {
			return nil, err
		}
		logger.Errorf(ctx, "Send email error: %v", err)
		return nil, errors.New("failed to send email, please try again or contact support")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Publish auth code sent event
	if s.ep != nil {
		ip, userAgent, _ := ctxutil.GetClientInfo(ctx)
		metadata := &types.JSON{
			"ip_address": ip,
			"user_agent": userAgent,
			"email":      body.Email,
			"phone":      body.Phone,
			"timestamp":  time.Now().UnixMilli(),
		}
		s.ep.PublishAuthCodeSent(ctx, body.Email, metadata)
	}

	return &types.JSON{"registered": user != nil}, nil
}

// isCodeExpired checks if the code has expired
func isCodeExpired(createdAt int64) bool {
	createdTime := time.UnixMilli(createdAt)
	expirationTime := createdTime.Add(24 * time.Hour)
	return time.Now().After(expirationTime)
}

// markCodeAsUsed marks the code as used
func (s *codeAuthService) markCodeAsUsed(ctx context.Context, codeID string) error {
	client := s.d.GetMasterEntClient()
	_, err := client.CodeAuth.UpdateOneID(codeID).SetLogged(true).Save(ctx)
	return err
}
