package service

import (
	"context"
	"fmt"
	"ncobase/internal/data/ent"
	codeAuthEnt "ncobase/internal/data/ent/codeauth"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"
	"ncobase/internal/server/middleware"
	"strings"
	"time"

	"ncobase/common/config"
	"ncobase/common/cookie"
	"ncobase/common/ecode"
	"ncobase/common/email"
	"ncobase/common/jwt"
	"ncobase/common/log"
	"ncobase/common/nanoid"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

// RegisterService register service
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

	// Create user, profile, tenant and tokens in a transaction
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

	if _, err := svc.isCreateTenant(ctx, &structs.CreateTenantBody{
		TenantBody: structs.TenantBody{Name: body.Tenant, CreatedBy: user.ID},
	}); err != nil {
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
		return "Username already exists"
	case exists.Phone == body.Phone:
		return "Phone already exists"
	default:
		return "Email already exists"
	}
}

func disableCodeAuth(ctx context.Context, client *ent.Client, id string) error {
	_, err := client.CodeAuth.Update().Where(codeAuthEnt.ID(id)).SetLogged(true).Save(ctx)
	return err
}

func createUserAndProfile(ctx context.Context, svc *Service, body *structs.RegisterBody, payload types.JSON) (*ent.User, error) {
	user, err := svc.user.Create(ctx, &structs.UserBody{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})
	if err != nil {
		return nil, err
	}
	_, err = svc.userProfile.Create(ctx, &structs.UserProfileBody{
		ID:          user.ID,
		DisplayName: body.DisplayName,
		ShortBio:    body.ShortBio,
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
}

// CodeAuthService code auth service
func (svc *Service) CodeAuthService(c *gin.Context, code string) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	conf := helper.GetConfig(c)
	client := svc.d.GetEntClient()

	codeAuth, err := client.CodeAuth.Query().Where(codeAuthEnt.CodeEQ(code)).Only(ctx)
	if exception, err := handleError("Code", err); exception != nil {
		return exception, err
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

func generateTokensForUser(c *gin.Context, client *ent.Client, user *ent.User, tenant string) (*resp.Exception, error) {
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
		return resp.InternalServer("Authorize is not created"), nil
	}
	cookie.Set(c.Writer, accessToken, refreshToken, tenant)
	return &resp.Exception{
		Data: types.JSON{
			"id":           user.ID,
			"access_token": accessToken,
		},
	}, tx.Commit()
}

// SendCodeService send code service
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
		log.Errorf(context.Background(), "send mail error: %v", err)
		return resp.BadRequest("send mail failed, please try again or contact support"), nil
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
	_, err := helper.SendEmailWithTemplate(c, e, template)
	return err
}

// LoginService login service
func (svc *Service) LoginService(c *gin.Context, body *structs.LoginBody) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	conf := helper.GetConfig(c)
	client := svc.d.GetEntClient()

	user, err := svc.user.FindUser(ctx, &structs.FindUser{Username: body.Username})
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	if user.Status != 0 {
		return resp.Forbidden("Your account has not been activated"), nil
	}

	verifyResult := svc.verifyUserPassword(c, user.ID, body.Password)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return resp.BadRequest(v.Error), nil
		} else if v.Valid && v.NeedsPasswordSet == true {
			// The user has not set a password and the mailbox is empty
			if validator.IsEmpty(user.Email) {
				return resp.BadRequest("Has not set a password, and the mailbox is empty, please contact the administrator"), nil
			}
			return svc.SendCodeService(c, &structs.SendCodeBody{Email: user.Email})
		}
	case error:
		return resp.InternalServer(v.Error()), nil
	}

	return generateTokensForUser(c, client, user, conf.Domain)
}

// GenerateCaptchaService generates a new captcha ID and image URL.
func (svc *Service) GenerateCaptchaService(_ *gin.Context, ext string) (*resp.Exception, error) {
	captchaID := captcha.New()
	captchaURL := "/v1/captcha/" + captchaID + ext

	// Set captcha ID in cache
	if err := svc.captcha.Set(context.Background(), captchaID, &types.JSON{"id": captchaID, "url": captchaURL}); err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{
		Data: &types.JSON{"url": captchaURL},
	}, nil
}

// GetCaptchaService gets the captcha from the cache.
func (svc *Service) GetCaptchaService(_ *gin.Context, id string) *resp.Exception {
	cached, err := svc.captcha.Get(context.Background(), id)
	if err != nil {
		return resp.NotFound(ecode.NotExist("captcha"))
	}
	return &resp.Exception{
		Data: cached,
	}
}

// ValidateCaptchaService validates the captcha code.
func (svc *Service) ValidateCaptchaService(_ *gin.Context, body *structs.Captcha) *resp.Exception {
	if body == nil || !captcha.VerifyString(body.ID, body.Solution) {
		return resp.BadRequest(ecode.FieldIsInvalid("captcha"))
	}

	// Delete captcha after verification
	if err := svc.captcha.Delete(context.Background(), body.ID); err != nil {
		return resp.InternalServer(err.Error())
	}

	return &resp.Exception{}
}
