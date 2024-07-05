package service

import (
	"context"
	"fmt"
	"ncobase/core/data/ent"
	oauthUserEnt "ncobase/core/data/ent/oauthuser"
	userEnt "ncobase/core/data/ent/user"
	"ncobase/core/data/structs"
	"ncobase/helper"
	"ncobase/middleware"
	"net/http"

	"ncobase/common/cookie"
	"ncobase/common/ecode"
	"ncobase/common/jwt"
	"ncobase/common/oauth"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"

	match "github.com/alexpantyukhin/go-pattern-match"
)

// OAuthRegisterService OAuth register service
func (svc *Service) OAuthRegisterService(ctx context.Context, body *structs.OAuthRegisterBody) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)
	registerToken, err := svc.getRegisterToken(ctx, body.RegisterToken)
	if err != nil {
		return resp.Forbidden("register authorize is empty or invalid", nil), nil
	}

	decoded, err := jwt.DecodeToken(conf.Auth.JWT.Secret, registerToken)
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
		return svc.generateAndSetTokens(ctx, tx, &structs.UserBody{ID: user.ID}), tx.Commit()
	}
}

// checkUserExistence Check if user exists
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

// createUserEntities Create user and related entities
func (svc *Service) createUserEntities(ctx context.Context, tx *ent.Tx, body *structs.OAuthRegisterBody, payload structs.RegisterTokenBody) (*ent.User, *resp.Exception) {
	user, err := svc.user.Create(ctx, &structs.UserBody{
		Username: body.Username,
		Email:    payload.Profile.Email,
		Phone:    body.Phone,
	})
	if exception, _ := helper.HandleError("User", err); exception != nil {
		return nil, exception
	}

	if err := svc.createOAuthUser(ctx, tx, payload, user.ID); err != nil {
		return nil, err
	}

	if err := svc.createUserProfile(ctx, &structs.UserProfileBody{
		ID:          user.ID,
		DisplayName: body.DisplayName,
		ShortBio:    body.ShortBio,
	}); err != nil {
		return nil, err
	}

	if _, err := svc.isCreateTenant(ctx, &structs.CreateTenantBody{
		TenantBody: structs.TenantBody{Name: body.Tenant, OperatorBy: structs.OperatorBy{
			CreatedBy: &user.ID,
			UpdatedBy: &user.ID,
		}},
	}); err != nil {
		return nil, resp.InternalServer(err.Error())
	}

	return user, nil
}

// createOAuthUser Create OAuth user
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

// createUserProfile Create user profile
func (svc *Service) createUserProfile(ctx context.Context, body *structs.UserProfileBody) *resp.Exception {
	_, err := svc.userProfile.Create(ctx, body)
	if exception, _ := helper.HandleError("User", err); exception != nil {
		return exception
	}

	return nil
}

// generateAndSetTokens Generate and set tokens
func (svc *Service) generateAndSetTokens(ctx context.Context, tx *ent.Tx, user *structs.UserBody) *resp.Exception {
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

	// cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain) // TODO: move to handler

	return &resp.Exception{
		Data: types.JSON{
			"id":           user.ID,
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
	}
}

// GetOAuthProfileInfoService Get OAuth profile info service
func (svc *Service) GetOAuthProfileInfoService(ctx context.Context, r string) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)

	decoded, err := jwt.DecodeToken(conf.Auth.JWT.Secret, r)
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

// OAuthCallbackService OAuth callback service
func (svc *Service) OAuthCallbackService(ctx context.Context, provider, code string) (*resp.Exception, error) {
	if code == "" {
		return resp.BadRequest("CODE IS EMPTY"), nil
	}

	result := svc.getOAuthInfo(ctx, provider, code).(types.JSON)
	helper.SetProvider(ctx, provider)
	helper.SetToken(ctx, result["authorize"].(string))
	helper.SetProfile(ctx, result["profile"])

	return &resp.Exception{}, nil
}

// OAuthAuthenticationService OAuth authentication service
func (svc *Service) OAuthAuthenticationService(ctx context.Context) (*resp.Exception, error) {
	token := helper.GetToken(ctx)
	provider := helper.GetProvider(ctx)
	profile := helper.GetProfile(ctx).(*oauth.Profile)

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
		return svc.handleExistingOAuthUser(ctx, tx, oauthUser)
	}

	return svc.handleNewOAuthUser(ctx, tx, profile, token, provider)
}

// Handle existing OAuth user
func (svc *Service) handleExistingOAuthUser(ctx context.Context, tx *ent.Tx, oauthUser *ent.OAuthUser) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)
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

	// cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain) // TODO: move to handler
	return &resp.Exception{
		Status: http.StatusMovedPermanently,
		Code:   http.StatusMovedPermanently,
		Data: types.JSON{
			"redirectUrl": helper.GetHost(conf, conf.Domain),
		},
	}, tx.Commit()
}

// Handle new OAuth user
func (svc *Service) handleNewOAuthUser(ctx context.Context, tx *ent.Tx, profile *oauth.Profile, token, provider string) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)
	user, err := svc.user.FindUser(ctx, &structs.FindUser{Email: profile.Email})
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	if validator.IsNil(user) {
		return svc.generateAndSetTokens(ctx, tx, &structs.UserBody{ID: user.ID}), tx.Commit()
	}

	subject := "email-register"
	payload := types.JSON{
		"profile":     profile,
		"provider":    provider,
		"accessToken": token,
	}

	registerToken, err := jwt.GenerateRegisterToken(conf.Auth.JWT.Secret, user.ID, payload, subject)
	if registerToken == "" || err != nil {
		return resp.Forbidden("authorize is not created"), nil
	}

	// cookie.SetRegister(c.Writer, registerToken, conf.Domain) // TODO: move to handler

	return &resp.Exception{
		Status: http.StatusMovedPermanently,
		Code:   http.StatusMovedPermanently,
		Data: types.JSON{
			"redirectUrl": conf.Frontend.SignUpURL + "?oauth=1",
		},
	}, tx.Commit()
}

// Get register token from request
func (svc *Service) getRegisterToken(ctx context.Context, bodyToken string) (string, error) {
	if bodyToken != "" {
		return bodyToken, nil
	}

	if c, ok := helper.GetGinContext(ctx); ok {
		cookieRegisterToken, err := cookie.Get(c.Request, "register_token")
		if err != nil {
			return "", err
		}
		return cookieRegisterToken, nil
	}

	return "", nil
}

// Get OAuth provider information
func (svc *Service) getOAuthInfo(ctx context.Context, provider, code string) any {
	conf := helper.GetConfig(ctx)
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
