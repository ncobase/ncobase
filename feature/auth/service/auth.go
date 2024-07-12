package service

import (
	"context"
	"fmt"
	"ncobase/common/jwt"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/auth/data"
	"ncobase/feature/auth/data/ent"
	codeAuthEnt "ncobase/feature/auth/data/ent/codeauth"
	"ncobase/feature/auth/middleware"
	"ncobase/feature/auth/structs"
	tenantService "ncobase/feature/tenant/service"
	tenantStructs "ncobase/feature/tenant/structs"
	userService "ncobase/feature/user/service"
	userStructs "ncobase/feature/user/structs"
	"ncobase/helper"
)

// AuthServiceInterface is the interface for the service.
type AuthServiceInterface interface {
	LoginService(ctx context.Context, body *structs.LoginBody) (*resp.Exception, error)
	RegisterService(ctx context.Context, body *structs.RegisterBody) (*resp.Exception, error)
}

// authService is the struct for the service.
type authService struct {
	d   *data.Data
	cas CodeAuthServiceInterface
	us  *userService.Service
	ts  *tenantService.Service
}

// NewAuthService creates a new service.
func NewAuthService(d *data.Data, cas CodeAuthServiceInterface, us *userService.Service, ts *tenantService.Service) AuthServiceInterface {
	return &authService{
		d:   d,
		cas: cas,
		us:  us,
		ts:  ts,
	}
}

// LoginService login service
func (s *authService) LoginService(ctx context.Context, body *structs.LoginBody) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	rst, err := s.us.User.FindUser(ctx, &userStructs.FindUser{Username: body.Username})
	if exception, err := handleEntError("User", err); exception != nil {
		return exception, err
	}

	if rst.User.Status != 0 {
		return resp.Forbidden("Your account has not been activated"), nil
	}

	verifyResult := s.us.User.VerifyUserPassword(ctx, rst.User.ID, body.Password)
	switch v := verifyResult.(type) {
	case userService.VerifyPasswordResult:
		if v.Valid == false {
			return resp.BadRequest(v.Error), nil
		} else if v.Valid && v.NeedsPasswordSet == true {
			// The user has not set a password and the mailbox is empty
			if validator.IsEmpty(rst.User.Email) {
				return resp.BadRequest("Has not set a password, and the mailbox is empty, please contact the administrator"), nil
			}
			return s.cas.SendCodeService(ctx, &structs.SendCodeBody{Email: rst.User.Email})
		}
	case error:
		return resp.InternalServer(v.Error()), nil
	}

	return generateTokensForUser(ctx, conf, client, rst.User)
}

// RegisterService register service
func (s *authService) RegisterService(ctx context.Context, body *structs.RegisterBody) (*resp.Exception, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	// Decode register token
	payload, err := decodeRegisterToken(conf.Auth.JWT.Secret, body.RegisterToken)
	if err != nil {
		return resp.Forbidden("Register token decode failed"), nil
	}

	// Verify user existence
	exists, err := s.us.User.FindUser(ctx, &userStructs.FindUser{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})
	if err != nil && exists.User != nil {
		return resp.Conflict(getExistMessage(&userStructs.FindUser{
			Username: exists.User.Username,
			Email:    exists.User.Email,
			Phone:    exists.User.Phone,
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

	rst, err := createUserAndProfile(ctx, s, body, payload)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.InternalServer(err.Error()), nil
	}

	if _, err := s.ts.Tenant.IsCreateTenant(ctx, &tenantStructs.CreateTenantBody{
		TenantBody: tenantStructs.TenantBody{Name: body.Tenant, CreatedBy: &rst.User.ID, UpdatedBy: &rst.User.ID},
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.InternalServer(err.Error()), nil
	}

	authToken, err := createAuthToken(ctx, tx, rst.User.ID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.Transactions(err.Error()), nil
	}

	accessToken, refreshToken := middleware.GenerateUserToken(conf.Auth.JWT.Secret, rst.User.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return resp.InternalServer(err.Error()), nil
		}
		return resp.InternalServer("authorize is not created", nil), nil
	}

	// cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain) // TODO: move to handler
	return &resp.Exception{
		Data: types.JSON{
			"id":           rst.User.ID,
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

func getExistMessage(exists *userStructs.FindUser, body *structs.RegisterBody) string {
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

func createUserAndProfile(ctx context.Context, svc *authService, body *structs.RegisterBody, payload types.JSON) (*userStructs.UserMeshes, error) {
	rst, err := svc.us.User.CreateUserService(ctx, &userStructs.UserMeshes{
		User: &userStructs.UserBody{
			Username: body.Username,
			Email:    payload["email"].(string),
			Phone:    body.Phone,
		},
		Profile: &userStructs.UserProfileBody{
			DisplayName: body.DisplayName,
			ShortBio:    body.ShortBio,
		},
	})
	if err != nil {
		return nil, err
	}
	user := rst.Data.(*userStructs.UserMeshes)
	return user, nil
}

func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
}
