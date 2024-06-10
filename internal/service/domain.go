package service

import (
	"context"
	"encoding/json"
	"errors"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/types"
	"stocms/pkg/validator"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

// AccountDomainService account domain service
func (svc *Service) AccountDomainService(c *gin.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("INVALID_USER_ID")
	}

	domain, err := svc.domain.GetByUser(c, userID)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, false),
	}, nil
}

// UserDomainService user domain service
func (svc *Service) UserDomainService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	domain, err := svc.domain.GetByUser(c, user.ID)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, false),
	}, nil
}

// CreateDomainService create domain service
func (svc *Service) CreateDomainService(c *gin.Context, body *structs.CreateDomainBody) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return resp.UnAuthorized("INVALID_USER_ID"), nil
	}

	user, err := svc.user.GetByID(c, userID)
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	body.CreatedBy = user.ID
	domain, err := svc.domain.Create(helper.FromGinContext(c), body)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: types.JSON{
			"id": domain.ID,
		},
	}, nil
}

// UpdateDomainService update domain service (full and partial).
func (svc *Service) UpdateDomainService(c *gin.Context, body *structs.UpdateDomainBody) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("INVALID_USER_ID")
	}

	if body.CreatedBy != "" {
		_, err := svc.domain.GetByUser(helper.FromGinContext(c), body.CreatedBy)
		if exception, err := handleError("Domain", err); exception != nil {
			return exception, err
		}
	}

	if body.ID == "" {
		body.ID, _ = svc.domain.GetIDByUser(helper.FromGinContext(c), userID)
	}

	domain, err := svc.domain.GetByID(helper.FromGinContext(c), body.ID)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	if domain.CreatedBy != userID {
		return resp.Forbidden("This domain is not yours"), nil
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	bodyData := types.JSON{}
	if err := json.Unmarshal(bodyJSON, &bodyData); err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	_, err = svc.domain.Update(helper.FromGinContext(c), domain.ID, bodyData)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: types.JSON{
			"id": body.ID,
		},
	}, nil
}

// ReadDomainService read domain service
func (svc *Service) ReadDomainService(c *gin.Context, id string) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("INVALID_USER_ID")
	}

	if id == "" {
		id, _ = svc.domain.GetIDByUser(helper.FromGinContext(c), userID)
	}

	domain, err := svc.domain.GetByID(helper.FromGinContext(c), id)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	if domain.CreatedBy != userID {
		return resp.Forbidden("This domain is not yours"), nil
	}

	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, true),
	}, nil
}

// ****** Internal methods of service

// isCreateDomain user count <= 1, create domain
func (svc *Service) isCreateDomain(ctx context.Context, body *structs.CreateDomainBody) (*ent.Domain, error) {
	if body.CreatedBy == "" {
		return nil, errors.New("INVALID_USER_ID")
	}

	client := svc.d.GetEntClient()
	countUser := client.User.Query().CountX(context.Background())
	if countUser <= 1 {
		domain, err := svc.domain.Create(ctx, body)
		if err != nil {
			return nil, err
		}

		_, err = client.UserDomain.Create().SetID(body.CreatedBy).SetDomainID(domain.ID).Save(ctx)
		if err != nil {
			return nil, err
		}

		return domain, nil
	}

	// check if user already have domain
	if validator.IsEmpty(body.Name) {
		if domain, _ := svc.domain.GetByUser(ctx, body.CreatedBy); domain != nil {
			return domain, nil
		}
		return svc.domain.Create(ctx, body)
	}

	return nil, nil
}

// serializeDomain  Serialize domain
func (svc *Service) serializeDomain(c *gin.Context, domain *ent.Domain, withUser bool) *structs.ReadDomain {
	readDomain := &structs.ReadDomain{
		ID:          domain.ID,
		Name:        domain.Name,
		Title:       domain.Title,
		URL:         domain.URL,
		Logo:        domain.Logo,
		LogoAlt:     domain.LogoAlt,
		Keywords:    strings.Split(domain.Keywords, ","),
		Copyright:   domain.Copyright,
		Description: domain.Description,
		Order:       domain.Order,
		Disabled:    domain.Disabled,
		Extras:      domain.Extras,
	}

	if withUser {
		user, err := svc.user.GetByID(c, domain.CreatedBy)
		if err == nil {
			readDomain.User = new(structs.User)
			_ = copier.Copy(&readDomain.User, user)
		}
	}

	return readDomain
}
