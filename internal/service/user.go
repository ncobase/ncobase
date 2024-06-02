package service

import (
	"context"
	"errors"
	"fmt"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/crypto"
	"stocms/pkg/log"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// ReadMeService - Read current user service
func (svc *Service) ReadMeService(c *gin.Context) (*resp.Exception, error) {
	return svc.readUser(c, helper.GetUserID(c))
}

// ReadUserService - Read user service
func (svc *Service) ReadUserService(c *gin.Context, userID string) (*resp.Exception, error) {
	return svc.readUser(c, userID)
}

// UpdatePasswordService - Update user password service
func (svc *Service) UpdatePasswordService(c *gin.Context, body *structs.UserRequestBody) (*resp.Exception, error) {
	verifyResult := svc.verifyUserPassword(c, helper.GetUserID(c), body.OldPassword)
	switch v := verifyResult.(type) {
	case VerifyPasswordResult:
		if v.Valid == false {
			return resp.BadRequest(v.Error), nil
		} else if v.Valid && v.NeedsPasswordSet == true { // 记录第一次设置密码的日志
			log.Printf(nil, "User %s is setting password for the first time", helper.GetUserID(c))
		}
	case error:
		return resp.InternalServer(v.Error()), nil
	}

	err := svc.updateUserPassword(c, helper.GetUserID(c), body.NewPassword)
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	return nil, nil
}

// ****** Internal methods of service

// readUser - Internal method to read user by ID
func (svc *Service) readUser(c *gin.Context, userID string) (*resp.Exception, error) {
	if userID == "" {
		return nil, errors.New("INVALID_USER_ID")
	}

	user, err := svc.findUserByID(c, userID)
	if ent.IsNotFound(err) {
		return resp.NotFound("User is not found"), nil
	} else if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{
		Data: user,
	}, nil
}

// VerifyPasswordResult - Verify password result
type VerifyPasswordResult struct {
	Valid            bool
	NeedsPasswordSet bool
	Error            string
}

// verifyUserPassword - Internal method to verify user password
func (svc *Service) verifyUserPassword(c *gin.Context, userID string, password string) any {
	user, err := svc.user.FindUser(c, &structs.FindUser{ID: userID})
	if ent.IsNotFound(err) {
		return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: "user not found"}
	} else if err != nil {
		return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: fmt.Sprintf("error getting user by ID: %v", err)}
	}

	if user.Password == "" {
		return VerifyPasswordResult{Valid: true, NeedsPasswordSet: true, Error: "user password not set"}
	}

	if crypto.ComparePassword(user.Password, password) {
		return VerifyPasswordResult{Valid: true, NeedsPasswordSet: false, Error: ""}
	}

	return VerifyPasswordResult{Valid: false, NeedsPasswordSet: false, Error: "wrong password"}
}

// updateUserPassword - Internal method to update user password
func (svc *Service) updateUserPassword(ctx context.Context, userID string, password string) error {
	err := svc.user.UpdatePassword(ctx, &structs.UserRequestBody{
		UserID:      userID,
		NewPassword: password,
	})

	if err != nil {
		log.Printf(nil, "Error updating password for user %s: %v", userID, err)
	}

	return err
}

// findUserByID - Internal method to find user by ID
func (svc *Service) findUserByID(ctx context.Context, id string) (structs.User, error) {
	user, err := svc.user.GetByID(ctx, id)
	if err != nil {
		return structs.User{}, err
	}

	profile, _ := svc.user.GetProfile(ctx, user.ID)
	return svc.serializeUser(user, profile), nil
}

// findUser - Internal method to find user by username, email, or phone
func (svc *Service) findUser(c *gin.Context, m *structs.FindUser) (structs.User, error) {
	ctx := helper.FromGinContext(c)
	user, err := svc.user.Find(ctx, &structs.FindUser{
		Username: m.Username,
		Email:    m.Email,
		Phone:    m.Phone,
	})
	if err != nil {
		return structs.User{}, err
	}

	profile, _ := svc.user.GetProfile(ctx, user.ID)
	return svc.serializeUser(user, profile), nil
}

// serializeUser - Serialize user
func (svc *Service) serializeUser(user *ent.User, profile *ent.UserProfile) structs.User {
	readUser := structs.User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Phone:       user.Phone,
		IsCertified: user.IsCertified,
		IsAdmin:     user.IsAdmin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	if profile != nil {
		readUser.Profile = &structs.UserProfile{
			DisplayName: profile.DisplayName,
			ShortBio:    profile.ShortBio,
		}
	}

	return readUser
}
