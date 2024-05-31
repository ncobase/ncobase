package service

import (
	"context"
	"fmt"
	"net/http"
	"stocms/internal/data/ent"
	oauthUserEnt "stocms/internal/data/ent/oauthuser"
	userEnt "stocms/internal/data/ent/user"
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/internal/server/middleware"
	"stocms/pkg/cookie"
	"stocms/pkg/ecode"
	"stocms/pkg/jwt"
	"stocms/pkg/oauth"
	"stocms/pkg/resp"
	"stocms/pkg/types"
	"stocms/pkg/validator"

	match "github.com/alexpantyukhin/go-pattern-match"
	"github.com/gin-gonic/gin"
)

// OAuthRegisterService - OAuth register service
func (svc *Service) OAuthRegisterService(c *gin.Context, body *structs.OAuthRegisterBody) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	conf := helper.GetConfig(c)
	registerToken, err := svc.getRegisterToken(c, body.RegisterToken)
	if err != nil {
		return resp.Forbidden("register authorize is empty or invalid", nil), nil
	}

	decoded, err := jwt.DecodeToken(conf.JWTSecret, registerToken)
	if err != nil {
		return resp.NotFound("decoded parsing is missing", nil), nil
	}

	fmt.Println(decoded)
	payload := decoded["payload"].(structs.RegisterTokenBody)

	bg := context.Background()
	client := svc.d.GetEntClient()

	tx, err := client.Tx(bg)
	if err != nil {
		return resp.Transactions(err.Error()), nil
	}

	// Check if user already exists
	if exists, err := svc.checkUserExistence(tx, body.Username, payload.Profile.Email); err != nil {
		return resp.InternalServer(err.Error()), nil
	} else if exists != "" {
		return &resp.Exception{
			Status:  http.StatusConflict,
			Code:    ecode.Conflict,
			Message: exists,
		}, nil
	}

	// Create user and related entities
	if user, err := svc.createUserEntities(ctx, tx, body, payload); err != nil {
		return err, nil
	} else {
		return svc.generateAndSetTokens(c, tx, &structs.ReadUser{ID: user.ID}), tx.Commit()
	}
}

// checkUserExistence - Check if user exists
func (svc *Service) checkUserExistence(tx *ent.Tx, username, email string) (string, error) {
	exists, err := tx.User.
		Query().
		Where(userEnt.Or(
			userEnt.UsernameEQ(username),
			userEnt.EmailEQ(email),
		)).
		Only(context.Background())

	if err != nil && !ent.IsNotFound(err) {
		return "", err
	}

	if exists != nil {
		if exists.Username == username {
			return "username", nil
		}
		return "email", nil
	}
	return "", nil
}

// createUserEntities - Create user and related entities
func (svc *Service) createUserEntities(ctx context.Context, tx *ent.Tx, body *structs.OAuthRegisterBody, payload structs.RegisterTokenBody) (*ent.User, *resp.Exception) {
	user, err := svc.user.Create(ctx, &structs.CreateUserBody{
		Username: body.Username,
		Email:    payload.Profile.Email,
		Phone:    body.Phone,
	})
	if err != nil {
		return nil, resp.InternalServer(err.Error())
	}

	if err := svc.createOAuthUser(ctx, tx, payload, user.ID); err != nil {
		return nil, err
	}

	if err := svc.createUserProfile(ctx, user.ID, body.DisplayName, body.ShortBio); err != nil {
		return nil, err
	}

	if _, err := svc.createDomain(ctx, &structs.CreateDomainBody{UserID: user.ID}); err != nil {
		return nil, resp.InternalServer(err.Error())
	}

	return user, nil
}

// createOAuthUser - Create OAuth user
func (svc *Service) createOAuthUser(ctx context.Context, tx *ent.Tx, payload structs.RegisterTokenBody, userID string) *resp.Exception {
	_, err := tx.OAuthUser.
		Create().
		SetAccessToken(payload.Token).
		SetProvider(payload.Provider).
		SetUserID(userID).
		SetOauthID(payload.Profile.ID).
		Save(ctx)
	if err != nil {
		if rear := tx.Rollback(); rear != nil {
			return resp.Transactions(rear.Error())
		}
		return resp.InternalServer(err.Error())
	}
	return nil
}

// createUserProfile - Create user profile
func (svc *Service) createUserProfile(ctx context.Context, userID, displayName, shortBio string) *resp.Exception {
	_, err := svc.user.CreateProfile(ctx, &structs.CreateProfileBody{
		UserID:      userID,
		DisplayName: displayName,
		ShortBio:    shortBio,
	})
	if err != nil {
		return resp.InternalServer(err.Error())
	}
	return nil
}

// generateAndSetTokens - Generate and set tokens
func (svc *Service) generateAndSetTokens(c *gin.Context, tx *ent.Tx, user *structs.ReadUser) *resp.Exception {
	conf := helper.GetConfig(c)
	authToken, err := tx.AuthToken.Create().SetUserID(user.ID).Save(context.Background())
	if err != nil {
		if rear := tx.Rollback(); rear != nil {
			return resp.Transactions(rear.Error())
		}
		return resp.InternalServer(err.Error())
	}

	accessToken, refreshToken := middleware.GenerateUserToken(user.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return resp.Transactions(err.Error())
		}
		return resp.InternalServer("authorize is not created")
	}

	cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain)
	return &resp.Exception{
		Data: types.JSON{
			"id":           user.ID,
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
	}
}

