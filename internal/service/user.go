package service

import (
	"errors"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/internal/helper"
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

// findUserByID - Internal method to find user by ID
func (svc *Service) findUserByID(c *gin.Context, id string) (structs.User, error) {
	ctx := helper.FromGinContext(c)
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
