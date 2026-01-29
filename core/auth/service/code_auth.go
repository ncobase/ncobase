package service

import (
	"context"
	"errors"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/repository"
	"ncobase/core/auth/event"
	"ncobase/core/auth/structs"
	"ncobase/core/auth/wrapper"
	userStructs "ncobase/core/user/structs"
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
	d             *data.Data
	jtm           *jwt.TokenManager
	ep            event.PublisherInterface
	codeAuthRepo  repository.CodeAuthRepositoryInterface
	authTokenRepo repository.AuthTokenRepositoryInterface

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.SpaceServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// NewCodeAuthService creates a new service
func NewCodeAuthService(d *data.Data, jtm *jwt.TokenManager, ep event.PublisherInterface, usw *wrapper.UserServiceWrapper, tsw *wrapper.SpaceServiceWrapper, asw *wrapper.AccessServiceWrapper) CodeAuthServiceInterface {
	return &codeAuthService{
		d:             d,
		jtm:           jtm,
		ep:            ep,
		codeAuthRepo:  repository.NewCodeAuthRepository(d),
		authTokenRepo: repository.NewAuthTokenRepository(d),
		usw:           usw,
		tsw:           tsw,
		asw:           asw,
	}
}

// CodeAuth handles code authentication
func (s *codeAuthService) CodeAuth(ctx context.Context, code string) (*AuthResponse, error) {
	// Find and validate code via repository
	codeAuth, err := s.codeAuthRepo.FindByCode(ctx, code)
	if err = handleEntError(ctx, "Code", err); err != nil {
		return nil, err
	}

	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return nil, errors.New("code expired or already used")
	}

	// Check if user exists
	user, err := s.usw.FindUser(ctx, &userStructs.FindUser{Email: codeAuth.Email})
	if repository.IsNotFound(err) {
		// User doesn't exist, return register token
		registerResult, err := sendRegisterMail(ctx, s.jtm, codeAuth.Email, codeAuth.ID)
		if err != nil {
			return nil, err
		}

		registerToken, err := safeGetStringFromResult(*registerResult, "register_token")
		if err != nil {
			return nil, errors.New("failed to extract register token: " + err.Error())
		}

		return &AuthResponse{
			AccessToken: registerToken,
			TokenType:   "Register",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// User exists, proceed with login
	userSpaces, _ := s.tsw.GetUserSpaces(ctx, user.ID)
	var spaceIDs []string
	if len(userSpaces) > 0 {
		for _, t := range userSpaces {
			spaceIDs = append(spaceIDs, t.ID)
		}
	}

	defaultSpace, err := s.tsw.GetSpaceByUser(ctx, user.ID)
	if err == nil && defaultSpace != nil {
		ctx = ctxutil.SetSpaceID(ctx, defaultSpace.ID)
	}

	payload, err := CreateUserTokenPayload(ctx, user, spaceIDs, s.asw, s.tsw)
	if err != nil {
		return nil, err
	}

	authResp, err := generateAuthResponse(ctx, s.jtm, s.authTokenRepo, payload, nil, "email_code")
	if err != nil {
		return nil, err
	}

	authResp.SpaceIDs = spaceIDs
	if defaultSpace != nil {
		authResp.DefaultSpace = &types.JSON{
			"id":   defaultSpace.ID,
			"name": defaultSpace.Name,
		}
	}

	// Mark code as used via repository
	if err := s.codeAuthRepo.MarkAsUsed(ctx, codeAuth.ID); err != nil {
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
	// Check if user exists
	user, _ := s.usw.FindUser(ctx, &userStructs.FindUser{Email: body.Email, Phone: body.Phone})

	// Generate auth code
	authCode := nanoid.String(6)

	// Create code auth record via repository
	_, err := s.codeAuthRepo.Create(ctx, strings.ToLower(body.Email), authCode)
	if err != nil {
		return nil, err
	}

	// Send email with code
	if err := sendAuthEmail(ctx, body.Email, authCode, user != nil); err != nil {
		logger.Errorf(ctx, "Send email error: %v", err)
		return nil, errors.New("failed to send email, please try again or contact support")
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

// safeGetStringFromResult safely extracts a string value from a result map with type checking
func safeGetStringFromResult(data types.JSON, key string) (string, error) {
	val, exists := data[key]
	if !exists {
		return "", errors.New("missing required field: " + key)
	}

	str, ok := val.(string)
	if !ok {
		return "", errors.New("field " + key + " is not a string")
	}

	if str == "" {
		return "", errors.New("field " + key + " cannot be empty")
	}

	return str, nil
}
