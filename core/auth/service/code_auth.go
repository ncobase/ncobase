package service

import (
	"context"
	"errors"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/ent"
	codeAuthEnt "ncobase/core/auth/data/ent/codeauth"
	"ncobase/core/auth/structs"
	userService "ncobase/core/user/service"
	userStructs "ncobase/core/user/structs"
	"ncore/pkg/config"
	"ncore/pkg/email"
	"ncore/pkg/helper"
	"ncore/pkg/jwt"
	"ncore/pkg/logger"
	"ncore/pkg/nanoid"
	"ncore/pkg/types"
	"strings"
	"time"
)

// CodeAuthServiceInterface is the interface for the service.
type CodeAuthServiceInterface interface {
	SendCode(ctx context.Context, body *structs.SendCodeBody) (*types.JSON, error)
	CodeAuth(ctx context.Context, code string) (*types.JSON, error)
}

// codeAuth is the struct for the service.
type codeAuthService struct {
	d  *data.Data
	us *userService.Service
}

// NewCodeAuthService creates a new service.
func NewCodeAuthService(d *data.Data, us *userService.Service) CodeAuthServiceInterface {
	return &codeAuthService{
		d:  d,
		us: us,
	}
}

// CodeAuth code auth service
func (s *codeAuthService) CodeAuth(ctx context.Context, code string) (*types.JSON, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if err := handleEntError(ctx, "Code", err); err != nil {
		return nil, err
	}
	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return nil, errors.New("code expired")
	}

	user, err := s.us.User.FindUser(ctx, &userStructs.FindUser{Email: codeAuth.Email})
	if ent.IsNotFound(err) {
		return sendRegisterMail(ctx, conf, codeAuth)
	}

	return generateTokensForUser(ctx, conf, client, user)
}

// Helper functions for codeAuthService
func isCodeExpired(createdAt int64) bool {
	createdTime := time.UnixMilli(createdAt)
	expirationTime := createdTime.Add(24 * time.Hour)
	return time.Now().After(expirationTime)
}

func sendRegisterMail(_ context.Context, conf *config.Config, codeAuth *ent.CodeAuth) (*types.JSON, error) {
	subject := "email-register"
	payload := types.JSON{"email": codeAuth.Email, "id": codeAuth.ID}
	registerToken, err := jwt.GenerateRegisterToken(conf.Auth.JWT.Secret, codeAuth.ID, payload, subject)
	if err != nil {
		return nil, err
	}
	return &types.JSON{"email": codeAuth.Email, "register_token": registerToken}, nil
}

func generateTokensForUser(ctx context.Context, conf *config.Config, client *ent.Client, user *userStructs.ReadUser) (*types.JSON, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	authToken, err := createAuthToken(ctx, tx, user.ID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}
	accessToken, refreshToken := generateUserToken(conf.Auth.JWT.Secret, user.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, errors.New("authorize is not created")
	}
	return &types.JSON{
		"id":           user.ID,
		"access_token": accessToken,
	}, tx.Commit()
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
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := sendAuthEmail(ctx, body.Email, authCode, user != nil); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		logger.Errorf(ctx, "send mail error: %v", err)
		return nil, errors.New("send mail failed, please try again or contact support")
	}

	return &types.JSON{"registered": user != nil}, tx.Commit()
}

// Helper functions for SendCode
func sendAuthEmail(ctx context.Context, e, code string, registered bool) error {
	conf := helper.GetConfig(ctx)
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
	_, err := helper.SendEmailWithTemplate(ctx, e, template)
	return err
}
