package service

//
// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"ncobase/core/auth/data"
// 	"ncobase/core/auth/data/ent"
// 	oauthUserEnt "ncobase/core/auth/data/ent/oauthuser"
// 	"ncobase/core/auth/middleware"
// 	authStructs "ncobase/core/auth/structs"
// 	tenantStructs "ncobase/core/tenant/structs"
// 	userEnt "ncobase/core/user/data/ent/user"
// 	userStructs "ncobase/core/user/structs"
// 	"ncobase/helper"
// 	"net/http"
//
// 	"github.com/ncobase/ncore/net/cookie"
// 	"github.com/ncobase/ncore/ecode"
// 	"github.com/ncobase/ncore/security/jwt"
// 	"github.com/ncobase/ncore/security/oauth"
// 	"github.com/ncobase/ncore/net/resp"
// 	"github.com/ncobase/ncore/types"
// 	"github.com/ncobase/ncore/validation/validator"
//
// 	match "github.com/alexpantyukhin/go-pattern-match"
// )
//
// // OAuthServiceInterface is the interface for the service.
// type OAuthServiceInterface interface {
// 	OAuthRegister(ctx context.Context, body *authStructs.OAuthRegisterBody) (*types.JSON, error)
// 	GetOAuthProfileInfo(ctx context.Context, r string) (*types.JSON, error)
// 	OAuthCallback(ctx context.Context, provider, code string) (*types.JSON, error)
// 	OAuthAuthentication(ctx context.Context) (*types.JSON, error)
// }
//
// // oAuthService is the struct for the service.
// type oAuthService struct {
// 	d *data.Data
// }
//
// // NewOAuth creates a new service.
// func NewOAuthService(d *data.Data) OAuthServiceInterface {
// 	return &oAuthService{
// 		d: d,
// 	}
// }
//
// // OAuthRegister OAuth register service
// func (svc *oAuthService) OAuthRegister(ctx context.Context, body *authStructs.OAuthRegisterBody) (*types.JSON, error) {
// 	conf := helper.GetConfig(ctx)
// 	registerToken, err := svc.getRegisterToken(ctx, body.RegisterToken)
// 	if err != nil {
// 		return nil, errors.New("register authorize is empty or invalid")
// 	}
//
// 	decoded, err := jwt.DecodeToken(conf.Auth.JWT.Secret, registerToken)
// 	if err != nil {
// 		return nil, errors.New("decoded parsing is missing")
// 	}
//
// 	fmt.Println(decoded)
// 	payload := decoded["payload"].(authStructs.RegisterTokenBody)
//
// 	bg := context.Background()
// 	client := svc.d.GetEntClient()
//
// 	tx, err := client.Tx(bg)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Check if user already exists
// 	if exists, err := svc.checkUserExistence(tx, body.Username, payload.Profile.Email); err != nil {
// 		return nil, err
// 	} else if exists != "" {
// 		return &types.JSON{
// 			"status":  http.StatusConflict,
// 			"code":    ecode.Conflict,
// 			"message": exists,
// 		}, nil
// 	}
//
// 	// Create user and related entities
// 	if user, err := svc.createUserEntities(ctx, tx, body, payload); err != nil {
// 		return err, nil
// 	} else {
// 		return svc.generateAndSetTokens(ctx, tx, &userStructs.UserBody{ID: user.ID}), tx.Commit()
// 	}
// }
//
// // checkUserExistence Check if user exists
// func (svc *oAuthService) checkUserExistence(tx *ent.Tx, username, email string) (string, error) {
// 	exists, err := tx.User.
// 		Query().
// 		Where(userEnt.Or(
// 			userEnt.UsernameEQ(username),
// 			userEnt.EmailEQ(email),
// 		)).
// 		Only(context.Background())
//
// 	if err != nil && !ent.IsNotFound(err) {
// 		return "", err
// 	}
//
// 	if exists != nil {
// 		if exists.Username == username {
// 			return "username", nil
// 		}
// 		return "email", nil
// 	}
// 	return "", nil
// }
//
// // createUserEntities Create user and related entities
// func (svc *oAuthService) createUserEntities(ctx context.Context, tx *ent.Tx, body *authStructs.OAuthRegisterBody, payload authStructs.RegisterTokenBody) (*userStructs.AccountMeshes, error) {
// 	user, err := svc.user.Create(ctx, &userStructs.UserBody{
// 		Username: body.Username,
// 		Email:    payload.Profile.Email,
// 		Phone:    body.Phone,
// 	})
// 	if err := handleEntError(ctx, "User", err); err != nil {
// 		return nil, err
// 	}
//
// 	if err := svc.createOAuthUser(ctx, tx, payload, user.ID); err != nil {
// 		return nil, err
// 	}
//
// 	if err := svc.createUserProfile(ctx, &userStructs.UserProfileBody{
// 		ID:          user.ID,
// 		DisplayName: body.DisplayName,
// 		ShortBio:    body.ShortBio,
// 	}); err != nil {
// 		return nil, err
// 	}
//
// 	if _, err := svc.isCreateTenant(ctx, &tenantStructs.CreateTenantBody{
// 		TenantBody: tenantStructs.TenantBody{Name: body.Tenant, CreatedBy: &user.ID, UpdatedBy: &user.ID},
// 	}); err != nil {
// 		return nil, resp.InternalServer(err.Error())
// 	}
//
// 	return user, nil
// }
//
// // createOAuthUser Create OAuth user
// func (svc *oAuthService) createOAuthUser(ctx context.Context, tx *ent.Tx, payload authStructs.RegisterTokenBody, userID string) *types.JSON {
// 	_, err := tx.OAuthUser.
// 		Create().
// 		SetAccessToken(payload.Token).
// 		SetProvider(payload.Provider).
// 		SetUserID(userID).
// 		SetOauthID(payload.Profile.ID).
// 		Save(ctx)
// 	if err != nil {
// 		if rear := tx.Rollback(); rear != nil {
// 			return resp.Transactions(rear.Error())
// 		}
// 		return resp.InternalServer(err.Error())
// 	}
// 	return nil
// }
//
// // createUserProfile Create user profile
// func (svc *oAuthService) createUserProfile(ctx context.Context, body *userStructs.UserProfileBody) *types.JSON {
// 	_, err := svc.userProfile.Create(ctx, body)
// 	if exception, _ := handleEntError(ctx, "User", err); exception != nil {
// 		return exception
// 	}
//
// 	return nil
// }
//
// // generateAndSetTokens Generate and set tokens
// func (svc *oAuthService) generateAndSetTokens(ctx context.Context, tx *ent.Tx, user *userStructs.UserBody) *types.JSON {
// 	authToken, err := tx.AuthToken.Create().SetUserID(user.ID).Save(context.Background())
// 	if err != nil {
// 		if rear := tx.Rollback(); rear != nil {
// 			return resp.Transactions(rear.Error())
// 		}
// 		return resp.InternalServer(err.Error())
// 	}
//
// 	accessToken, refreshToken := middleware.GenerateUserToken(user.ID, authToken.ID)
// 	if accessToken == "" || refreshToken == "" {
// 		if err := tx.Rollback(); err != nil {
// 			return resp.Transactions(err.Error())
// 		}
// 		return resp.InternalServer("authorize is not created")
// 	}
//
// 	return &types.JSON{
// 		"id":           user.ID,
// 		"accessToken":  accessToken,
// 		"refreshToken": refreshToken,
// 	}
// }
//
// // GetOAuthProfileInfo Get OAuth profile info service
// func (svc *oAuthService) GetOAuthProfileInfo(ctx context.Context, r string) (*types.JSON, error) {
// 	conf := helper.GetConfig(ctx)
//
// 	decoded, err := jwt.DecodeToken(conf.Auth.JWT.Secret, r)
// 	if err != nil {
// 		return nil, errors.New("decoded parsing is missing")
// 	}
//
// 	return &types.JSON{
// 		"profile": decoded,
// 	}, nil
// }
//
// // OAuthCallback OAuth callback service
// func (svc *oAuthService) OAuthCallback(ctx context.Context, provider, code string) (*types.JSON, error) {
// 	if code == "" {
// 		return resp.BadRequest("CODE IS EMPTY"), nil
// 	}
//
// 	result := svc.getOAuthInfo(ctx, provider, code).(types.JSON)
// 	helper.SetProvider(ctx, provider)
// 	helper.SetToken(ctx, result["authorize"].(string))
// 	helper.SetProfile(ctx, result["profile"])
//
// 	return &resp.Exception{}, nil
// }
//
// // OAuthAuthentication OAuth authentication service
// func (svc *oAuthService) OAuthAuthentication(ctx context.Context) (*types.JSON, error) {
// 	token := helper.GetToken(ctx)
// 	provider := helper.GetProvider(ctx)
// 	profile := helper.GetProfile(ctx).(*oauth.Profile)
//
// 	if profile == nil || token == "" {
// 		return resp.Forbidden("profile and authorize is empty"), nil
// 	}
//
// 	client := svc.d.GetEntClient()
//
// 	tx, err := client.Tx(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Check if OAuth user exists
// 	oauthUser, err := tx.OAuthUser.Query().Where(
// 		oauthUserEnt.And(
// 			oauthUserEnt.OauthID(profile.ID),
// 			oauthUserEnt.ProviderEQ(provider),
// 		),
// 	).Only(ctx)
//
// 	if !ent.IsNotFound(err) {
// 		return svc.handleExistingOAuthUser(ctx, tx, oauthUser)
// 	}
//
// 	return svc.handleNewOAuthUser(ctx, tx, profile, token, provider)
// }
//
// // Handle existing OAuth user
// func (svc *oAuthService) handleExistingOAuthUser(ctx context.Context, tx *ent.Tx, oauthUser *ent.OAuthUser) (*types.JSON, error) {
// 	conf := helper.GetConfig(ctx)
// 	user, err := tx.User.Query().Where(userEnt.IDEQ(oauthUser.UserID)).Only(context.Background())
// 	if ent.IsNotFound(err) {
// 		return resp.NotFound("User is missing"), nil
// 	}
//
// 	authToken, err := tx.AuthToken.Create().SetUserID(user.ID).Save(context.Background())
// 	if err != nil {
// 		if rear := tx.Rollback(); rear != nil {
// 			return resp.Transactions(rear.Error()), nil
// 		}
// 		return nil, err
// 	}
//
// 	accessToken, refreshToken := middleware.GenerateUserToken(user.ID, authToken.ID)
// 	if accessToken == "" || refreshToken == "" {
// 		if err := tx.Rollback(); err != nil {
// 			return nil, err
// 		}
// 		return resp.InternalServer("authorize is not created", nil), nil
// 	}
//
// 	return &resp.Exception{
// 		Status: http.StatusMovedPermanently,
// 		Code:   http.StatusMovedPermanently,
// 		Data: types.JSON{
// 			"redirectUrl": helper.GetHost(conf, conf.Domain),
// 		},
// 	}, tx.Commit()
// }
//
// // Handle new OAuth user
// func (svc *oAuthService) handleNewOAuthUser(ctx context.Context, tx *ent.Tx, profile *oauth.Profile, token, provider string) (*types.JSON, error) {
// 	conf := helper.GetConfig(ctx)
// 	user, err := svc.user.FindUser(ctx, &userStructs.FindUser{Email: profile.Email})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if validator.IsNil(user) {
// 		return svc.generateAndSetTokens(ctx, tx, &userStructs.UserBody{ID: user.ID}), tx.Commit()
// 	}
//
// 	subject := "email-register"
// 	payload := types.JSON{
// 		"profile":     profile,
// 		"provider":    provider,
// 		"accessToken": token,
// 	}
//
// 	registerToken, err := jwt.GenerateRegisterToken(conf.Auth.JWT.Secret, user.ID, payload, subject)
// 	if registerToken == "" || err != nil {
// 		return resp.Forbidden("authorize is not created"), nil
// 	}
//
// 	return &resp.Exception{
// 		Status: http.StatusMovedPermanently,
// 		Code:   http.StatusMovedPermanently,
// 		Data: types.JSON{
// 			"redirectUrl": conf.Frontend.SignUpURL + "?oauth=1",
// 		},
// 	}, tx.Commit()
// }
//
// // Get register token from request
// func (svc *oAuthService) getRegisterToken(ctx context.Context, bodyToken string) (string, error) {
// 	if bodyToken != "" {
// 		return bodyToken, nil
// 	}
//
// 	if c, ok := helper.GetGinContext(ctx); ok {
// 		cookieRegisterToken, err := cookie.Get(c.Request, "register_token")
// 		if err != nil {
// 			return "", err
// 		}
// 		return cookieRegisterToken, nil
// 	}
//
// 	return "", nil
// }
//
// // Get OAuth provider information
// func (svc *oAuthService) getOAuthInfo(ctx context.Context, provider, code string) any {
// 	conf := helper.GetConfig(ctx)
// 	_, result := match.Match(provider).
// 		When("facebook", func() types.JSON {
// 			accessToken, _ := oauth.GetFacebookAccessToken(helper.GetHost(conf, conf.Domain), code)
// 			profile, _ := oauth.GetFacebookProfile(accessToken)
// 			return types.JSON{"authorize": accessToken, "profile": profile}
// 		}).
// 		When("github", func() types.JSON {
// 			accessToken, _ := oauth.GetGithubAccessToken(code)
// 			profile, _ := oauth.GetGithubProfile(accessToken)
// 			return types.JSON{"authorize": accessToken, "profile": profile}
// 		}).
// 		Result()
//
// 	return result
// }
