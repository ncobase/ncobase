package service

import (
	"context"
	"fmt"
	"stocms/internal/config"
	"stocms/internal/data/ent"
	codeAuthEnt "stocms/internal/data/ent/codeauth"
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/internal/server/middleware"
	"stocms/pkg/cookie"
	"stocms/pkg/email"
	"stocms/pkg/jwt"
	"stocms/pkg/nanoid"
	"stocms/pkg/resp"
	"stocms/pkg/types"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterService - Register service
func (svc *Service) RegisterService(c *gin.Context, body *structs.RegisterBody) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	conf := helper.GetConfig(c)
	client := svc.d.GetEntClient()

	// Decode register token
	payload, err := decodeRegisterToken(conf.JWTSecret, body.RegisterToken)
	if err != nil {
		return resp.Forbidden("Register token decode failed"), nil
	}

	// Verify user existence
	exists, err := svc.user.Find(ctx, &structs.FindUser{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})
	if exists != nil {
		return resp.Conflict(getExistMessage(&structs.FindUser{
			Username: exists.Username,
			Email:    exists.Email,
			Phone:    exists.Phone,
		}, body)), nil
	}

	// Disable code
	if err := disableCodeAuth(ctx, client, payload["id"].(string)); err != nil {
		return resp.DBQuery(err.Error()), nil
	}

	// Create user, profile, domain and tokens in a transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		return resp.Transactions(err.Error()), nil
	}

	user, err := createUserAndProfile(ctx, svc, body, payload)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.InternalServer(err.Error()), nil
	}

	if _, err := svc.createDomain(c, &structs.CreateDomainBody{UserID: user.ID}); err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.InternalServer(err.Error()), nil
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
		return resp.InternalServer("authorize is not created", nil), nil
	}

	cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain)
	return &resp.Exception{
		Data: types.JSON{
			"id":           user.ID,
			"access_token": accessToken,
		},
	}, tx.Commit()
}

// Helper functions for RegisterService
func decodeRegisterToken(secret, token string) (types.JSON, error) {
	decoded, err := jwt.DecodeToken(secret, token)
	if err != nil {
		return nil, err
	}
	if decoded["sub"] != "email-register" {
		return nil, fmt.Errorf("not valid authorize information")
	}
	return decoded["payload"].(types.JSON), nil
}

func getExistMessage(exists *structs.FindUser, body *structs.RegisterBody) string {
	switch {
	case exists.Username == body.Username:
		return "username already exists"
	case exists.Phone == body.Phone:
		return "phone already exists"
	default:
		return "email already exists"
	}
}

func disableCodeAuth(ctx context.Context, client *ent.Client, id string) error {
	_, err := client.CodeAuth.Update().Where(codeAuthEnt.ID(id)).SetLogged(true).Save(ctx)
	return err
}

func createUserAndProfile(ctx context.Context, svc *Service, body *structs.RegisterBody, payload types.JSON) (*ent.User, error) {
	user, err := svc.user.Create(ctx, &structs.UserRequestBody{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
		Action:   "create",
	})
	if err != nil {
		return nil, err
	}
	_, err = svc.user.CreateProfile(ctx, &structs.UserRequestBody{
		UserID:      user.ID,
		DisplayName: body.DisplayName,
		ShortBio:    body.ShortBio,
		Action:      "profile",
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
}

// CodeAuthService - Code auth service
func (svc *Service) CodeAuthService(c *gin.Context, code string) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	conf := helper.GetConfig(c)
	client := svc.d.GetEntClient()

	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if ent.IsNotFound(err) {
		return resp.NotFound("Code is not found"), nil
	}
	if codeAuth.Logged || isCodeExpired(codeAuth.CreatedAt) {
		return resp.Forbidden("EXPIRED_CODE"), nil
	}

	user, err := svc.user.Find(ctx, &structs.FindUser{Email: codeAuth.Email})
	if ent.IsNotFound(err) {
		return sendRegisterMail(c, conf, codeAuth)
	}

	return generateTokensForUser(c, client, user, conf.Domain)
}

// Helper functions for CodeAuthService
func isCodeExpired(createdAt time.Time) bool {
	return time.Now().After(createdAt.Add(24 * time.Hour))
}

func sendRegisterMail(c *gin.Context, conf *config.Config, codeAuth *ent.CodeAuth) (*resp.Exception, error) {
	subject := "email-register"
	payload := types.JSON{"email": codeAuth.Email, "id": codeAuth.ID}
	registerToken, err := jwt.GenerateRegisterToken(conf.JWTSecret, codeAuth.ID, payload, subject)
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}
	cookie.SetRegister(c.Writer, registerToken, conf.Domain)
	return &resp.Exception{Data: types.JSON{"email": codeAuth.Email, "register_token": registerToken}}, nil
}

func generateTokensForUser(c *gin.Context, client *ent.Client, user *ent.User, domain string) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
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
		return resp.InternalServer("authorize is not created"), nil
	}
	cookie.Set(c.Writer, accessToken, refreshToken, domain)
	return &resp.Exception{
		Data: types.JSON{
			"id":           user.ID,
			"access_token": accessToken,
		},
	}, tx.Commit()
}

// SendCodeService - Send code service
func (svc *Service) SendCodeService(c *gin.Context, body *structs.SendCodeBody) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	client := svc.d.GetEntClient()

	user, _ := svc.user.Find(ctx, &structs.FindUser{Email: body.Email, Phone: body.Phone})
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

	if err := sendAuthEmail(c, body.Email, authCode, user != nil); err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.BadRequest("Send mail failed"), nil
	}

	return &resp.Exception{Data: types.JSON{"registered": user != nil}}, tx.Commit()
}

// Helper functions for SendCodeService
func sendAuthEmail(c *gin.Context, e, code string, registered bool) error {
	conf := helper.GetConfig(c)
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
	_, err := email.SendTemplateEmailWithMailgun(&email.MailgunConfig{
		APIKey: conf.Mailgun.Key,
		Domain: conf.Mailgun.Domain,
		From:   conf.Mailgun.From,
	}, e, template)
	return err
}
