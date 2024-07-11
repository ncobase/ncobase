package service

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/email"
	"ncobase/common/jwt"
	"ncobase/common/log"
	"ncobase/common/nanoid"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/feature/auth/data"
	"ncobase/feature/auth/data/ent"
	codeAuthEnt "ncobase/feature/auth/data/ent/codeauth"
	"ncobase/feature/auth/structs"
	userService "ncobase/feature/user/service"
	userStructs "ncobase/feature/user/structs"
	"ncobase/helper"
	"ncobase/middleware"
	"strings"
	"time"
)

// CodeAuthServiceInterface is the interface for the service.
type CodeAuthServiceInterface interface {
	SendCodeService(ctx context.Context, body *structs.SendCodeBody) (*resp.Exception, error)
	CodeAuthService(ctx context.Context, code string) (*resp.Exception, error)
}

// codeAuthService is the struct for the service.
type codeAuthService struct {
	d   *data.Data
	usi userService.UserServiceInterface
}

// NewCodeAuthService creates a new service.
func NewCodeAuthService(d *data.Data, usi userService.UserServiceInterface) CodeAuthServiceInterface {
	return &codeAuthService{
		d:   d,
		usi: usi,
	}
}

// CodeAuthService code auth service
func (s *codeAuthService) CodeAuthService(ctx context.Context, code string) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if exception, err := handleEntError("Code", err); exception != nil {
		return exception, err
	}
	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return resp.Forbidden("EXPIRED_CODE"), nil
	}

	rst, err := s.usi.FindUser(ctx, &userStructs.FindUser{Email: codeAuth.Email})
	if ent.IsNotFound(err) {
		return sendRegisterMail(ctx, conf, codeAuth)
	}

	return generateTokensForUser(ctx, client, rst.User, conf.Domain)
}

// Helper functions for codeAuthService
func isCodeExpired(createdAt time.Time) bool {
	return time.Now().After(createdAt.Add(24 * time.Hour))
}

func sendRegisterMail(ctx context.Context, conf *config.Config, codeAuth *ent.CodeAuth) (*resp.Exception, error) {
	subject := "email-register"
	payload := types.JSON{"email": codeAuth.Email, "id": codeAuth.ID}
	registerToken, err := jwt.GenerateRegisterToken(conf.Auth.JWT.Secret, codeAuth.ID, payload, subject)
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}
	// cookie.SetRegister(c.Writer, registerToken, conf.Domain) // TODO: move to handler
	return &resp.Exception{Data: types.JSON{"email": codeAuth.Email, "register_token": registerToken}}, nil
}

func generateTokensForUser(ctx context.Context, client *ent.Client, user *userStructs.UserBody, tenant string) (*resp.Exception, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return resp.Transactions(err.Error()), nil
	}
	authToken, err := createAuthToken(ctx, tx, user.ID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.Transactions(err.Error()), nil
	}
	accessToken, refreshToken := middleware.GenerateUserToken(user.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.InternalServer("Authorize is not created"), nil
	}
	// cookie.Set(c.Writer, accessToken, refreshToken, tenant) // TODO: move to handler
	return &resp.Exception{
		Data: types.JSON{
			"id":           user.ID,
			"access_token": accessToken,
		},
	}, tx.Commit()
}

// SendCodeService send code service
func (s *codeAuthService) SendCodeService(ctx context.Context, body *structs.SendCodeBody) (*resp.Exception, error) {
	client := s.d.GetEntClient()

	user, _ := s.usi.FindUser(ctx, &userStructs.FindUser{Email: body.Email, Phone: body.Phone})
	tx, err := client.Tx(ctx)
	if err != nil {
		return resp.Transactions(err.Error()), nil
	}

	authCode := nanoid.String(6)
	_, err = tx.CodeAuth.Create().SetEmail(strings.ToLower(body.Email)).SetCode(authCode).Save(ctx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.Transactions(err.Error()), nil
	}

	if err := sendAuthEmail(ctx, body.Email, authCode, user.User != nil); err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		log.Errorf(context.Background(), "send mail error: %v", err)
		return resp.BadRequest("send mail failed, please try again or contact support"), nil
	}

	return &resp.Exception{Data: types.JSON{"registered": user.User != nil}}, tx.Commit()
}

// Helper functions for SendCodeService
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