// GetOAuthProfileInfoService - Get OAuth profile info service
func (svc *Service) GetOAuthProfileInfoService(c *gin.Context) (*resp.Exception, error) {
	conf := helper.GetConfig(c)
	registerToken, err := cookie.Get(c.Request, "register_token")
	if err != nil {
		return resp.Forbidden("register authorize is empty or invalid"), nil
	}

	decoded, err := jwt.DecodeToken(conf.JWTSecret, registerToken)
	if err != nil {
		return resp.NotFound("decoded parsing is missing", nil), nil
	}

	return &resp.Exception{
		Status: http.StatusMovedPermanently,
		Code:   http.StatusMovedPermanently,
		Data: types.JSON{
			"profile": decoded,
		},
	}, nil
}

// OAuthCallbackService - OAuth callback service
func (svc *Service) OAuthCallbackService(c *gin.Context, provider, code string) (*resp.Exception, error) {
	if code == "" {
		return resp.BadRequest("CODE IS EMPTY"), nil
	}

	result := svc.getOAuthInfo(c, provider, code).(types.JSON)
	helper.SetProvider(c, provider)
	helper.SetToken(c, result["authorize"].(string))
	helper.SetProfile(c, result["profile"])

	return &resp.Exception{}, nil
}

// OAuthAuthenticationService - OAuth authentication service
func (svc *Service) OAuthAuthenticationService(c *gin.Context) (*resp.Exception, error) {
	ctx := helper.FromGinContext(c)
	token := helper.GetToken(c)
	provider := helper.GetProvider(c)
	profile := helper.GetProfile(c).(*oauth.Profile)

	if profile == nil || token == "" {
		return resp.Forbidden("profile and authorize is empty"), nil
	}

	client := svc.d.GetEntClient()

	tx, err := client.Tx(ctx)
	if err != nil {
		return resp.Transactions(err.Error()), nil
	}

	// Check if OAuth user exists
	oauthUser, err := tx.OAuthUser.Query().Where(
		oauthUserEnt.And(
			oauthUserEnt.OauthID(profile.ID),
			oauthUserEnt.ProviderEQ(provider),
		),
	).Only(ctx)

	if !ent.IsNotFound(err) {
		return svc.handleExistingOAuthUser(c, tx, oauthUser)
	}

	return svc.handleNewOAuthUser(c, tx, profile, token, provider)
}

// handleExistingOAuthUser - Handle existing OAuth user
func (svc *Service) handleExistingOAuthUser(c *gin.Context, tx *ent.Tx, oauthUser *ent.OAuthUser) (*resp.Exception, error) {
	conf := helper.GetConfig(c)
	user, err := tx.User.Query().Where(userEnt.IDEQ(oauthUser.UserID)).Only(context.Background())
	if ent.IsNotFound(err) {
		return resp.NotFound("User is missing"), nil
	}

	authToken, err := tx.AuthToken.Create().SetUserID(user.ID).Save(context.Background())
	if err != nil {
		if rear := tx.Rollback(); rear != nil {
			return resp.Transactions(rear.Error()), nil
		}
		return resp.InternalServer(err.Error()), nil
	}

	accessToken, refreshToken := middleware.GenerateUserToken(user.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return resp.Transactions(err.Error()), nil
		}
		return resp.InternalServer("authorize is not created", nil), nil
	}

	cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain)
	return &resp.Exception{
		Status: http.StatusMovedPermanently,
		Code:   http.StatusMovedPermanently,
		Data: types.JSON{
			"redirectUrl": helper.GetHost(conf, conf.Domain),
		},
	}, tx.Commit()
}

// handleNewOAuthUser - Handle new OAuth user
func (svc *Service) handleNewOAuthUser(c *gin.Context, tx *ent.Tx, profile *oauth.Profile, token, provider string) (*resp.Exception, error) {
	conf := helper.GetConfig(c)
	user, err := svc.findUser(c, &structs.UserKey{Email: profile.Email})
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	if validator.IsNil(user) {
		return svc.generateAndSetTokens(c, tx, &user), tx.Commit()
	}

	subject := "email-register"
	payload := types.JSON{
		"profile":     profile,
		"provider":    provider,
		"accessToken": token,
	}

	registerToken, err := jwt.GenerateRegisterToken(conf.JWTSecret, user.ID, payload, subject)
	if registerToken == "" || err != nil {
		return resp.Forbidden("authorize is not created"), nil
	}

	cookie.SetRegister(c.Writer, registerToken, conf.Domain)
	return &resp.Exception{
		Status: http.StatusMovedPermanently,
		Code:   http.StatusMovedPermanently,
		Data: types.JSON{
			"redirectUrl": conf.Frontend.SignUpURL + "?oauth=1",
		},
	}, tx.Commit()
}

// getRegisterToken - Get register token from request
func (svc *Service) getRegisterToken(c *gin.Context, bodyToken string) (string, error) {
	if bodyToken != "" {
		return bodyToken, nil
	}

	cookieRegisterToken, err := cookie.Get(c.Request, "register_token")
	if err != nil {
		return "", err
	}
	return cookieRegisterToken, nil
}

// getOAuthInfo - Get OAuth provider information
func (svc *Service) getOAuthInfo(c *gin.Context, provider, code string) any {
	conf := helper.GetConfig(c)
	_, result := match.Match(provider).
		When("facebook", func() types.JSON {
			accessToken, _ := oauth.GetFacebookAccessToken(helper.GetHost(conf, conf.Domain), code)
			profile, _ := oauth.GetFacebookProfile(accessToken)
			return types.JSON{"authorize": accessToken, "profile": profile}
		}).
		When("github", func() types.JSON {
			accessToken, _ := oauth.GetGithubAccessToken(code)
			profile, _ := oauth.GetGithubProfile(accessToken)
			return types.JSON{"authorize": accessToken, "profile": profile}
		}).
		Result()

	return result
}
