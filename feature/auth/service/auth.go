package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/common/jwt"
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
	Login(ctx context.Context, body *structs.LoginBody) (*types.JSON, error)
	Register(ctx context.Context, body *structs.RegisterBody) (*types.JSON, error)
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

// Login login service
func (s *authService) Login(ctx context.Context, body *structs.LoginBody) (*types.JSON, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	rst, err := s.us.User.FindUser(ctx, &userStructs.FindUser{Username: body.Username})
	if err := handleEntError("User", err); err != nil {
		return nil, err
	}

	if rst.User.Status != 0 {
		return nil, errors.New("account has been disabled, please contact the administrator")
	}

	verifyResult := s.us.User.VerifyPassword(ctx, rst.User.ID, body.Password)
	switch v := verifyResult.(type) {
	case userService.VerifyPasswordResult:
		if v.Valid == false {
			return nil, errors.New(v.Error)
		} else if v.Valid && v.NeedsPasswordSet == true {
			// The user has not set a password and the mailbox is empty
			if validator.IsEmpty(rst.User.Email) {
				return nil, errors.New("has not set a password, and the mailbox is empty, please contact the administrator")
			}
			return s.cas.SendCode(ctx, &structs.SendCodeBody{Email: rst.User.Email})
		}
	case error:
		return nil, v
	}

	return generateTokensForUser(ctx, conf, client, rst.User)
}

// Register register service
func (s *authService) Register(ctx context.Context, body *structs.RegisterBody) (*types.JSON, error) {
	conf := helper.GetConfig(ctx)
	client := s.d.GetEntClient()

	// Decode register token
	payload, err := decodeRegisterToken(conf.Auth.JWT.Secret, body.RegisterToken)
	if err != nil {
		return nil, errors.New("register token decode failed")
	}

	// Verify user existence
	exists, err := s.us.User.FindUser(ctx, &userStructs.FindUser{
		Username: body.Username,
		Email:    payload["email"].(string),
		Phone:    body.Phone,
	})
	if err != nil && exists.User != nil {
		return nil, errors.New(getExistMessage(&userStructs.FindUser{
			Username: exists.User.Username,
			Email:    exists.User.Email,
			Phone:    exists.User.Phone,
		}, body))
	}

	// Disable code
	if err := disableCodeAuth(ctx, client, payload["id"].(string)); err != nil {
		return nil, err
	}

	// Create user, profile, tenant and tokens in a transaction
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	rst, err := createUserAndProfile(ctx, s, body, payload)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	if _, err := s.ts.Tenant.IsCreate(ctx, &tenantStructs.CreateTenantBody{
		TenantBody: tenantStructs.TenantBody{Name: body.Tenant, CreatedBy: &rst.User.ID, UpdatedBy: &rst.User.ID},
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	authToken, err := createAuthToken(ctx, tx, rst.User.ID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	accessToken, refreshToken := middleware.GenerateUserToken(conf.Auth.JWT.Secret, rst.User.ID, authToken.ID)
	if accessToken == "" || refreshToken == "" {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, errors.New("authorize is not created")
	}

	// cookie.Set(c.Writer, accessToken, refreshToken, conf.Domain) // TODO: move to handler
	return &types.JSON{
		"id":           rst.User.ID,
		"access_token": accessToken,
	}, tx.Commit()
}

// Helper functions for Register
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
	rst, err := svc.us.User.CreateUser(ctx, &userStructs.UserMeshes{
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
	return rst, nil
}

func createAuthToken(ctx context.Context, tx *ent.Tx, userID string) (*ent.AuthToken, error) {
	return tx.AuthToken.Create().SetUserID(userID).Save(ctx)
}
